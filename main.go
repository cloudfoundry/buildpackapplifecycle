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

	if err := theLinuxSmeltingConfig.Validate(); err != nil {
		println(err.Error())
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
