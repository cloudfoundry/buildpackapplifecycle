package buildpackrunner_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBuildpackrunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Buildpackrunner Suite")
}

var tmpDir string
var httpServer *httptest.Server
var gitUrl url.URL

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	tmpDir, err = ioutil.TempDir("", "tmpDir")
	Ω(err).ShouldNot(HaveOccurred())
	buildpackDir := filepath.Join(tmpDir, "fake-buildpack")
	err = os.MkdirAll(buildpackDir, os.ModePerm)
	Ω(err).ShouldNot(HaveOccurred())

	execute(buildpackDir, "rm", "-rf", ".git")
	execute(buildpackDir, "git", "init")

	err = ioutil.WriteFile(filepath.Join(buildpackDir, "content"),
		[]byte("some content"), os.ModePerm)
	Ω(err).ShouldNot(HaveOccurred())

	execute(buildpackDir, "git", "add", ".")
	execute(buildpackDir, "git", "add", "-A")
	execute(buildpackDir, "git", "commit", "-am", "fake commit")
	execute(buildpackDir, "git", "branch", "a_branch")
	execute(buildpackDir, "git", "tag", "-m", "annotated tag", "a_tag")
	execute(buildpackDir, "git", "tag", "a_lightweight_tag")
	execute(buildpackDir, "git", "update-server-info")

	httpServer = httptest.NewServer(http.FileServer(http.Dir(tmpDir)))

	gitUrl = url.URL{
		Scheme: "http",
		Host:   httpServer.Listener.Addr().String(),
		Path:   "/fake-buildpack/.git",
	}
	return nil
}, func(data []byte) {

})

var _ = SynchronizedAfterSuite(func() {
	httpServer.CloseClientConnections()
	httpServer.Close()
	os.RemoveAll(tmpDir)
}, func() {

})

func execute(dir string, execCmd string, args ...string) {
	cmd := exec.Command(execCmd, args...)
	cmd.Dir = dir
	err := cmd.Run()
	Ω(err).ShouldNot(HaveOccurred())
}
