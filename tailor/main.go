package main

import (
	"flag"
	"os"

	"github.com/cloudfoundry-incubator/linux-circus/buildpackrunner"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

func main() {
	config := models.NewCircusTailorConfig([]string{})

	if err := config.Parse(os.Args[1:len(os.Args)]); err != nil {
		println(err.Error())
		os.Exit(1)
	}

	if err := config.Validate(); err != nil {
		println(err.Error())
		usage()
	}

	runner := buildpackrunner.New(&config)

	err := runner.Run()
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
