package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

const (
	VERSION = "0.0.1"
)

var (
	// configurations
	outdir   = "."
	tmpldir  = "templates"
	manifile = "manifest.yml"

	remove     bool
	push       bool
	verify     bool
	skipVerify bool
	skipBuild  bool
	skipGen    bool

	image   string
	version string
)

func main() {
	app := cli.NewApp()
	app.Name = "imagend"
	app.Usage = "(ima)ge (gen)erator for (d)ocker"
	app.Version = VERSION

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "o",
			Usage:       "base directory for used in outputting generated Dockerfiles",
			EnvVar:      "IMAGEND_OUTDIR",
			Value:       outdir,
			Destination: &outdir,
		},
		cli.StringFlag{
			Name:        "t",
			Usage:       "dockerfile templates directory",
			Destination: &tmpldir,
			Value:       tmpldir,
			EnvVar:      "IMAGEND_TEMPLATES",
		},
		cli.StringFlag{
			Name:        "m",
			Usage:       "manifest.yml file path",
			Destination: &manifile,
			Value:       manifile,
			EnvVar:      "IMAGEND_MANIFEST",
		},
		cli.StringFlag{
			Name:        "i",
			Usage:       "image (e.g. language or base) to build, empty builds all",
			Destination: &image,
		},
		cli.StringFlag{
			Name:        "V",
			Usage:       "version to build, requires 'i' flag, empty builds all",
			Destination: &version,
		},
		cli.BoolFlag{
			Name:        "r",
			Usage:       "remove, if they exists",
			Destination: &remove,
		},
		cli.BoolFlag{
			Name:        "skip-gen",
			Usage:       "skip template generation",
			Destination: &skipGen,
		},
		cli.BoolFlag{
			Name:        "skip-verify",
			Usage:       "skip image verification check",
			Destination: &skipVerify,
		},
		cli.BoolFlag{
			Name:        "skip-build",
			Usage:       "skip image build step, prevents removal",
			Destination: &skipBuild,
		},
		cli.BoolFlag{
			Name:        "p",
			Usage:       "push images to docker hub",
			Destination: &push,
		},
		cli.BoolFlag{
			Name:        "verify",
			Usage:       "verify images, same as '--skip-gen --skip-build --p=false -r=false'",
			Destination: &verify,
		},
	}

	app.Action = func(c *cli.Context) error {
		if verify {
			skipBuild = true
			skipGen = true
			push = false
		}

		outdir = expand(outdir, true)
		tmpldir = expand(tmpldir, false)
		manifile = expand(manifile, false)

		if image == "" && version != "" {
			log.Println("--- ERROR image is required when a version is specified\n---")
			cli.ShowAppHelpAndExit(c, 1)
		}

		log.Println("loading manifest:", manifile)
		loadManifest(manifile).builds().generate()

		return nil
	}

	app.Run(os.Args)

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
