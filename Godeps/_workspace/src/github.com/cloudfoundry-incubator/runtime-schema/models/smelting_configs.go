package models

import (
	"flag"
	"fmt"
	"path"
	"strings"
)

type LinuxSmeltingConfig struct {
	*flag.FlagSet

	compilerPath string

	values map[string]*string

	appDir         *string
	outputDir      *string
	resultDir      *string
	buildpacksDir  *string
	cacheDir       *string
	buildpackOrder *string
}

const (
	LinuxSmeltingAppDirFlag         = "appDir"
	LinuxSmeltingOutputDirFlag      = "outputDir"
	LinuxSmeltingResultDirFlag      = "resultDir"
	LinuxSmeltingBuildpacksDirFlag  = "buildpacksDir"
	LinuxSmeltingCacheDirFlag       = "cacheDir"
	LinuxSmeltingBuildpackOrderFlag = "buildpackOrder"
)

var LinuxSmeltingDefaults = map[string]string{
	LinuxSmeltingAppDirFlag:        "/app",
	LinuxSmeltingOutputDirFlag:     "/tmp/droplet",
	LinuxSmeltingResultDirFlag:     "/tmp/result",
	LinuxSmeltingBuildpacksDirFlag: "/tmp/buildpacks",
	LinuxSmeltingCacheDirFlag:      "/tmp/cache",
}

func NewLinuxSmeltingConfig(buildpacks []string) LinuxSmeltingConfig {
	flagSet := flag.NewFlagSet("linux-smelter", flag.ExitOnError)

	appDir := flagSet.String(
		LinuxSmeltingAppDirFlag,
		LinuxSmeltingDefaults[LinuxSmeltingAppDirFlag],
		"directory containing raw app bits",
	)

	outputDir := flagSet.String(
		LinuxSmeltingOutputDirFlag,
		LinuxSmeltingDefaults[LinuxSmeltingOutputDirFlag],
		"directory in which to write the smelted app bits",
	)

	resultDir := flagSet.String(
		LinuxSmeltingResultDirFlag,
		LinuxSmeltingDefaults[LinuxSmeltingResultDirFlag],
		"directory in which to place smelting result metadata",
	)

	buildpacksDir := flagSet.String(
		LinuxSmeltingBuildpacksDirFlag,
		LinuxSmeltingDefaults[LinuxSmeltingBuildpacksDirFlag],
		"directory containing the buildpacks to try",
	)

	cacheDir := flagSet.String(
		LinuxSmeltingCacheDirFlag,
		LinuxSmeltingDefaults[LinuxSmeltingCacheDirFlag],
		"directory to store cached artifacts to buildpacks",
	)

	buildpackOrder := flagSet.String(
		LinuxSmeltingBuildpackOrderFlag,
		strings.Join(buildpacks, ","),
		"comma-separated list of buildpacks, to be tried in order",
	)

	compilerPath := "/tmp/compiler"

	return LinuxSmeltingConfig{
		FlagSet: flagSet,

		compilerPath: compilerPath,

		appDir:         appDir,
		outputDir:      outputDir,
		resultDir:      resultDir,
		buildpacksDir:  buildpacksDir,
		cacheDir:       cacheDir,
		buildpackOrder: buildpackOrder,

		values: map[string]*string{
			"appDir":         appDir,
			"outputDir":      outputDir,
			"resultDir":      resultDir,
			"buildpacksDir":  buildpacksDir,
			"cacheDir":       cacheDir,
			"buildpackOrder": buildpackOrder,
		},
	}
}

func (s LinuxSmeltingConfig) Script() string {
	argv := []string{s.compilerCommand()}

	s.FlagSet.VisitAll(func(flag *flag.Flag) {
		argv = append(argv, fmt.Sprintf("-%s='%s'", flag.Name, *s.values[flag.Name]))
	})

	return strings.Join(argv, " ")
}

func (s LinuxSmeltingConfig) Validate() error {
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

func (s LinuxSmeltingConfig) AppDir() string {
	return *s.appDir
}

func (s LinuxSmeltingConfig) BuildpackPath(buildpackName string) string {
	return path.Join(s.BuildpacksDir(), buildpackName)
}

func (s LinuxSmeltingConfig) BuildpackOrder() []string {
	return strings.Split(*s.buildpackOrder, ",")
}

func (s LinuxSmeltingConfig) BuildpacksDir() string {
	return *s.buildpacksDir
}

func (s LinuxSmeltingConfig) CacheDir() string {
	return *s.cacheDir
}

func (s LinuxSmeltingConfig) CompilerPath() string {
	return s.compilerPath
}

func (s LinuxSmeltingConfig) compilerCommand() string {
	return path.Join(s.CompilerPath(), "run")
}

func (s LinuxSmeltingConfig) DropletArchivePath() string {
	return path.Join(s.OutputDir(), "droplet.tgz")
}

func (s LinuxSmeltingConfig) OutputDir() string {
	return *s.outputDir
}

func (s LinuxSmeltingConfig) ResultJsonDir() string {
	return *s.resultDir
}

func (s LinuxSmeltingConfig) ResultJsonPath() string {
	return path.Join(s.ResultJsonDir(), "result.json")
}
