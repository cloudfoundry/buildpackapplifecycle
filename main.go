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
	"",
	"directory containing raw app bits, settable as $APP_DIR",
)

var outputDir = flag.String(
	"outputDir",
	"",
	"directory in which to write the smelted app bits, settable as $OUTPUT_DIR",
)

var resultDir = flag.String(
	"resultDir",
	"",
	"directory in which to place smelting result metadata, settable as $RESULT_DIR",
)

var buildpacksDir = flag.String(
	"buildpacksDir",
	"",
	"directory containing the buildpacks to try, settable as $BUILDPACKS_DIR",
)

var cacheDir = flag.String(
	"cacheDir",
	"",
	"directory to store cached artifacts to buildpacks, settable as $CACHE_DIR",
)

var buildpackOrder = flag.String(
	"buildpackOrder",
	"",
	"comma-separated list of buildpacks, to be tried in order, settable as $BUILDPACK_ORDER",
)

var debug = flag.Bool(
	"debug",
	false,
	"print the output of commands as they're executed",
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

	if *resultDir == "" {
		println("missing -resultDir")
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

	smelter := smelter.New(
		*appDir,
		*outputDir,
		*resultDir,
		buildpacks,
		*cacheDir,
		command_runner.New(*debug),
	)

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
