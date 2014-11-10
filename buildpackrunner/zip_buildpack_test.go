package buildpackrunner_test

import (
	"archive/zip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"

	. "github.com/cloudfoundry-incubator/linux-circus/buildpackrunner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ZipBuildpack", func() {
	var destination string

	BeforeEach(func() {
		var err error
		destination, err = ioutil.TempDir("", "unzipdir")
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(destination)
	})

	Describe("IsZipFile", func() {
		It("returns true with .zip extension", func() {
			Ω(IsZipFile("abc.zip")).Should(BeTrue())
		})

		It("returns false without .zip extension", func() {
			Ω(IsZipFile("abc.tar")).Should(BeFalse())
		})
	})

	Describe("DownloadZipAndExtract", func() {
		var fileserver *httptest.Server

		BeforeEach(func() {
			fileserver = httptest.NewServer(http.FileServer(http.Dir(os.TempDir())))
		})

		AfterEach(func() {
			fileserver.Close()
		})

		Context("with a valid zip file", func() {
			var zipfile string

			BeforeEach(func() {
				var err error
				z, err := ioutil.TempFile("", "zipfile")
				Ω(err).ShouldNot(HaveOccurred())
				zipfile = z.Name()

				w := zip.NewWriter(z)
				f, err := w.Create("contents")
				Ω(err).ShouldNot(HaveOccurred())
				f.Write([]byte("stuff"))
				err = w.Close()
				Ω(err).ShouldNot(HaveOccurred())
			})

			AfterEach(func() {
				os.Remove(zipfile)
			})

			It("downloads and extracts", func() {
				u, _ := url.Parse(fileserver.URL)
				u.Path = filepath.Base(zipfile)
				err := DownloadZipAndExtract(u, destination)
				Ω(err).ShouldNot(HaveOccurred())
				file, err := os.Open(filepath.Join(destination, "contents"))
				Ω(err).ShouldNot(HaveOccurred())
				defer file.Close()

				bytes, err := ioutil.ReadAll(file)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(bytes).Should(Equal([]byte("stuff")))
			})
		})

		It("fails when the zip file does not exist", func() {
			u, _ := url.Parse("file:///foobar_not_there")
			err := DownloadZipAndExtract(u, destination)
			Ω(err).Should(HaveOccurred())
		})

		It("fails when the file is not a zip file", func() {
			u, _ := url.Parse(fileserver.URL)
			u.Path = filepath.Base(destination)
			err := DownloadZipAndExtract(u, destination)
			Ω(err).Should(HaveOccurred())
		})
	})
})
