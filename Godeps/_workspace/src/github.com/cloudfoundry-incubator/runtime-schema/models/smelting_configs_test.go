package models_test

import (
	. "github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LinuxSmeltingConfig", func() {
	var smeltingConfig LinuxSmeltingConfig
	BeforeEach(func() {
		smeltingConfig = NewLinuxSmeltingConfig([]string{"ocaml-buildpack", "haskell-buildpack", "bash-buildpack"})
	})

	It("generates a script for running its smelter", func() {
		expectedScript := "/tmp/compiler/run" +
			" -appDir /app" +
			" -outputDir /tmp/droplet" +
			" -resultDir /tmp/result" +
			" -buildpacksDir /tmp/buildpacks" +
			" -buildpackOrder ocaml-buildpack,haskell-buildpack,bash-buildpack" +
			" -cacheDir /tmp/cache"

		Ω(smeltingConfig.Script()).To(Equal(expectedScript))
	})

	It("returns the path to the compiler", func() {
		Ω(smeltingConfig.CompilerPath()).To(Equal("/tmp/compiler"))
	})

	It("returns the path to the app bits", func() {
		Ω(smeltingConfig.AppPath()).To(Equal("/app"))
	})

	It("returns the path to a given buildpack", func() {
		Ω(smeltingConfig.BuildpackPath("my-buildpack")).To(Equal("/tmp/buildpacks/my-buildpack"))
	})

	It("returns the path to the droplet.tgz", func() {
		Ω(smeltingConfig.DropletArchivePath()).To(Equal("/tmp/droplet/droplet.tgz"))
	})

	It("returns the path to the result.json", func() {
		Ω(smeltingConfig.ResultJsonPath()).To(Equal("/tmp/result/result.json"))
	})
})
