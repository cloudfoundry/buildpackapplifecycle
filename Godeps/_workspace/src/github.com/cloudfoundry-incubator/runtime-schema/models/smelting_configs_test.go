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
			command := "/tmp/compiler/run"
			commandFlags := []string{
				"-appDir='/app'",
				"-buildpackOrder='ocaml-buildpack,haskell-buildpack,bash-buildpack'",
				"-buildpacksDir='/tmp/buildpacks'",
				"-buildArtifactsCacheDir='/tmp/cache'",
				"-outputDir='/tmp/droplet'",
				"-resultDir='/tmp/result'",
			}

			Ω(strings.HasPrefix(smeltingConfig.Script(), command)).To(BeTrue())
			for _, commandFlag := range commandFlags {
				Ω(smeltingConfig.Script()).To(ContainSubstring(commandFlag))
			}
		})
	})

	Context("with overrides", func() {
		BeforeEach(func() {
			smeltingConfig.Set(LinuxSmeltingAppDirFlag, "/some/app/dir")
			smeltingConfig.Set(LinuxSmeltingOutputDirFlag, "/some/droplet/dir")
			smeltingConfig.Set(LinuxSmeltingResultDirFlag, "/some/result/dir")
			smeltingConfig.Set(LinuxSmeltingBuildpacksDirFlag, "/some/buildpacks/dir")
			smeltingConfig.Set(LinuxSmeltingBuildArtifactsCacheDirFlag, "/some/cache/dir")
		})

		It("generates a script for running its smelter", func() {
			command := "/tmp/compiler/run"
			commandFlags := []string{
				"-appDir='/some/app/dir'",
				"-buildpackOrder='ocaml-buildpack,haskell-buildpack,bash-buildpack'",
				"-buildpacksDir='/some/buildpacks/dir'",
				"-buildArtifactsCacheDir='/some/cache/dir'",
				"-outputDir='/some/droplet/dir'",
				"-resultDir='/some/result/dir'",
			}

			Ω(strings.HasPrefix(smeltingConfig.Script(), command)).To(BeTrue())
			for _, commandFlag := range commandFlags {
				Ω(smeltingConfig.Script()).To(ContainSubstring(commandFlag))
			}
		})
	})

	It("returns the path to the compiler", func() {
		Ω(smeltingConfig.CompilerPath()).To(Equal("/tmp/compiler"))
	})

	It("returns the path to the app bits", func() {
		Ω(smeltingConfig.AppDir()).To(Equal("/app"))
	})

	It("returns the path to a given buildpack", func() {
		key := "my-buildpack/key/::"
		Ω(smeltingConfig.BuildpackPath(key)).To(Equal("/tmp/buildpacks/8b2f72a0702aed614f8b5d8f7f5b431b"))
	})

	It("returns the path to the result.json", func() {
		Ω(smeltingConfig.ResultJsonPath()).To(Equal("/tmp/result/result.json"))
	})
})
