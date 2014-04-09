package models_test

import (
	"strings"

	. "github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LinuxSmeltingConfig", func() {
	var smeltingConfig LinuxSmeltingConfig

	BeforeEach(func() {
		smeltingConfig = NewLinuxSmeltingConfig([]string{"ocaml-buildpack", "haskell-buildpack", "bash-buildpack"})
	})

	Context("with defaults", func() {
		It("generates a script for running its smelter", func() {
			expectedScript := strings.Join(
				[]string{
					"/tmp/compiler/run",
					"-appDir='/app'",
					"-buildpackOrder='ocaml-buildpack,haskell-buildpack,bash-buildpack'",
					"-buildpacksDir='/tmp/buildpacks'",
					"-cacheDir='/tmp/cache'",
					"-outputDir='/tmp/droplet'",
					"-resultDir='/tmp/result'",
				},
				" ",
			)

			Ω(smeltingConfig.Script()).To(Equal(expectedScript))
		})
	})

	Context("with overrides", func() {
		BeforeEach(func() {
			smeltingConfig.Set(LinuxSmeltingAppDirFlag, "/some/app/dir")
			smeltingConfig.Set(LinuxSmeltingOutputDirFlag, "/some/droplet/dir")
			smeltingConfig.Set(LinuxSmeltingResultDirFlag, "/some/result/dir")
			smeltingConfig.Set(LinuxSmeltingBuildpacksDirFlag, "/some/buildpacks/dir")
			smeltingConfig.Set(LinuxSmeltingCacheDirFlag, "/some/cache/dir")
		})

		It("generates a script for running its smelter", func() {
			expectedScript := strings.Join(
				[]string{
					"/tmp/compiler/run",
					"-appDir='/some/app/dir'",
					"-buildpackOrder='ocaml-buildpack,haskell-buildpack,bash-buildpack'",
					"-buildpacksDir='/some/buildpacks/dir'",
					"-cacheDir='/some/cache/dir'",
					"-outputDir='/some/droplet/dir'",
					"-resultDir='/some/result/dir'",
				},
				" ",
			)

			Ω(smeltingConfig.Script()).To(Equal(expectedScript))
		})
	})

	It("returns the path to the compiler", func() {
		Ω(smeltingConfig.CompilerPath()).To(Equal("/tmp/compiler"))
	})

	It("returns the path to the app bits", func() {
		Ω(smeltingConfig.AppDir()).To(Equal("/app"))
	})

	It("returns the path to a given buildpack", func() {
		Ω(smeltingConfig.BuildpackPath("my-buildpack")).To(Equal("/tmp/buildpacks/my-buildpack"))
	})

	It("returns the path to the result.json", func() {
		Ω(smeltingConfig.ResultJsonPath()).To(Equal("/tmp/result/result.json"))
	})
})
