package main

import (
	"flag"
	"os"

	"github.com/cloudfoundry/gunk/command_runner"

	"github.com/cloudfoundry-incubator/linux-smelter/smelter"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func main() {
	theLinuxSmeltingConfig := models.NewLinuxSmeltingConfig([]string{})

	debug := theLinuxSmeltingConfig.Bool(
		"debug",
		false,
		"print the output of commands as they're executed",
	)

	if err := theLinuxSmeltingConfig.Parse(os.Args[1:len(os.Args)]); err != nil {
		println(err.Error())
		os.Exit(1)
	}

	if theLinuxSmeltingConfig.AppDir() == "" {
		println("missing -appDir")
		usage()
	}

	if theLinuxSmeltingConfig.OutputDir() == "" {
		println("missing -outputDir")
		usage()
	}

	if theLinuxSmeltingConfig.BuildpacksDir() == "" {
		println("missing -buildpacksDir")
		usage()
	}

	if theLinuxSmeltingConfig.ResultJsonDir() == "" {
		println("missing -resultDir")
		usage()
	}

	if len(theLinuxSmeltingConfig.BuildpackOrder()) == 0 {
		println("missing -buildpackOrder")
		usage()
	}

	smelter := smelter.New(&theLinuxSmeltingConfig, command_runner.New(*debug))

	err := smelter.Smelt()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

func usage() {
	flag.PrintDefaults()
	os.Exit(1)
}
