package platformoptions_test

import (
	"os"

	"code.cloudfoundry.org/buildpackapplifecycle/platformoptions"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Platformoptions", func() {
	var (
		platformOptions     *platformoptions.PlatformOptions
		err                 error
		vcapPlatformOptions string
	)

	BeforeEach(func() {
		vcapPlatformOptions = os.Getenv("VCAP_PLATFORM_OPTIONS")
	})

	JustBeforeEach(func() {
		platformOptions, err = platformoptions.Get()
	})

	AfterEach(func() {
		os.Setenv("VCAP_PLATFORM_OPTIONS", vcapPlatformOptions)
	})

	assertRemovesVcapPlatformOptions := func() {
		It("removes the VCAP_PLATFORM_OPTIONS environment variable", func() {
			_, envVarExists := os.LookupEnv("VCAP_PLATFORM_OPTIONS")
			Expect(envVarExists).To(BeFalse())
		})
	}

	Context("when VCAP_PLATFORM_OPTIONS is not set", func() {
		BeforeEach(func() {
			os.Unsetenv("VCAP_PLATFORM_OPTIONS")
		})

		It("returns nil PlatformOptions without error", func() {
			Expect(platformOptions).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		assertRemovesVcapPlatformOptions()
	})

	Context("when VCAP_PLATFORM_OPTIONS is an empty string", func() {
		BeforeEach(func() {
			os.Setenv("VCAP_PLATFORM_OPTIONS", "")
		})

		It("returns nil PlatformOptions without error", func() {
			Expect(platformOptions).To(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

		assertRemovesVcapPlatformOptions()
	})

	Context("when VCAP_PLATFORM_OPTIONS is an empty JSON object", func() {
		BeforeEach(func() {
			os.Setenv("VCAP_PLATFORM_OPTIONS", "{}")
		})

		It("returns an unset PlatformOptions", func() {
			Expect(platformOptions).NotTo(BeNil())
			Expect(err).ToNot(HaveOccurred())
			Expect(platformOptions).To(Equal(&platformoptions.PlatformOptions{}))
		})

		assertRemovesVcapPlatformOptions()
	})

	Context("when VCAP_PLATFORM_OPTIONS is an invalid JSON object", func() {
		BeforeEach(func() {
			os.Setenv("VCAP_PLATFORM_OPTIONS", `{"credhub-uri":"missing quote and brace`)
		})

		It("returns a nil PlatformOptions with an error", func() {
			Expect(platformOptions).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		assertRemovesVcapPlatformOptions()
	})

	Context("when VCAP_PLATFORM_OPTIONS is a valid JSON object", func() {
		BeforeEach(func() {
			os.Setenv("VCAP_PLATFORM_OPTIONS", `{"credhub-uri":"valid_json"}`)
		})

		It("returns populated PlatformOptions", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(platformOptions.CredhubURI).To(Equal("valid_json"))
		})

		It("returns the same populated PlatformOptions on subsequent invocations", func() {
			platformOptions, err = platformoptions.Get()
			Expect(err).ToNot(HaveOccurred())
			Expect(platformOptions).NotTo(BeNil())
			Expect(platformOptions.CredhubURI).To(Equal("valid_json"))
		})

		assertRemovesVcapPlatformOptions()
	})
})
