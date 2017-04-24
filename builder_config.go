package buildpackapplifecycle

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"math"
	"path"
	"strings"
)

type LifecycleBuilderConfig struct {
	*flag.FlagSet

	ExecutablePath string
}

const (
	lifecycleBuilderBuildDirFlag                  = "buildDir"
	lifecycleBuilderOutputDropletFlag             = "outputDroplet"
	lifecycleBuilderOutputMetadataFlag            = "outputMetadata"
	lifecycleBuilderOutputBuildArtifactsCacheFlag = "outputBuildArtifactsCache"
	lifecycleBuilderBuildpacksDirFlag             = "buildpacksDir"
	lifecycleBuilderBuildArtifactsCacheDirFlag    = "buildArtifactsCacheDir"
	lifecycleBuilderBuildpackOrderFlag            = "buildpackOrder"
	lifecycleBuilderSkipDetect                    = "skipDetect"
	lifecycleBuilderSkipCertVerify                = "skipCertVerify"
)

var lifecycleBuilderDefaults = map[string]string{
	lifecycleBuilderBuildDirFlag:                  "/tmp/app",
	lifecycleBuilderOutputDropletFlag:             "/tmp/droplet",
	lifecycleBuilderOutputMetadataFlag:            "/tmp/result.json",
	lifecycleBuilderOutputBuildArtifactsCacheFlag: "/tmp/output-cache",
	lifecycleBuilderBuildpacksDirFlag:             "/tmp/buildpacks",
	lifecycleBuilderBuildArtifactsCacheDirFlag:    "/tmp/cache",
}

func NewLifecycleBuilderConfig(buildpacks []string, skipDetect bool, skipCertVerify bool) LifecycleBuilderConfig {
	flagSet := flag.NewFlagSet("builder", flag.ExitOnError)

	flagSet.String(
		lifecycleBuilderBuildDirFlag,
		lifecycleBuilderDefaults[lifecycleBuilderBuildDirFlag],
		"directory containing raw app bits",
	)

	flagSet.String(
		lifecycleBuilderOutputDropletFlag,
		lifecycleBuilderDefaults[lifecycleBuilderOutputDropletFlag],
		"file where compressed droplet should be written",
	)

	flagSet.String(
		lifecycleBuilderOutputMetadataFlag,
		lifecycleBuilderDefaults[lifecycleBuilderOutputMetadataFlag],
		"directory in which to write the app metadata",
	)

	flagSet.String(
		lifecycleBuilderOutputBuildArtifactsCacheFlag,
		lifecycleBuilderDefaults[lifecycleBuilderOutputBuildArtifactsCacheFlag],
		"file where compressed contents of new cached build artifacts should be written",
	)

	flagSet.String(
		lifecycleBuilderBuildpacksDirFlag,
		lifecycleBuilderDefaults[lifecycleBuilderBuildpacksDirFlag],
		"directory containing the buildpacks to try",
	)

	flagSet.String(
		lifecycleBuilderBuildArtifactsCacheDirFlag,
		lifecycleBuilderDefaults[lifecycleBuilderBuildArtifactsCacheDirFlag],
		"directory where previous cached build artifacts should be extracted",
	)

	flagSet.String(
		lifecycleBuilderBuildpackOrderFlag,
		strings.Join(buildpacks, ","),
		"comma-separated list of buildpacks, to be tried in order",
	)

	flagSet.Bool(
		lifecycleBuilderSkipDetect,
		skipDetect,
		"skip buildpack detect",
	)

	flagSet.Bool(
		lifecycleBuilderSkipCertVerify,
		skipCertVerify,
		"skip SSL certificate verification",
	)

	return LifecycleBuilderConfig{
		FlagSet: flagSet,

		ExecutablePath: "/tmp/lifecycle/builder",
	}
}

func (s LifecycleBuilderConfig) Path() string {
	return s.ExecutablePath
}

