package cacheddownloader_test

import (
	. "github.com/cloudfoundry-incubator/buildpack_app_lifecycle/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry-incubator/buildpack_app_lifecycle/Godeps/_workspace/src/github.com/onsi/gomega"
	"testing"
)

func TestCachedDownloader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CachedDownloader Suite")
}
