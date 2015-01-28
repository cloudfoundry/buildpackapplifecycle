package buildpack_app_lifecycle_test

import (
	. "github.com/cloudfoundry-incubator/buildpack_app_lifecycle"
	. "github.com/cloudfoundry-incubator/buildpack_app_lifecycle/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry-incubator/buildpack_app_lifecycle/Godeps/_workspace/src/github.com/onsi/gomega"
)

var _ = Describe("LifecycleBuilderConfig", func() {
	var builderConfig LifecycleBuilderConfig

	BeforeEach(func() {
		builderConfig = NewLifecycleBuilderConfig([]string{"ocaml-buildpack", "haskell-buildpack", "bash-buildpack"}, false)
	})

	Context("with defaults", func() {
		It("generates a script for running its builder", func() {
			commandFlags := []string{
				"-buildDir=/tmp/app",
				"-buildpackOrder=ocaml-buildpack,haskell-buildpack,bash-buildpack",
				"-buildpacksDir=/tmp/buildpacks",
				"-buildArtifactsCacheDir=/tmp/cache",
				"-outputDroplet=/tmp/droplet",
				"-outputMetadata=/tmp/result.json",
				"-outputBuildArtifactsCache=/tmp/output-cache",
				"-skipCertVerify=false",
			}

			Ω(builderConfig.Path()).Should(Equal("/tmp/lifecycle/builder"))
			Ω(builderConfig.Args()).Should(ConsistOf(commandFlags))
		})
	})

	Context("with overrides", func() {
		BeforeEach(func() {
			builderConfig.Set("buildDir", "/some/build/dir")
			builderConfig.Set("outputDroplet", "/some/droplet")
			builderConfig.Set("outputMetadata", "/some/result/dir")
			builderConfig.Set("buildpacksDir", "/some/buildpacks/dir")
			builderConfig.Set("buildArtifactsCacheDir", "/some/cache/dir")
			builderConfig.Set("outputBuildArtifactsCache", "/some/cache-file")
			builderConfig.Set("skipCertVerify", "true")
		})

		It("generates a script for running its builder", func() {
			commandFlags := []string{
				"-buildDir=/some/build/dir",
				"-buildpackOrder=ocaml-buildpack,haskell-buildpack,bash-buildpack",
				"-buildpacksDir=/some/buildpacks/dir",
				"-buildArtifactsCacheDir=/some/cache/dir",
				"-outputDroplet=/some/droplet",
				"-outputMetadata=/some/result/dir",
				"-outputBuildArtifactsCache=/some/cache-file",
				"-skipCertVerify=true",
			}

			Ω(builderConfig.Path()).Should(Equal("/tmp/lifecycle/builder"))
			Ω(builderConfig.Args()).Should(ConsistOf(commandFlags))
		})
	})

	It("returns the path to the app bits", func() {
		Ω(builderConfig.BuildDir()).To(Equal("/tmp/app"))
	})

	It("returns the path to a given buildpack", func() {
		key := "my-buildpack/key/::"
		Ω(builderConfig.BuildpackPath(key)).To(Equal("/tmp/buildpacks/8b2f72a0702aed614f8b5d8f7f5b431b"))
	})

	It("returns the path to the staging metadata", func() {
		Ω(builderConfig.OutputMetadata()).To(Equal("/tmp/result.json"))
	})
})