func (s LifecycleBuilderConfig) Args() []string {
	argv := []string{}

	s.FlagSet.VisitAll(func(flag *flag.Flag) {
		argv = append(argv, fmt.Sprintf("-%s=%s", flag.Name, flag.Value.String()))
	})

	return argv
}

func (s LifecycleBuilderConfig) Validate() error {
	var validationError ValidationError

	s.FlagSet.VisitAll(func(flag *flag.Flag) {
		value := flag.Value.String()
		if value == "" {
			validationError = validationError.Append(fmt.Errorf("missing flag: -%s", flag.Name))
		}
	})

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (s LifecycleBuilderConfig) BuildDir() string {
	return s.Lookup(lifecycleBuilderBuildDirFlag).Value.String()
}

func (s LifecycleBuilderConfig) BuildpackPath(buildpackName string) string {
	return path.Join(s.BuildpacksDir(), fmt.Sprintf("%x", md5.Sum([]byte(buildpackName))))
}

func (s LifecycleBuilderConfig) BuildpackOrder() []string {
	buildpackOrder := s.Lookup(lifecycleBuilderBuildpackOrderFlag).Value.String()
	return strings.Split(buildpackOrder, ",")
}

func (s LifecycleBuilderConfig) NumBuildpacks() int {
	return len(s.BuildpackOrder())
}

func (s LifecycleBuilderConfig) SupplyBuildpacks() []string {
	return s.BuildpackOrder()[0 : s.NumBuildpacks()-1]
}

func (s LifecycleBuilderConfig) FinalBuildpack() string {
	return s.BuildpackOrder()[s.NumBuildpacks()-1]
}

func (s LifecycleBuilderConfig) DepsIndices() []string {
	var indices []string
	padDigits := 1

	if s.NumBuildpacks() > 0 {
		padDigits = int(math.Log10(float64(s.NumBuildpacks()))) + 1
	}

	indexFormat := fmt.Sprintf("%%0%dd", padDigits)
	for i := 0; i < s.NumBuildpacks(); i++ {
		indices = append(indices, fmt.Sprintf(indexFormat, i))
	}

	return indices
}

func (s LifecycleBuilderConfig) FinalDepsIndex() string {
	return s.DepsIndices()[s.NumBuildpacks()-1]
}

func (s LifecycleBuilderConfig) BuildpacksDir() string {
	return s.Lookup(lifecycleBuilderBuildpacksDirFlag).Value.String()
}

func (s LifecycleBuilderConfig) BuildArtifactsCacheDir() string {
	return s.Lookup(lifecycleBuilderBuildArtifactsCacheDirFlag).Value.String()
}

func (s LifecycleBuilderConfig) OutputDroplet() string {
	return s.Lookup(lifecycleBuilderOutputDropletFlag).Value.String()
}

func (s LifecycleBuilderConfig) OutputMetadata() string {
	return s.Lookup(lifecycleBuilderOutputMetadataFlag).Value.String()
}

func (s LifecycleBuilderConfig) OutputBuildArtifactsCache() string {
	return s.Lookup(lifecycleBuilderOutputBuildArtifactsCacheFlag).Value.String()
}

func (s LifecycleBuilderConfig) SkipCertVerify() bool {
	return s.Lookup(lifecycleBuilderSkipCertVerify).Value.String() == "true"
}

func (s LifecycleBuilderConfig) SkipDetect() bool {
	return s.Lookup(lifecycleBuilderSkipDetect).Value.String() == "true"
}

func (s LifecycleBuilderConfig) IsMultiBuildpack() bool {
	return s.SkipDetect() && len(s.BuildpackOrder()) > 1
}

type ValidationError []error

func (ve ValidationError) Append(err error) ValidationError {
	switch err := err.(type) {
	case ValidationError:
		return append(ve, err...)
	default:
		return append(ve, err)
	}
}

func (ve ValidationError) Error() string {
	var buffer bytes.Buffer

	for i, err := range ve {
		if err == nil {
			continue
		}
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(err.Error())
	}

	return buffer.String()
}

func (ve ValidationError) Empty() bool {
	return len(ve) == 0
}
