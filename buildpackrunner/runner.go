package buildpackrunner

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"code.cloudfoundry.org/buildpackapplifecycle"
	"code.cloudfoundry.org/bytefmt"
	"github.com/google/uuid"
)

const DOWNLOAD_TIMEOUT = 10 * time.Minute

type Runner interface {
	Run(config *buildpackapplifecycle.LifecycleBuilderConfig) (string, error)
}

type runner struct {
	config        *buildpackapplifecycle.LifecycleBuilderConfig
	zipDownloader ZipDownloader
}

type descriptiveError struct {
	message string
	err     error
}

type Release struct {
	DefaultProcessTypes buildpackapplifecycle.ProcessTypes `yaml:"default_process_types"`
}

func (e descriptiveError) Error() string {
	if e.err == nil {
		return e.message
	}
	return fmt.Sprintf("%s: %s", e.message, e.err.Error())
}

func newDescriptiveError(err error, message string, args ...interface{}) error {
	if len(args) == 0 {
		return descriptiveError{message: message, err: err}
	}
	return descriptiveError{message: fmt.Sprintf(message, args...), err: err}
}

func New(zipDownloader ZipDownloader) Runner {
	return &runner{
		zipDownloader: zipDownloader,
	}
}

func (runner *runner) Run(config *buildpackapplifecycle.LifecycleBuilderConfig) (string, error) {
	runner.config = config

	//set up the world
	err := runner.makeDirectories()
	if err != nil {
		return "", newDescriptiveError(err, "Failed to set up filesystem when generating droplet")
	}

	err = runner.downloadBuildpacks()
	if err != nil {
		return "", err
	}

	//detect, compile, release
	var detectedBuildpack, detectOutput, detectedBuildpackDir string
	var ok bool

	if config.SkipDetect() {
		detectedBuildpackDir, ok = runner.supply()
	} else {
		detectedBuildpack, detectedBuildpackDir, detectOutput, ok = runner.detect()
	}
	if !ok {
		return "", newDescriptiveError(nil, buildpackapplifecycle.DetectFailMsg)
	}

	err = runner.compile(detectedBuildpackDir)
	if err != nil {
		return "", newDescriptiveError(nil, buildpackapplifecycle.CompileFailMsg)
	}

	startCommands, err := runner.readProcfile()
	if err != nil {
		return "", newDescriptiveError(err, "Failed to read command from Procfile")
	}

	releaseInfo, err := runner.release(detectedBuildpackDir, startCommands)
	if err != nil {
		return "", newDescriptiveError(err, buildpackapplifecycle.ReleaseFailMsg)
	}

	if releaseInfo.DefaultProcessTypes["web"] == "" {
		printError("No start command specified by buildpack or via Procfile.")
		printError("App will not start unless a command is provided at runtime.")
	}

	tarPath, err := exec.LookPath("tar")
	if err != nil {
		return "", err
	}

	contentsDir := runner.config.BuildRootDir()

	//generate staging_info.yml and result json file
	infoFilePath := path.Join(contentsDir, "staging_info.yml")
	err = runner.saveInfo(infoFilePath, detectedBuildpack, detectOutput, releaseInfo)
	if err != nil {
		return "", newDescriptiveError(err, "Failed to encode generated metadata")
	}

	for _, name := range []string{"tmp", "logs"} {
		if err := os.RemoveAll(path.Join(contentsDir, name)); err != nil {
			return "", newDescriptiveError(err, "Failed to set up droplet filesystem")
		}

		if err := os.MkdirAll(path.Join(contentsDir, name), 0755); err != nil {
			return "", newDescriptiveError(err, "Failed to set up droplet filesystem")
		}
	}

	if path.Base(runner.config.BuildDir()) != "app" {
		os.RemoveAll(filepath.Join(runner.config.BuildRootDir(), "app"))

		err = os.Rename(runner.config.BuildDir(), filepath.Join(runner.config.BuildRootDir(), "app"))
		if err != nil {
			return "", newDescriptiveError(err, "Failed to set up droplet filesystem")
		}
	}

	err = exec.Command(tarPath, "-czf", runner.config.OutputDroplet(), "-C", contentsDir, "./app", "./deps", "./staging_info.yml", "./tmp", "./logs").Run()
	if err != nil {
		return "", newDescriptiveError(err, "Failed to compress droplet filesystem")
	}

	//prepare the build artifacts cache output directory
	err = os.MkdirAll(filepath.Dir(runner.config.OutputBuildArtifactsCache()), 0755)
	if err != nil {
		return "", newDescriptiveError(err, "Failed to create output build artifacts cache dir")
	}

	err = exec.Command(tarPath, "-czf", runner.config.OutputBuildArtifactsCache(), "-C", runner.config.BuildArtifactsCacheDir(), ".").Run()
	if err != nil {
		return "", newDescriptiveError(err, "Failed to compress build artifacts")
	}

	return infoFilePath, nil
}

