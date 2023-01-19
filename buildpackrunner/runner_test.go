package buildpackrunner_test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"code.cloudfoundry.org/buildpackapplifecycle"
	"code.cloudfoundry.org/buildpackapplifecycle/buildpackrunner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Runner", func() {
	Context("StartCommand", func() {

		var runner *buildpackrunner.Runner
		var buildpacks = []string{"haskell-buildpack", "bash-buildpack"}
		var builderConfig buildpackapplifecycle.LifecycleBuilderConfig

		BeforeEach(func() {
			builderConfig = makeBuilderConfig(buildpacks)
			runner = buildpackrunner.New(&builderConfig)
			Expect(runner.Setup()).To(Succeed())
		})

		When("There is NO procfile and NO launch.yml file", func() {
			It("should use the default start command", func() {
				resultsJSON, stagingInfo, err := runner.GoLikeLightning()

				Expect(err).NotTo(HaveOccurred())
				Expect(stagingInfo).To(ContainSubstring("staging_info.yml"))
				Expect(stagingInfo).To(BeAnExistingFile())

				stagingInfoContents, err := ioutil.ReadFile(stagingInfo)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(stagingInfoContents)).To(ContainSubstring(`{"detected_buildpack":"","start_command":"I wish I was a baller"}`))

				resultsJSONContents, err := os.ReadFile(resultsJSON)
				Expect(err).ToNot(HaveOccurred())

				actualStagingResult := buildpackapplifecycle.StagingResult{}
				Expect(json.Unmarshal(resultsJSONContents, &actualStagingResult)).To(Succeed())

				Expect(actualStagingResult.ProcessTypes).To(Equal(buildpackapplifecycle.ProcessTypes{"web": "I wish I was a baller"}))
				Expect(actualStagingResult.ProcessList).To(Equal([]buildpackapplifecycle.Process{{Type: "web", Command: "I wish I was a baller"}}))
				//TODO: Find the origin of the default start command "I wish I was a baller"
			})
		})

		When("A launch.yml is present and there is NO procfile", func() {
			var launchContent = []string{`
processes:
- type: "web"
  command: "do something forever"
- type: "worker"
  command: "do something and then quit"
- type: "newrelic"
  command: "run new relic"
  platforms:
    cloudfoundry:
      sidecar_for: [ "web" , "worker" ] `, `
processes:
- type: "web"
  command: "do something else forever"
- type: "oldrelic"
  command: "run new relic"
  limits:
    memory: 10
  platforms:
    cloudfoundry:
      sidecar_for: [ "web" ] `}

			BeforeEach(func() {
				Expect(os.MkdirAll(runner.GetDepsDir(), os.ModePerm)).To(Succeed())

				for index := range buildpacks {
					depsIdxPath := filepath.Join(runner.GetDepsDir(), strconv.Itoa(index))
					Expect(os.MkdirAll(depsIdxPath, os.ModePerm)).To(Succeed())
					launchPath := filepath.Join(depsIdxPath, "launch.yml")
					Expect(ioutil.WriteFile(launchPath, []byte(launchContent[index]), os.ModePerm)).To(Succeed())
				}
			})

			AfterEach(func() {
				os.RemoveAll(runner.GetDepsDir())
			})

			It("Should use the start command from launch.yml", func() {
				resultsJSON, stagingInfo, err := runner.GoLikeLightning()

				Expect(err).NotTo(HaveOccurred())
				Expect(stagingInfo).To(ContainSubstring("staging_info.yml"))
				Expect(stagingInfo).To(BeAnExistingFile())

				stagingInfoContents, err := ioutil.ReadFile(stagingInfo)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(stagingInfoContents)).To(ContainSubstring(`{"detected_buildpack":"","start_command":"do something else forever"}`))

				resultsJSONContents, err := ioutil.ReadFile(resultsJSON)
				Expect(err).ToNot(HaveOccurred())

				actualStagingResult := buildpackapplifecycle.StagingResult{}
				Expect(json.Unmarshal(resultsJSONContents, &actualStagingResult)).To(Succeed())

				Expect(actualStagingResult.ProcessTypes).To(Equal(buildpackapplifecycle.ProcessTypes{
					"web":    "do something else forever",
					"worker": "do something and then quit",
				}))

				Expect(actualStagingResult.ProcessList).To(Equal([]buildpackapplifecycle.Process{
					{Type: "web", Command: "do something else forever"},
					{Type: "worker", Command: "do something and then quit"},
				}))

				Expect(actualStagingResult.Sidecars).To(Equal([]buildpackapplifecycle.Sidecar{
					{Name: "newrelic", ProcessTypes: []string{"web", "worker"}, Command: "run new relic"},
					{Name: "oldrelic", ProcessTypes: []string{"web"}, Command: "run new relic", Memory: 10},
				}))

			})
		})

		When("A procfile is present and there is NO launch.yml", func() {
			var procFilePath string
			BeforeEach(func() {
				procFilePath = filepath.Join(builderConfig.BuildDir(), "Procfile")
				Expect(ioutil.WriteFile(procFilePath, []byte("web: gunicorn server:app"), os.ModePerm)).To(Succeed())
			})

			AfterEach(func() {
				os.Remove(procFilePath)
			})

			It("Should always use the start command from the procfile", func() {
				resultsJSON, stagingInfo, err := runner.GoLikeLightning()

				Expect(err).NotTo(HaveOccurred())
				Expect(stagingInfo).To(ContainSubstring("staging_info.yml"))
				Expect(stagingInfo).To(BeAnExistingFile())

				contents, err := ioutil.ReadFile(stagingInfo)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring(`{"detected_buildpack":"","start_command":"gunicorn server:app"}`))

				resultsJSONContents, err := ioutil.ReadFile(resultsJSON)
				Expect(err).ToNot(HaveOccurred())

				actualStagingResult := buildpackapplifecycle.StagingResult{}
				Expect(json.Unmarshal(resultsJSONContents, &actualStagingResult)).To(Succeed())

				Expect(actualStagingResult.ProcessTypes).To(Equal(buildpackapplifecycle.ProcessTypes{"web": "gunicorn server:app"}))
				Expect(actualStagingResult.ProcessList).To(Equal([]buildpackapplifecycle.Process{{Type: "web", Command: "gunicorn server:app"}}))
			})
		})

		When("there is NO procfile present and there is launch.yml provided by supply buildpacks", func() {
			var launchContents = `
processes:
- type: "web"
  command: "do something forever"
- type: "worker"
  command: "do something and then quit"
- type: "lightning"
  command: "go forth"
- type: "newrelic"
  command: "run new relic"
  platforms:
    cloudfoundry:
      sidecar_for: [ "web" ]`

			BeforeEach(func() {
				depsIdxPath := filepath.Join(runner.GetDepsDir(), strconv.Itoa(0))
				Expect(os.MkdirAll(depsIdxPath, os.ModePerm)).To(Succeed())
				launchPath := filepath.Join(depsIdxPath, "launch.yml")
				Expect(ioutil.WriteFile(launchPath, []byte(launchContents), os.ModePerm)).To(Succeed())
			})

			It("Should always use the start command from the bin/release", func() {
				resultsJSON, stagingInfo, err := runner.GoLikeLightning()

				Expect(err).NotTo(HaveOccurred())
				Expect(stagingInfo).To(ContainSubstring("staging_info.yml"))
				Expect(stagingInfo).To(BeAnExistingFile())

				stagingInfoContents, err := ioutil.ReadFile(stagingInfo)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(stagingInfoContents)).To(ContainSubstring(`{"detected_buildpack":"","start_command":"I wish I was a baller"}`))

				resultsJSONContents, err := ioutil.ReadFile(resultsJSON)
				Expect(err).ToNot(HaveOccurred())

				actualStagingResult := buildpackapplifecycle.StagingResult{}
				Expect(json.Unmarshal(resultsJSONContents, &actualStagingResult)).To(Succeed())

				Expect(actualStagingResult.ProcessTypes).To(Equal(buildpackapplifecycle.ProcessTypes{
					"lightning": "go forth",
					"web":       "I wish I was a baller",
					"worker":    "do something and then quit",
				}))

				Expect(actualStagingResult.ProcessList).To(Equal([]buildpackapplifecycle.Process{
					{Type: "web", Command: "I wish I was a baller"},
					{Type: "worker", Command: "do something and then quit"},
					{Type: "lightning", Command: "go forth"},
				}))

				Expect(actualStagingResult.Sidecars).To(Equal([]buildpackapplifecycle.Sidecar{
					{Name: "newrelic", ProcessTypes: []string{"web"}, Command: "run new relic"},
				}))
			})
		})

		When("A procfile is present and there is launch.yml provided by all buildpacks", func() {
			var procFilePath string
			var launchContent = []string{`
processes:
- type: "web"
  command: "do something forever"
- type: "worker"
  command: "do something and then quit"
- type: "lightning"
  command: "go forth"
- type: "newrelic"
  command: "run new relic"
  platforms:
    cloudfoundry:
      sidecar_for: [ "web" ] `, `
processes:
- type: "worker"
  command: "do something else forever"
- type: "oldrelic"
  command: "run new relic"
  limits:
    memory: 10
  platforms:
    cloudfoundry:
      sidecar_for: [ "worker" ] `}

			BeforeEach(func() {
				procFilePath := filepath.Join(builderConfig.BuildDir(), "Procfile")
				Expect(ioutil.WriteFile(procFilePath, []byte("web: gunicorn server:app"), os.ModePerm)).To(Succeed())

				for index := range buildpacks {
					depsIdxPath := filepath.Join(runner.GetDepsDir(), strconv.Itoa(index))
					Expect(os.MkdirAll(depsIdxPath, os.ModePerm)).To(Succeed())
					launchPath := filepath.Join(depsIdxPath, "launch.yml")
					Expect(ioutil.WriteFile(launchPath, []byte(launchContent[index]), os.ModePerm)).To(Succeed())
				}
			})

			AfterEach(func() {
				os.Remove(procFilePath)
			})

			It("Should always use the start command from the procfile", func() {
				resultsJSON, stagingInfo, err := runner.GoLikeLightning()

				Expect(err).NotTo(HaveOccurred())
				Expect(stagingInfo).To(ContainSubstring("staging_info.yml"))
				Expect(stagingInfo).To(BeAnExistingFile())

				stagingInfoContents, err := ioutil.ReadFile(stagingInfo)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(stagingInfoContents)).To(ContainSubstring(`{"detected_buildpack":"","start_command":"gunicorn server:app"}`))

				resultsJSONContents, err := ioutil.ReadFile(resultsJSON)
				Expect(err).ToNot(HaveOccurred())

				actualStagingResult := buildpackapplifecycle.StagingResult{}
				Expect(json.Unmarshal(resultsJSONContents, &actualStagingResult)).To(Succeed())

				Expect(actualStagingResult.ProcessTypes).To(Equal(buildpackapplifecycle.ProcessTypes{
					"lightning": "go forth",
					"web":       "gunicorn server:app",
					"worker":    "do something else forever",
				}))

				Expect(actualStagingResult.ProcessList).To(Equal([]buildpackapplifecycle.Process{
					{Type: "web", Command: "gunicorn server:app"},
					{Type: "worker", Command: "do something else forever"},
					{Type: "lightning", Command: "go forth"},
				}))

				Expect(actualStagingResult.Sidecars).To(Equal([]buildpackapplifecycle.Sidecar{
					{Name: "newrelic", ProcessTypes: []string{"web"}, Command: "run new relic"},
					{Name: "oldrelic", ProcessTypes: []string{"worker"}, Command: "run new relic", Memory: 10},
				}))
			})
		})
	})
})

