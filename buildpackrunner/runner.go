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

	"github.com/cloudfoundry-incubator/candiedyaml"

	"code.cloudfoundry.org/buildpackapplifecycle"
	"code.cloudfoundry.org/bytefmt"
)

const DOWNLOAD_TIMEOUT = 10 * time.Minute

type Runner struct {
	config      *buildpackapplifecycle.LifecycleBuilderConfig
	depsDir     string
	contentsDir string
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

func New(config *buildpackapplifecycle.LifecycleBuilderConfig) *Runner {
	return &Runner{
		config: config,
	}
}

func (runner *Runner) Run() (string, error) {
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

	if runner.config.IsMultiBuildpack() {
		err = runner.cleanCacheDir()
		if err != nil {
			return "", err
		}

		detectedBuildpackDir, err = runner.runMultiBuildpacks()
		if err != nil {
			return "", err
		}
	} else {
		detectedBuildpack, detectedBuildpackDir, detectOutput, ok = runner.detect()
		if !ok {
			return "", newDescriptiveError(nil, buildpackapplifecycle.DetectFailMsg)
		}

		err = runner.compile(detectedBuildpackDir, runner.config.BuildArtifactsCacheDir())

		if err != nil {
			return "", newDescriptiveError(nil, buildpackapplifecycle.CompileFailMsg)
		}
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

	//generate staging_info.yml and result json file
	infoFilePath := path.Join(runner.contentsDir, "staging_info.yml")
	err = runner.saveInfo(infoFilePath, detectedBuildpack, detectOutput, releaseInfo)
	if err != nil {
		return "", newDescriptiveError(err, "Failed to encode generated metadata")
	}

	for _, name := range []string{"tmp", "logs"} {
		if err := os.MkdirAll(path.Join(runner.contentsDir, name), 0755); err != nil {
			return "", newDescriptiveError(err, "Failed to set up droplet filesystem")
		}
	}

	appDir := path.Join(runner.contentsDir, "app")
	err = runner.copyApp(runner.config.BuildDir(), appDir)
	if err != nil {
		return "", newDescriptiveError(err, "Failed to copy compiled droplet")
	}

	err = exec.Command(tarPath, "-czf", runner.config.OutputDroplet(), "-C", runner.contentsDir, ".").Run()
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

func (runner *Runner) makeDirectories() error {
	if err := os.MkdirAll(filepath.Dir(runner.config.OutputDroplet()), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(runner.config.OutputMetadata()), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(runner.config.BuildArtifactsCacheDir(), 0755); err != nil {
		return err
	}

	var err error
	runner.contentsDir, err = ioutil.TempDir("", "contents")
	if err != nil {
		return err
	}

	runner.depsDir = filepath.Join(runner.contentsDir, "deps")
	if err := os.MkdirAll(runner.depsDir, 0755); err != nil {
		return err
	}

	if runner.config.IsMultiBuildpack() {
		if err := os.MkdirAll(filepath.Join(runner.config.BuildArtifactsCacheDir(), "primary"), 0755); err != nil {
			return err
		}

		for _, buildpack := range runner.config.SupplyBuildpacks() {
			if err := os.MkdirAll(runner.supplyCachePath(buildpack), 0755); err != nil {
				return err
			}
		}

		for _, index := range runner.config.DepsIndices() {
			if err := os.MkdirAll(path.Join(runner.depsDir, index), 0755); err != nil {
				return err
			}
		}
	}

	return nil
}

func (runner *Runner) downloadBuildpacks() error {
	// Do we have a custom buildpack?
	for _, buildpackName := range runner.config.BuildpackOrder() {
		buildpackURL, err := url.Parse(buildpackName)
		if err != nil {
			return fmt.Errorf("Invalid buildpack url (%s): %s", buildpackName, err.Error())
		}
		if !buildpackURL.IsAbs() {
			continue
		}

		destination := runner.config.BuildpackPath(buildpackName)

		if IsZipFile(buildpackURL.Path) {
			zipDownloader := NewZipDownloader(runner.config.SkipCertVerify())
			size, err := zipDownloader.DownloadAndExtract(buildpackURL, destination)
			if err == nil {
				fmt.Printf("Downloaded buildpack `%s` (%s)\n", buildpackURL.String(), bytefmt.ByteSize(size))
			}
		} else {
			err = GitClone(*buildpackURL, destination)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (runner *Runner) cleanCacheDir() error {
	neededCacheDirs := map[string]bool{
		filepath.Join(runner.config.BuildArtifactsCacheDir(), "primary"): true,
	}

	for _, bp := range runner.config.SupplyBuildpacks() {
		neededCacheDirs[runner.supplyCachePath(bp)] = true
	}

	dirs, err := ioutil.ReadDir(runner.config.BuildArtifactsCacheDir())
	if err != nil {
		return err
	}

	for _, dirInfo := range dirs {
		dir := filepath.Join(runner.config.BuildArtifactsCacheDir(), dirInfo.Name())
		if !neededCacheDirs[dir] {
			err = os.RemoveAll(dir)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (runner *Runner) buildpackPath(buildpack string) (string, error) {
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

func (runner *Runner) pathHasBinDirectory(pathToTest string) bool {
	_, err := os.Stat(path.Join(pathToTest, "bin"))
	return err == nil
}

func (runner *Runner) supplyCachePath(buildpack string) string {
	return filepath.Join(runner.config.BuildArtifactsCacheDir(), fmt.Sprintf("%x", md5.Sum([]byte(buildpack))))
}

func hasFinalize(buildpackPath string) (bool, error) {
	_, err := os.Stat(filepath.Join(buildpackPath, "bin", "finalize"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func hasSupply(buildpackPath string) (bool, error) {
	_, err := os.Stat(filepath.Join(buildpackPath, "bin", "supply"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// returns buildpack path, ok
func (runner *Runner) runMultiBuildpacks() (string, error) {
	for i, buildpack := range runner.config.SupplyBuildpacks() {
		buildpackPath, err := runner.buildpackPath(buildpack)
		if err != nil {
			printError(err.Error())
			return "", newDescriptiveError(err, buildpackapplifecycle.SupplyFailMsg)
		}

		err = runner.run(exec.Command(path.Join(buildpackPath, "bin", "supply"), runner.config.BuildDir(), runner.supplyCachePath(buildpack), runner.depsDir, runner.config.DepsIndices()[i]), os.Stdout)
		if err != nil {
			return "", newDescriptiveError(err, buildpackapplifecycle.SupplyFailMsg)
		}
	}

	return runner.runFinalBuildpack()
}

func (runner *Runner) runFinalBuildpack() (string, error) {
	buildpackPath, err := runner.buildpackPath(runner.config.FinalBuildpack())
	if err != nil {
		printError(err.Error())
		return "", newDescriptiveError(err, buildpackapplifecycle.FinalizeFailMsg)
	}

	depsIndex := runner.config.FinalDepsIndex()
	cacheDir := filepath.Join(runner.config.BuildArtifactsCacheDir(), "primary")

	hasFinalize, err := hasFinalize(buildpackPath)
	if err != nil {
		return "", newDescriptiveError(err, buildpackapplifecycle.FinalizeFailMsg)
	}

	if hasFinalize {
		hasSupply, err := hasSupply(buildpackPath)
		if err != nil {
			return "", newDescriptiveError(err, buildpackapplifecycle.SupplyFailMsg)
		}

		if hasSupply {
			if err := runner.run(exec.Command(path.Join(buildpackPath, "bin", "supply"), runner.config.BuildDir(), cacheDir, runner.depsDir, depsIndex), os.Stdout); err != nil {
				return "", newDescriptiveError(err, buildpackapplifecycle.SupplyFailMsg)
			}
		}

		if err := runner.run(exec.Command(path.Join(buildpackPath, "bin", "finalize"), runner.config.BuildDir(), cacheDir, runner.depsDir, depsIndex), os.Stdout); err != nil {
			return "", newDescriptiveError(err, buildpackapplifecycle.FinalizeFailMsg)
		}
	} else {
		// remove unused deps sub dir
		if err := os.RemoveAll(filepath.Join(runner.depsDir, depsIndex)); err != nil {
			return "", newDescriptiveError(err, buildpackapplifecycle.CompileFailMsg)
		}

		if err := runner.compile(buildpackPath, cacheDir); err != nil {
			return "", newDescriptiveError(err, buildpackapplifecycle.CompileFailMsg)
		}
	}

	return buildpackPath, nil
}

// returns buildpack name,  buildpack path, buildpack detect output, ok
func (runner *Runner) detect() (string, string, string, bool) {
	for _, buildpack := range runner.config.BuildpackOrder() {

		buildpackPath, err := runner.buildpackPath(buildpack)
		if err != nil {
			printError(err.Error())
			continue
		}

		if runner.config.SkipDetect() {
			return buildpack, buildpackPath, "", true
		}

		output := new(bytes.Buffer)
		err = runner.run(exec.Command(path.Join(buildpackPath, "bin", "detect"), runner.config.BuildDir()), output)

		if err == nil {
			return buildpack, buildpackPath, strings.TrimRight(output.String(), "\n"), true
		}
	}

	return "", "", "", false
}

func (runner *Runner) readProcfile() (map[string]string, error) {
	processes := map[string]string{}

	procFile, err := os.Open(filepath.Join(runner.config.BuildDir(), "Procfile"))
	if err != nil {
		if os.IsNotExist(err) {
			// Procfiles are optional
			return processes, nil
		}

		return processes, err
	}
	defer procFile.Close()

	err = candiedyaml.NewDecoder(procFile).Decode(&processes)
	if err != nil {
		// clobber yaml parsing  error
		return processes, errors.New("invalid YAML")
	}

	return processes, nil
}

func (runner *Runner) compile(buildpackDir, cacheDir string) error {
	return runner.run(exec.Command(path.Join(buildpackDir, "bin", "compile"), runner.config.BuildDir(), cacheDir), os.Stdout)
}

func (runner *Runner) release(buildpackDir string, startCommands map[string]string) (Release, error) {
	output := new(bytes.Buffer)

	err := runner.run(exec.Command(path.Join(buildpackDir, "bin", "release"), runner.config.BuildDir()), output)
	if err != nil {
		return Release{}, err
	}

	decoder := candiedyaml.NewDecoder(output)

	parsedRelease := Release{}

	err = decoder.Decode(&parsedRelease)
	if err != nil {
		return Release{}, newDescriptiveError(err, "buildpack's release output invalid")
	}

	if len(startCommands) > 0 {
		parsedRelease.DefaultProcessTypes = startCommands
	}

	return parsedRelease, nil
}

func (runner *Runner) saveInfo(infoFilePath, buildpack, detectOutput string, releaseInfo Release) error {
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

func (runner *Runner) copyApp(buildDir, stageDir string) error {
	return runner.run(exec.Command("cp", "-a", buildDir, stageDir), os.Stdout)
}

func (runner *Runner) run(cmd *exec.Cmd, output io.Writer) error {
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func printError(message string) {
	println(message)
}
