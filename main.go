package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

const (
	VERSION = "0.0.5"
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

	images   []string
	versions []string
)

func main() {
	app := cli.NewApp()
	app.Name = "imagend"
	app.Usage = "(ima)ge (gen)erator for (d)ocker"
	app.Version = VERSION

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "output,o",
			Usage:       "output dir for generated Dockerfiles",
			Value:       outdir,
			Destination: &outdir,
		},
		cli.StringFlag{
			Name:        "target,t",
			Usage:       "templates directory",
			Destination: &tmpldir,
			Value:       tmpldir,
		},
		cli.StringFlag{
			Name:        "manifest,m",
			Usage:       "manifest file path",
			Destination: &manifile,
			Value:       manifile,
		},
		cli.StringSliceFlag{
			Name:  "images,i",
			Usage: "image names to build, empty builds all",
		},
		cli.StringSliceFlag{
			Name:  "versions,V",
			Usage: "versions to build, empty builds all",
		},
		cli.BoolFlag{
			Name:        "remove,r",
			Usage:       "remove, if they exists",
			Destination: &remove,
		},
		cli.BoolFlag{
			Name:        "push,p",
			Usage:       "push images to docker hub",
			Destination: &push,
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
			Name:        "push-only",
			Usage:       "only push images, don't generate or build, will verify",
			Destination: &pushOnly,
		},
		cli.BoolFlag{
			Name:        "verify,T",
			Usage:       "verify images, same as '--skip-gen --skip-build --p=false -r=false'",
			Destination: &verify,
		},
		cli.BoolFlag{
			Name:        "docs,D",
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

		images = c.StringSlice("images")
		versions = c.StringSlice("versions")

		if len(images) == 0 && len(versions) != 0 {
			log.Println("--- ERROR image is required when versions are specified\n---")
			cli.ShowAppHelpAndExit(c, 1)
		}

		if len(images) > 1 && len(versions) > 1 {
			log.Println("--- ERROR a single image is required when versions are specified\n---")
			cli.ShowAppHelpAndExit(c, 1)
		}

		log.Println("loading manifest:", manifile)
		loadManifest(manifile).builds().generate()

		return nil
	}

	app.Run(os.Args)

}
