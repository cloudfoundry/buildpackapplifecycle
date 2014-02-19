package main

import (
	"flag"
	"os"
	"path"
	"strings"

	"github.com/cloudfoundry-incubator/linux-smelter/smelter"
	"github.com/cloudfoundry/gunk/command_runner"
)

var appDir = flag.String(
	"appDir",
	os.Getenv("APP_DIR"),
	"directory containing raw app bits, settable as $APP_DIR",
)

var outputDir = flag.String(
	"outputDir",
	os.Getenv("OUTPUT_DIR"),
	"directory in which to write the smelted app bits, settable as $OUTPUT_DIR",
)

var buildpacksDir = flag.String(
	"buildpacksDir",
	os.Getenv("BUILDPACKS_DIR"),
	"directory containing the buildpacks to try, settable as $BUILDPACKS_DIR",
)

var buildpackOrder = flag.String(
	"buildpackOrder",
	os.Getenv("BUILDPACK_ORDER"),
	"comma-separated list of buildpacks, to be tried in order, settable as $BUILDPACK_ORDER",
)

func main() {
	flag.Parse()

	if *appDir == "" {
		println("missing -appDir")
		usage()
	}

	if *outputDir == "" {
		println("missing -outputDir")
		usage()
	}

	if *buildpacksDir == "" {
		println("missing -buildpacksDir")
		usage()
	}

	if *buildpackOrder == "" {
		println("missing -buildpackOrder")
		usage()
	}

	buildpacks := []string{}

	for _, name := range strings.Split(*buildpackOrder, ",") {
		buildpacks = append(buildpacks, path.Join(*buildpacksDir, name))
	}

	smelter := smelter.New(*appDir, *outputDir, buildpacks, command_runner.New(false))

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
