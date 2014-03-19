package models

import (
	"path"
	"strings"
)

type LinuxSmeltingConfig struct {
	buildpackOrder []string
}

func NewLinuxSmeltingConfig(buildpackOrder []string) LinuxSmeltingConfig {
	return LinuxSmeltingConfig{
		buildpackOrder: buildpackOrder,
	}
}

func (s LinuxSmeltingConfig) Script() string {
	return strings.Join([]string{
		s.compilerCommand(),
		"-appDir", s.AppPath(),
		"-outputDir", s.dropletDirPath(),
		"-resultDir", s.resultJsonDir(),
		"-buildpacksDir", s.buildpacksDir(),
		"-buildpackOrder", strings.Join(s.buildpackOrder, ","),
		"-cacheDir", s.cacheDir(),
	}, " ")
}

func (s LinuxSmeltingConfig) AppPath() string {
	return "/app"
}

func (s LinuxSmeltingConfig) buildpacksDir() string {
	return path.Join("/tmp", "buildpacks")
}

func (s LinuxSmeltingConfig) BuildpackPath(buildpackName string) string {
	return path.Join(s.buildpacksDir(), buildpackName)
}

func (s LinuxSmeltingConfig) compilerCommand() string {
	return path.Join(s.CompilerPath(), "run")
}

func (s LinuxSmeltingConfig) CompilerPath() string {
	return path.Join("/tmp", "compiler")
}

func (s LinuxSmeltingConfig) cacheDir() string {
	return path.Join("/tmp", "cache")
}

func (s LinuxSmeltingConfig) dropletDirPath() string {
	return path.Join("/tmp", "droplet")
}

func (s LinuxSmeltingConfig) DropletArchivePath() string {
	return path.Join(s.dropletDirPath(), "droplet.tgz")
}

func (s LinuxSmeltingConfig) resultJsonDir() string {
	return path.Join("/tmp", "result")
}

func (s LinuxSmeltingConfig) ResultJsonPath() string {
	return path.Join(s.resultJsonDir(), "result.json")
}