func makeBuilderConfig(buildpacks []string) buildpackapplifecycle.LifecycleBuilderConfig {
	skipDetect := true
	builderConfig := buildpackapplifecycle.NewLifecycleBuilderConfig(buildpacks, skipDetect, false)
	outputMetadataPath, err := ioutil.TempDir(os.TempDir(), "results")
	Expect(err).ToNot(HaveOccurred())
	Expect(builderConfig.Set("outputMetadata", filepath.Join(outputMetadataPath, "results.json"))).To(Succeed())

	buildDirPath, err := ioutil.TempDir(os.TempDir(), "app")
	Expect(err).ToNot(HaveOccurred())
	Expect(builderConfig.Set("buildDir", buildDirPath)).To(Succeed())

	buildpacksDirPath, err := ioutil.TempDir(os.TempDir(), "buildpack")
	Expect(err).ToNot(HaveOccurred())
	Expect(builderConfig.Set("buildpacksDir", buildpacksDirPath)).To(Succeed())

	for _, bp := range buildpacks {
		bpPath := builderConfig.BuildpackPath(bp)
		Expect(genFakeBuildpack(bpPath)).To(Succeed())
	}

	err = os.MkdirAll(builderConfig.BuildDir(), os.ModePerm)
	Expect(err).ToNot(HaveOccurred())

	if runtime.GOOS == "windows" {
		copyDst := filepath.Join(filepath.Dir(builderConfig.Path()), "tar.exe")
		CopyFileWindows(tmpTarPath, copyDst)
	}

	return builderConfig
}

