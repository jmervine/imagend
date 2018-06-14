package main

import (
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
	Source   string   `yaml:"source"`
	Native   bool     `yaml:"native"` // indicates that it's included in the source image
	Priority int      `yaml:"priority"`
	Aliases  []string `yaml:"aliases"`
	Tags     []string `yaml:"tags"`
	Verify   struct {
		Exec   string `yaml:"exec"`
		Expect string `yaml:"expect"`
		Exact  bool   `yaml:"exact"`
		Skip   bool   `yaml:"skip"`
	} `yaml:"verify"`
}

type Manifest []Version

func loadManifest(file string) Manifest {
	var m Manifest
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(data, &m)

	return m
}

func (m Manifest) builds() Manifest {
	// build all
	if len(images) == 0 && len(versions) == 0 {
		return m
	}

	builds := make(Manifest, 0)

	for _, v := range m {
		if len(images) > 0 && containsAny(images, v.Name, v.Image) {
			if len(versions) == 0 || containsAny(versions, v.Version) {
				builds = append(builds, v)
				continue
			}
		}
	}

	if len(builds) == 0 {
		if len(versions) > 0 {
			log.Fatal("ERROR requested version(s) not found: ", strings.Join(versions, ", "))
		} else {
			log.Fatal("ERROR requested image(s) not found: ", strings.Join(images, ", "))
		}
	}

	return builds
}

func (m Manifest) sort() ([]string, map[string][]Version) {
	order := make([]string, 0)
	sets := make(map[string][]Version)

	// Order manifest by priority
	sort.Slice(m, func(i, j int) bool {
		return m[i].Priority < m[j].Priority
	})

	// Build order matters for image sets, not version.
	for _, v := range m {
		if len(sets[v.Name]) == 0 {
			order = append(order, v.Name)
			sets[v.Name] = make([]Version, 0)
		}

		sets[v.Name] = append(sets[v.Name], v)
	}

	return order, sets
}

func (m Manifest) generate() {
	order, sets := m.sort()

	for _, n := range order {
		vers := sets[n]

		var wg sync.WaitGroup
		wg.Add(len(vers))

		for _, ver := range vers {
			act := func(v Version) {
				defer wg.Done()
				log.Println("IMAGE:", v.tag())

				if !skipGen || !skipBuild || remove || push {
					log.Println("- log:", v.logpath())
				}

				if !skipGen {
					v.render()
				}

				if remove && !skipBuild {
					v.rmi()
				}

				if !skipBuild {
					v.build()
				}

				if !skipVerify {
					v.verify()
				}

				if push {
					v.push()
				}
			}

			if serialized {
				act(ver)
			} else {
				go act(ver)
			}
		}

		wg.Wait()
	}
}

func (v *Version) render() {
	if exists(v.dockerfile()) {
		log.Println("--- removing:", v.dockerfile())
		os.Remove(v.dockerfile())
	}

	log.Println("--- rendering:", v.dockerfile())
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

func (v *Version) rmi() {
	for _, tag := range v.tags() {
		log.Println("--- removing:", tag)
		v.execute(exec.Command("docker", "rmi", "-f", tag), true)
	}
}

func (v *Version) build() {
	log.Println("--- building:", v.tag())
	v.execute(exec.Command("docker", "build", v.dockerbase(), "-t", v.tag()), false)
	log.Println("--- complete:", v.tag())

	for _, tag := range v.tags() {
		if tag != v.tag() {
			log.Println("--- tagging: ", tag)
			v.execute(exec.Command("docker", "tag", v.tag(), tag), false)
		}
	}
}

func (v *Version) push() {
	for _, tag := range v.tags() {
		log.Println("--- pushing: ", tag)
		v.execute(exec.Command("docker", "push", tag), false)
	}
}

func (v *Version) verify() {
	if !v.Verify.Skip {
		for _, tag := range v.tags() {
			log.Println("--- verifying: ", tag)

			var x string
			if v.Verify.Exec != "" {
				x = fmt.Sprintf("docker run --rm %s sh -c '%s'", tag, v.Verify.Exec)
			} else {
				x = fmt.Sprintf("docker run --rm %s sh -c '%s --version'", tag, v.Name)
			}

			cmd := exec.Command("sh", "-c", x)
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Println("------ failure:", tag)
				log.Fatal("------- with: ", err.Error())
			}

			check := v.Version
			var pass bool

			if v.Verify.Expect != "" {
				check = v.Verify.Expect
			}

			if v.Verify.Exact {
				if string(out) == check {
					pass = true
				}
			} else {
				pass = strings.Contains(string(out), check)
			}

			if pass {
				log.Println("---- verified: ", tag)
			} else {
				log.Println("------ failure:", tag)
				log.Fatal("------- with \"", string(out), "\"")
			}
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

func (v *Version) tag() string {
	return fmt.Sprintf("%s:%s", v.imageName(), v.Version)
}

func (v *Version) tags() []string {
	subtags := []string{v.Version}
	subtags = append(subtags, v.Tags...)

	images := []string{v.imageName()}
	for _, alias := range v.Aliases {
		images = append(images, alias)
	}

	tags := make([]string, 0)
	for _, image := range images {
		for _, st := range subtags {
			tags = append(tags, fmt.Sprintf("%s:%s", image, st))
		}
	}

	return tags
}

func (v *Version) template() string {
	t := path.Join(tmpldir, "Dockerfile."+v.Name+".tmpl")
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
