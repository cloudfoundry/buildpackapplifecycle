package models

import (
	"crypto/md5"
	"flag"
	"fmt"
	"path"
	"strings"
)

type LinuxCircusTailorConfig struct {
	*flag.FlagSet

	compilerPath string

	values map[string]*string

	appDir                 *string
	outputDir              *string
	resultDir              *string
	buildpacksDir          *string
	buildArtifactsCacheDir *string
	buildpackOrder         *string
}

const (
	LinuxCircusTailorAppDirFlag                 = "appDir"
	LinuxCircusTailorOutputDirFlag              = "outputDir"
	LinuxCircusTailorResultDirFlag              = "resultDir"
	LinuxCircusTailorBuildpacksDirFlag          = "buildpacksDir"
	LinuxCircusTailorBuildArtifactsCacheDirFlag = "buildArtifactsCacheDir"
	LinuxCircusTailorBuildpackOrderFlag         = "buildpackOrder"
)

var LinuxCircusTailorDefaults = map[string]string{
	LinuxCircusTailorAppDirFlag:                 "/app",
	LinuxCircusTailorOutputDirFlag:              "/tmp/droplet",
	LinuxCircusTailorResultDirFlag:              "/tmp/result",
	LinuxCircusTailorBuildpacksDirFlag:          "/tmp/buildpacks",
	LinuxCircusTailorBuildArtifactsCacheDirFlag: "/tmp/cache",
}

func NewLinuxCircusTailorConfig(buildpacks []string) LinuxCircusTailorConfig {
	flagSet := flag.NewFlagSet("linux-smelter", flag.ExitOnError)

	appDir := flagSet.String(
		LinuxCircusTailorAppDirFlag,
		LinuxCircusTailorDefaults[LinuxCircusTailorAppDirFlag],
		"directory containing raw app bits",
	)

	outputDir := flagSet.String(
		LinuxCircusTailorOutputDirFlag,
		LinuxCircusTailorDefaults[LinuxCircusTailorOutputDirFlag],
		"directory in which to write the smelted app bits",
	)

	resultDir := flagSet.String(
		LinuxCircusTailorResultDirFlag,
		LinuxCircusTailorDefaults[LinuxCircusTailorResultDirFlag],
		"directory in which to place smelting result metadata",
	)

	buildpacksDir := flagSet.String(
		LinuxCircusTailorBuildpacksDirFlag,
		LinuxCircusTailorDefaults[LinuxCircusTailorBuildpacksDirFlag],
		"directory containing the buildpacks to try",
	)

	buildArtifactsCacheDir := flagSet.String(
		LinuxCircusTailorBuildArtifactsCacheDirFlag,
		LinuxCircusTailorDefaults[LinuxCircusTailorBuildArtifactsCacheDirFlag],
		"directory to store cached artifacts to buildpacks",
	)

	buildpackOrder := flagSet.String(
		LinuxCircusTailorBuildpackOrderFlag,
		strings.Join(buildpacks, ","),
		"comma-separated list of buildpacks, to be tried in order",
	)

	compilerPath := "/tmp/compiler"

	return LinuxCircusTailorConfig{
		FlagSet: flagSet,

		compilerPath: compilerPath,

		appDir:                 appDir,
		outputDir:              outputDir,
		resultDir:              resultDir,
		buildpacksDir:          buildpacksDir,
		buildArtifactsCacheDir: buildArtifactsCacheDir,
		buildpackOrder:         buildpackOrder,

		values: map[string]*string{
			LinuxCircusTailorAppDirFlag:                 appDir,
			LinuxCircusTailorOutputDirFlag:              outputDir,
			LinuxCircusTailorResultDirFlag:              resultDir,
			LinuxCircusTailorBuildpacksDirFlag:          buildpacksDir,
			LinuxCircusTailorBuildArtifactsCacheDirFlag: buildArtifactsCacheDir,
			LinuxCircusTailorBuildpackOrderFlag:         buildpackOrder,
		},
	}
}

func (s LinuxCircusTailorConfig) Script() string {
	argv := []string{s.compilerCommand()}

	s.FlagSet.VisitAll(func(flag *flag.Flag) {
		argv = append(argv, fmt.Sprintf("-%s='%s'", flag.Name, *s.values[flag.Name]))
	})

	return strings.Join(argv, " ")
}

func (s LinuxCircusTailorConfig) Validate() error {
	var missingFlags []string

	s.FlagSet.VisitAll(func(flag *flag.Flag) {
		schemaFlag, ok := s.values[flag.Name]
		if !ok {
			return
		}

		value := *schemaFlag
		if value == "" {
			missingFlags = append(missingFlags, "-"+flag.Name)
		}
	})

	if len(missingFlags) > 0 {
		return fmt.Errorf("missing flags: %s", strings.Join(missingFlags, ", "))
	}

	return nil
}

func (s LinuxCircusTailorConfig) AppDir() string {
	return *s.appDir
}

func (s LinuxCircusTailorConfig) BuildpackPath(buildpackName string) string {
	return path.Join(s.BuildpacksDir(), fmt.Sprintf("%x", md5.Sum([]byte(buildpackName))))
}

func (s LinuxCircusTailorConfig) BuildpackOrder() []string {
	return strings.Split(*s.buildpackOrder, ",")
}

func (s LinuxCircusTailorConfig) BuildpacksDir() string {
	return *s.buildpacksDir
}

func (s LinuxCircusTailorConfig) BuildArtifactsCacheDir() string {
	return *s.buildArtifactsCacheDir
}

func (s LinuxCircusTailorConfig) CompilerPath() string {
	return s.compilerPath
}

func (s LinuxCircusTailorConfig) compilerCommand() string {
	return path.Join(s.CompilerPath(), "run")
}

func (s LinuxCircusTailorConfig) OutputDir() string {
	return *s.outputDir
}

func (s LinuxCircusTailorConfig) ResultJsonDir() string {
	return *s.resultDir
}

func (s LinuxCircusTailorConfig) ResultJsonPath() string {
	return path.Join(s.ResultJsonDir(), "result.json")
}
