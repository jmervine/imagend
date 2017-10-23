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

const baseImageName = "jmervine/herokudev-"

type Version struct {
	Image   string `yaml:"image"`
	Version string `yaml:"version"`
	Latest  bool   `yaml:"latest"`
	Source  string `yaml:"source"`
	Native  bool   `yaml:"native"` // indicates that it's included in the source image
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
		if image != "" && v.Image == image {
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
	sets := make(map[string][]Version)
	keys := make([]string, 0)

	// Build order matters for image sets, not version.
	for _, v := range m {
		if len(sets[v.Image]) == 0 {
			keys = append(keys, v.Image)
			sets[v.Image] = make([]Version, 0)
		}
		sets[v.Image] = append(sets[v.Image], v)
	}

	// Enforce order
	sort.Strings(keys)

	for _, k := range keys {
		vers := sets[k]

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
	return baseImageName + v.Image + ":" + v.Version
}

func (v *Version) template() string {
	return template("Dockerfile." + v.Image)
}

func template(name string) string {
	t := path.Join(tmpldir, name+".tmpl")
	if !exists(t) {
		log.Fatal("ERROR template file not found: ", t)
	}
	return t
}

func (v *Version) dockerbase() string {
	b := path.Join(outdir, v.Image+"-"+v.Version)
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

func (v *Version) imageName() string {
	return baseImageName + v.Image
}

func (v *Version) rmi() {
	log.Println("--- removing:", v.tag())
	rmi := exec.Command("docker", "rmi", "-f", v.tag())
	v.execute(rmi, true)

	if v.Latest {
		log.Println("--- removing: ", v.imageName()+":latest")
		rmi := exec.Command("docker", "rmi", "-f", v.imageName()+":latest")
		v.execute(rmi, true)
	}
}

func (v *Version) build() {
	log.Println("--- building:", v.tag())
	bld := exec.Command("docker", "build", v.dockerbase(), "-t", v.tag())
	v.execute(bld, false)
	log.Println("--- complete:", v.tag())

	if v.Latest {
		log.Println("--- tagging: ", v.imageName()+":latest")
		tag := exec.Command("docker", "tag", v.tag(), v.imageName()+":latest")
		v.execute(tag, false)
	}
}

func (v *Version) push() {
	log.Println("--- pushing:", v.tag())
	push := exec.Command("docker", "push", v.tag())
	v.execute(push, false)

	if v.Latest {
		log.Println("--- pushing: ", v.imageName()+":latest")
		push = exec.Command("docker", "push", v.imageName()+":latest")
		v.execute(push, false)
	}
}

func (v *Version) verify() {
	if v.Image != "base" {
		check := func(t string) {
			log.Println("--- verifying: ", t)
			x := fmt.Sprintf("docker run --rm %s %s --version", t, v.Image)
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
			check(v.imageName() + ":latest")
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
