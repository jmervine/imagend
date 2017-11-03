package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

const (
	VERSION = "0.0.4"
)

var (
	// configurations
	outdir   = "."
	tmpldir  = "templates"
	manifile = "manifest.yml"

	docs       bool
	remove     bool
	push       bool
	pushOnly   bool
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
			Name:        "push-only",
			Usage:       "only push images, don't generate or build, will verify",
			Destination: &pushOnly,
		},
		cli.BoolFlag{
			Name:        "verify",
			Usage:       "verify images, same as '--skip-gen --skip-build --p=false -r=false'",
			Destination: &verify,
		},
		cli.BoolFlag{
			Name:        "docs",
			Usage:       "generate markdown docs from your manifest, does nothing else",
			Destination: &docs,
		},
	}

	app.Action = func(c *cli.Context) error {
		if docs {
			manifest2markdown(manifile)
			return nil
		}

		if verify {
			skipBuild = true
			skipGen = true
			push = false
		}

		if pushOnly {
			skipBuild = true
			skipGen = true
			push = true
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
