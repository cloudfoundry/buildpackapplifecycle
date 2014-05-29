package models_test

import (
	. "github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LinuxCircusTailorConfig", func() {
	var tailorConfig LinuxCircusTailorConfig

	BeforeEach(func() {
		tailorConfig = NewLinuxCircusTailorConfig([]string{"ocaml-buildpack", "haskell-buildpack", "bash-buildpack"})
	})

	Context("with defaults", func() {
		It("generates a script for running its tailor", func() {
			commandFlags := []string{
				"-appDir='/app'",
				"-buildpackOrder='ocaml-buildpack,haskell-buildpack,bash-buildpack'",
				"-buildpacksDir='/tmp/buildpacks'",
				"-buildArtifactsCacheDir='/tmp/cache'",
				"-outputDropletDir='/tmp/droplet'",
				"-outputMetadataDir='/tmp/result'",
			}

			Ω(tailorConfig.Script()).Should(MatchRegexp("^/tmp/circus/tailor"))
			for _, commandFlag := range commandFlags {
				Ω(tailorConfig.Script()).To(ContainSubstring(commandFlag))
			}
		})
	})

	Context("with overrides", func() {
		BeforeEach(func() {
			tailorConfig.Set(LinuxCircusTailorAppDirFlag, "/some/app/dir")
			tailorConfig.Set(LinuxCircusTailorOutputDropletDirFlag, "/some/droplet/dir")
			tailorConfig.Set(LinuxCircusTailorOutputMetadataDirFlag, "/some/result/dir")
			tailorConfig.Set(LinuxCircusTailorBuildpacksDirFlag, "/some/buildpacks/dir")
			tailorConfig.Set(LinuxCircusTailorBuildArtifactsCacheDirFlag, "/some/cache/dir")
		})

		It("generates a script for running its tailor", func() {
			commandFlags := []string{
				"-appDir='/some/app/dir'",
				"-buildpackOrder='ocaml-buildpack,haskell-buildpack,bash-buildpack'",
				"-buildpacksDir='/some/buildpacks/dir'",
				"-buildArtifactsCacheDir='/some/cache/dir'",
				"-outputDropletDir='/some/droplet/dir'",
				"-outputMetadataDir='/some/result/dir'",
			}

			Ω(tailorConfig.Script()).Should(MatchRegexp("^/tmp/circus/tailor"))
			for _, commandFlag := range commandFlags {
				Ω(tailorConfig.Script()).To(ContainSubstring(commandFlag))
			}
		})
	})

	It("returns the path to the app bits", func() {
		Ω(tailorConfig.AppDir()).To(Equal("/app"))
	})

	It("returns the path to a given buildpack", func() {
		key := "my-buildpack/key/::"
		Ω(tailorConfig.BuildpackPath(key)).To(Equal("/tmp/buildpacks/8b2f72a0702aed614f8b5d8f7f5b431b"))
	})

	It("returns the path to the staging metadata", func() {
		Ω(tailorConfig.OutputMetadataPath()).To(Equal("/tmp/result/result.json"))
	})
})
