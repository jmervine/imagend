package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

const (
	// defaults
	defaultOutdir   = "../"
	defaultTmpldir  = "../templates"
	defaultManifile = "../manifest.yml"
	defaultImage    = ""
	defaultVersion  = ""
	defaultRemove   = true
	defaultVerify   = true
	defaultPush     = true
	defaultBuild    = true
)

var (
	// configurations
	outdir   string
	tmpldir  string
	manifile string
	image    string
	version  string
	remove   bool
	verify   bool
	push     bool
	build    bool
)

func init() {
	flag.StringVar(&outdir, "r", defaultOutdir, "root directory, used in saving generated Dockerfiles")
	flag.StringVar(&tmpldir, "t", defaultTmpldir, "dockerfile templates directory")
	flag.StringVar(&manifile, "m", defaultManifile, "manifest file path")
	flag.StringVar(&image, "l", defaultImage, "image (e.g. language or base) to build, empty will build all")
	flag.StringVar(&version, "v", defaultVersion, "version to build, requires image, empty will build all")
	flag.BoolVar(&remove, "x", defaultRemove, "remove previous images, if they exist")
	flag.BoolVar(&verify, "c", defaultVerify, "verify built image(s)")
	flag.BoolVar(&push, "p", defaultPush, "push built image(s)")
	flag.BoolVar(&build, "b", defaultBuild, "build image(s)")
}

func main() {
	parseArgs()

	log.Println("loading manifest:", manifile)
	manifest, err := loadManifest(manifile)

	if err != nil {
		log.Fatal(err)
	}

	if manifest, err = manifest.builds(); err != nil {
		log.Fatal(err)
	}

	manifest.generate()
}

// parseArgs parses arguments using the flag package
func parseArgs() {
	flag.Parse()

	if !build { // don't remove if not building
		remove = false
	}

	outdir = expand(outdir, true)
	tmpldir = expand(tmpldir, false)
	manifile = expand(manifile, false)

	if image == "" && version != "" {
		log.Println("--- ERROR image is required when a version is specified\n---")
		flag.Usage()
		os.Exit(1)
	}
}

func expand(path string, mkdir bool) string {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	if !exists(path) {
		if !mkdir {
			log.Fatal("--- ERROR file or directory not found: ", path)
		}
		if err := os.Mkdir(path, 0755); err != nil {
			log.Fatal(err)
		}
	}

	return path
}
