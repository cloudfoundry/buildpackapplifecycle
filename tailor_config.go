package linux_circus

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"path"
	"strings"
)

type CircusTailorConfig struct {
	*flag.FlagSet

	ExecutablePath string
}

const (
	circusTailorBuildDirFlag                  = "buildDir"
	circusTailorOutputDropletFlag             = "outputDroplet"
	circusTailorOutputMetadataFlag            = "outputMetadata"
	circusTailorOutputBuildArtifactsCacheFlag = "outputBuildArtifactsCache"
	circusTailorBuildpacksDirFlag             = "buildpacksDir"
	circusTailorBuildArtifactsCacheDirFlag    = "buildArtifactsCacheDir"
	circusTailorBuildpackOrderFlag            = "buildpackOrder"
	circusTailorSkipCertVerify                = "skipCertVerify"
)

var circusTailorDefaults = map[string]string{
	circusTailorBuildDirFlag:                  "/tmp/app",
	circusTailorOutputDropletFlag:             "/tmp/droplet",
	circusTailorOutputMetadataFlag:            "/tmp/result.json",
	circusTailorOutputBuildArtifactsCacheFlag: "/tmp/output-cache",
	circusTailorBuildpacksDirFlag:             "/tmp/buildpacks",
	circusTailorBuildArtifactsCacheDirFlag:    "/tmp/cache",
}

func NewCircusTailorConfig(buildpacks []string, skipCertVerify bool) CircusTailorConfig {
	flagSet := flag.NewFlagSet("tailor", flag.ExitOnError)

	flagSet.String(
		circusTailorBuildDirFlag,
		circusTailorDefaults[circusTailorBuildDirFlag],
		"directory containing raw app bits",
	)

	flagSet.String(
		circusTailorOutputDropletFlag,
		circusTailorDefaults[circusTailorOutputDropletFlag],
		"file where compressed droplet should be written",
	)

	flagSet.String(
		circusTailorOutputMetadataFlag,
		circusTailorDefaults[circusTailorOutputMetadataFlag],
		"directory in which to write the app metadata",
	)

	flagSet.String(
		circusTailorOutputBuildArtifactsCacheFlag,
		circusTailorDefaults[circusTailorOutputBuildArtifactsCacheFlag],
		"file where compressed contents of new cached build artifacts should be written",
	)

	flagSet.String(
		circusTailorBuildpacksDirFlag,
		circusTailorDefaults[circusTailorBuildpacksDirFlag],
		"directory containing the buildpacks to try",
	)

	flagSet.String(
		circusTailorBuildArtifactsCacheDirFlag,
		circusTailorDefaults[circusTailorBuildArtifactsCacheDirFlag],
		"directory where previous cached build artifacts should be extracted",
	)

	flagSet.String(
		circusTailorBuildpackOrderFlag,
		strings.Join(buildpacks, ","),
		"comma-separated list of buildpacks, to be tried in order",
	)

	flagSet.Bool(
		circusTailorSkipCertVerify,
		skipCertVerify,
		"skip SSL certificate verification",
	)

	return CircusTailorConfig{
		FlagSet: flagSet,

		ExecutablePath: "/tmp/circus/tailor",
	}
}

func (s CircusTailorConfig) Path() string {
	return s.ExecutablePath
}

func (s CircusTailorConfig) Args() []string {
	argv := []string{}

	s.FlagSet.VisitAll(func(flag *flag.Flag) {
		argv = append(argv, fmt.Sprintf("-%s=%s", flag.Name, flag.Value.String()))
	})

	return argv
}

func (s CircusTailorConfig) Validate() error {
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

func (s CircusTailorConfig) BuildDir() string {
	return s.Lookup(circusTailorBuildDirFlag).Value.String()
}

func (s CircusTailorConfig) BuildpackPath(buildpackName string) string {
	return path.Join(s.BuildpacksDir(), fmt.Sprintf("%x", md5.Sum([]byte(buildpackName))))
}

func (s CircusTailorConfig) BuildpackOrder() []string {
	buildpackOrder := s.Lookup(circusTailorBuildpackOrderFlag).Value.String()
	return strings.Split(buildpackOrder, ",")
}

func (s CircusTailorConfig) BuildpacksDir() string {
	return s.Lookup(circusTailorBuildpacksDirFlag).Value.String()
}

func (s CircusTailorConfig) BuildArtifactsCacheDir() string {
	return s.Lookup(circusTailorBuildArtifactsCacheDirFlag).Value.String()
}

func (s CircusTailorConfig) OutputDroplet() string {
	return s.Lookup(circusTailorOutputDropletFlag).Value.String()
}

func (s CircusTailorConfig) OutputMetadata() string {
	return s.Lookup(circusTailorOutputMetadataFlag).Value.String()
}

func (s CircusTailorConfig) OutputBuildArtifactsCache() string {
	return s.Lookup(circusTailorOutputBuildArtifactsCacheFlag).Value.String()
}

func (s CircusTailorConfig) SkipCertVerify() bool {
	return s.Lookup(circusTailorSkipCertVerify).Value.String() == "true"
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
