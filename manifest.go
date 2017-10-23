package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	tmpl "text/template"

	"gopkg.in/yaml.v2"
)

type Version struct {
	Name     string   `yaml:"name"`
	Image    string   `yaml:"image"`
	Version  string   `yaml:"version"`
	Latest   bool     `yaml:"latest"`
	Source   string   `yaml:"source"`
	Native   bool     `yaml:"native"` // indicates that it's included in the source image
	Priority int      `yaml:"priority"`
	Aliases  []string `yaml:"aliases"`
}

type Manifest []Version

func loadManifest(file string) (Manifest, error) {
	var m Manifest
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return m, err
	}

	err = yaml.Unmarshal(data, &m)

	return m, err
}

func (m Manifest) builds() (Manifest, error) {
	builds := make(Manifest, 0)

	if image == "" && version == "" {
		builds = m
		return builds, nil
	}

	for _, v := range m {
		if image != "" && (v.Name == image || v.Image == image) {
			if version == "" || (version != "" && version == v.Version) {
				builds = append(builds, v)
				continue
			}
		}
	}

	if len(builds) == 0 {
		if version != "" {
			return builds, errors.New("ERROR requested version not found: " + version)
		}

		return builds, errors.New("ERROR requested image not found: " + image)
	}

	return builds, nil
}

func (m Manifest) generate() error {

	// Order manifest by priority
	sort.Slice(m, func(i, j int) bool {
		return m[i].Priority < m[j].Priority
	})

	sets := make(map[string][]Version)
	order := make([]string, 0)

	// Build order matters for image sets, not version.
	for _, v := range m {
		if len(sets[v.Name]) == 0 {
			order = append(order, v.Name)
			sets[v.Name] = make([]Version, 0)
		}
		sets[v.Name] = append(sets[v.Name], v)
	}

	for _, n := range order {
		vers := sets[n]

		var wg sync.WaitGroup
		wg.Add(len(vers))

		for _, ver := range vers {
			go func(v Version) {
				defer wg.Done()
				log.Println("IMAGE:", v.tag())
				log.Println("--- log:", v.logpath())
				v.render()

				if remove {
					v.rmi()
				}

				if build {
					v.build()
				}

				if verify {
					v.verify()
				}

				if push {
					v.push()
				}
			}(ver)
		}

		wg.Wait()
	}

	return nil
}

func (v *Version) render() {
	t, err := tmpl.ParseFiles(v.template())
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(v.dockerfile(), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := t.Execute(f, v); err != nil {
		log.Fatal(err)
	}
}

func (v *Version) tag() string {
	return fmt.Sprintf("%s:%s", v.imageName(), v.Version)
}

func (v *Version) template() string {
	return template("Dockerfile." + v.Name)
}

func template(name string) string {
	t := path.Join(tmpldir, name+".tmpl")
	if !exists(t) {
		log.Fatal("ERROR template file not found: ", t)
	}
	return t
}

func (v *Version) dockerbase() string {
	b := path.Join(outdir, v.Name+"-"+v.Version)
	if !exists(b) {
		if err := os.Mkdir(b, 0755); err != nil {
			log.Fatal(err)
		}
	}

	return b
}

func (v *Version) dockerfile() string {
	return path.Join(v.dockerbase(), "Dockerfile")
}

func (v *Version) latest() string {
	return fmt.Sprintf("%s:latest", v.imageName())
}

func (v *Version) rmi() {
	log.Println("--- removing:", v.tag())
	v.execute(exec.Command("docker", "rmi", "-f", v.tag()), true)

	if v.Latest {
		log.Println("--- removing: ", v.imageName()+":latest")
		v.execute(exec.Command("docker", "rmi", "-f", v.imageName()+":latest"), true)
	}

	for _, a := range v.aliases() {
		log.Println("--- removing: ", a)
		v.execute(exec.Command("docker", "rmi", "-f", a), true)
	}
}

func (v *Version) build() {
	log.Println("--- building:", v.tag())
	v.execute(exec.Command("docker", "build", v.dockerbase(), "-t", v.tag()), false)
	log.Println("--- complete:", v.tag())

	if v.Latest {
		log.Println("--- tagging: ", v.latest())
		v.execute(exec.Command("docker", "tag", v.tag(), v.latest()), false)
	}

	for _, a := range v.aliases() {
		log.Println("--- tagging: ", a)
		v.execute(exec.Command("docker", "tag", v.tag(), a), false)
	}
}

func (v *Version) aliases() []string {
	aliases := make([]string, 0)

	if len(v.Aliases) > 0 {
		for _, alias := range v.Aliases {
			aliases = append(aliases, fmt.Sprintf("%s:%s", alias, v.Version))

			if v.Latest {
				aliases = append(aliases, fmt.Sprintf("%s:latest", alias))
			}
		}
	}

	return aliases
}

func (v *Version) push() {
	log.Println("--- pushing:", v.tag())
	v.execute(exec.Command("docker", "push", v.tag()), false)

	if v.Latest {
		log.Println("--- pushing: ", v.latest())
		v.execute(exec.Command("docker", "push", v.latest()), false)
	}

	for _, a := range v.aliases() {
		log.Println("--- pushing: ", a)
		v.execute(exec.Command("docker", "push", a), false)
	}
}

func (v *Version) verify() {
	if v.Name != "base" {
		check := func(t string) {
			log.Println("--- verifying: ", t)
			x := fmt.Sprintf("docker run --rm %s %s --version", t, v.Name)
			cmd := exec.Command("sh", "-c", x)
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Println("------ failure:", t)
				log.Fatal("------- with: ", err.Error())
			}

			if strings.Contains(string(out), v.Version) {
				log.Println("---- verified: ", t)
			} else {
				log.Println("------ failure:", t)
				log.Fatal("------- with \"", string(out), "\"")
			}
		}

		check(v.tag())

		if v.Latest {
			check(v.latest())
		}

		for _, a := range v.aliases() {
			check(a)
		}
	}
}

func (v *Version) logpath() string {
	out, err := filepath.Abs(path.Join(".", "log", v.tag()+".log"))
	if err != nil {
		log.Fatal(err)
	}

	base := filepath.Dir(out)

	// Ensure log target directory exists.
	if !exists(base) {
		if err := os.MkdirAll(base, 0755); err != nil {
			log.Fatal(err)
		}
	}

	return out
}

func (v *Version) execute(cmd *exec.Cmd, ignoreErrors bool) {
	out := v.logpath()

	f, err := os.OpenFile(out, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	stdout := io.MultiWriter(f)
	stderr := io.MultiWriter(f, os.Stderr)

	cmd.Stderr = stderr
	cmd.Stdout = stdout

	err = cmd.Run()
	if err != nil && !ignoreErrors {
		log.Fatal(err)
	}
}

func (v *Version) imageName() string {
	if v.Image == "" {
		return v.Name
	}

	return v.Image
}
