// +build !windows2012R2

package containerpath_test

import (
	"os"
	"path/filepath"

	"code.cloudfoundry.org/buildpackapplifecycle/containerpath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("containerpath.For", func() {
	var userProfile string
	BeforeEach(func() {
		userProfile = os.Getenv("USERPROFILE")
		os.Setenv("USERPROFILE", "oooo")
	})

	AfterEach(func() {
		os.Setenv("USERPROFILE", userProfile)
	})

	It("returns paths relative to root", func() {
		Expect(containerpath.For(filepath.FromSlash("/foo/bar/baz"))).To(Equal(filepath.FromSlash("/foo/bar/baz")))
	})
})
