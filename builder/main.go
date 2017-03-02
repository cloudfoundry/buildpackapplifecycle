package main

import (
	"flag"
	"os"

	"code.cloudfoundry.org/buildpackapplifecycle"
	"code.cloudfoundry.org/buildpackapplifecycle/buildpackrunner"
)

func main() {
	config := buildpackapplifecycle.NewLifecycleBuilderConfig([]string{}, false, false)

	if err := config.Parse(os.Args[1:len(os.Args)]); err != nil {
		println(err.Error())
		os.Exit(1)
	}

	if err := config.Validate(); err != nil {
		println(err.Error())
		usage()
	}

	zipDownloader := buildpackrunner.NewZipDownloader(config.SkipCertVerify())
	runner := buildpackrunner.New(zipDownloader)

	_, err := runner.Run(&config)
	if err != nil {
		println(err)
		println(err.Error())
		os.Exit(buildpackapplifecycle.ExitCodeFromError(err))
	}

	os.Exit(0)
}

func usage() {
	flag.PrintDefaults()
	os.Exit(1)
}
