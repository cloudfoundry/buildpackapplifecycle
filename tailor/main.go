package main

import (
	"log"
	"os"

	"github.com/cloudfoundry-incubator/linux-circus/buildpackrunner"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "tailor"
	app.Usage = "run buildpacks, generate droplet"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			"appDir",
			"",
			"app source location",
		},
		cli.StringFlag{
			"outputDropletDir",
			"",
			"output droplet destination",
		},
		cli.StringFlag{
			"outputMetadataDir",
			"",
			"output metadata destination",
		},
		cli.StringFlag{
			"buildpacksDir",
			"",
			"directory containing all buildpacks",
		},
		cli.StringFlag{
			"buildArtifactsCacheDir",
			"",
			"location to store arbitrary data cached between staging",
		},
		cli.StringSliceFlag{
			"buildpack",
			&cli.StringSlice{},
			"buildpack to run (specify many with many flags, in order)",
		},
	}

	app.Action = func(c *cli.Context) {
		config := models.LinuxCircusTailorConfig{
			AppDir:                 c.String("appDir"),
			OutputDropletDir:       c.String("outputDropletDir"),
			OutputMetadataDir:      c.String("outputMetadataDir"),
			BuildpacksDir:          c.String("buildpacksDir"),
			BuildArtifactsCacheDir: c.String("buildArtifactsCacheDir"),
			BuildpackOrder:         c.StringSlice("buildpack"),
		}

		log.Println("CONFIG:", config)
		buildpackRunner := buildpackrunner.New(config)

		err := buildpackRunner.Run()
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}

	app.Run(os.Args)
}