func genFakeBuildpack(bpRoot string) error {
	err := os.MkdirAll(filepath.Join(bpRoot, "bin"), os.ModePerm)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		CopyDirectory(filepath.Join("testdata", "fake_windows_bp", "bin", "*"), filepath.Join(bpRoot, "bin"))
	} else {
		CopyDirectory(filepath.Join("testdata", "fake_unix_bp", "bin"), filepath.Join(bpRoot))
	}
	return nil
}

func CopyFileWindows(src string, dst string) {
	s, err := os.Open(src)
	Expect(err).ToNot(HaveOccurred())

	defer s.Close()

	i, err := s.Stat()
	Expect(err).ToNot(HaveOccurred())

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	Expect(err).ToNot(HaveOccurred())

	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, i.Mode())
	Expect(err).ToNot(HaveOccurred())

	defer f.Close()

	_, err = io.Copy(f, s)
	Expect(err).ToNot(HaveOccurred())

}

func CopyDirectory(src string, dst string) {
	var command *exec.Cmd
	if runtime.GOOS == "windows" {
		command = exec.Command("powershell", "-Command", "Copy-Item", "-Recurse", "-Force", src, dst)
	} else {
		command = exec.Command("cp", "-a", "-R", src, dst)
	}

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	session.Wait()
	Expect(session).Should(gexec.Exit())
}