func (runner *runner) makeDirectories() error {
	if err := os.MkdirAll(filepath.Dir(runner.config.OutputDroplet()), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(runner.config.OutputMetadata()), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(runner.config.BuildArtifactsCacheDir(), 0755); err != nil {
		return err
	}

	if runner.config.IsMultiBuildpack() {
		if err := os.MkdirAll(filepath.Join(runner.config.BuildArtifactsCacheDir(), "primary"), 0755); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(runner.config.DepsDir(), 0755); err != nil {
		return err
	}

	return nil
}

func (runner *runner) downloadBuildpacks() error {
	// Do we have a custom buildpack?
	for _, buildpackName := range runner.config.BuildpackOrder() {
		buildpackUrl, err := url.Parse(buildpackName)
		if err != nil {
			return fmt.Errorf("Invalid buildpack url (%s): %s", buildpackName, err.Error())
		}
		if !buildpackUrl.IsAbs() {
			continue
		}

		destination := runner.config.BuildpackPath(buildpackName)

		if IsZipFile(buildpackUrl.Path) {
			var size uint64
			size, err = runner.zipDownloader.DownloadAndExtract(buildpackUrl, destination)
			if err == nil {
				fmt.Printf("Downloaded buildpack `%s` (%s)\n", buildpackUrl.String(), bytefmt.ByteSize(size))
			}
		} else {
			err = GitClone(*buildpackUrl, destination)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (runner *runner) buildpackPath(buildpack string) (string, error) {
	buildpackPath := runner.config.BuildpackPath(buildpack)

	if runner.pathHasBinDirectory(buildpackPath) {
		return buildpackPath, nil
	}

	files, err := ioutil.ReadDir(buildpackPath)
	if err != nil {
		return "", newDescriptiveError(nil, "Failed to read buildpack directory '%s' for buildpack '%s'", buildpackPath, buildpack)
	}

	if len(files) == 1 {
		nestedPath := path.Join(buildpackPath, files[0].Name())

		if runner.pathHasBinDirectory(nestedPath) {
			return nestedPath, nil
		}
	}

	return "", newDescriptiveError(nil, "malformed buildpack does not contain a /bin dir: %s", buildpack)
}

func (runner *runner) pathHasBinDirectory(pathToTest string) bool {
	_, err := os.Stat(path.Join(pathToTest, "bin"))
	return err == nil
}

func (runner *runner) supplyCachePath(buildpack string) string {
	return filepath.Join(runner.config.BuildArtifactsCacheDir(), fmt.Sprintf("%x", md5.Sum([]byte(buildpack))))
}

// returns buildpack path, ok
func (runner *runner) supply() (string, bool) {
	buildpacks := runner.config.BuildpackOrder()
	supplyBuildpacks := buildpacks[0:(len(buildpacks) - 1)]
	compileBuildpack := buildpacks[len(buildpacks)-1]

	for _, buildpack := range supplyBuildpacks {
		buildpackPath, err := runner.buildpackPath(buildpack)
		if err != nil {
			printError(err.Error())
			continue
		}

		output := new(bytes.Buffer)
		guid := uuid.Must(uuid.NewRandom()).String()
		err = os.MkdirAll(path.Join(runner.config.DepsDir(), guid), 0755)
		if err != nil {
			printError(err.Error())
			continue
		}

		err = os.MkdirAll(runner.supplyCachePath(buildpack), 0755)
		if err != nil {
			printError(err.Error())
			continue
		}

		err = runner.run(exec.Command(path.Join(buildpackPath, "bin", "supply"), runner.config.BuildDir(), runner.supplyCachePath(buildpack), guid, runner.config.DepsDir()), output)
		if err != nil {
			printError(err.Error())
			continue
		}
	}

	buildpackPath, err := runner.buildpackPath(compileBuildpack)
	return buildpackPath, (err == nil)
}

// returns buildpack name,  buildpack path, buildpack detect output, ok
func (runner *runner) detect() (string, string, string, bool) {
	for _, buildpack := range runner.config.BuildpackOrder() {

		buildpackPath, err := runner.buildpackPath(buildpack)
		if err != nil {
			printError(err.Error())
			continue
		}

		output := new(bytes.Buffer)
		err = runner.run(exec.Command(path.Join(buildpackPath, "bin", "detect"), runner.config.BuildDir()), output)

		if err == nil {
			return buildpack, buildpackPath, strings.TrimRight(output.String(), "\n"), true
		}
	}

	return "", "", "", false
}

func (runner *runner) readProcfile() (map[string]string, error) {
	processes := map[string]string{}

	procFile, err := ioutil.ReadFile(filepath.Join(runner.config.BuildDir(), "Procfile"))
	if err != nil {
		if os.IsNotExist(err) {
			// Procfiles are optional
			return processes, nil
		}

		return processes, err
	}

	err = yaml.Unmarshal(procFile, &processes)
	if err != nil {
		// clobber yaml parsing  error
		return processes, errors.New("invalid YAML")
	}

	return processes, nil
}

func (runner *runner) compile(buildpackDir string) error {
	compileCacheDir := runner.config.BuildArtifactsCacheDir()
	if runner.config.IsMultiBuildpack() {
		compileCacheDir = filepath.Join(compileCacheDir, "primary")
	}

	return runner.run(exec.Command(path.Join(buildpackDir, "bin", "compile"), runner.config.BuildDir(), compileCacheDir, "", runner.config.DepsDir()), os.Stdout)
}

func (runner *runner) release(buildpackDir string, startCommands map[string]string) (Release, error) {
	output := new(bytes.Buffer)

	err := runner.run(exec.Command(path.Join(buildpackDir, "bin", "release"), runner.config.BuildDir()), output)
	if err != nil {
		return Release{}, err
	}

	parsedRelease := Release{}

	err = yaml.Unmarshal(output.Bytes(), &parsedRelease)
	if err != nil {
		return Release{}, newDescriptiveError(err, "buildpack's release output invalid")
	}

	if len(startCommands) > 0 {
		parsedRelease.DefaultProcessTypes = startCommands
	}

	return parsedRelease, nil
}

func (runner *runner) saveInfo(infoFilePath, buildpack, detectOutput string, releaseInfo Release) error {
	deaInfoFile, err := os.Create(infoFilePath)
	if err != nil {
		return err
	}
	defer deaInfoFile.Close()

	// JSON ⊂ YAML
	err = json.NewEncoder(deaInfoFile).Encode(DeaStagingInfo{
		DetectedBuildpack: detectOutput,
		StartCommand:      releaseInfo.DefaultProcessTypes["web"],
	})
	if err != nil {
		return err
	}

	resultFile, err := os.Create(runner.config.OutputMetadata())
	if err != nil {
		return err
	}
	defer resultFile.Close()

	err = json.NewEncoder(resultFile).Encode(buildpackapplifecycle.NewStagingResult(
		releaseInfo.DefaultProcessTypes,
		buildpackapplifecycle.LifecycleMetadata{
			BuildpackKey:      buildpack,
			DetectedBuildpack: detectOutput,
		},
	))
	if err != nil {
		return err
	}

	return nil
}

func (runner *runner) run(cmd *exec.Cmd, output io.Writer) error {
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func printError(message string) {
	println(message)
}
