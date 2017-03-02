package buildpackrunner_test

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/buildpackapplifecycle"
	"code.cloudfoundry.org/buildpackapplifecycle/buildpackrunner"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=zip_buildpack.go --destination=mocks_zip_buildpack_test.go --package=buildpackrunner_test

var _ = Describe("Runner", func() {
	var (
		config            buildpackapplifecycle.LifecycleBuilderConfig
		runner            buildpackrunner.Runner
		mockZipDownloader *MockZipDownloader
		buildpacks        []*url.URL
		rootDir           string
		buildDir          string
		infoFilePath      string
	)

	BeforeEach(func() {
		mockCtrl := gomock.NewController(GinkgoT())
		mockZipDownloader = NewMockZipDownloader(mockCtrl)
		buildpackStrings := []string{"http://example.com/buildpack_one.zip", "http://example.com/buildpack_two.zip", "http://example.com/buildpack_three.zip"}
		buildpacks = make([]*url.URL, len(buildpackStrings), len(buildpackStrings))
		for i, b := range buildpackStrings {
			buildpacks[i], _ = url.Parse(b)
		}

		config = buildpackapplifecycle.NewLifecycleBuilderConfig(buildpackStrings, false, false)
		rootDir, _ = ioutil.TempDir("", "buildpackapplifecycle.root")
		buildDir, _ = ioutil.TempDir(rootDir, "buildpackapplifecycle.buildir")

		config.Set("buildDir", buildDir)
		runner = buildpackrunner.New(mockZipDownloader)
	})

	AfterEach(func() {
		os.RemoveAll(buildDir)
	})

	JustBeforeEach(func() {
		old := os.Stdout
		os.Stdout = nil
		var err error
		infoFilePath, err = runner.Run(&config)
		os.Stdout = old

		Expect(err).To(BeNil())
	})

	Context("skipDetect == false and multiple buildpacks", func() {
		BeforeEach(func() {
			config.Set("skipDetect", "false")

			mockZipDownloader.EXPECT().DownloadAndExtract(buildpacks[0], gomock.Any()).Do(func(_ *url.URL, dir string) {
				os.MkdirAll(filepath.Join(dir, "bin"), 0755)
				ioutil.WriteFile(filepath.Join(dir, "bin", "detect"), []byte("#!/usr/bin/env bash\n\nexit 142\n"), 0755)
			}).Return(uint64(10), nil)
			mockZipDownloader.EXPECT().DownloadAndExtract(buildpacks[1], gomock.Any()).Do(func(_ *url.URL, dir string) {
				os.MkdirAll(filepath.Join(dir, "bin"), 0755)
				ioutil.WriteFile(filepath.Join(dir, "bin", "detect"), []byte("#!/usr/bin/env bash\n\necho Two\nexit 0\n"), 0755)
				ioutil.WriteFile(filepath.Join(dir, "bin", "compile"), []byte("#!/usr/bin/env bash\n\nexit 0\n"), 0755)
				ioutil.WriteFile(filepath.Join(dir, "bin", "release"), []byte("#!/usr/bin/env bash\n\necho ---\necho default_process_types:\necho '  web: echo Buildpack Two Run'\nexit 0\n"), 0755)
			}).Return(uint64(10), nil)
			mockZipDownloader.EXPECT().DownloadAndExtract(buildpacks[2], gomock.Any()).Return(uint64(10), nil)
		})

		It("runs compile/release on the first buildpack for which detect succeeds", func() {
			infoFile, err := ioutil.ReadFile(infoFilePath)
			Expect(err).To(BeNil())
			Expect(string(infoFile)).To(Equal(`{"detected_buildpack":"Two","start_command":"echo Buildpack Two Run"}` + "\n"))
		})
	})

	Context("skipDetect == true and multiple buildpacks", func() {
		BeforeEach(func() {
			config.Set("skipDetect", "true")

			mockZipDownloader.EXPECT().DownloadAndExtract(buildpacks[0], gomock.Any()).Do(func(_ *url.URL, dir string) {
				os.MkdirAll(filepath.Join(dir, "bin"), 0755)
				ioutil.WriteFile(filepath.Join(dir, "bin", "detect"), []byte("#!/usr/bin/env bash\n\necho One\nexit 0\n"), 0755)
				ioutil.WriteFile(filepath.Join(dir, "bin", "supply"), []byte("#!/usr/bin/env bash\n\necho 'name: One' > $4/$3/spec.yml\nexit 0\n"), 0755)
			}).Return(uint64(10), nil)
			mockZipDownloader.EXPECT().DownloadAndExtract(buildpacks[1], gomock.Any()).Do(func(_ *url.URL, dir string) {
				os.MkdirAll(filepath.Join(dir, "bin"), 0755)
				ioutil.WriteFile(filepath.Join(dir, "bin", "detect"), []byte("#!/usr/bin/env bash\n\necho Two\nexit 0\n"), 0755)
				ioutil.WriteFile(filepath.Join(dir, "bin", "supply"), []byte("#!/usr/bin/env bash\n\necho 'name: Two' > $4/$3/spec.yml\necho $4/$3/spec.yml Two\nexit 0\n"), 0755)
			}).Return(uint64(10), nil)
			mockZipDownloader.EXPECT().DownloadAndExtract(buildpacks[2], gomock.Any()).Do(func(_ *url.URL, dir string) {
				os.MkdirAll(filepath.Join(dir, "bin"), 0755)
				ioutil.WriteFile(filepath.Join(dir, "bin", "detect"), []byte("#!/usr/bin/env bash\n\necho Three\nexit 0\n"), 0755)
				ioutil.WriteFile(filepath.Join(dir, "bin", "supply"), []byte("#!/usr/bin/env bash\n\necho 'name: Three' > $4/$3/spec.yml\nexit 0\n"), 0755)
				ioutil.WriteFile(filepath.Join(dir, "bin", "compile"), []byte("#!/usr/bin/env bash\n\nexit 0\n"), 0755)
				ioutil.WriteFile(filepath.Join(dir, "bin", "release"), []byte("#!/usr/bin/env bash\n\necho ---\necho default_process_types:\necho '  web: echo Buildpack Three Run'\nexit 0\n"), 0755)
			}).Return(uint64(10), nil)
		})

		It("runs compile/release for only the last buildpack", func() {
			infoFile, err := ioutil.ReadFile(infoFilePath)
			Expect(err).To(BeNil())
			Expect(string(infoFile)).To(Equal(`{"detected_buildpack":"","start_command":"echo Buildpack Three Run"}` + "\n"))
		})

		It("runs supply for all but the last buildpack", func() {
			specyamls := make([]string, 0)
			files, _ := filepath.Glob(filepath.Join(rootDir, "deps", "*", "spec.yml"))
			for _, file := range files {
				body, _ := ioutil.ReadFile(file)
				specyamls = append(specyamls, string(body))
			}

			Expect(specyamls).To(ConsistOf("name: One\n", "name: Two\n"))
		})
	})
})
