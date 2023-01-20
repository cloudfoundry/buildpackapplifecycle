package buildpackrunner

const DeaStagingInfoFilename = "staging_info.yml"

// Used to generate YAML file read by the DEA
type DeaStagingInfo struct {
	DetectedBuildpack string                `json:"detected_buildpack" yaml:"detected_buildpack"`
	StartCommand      string                `json:"start_command" yaml:"start_command"`
	Config            *DeaStagingInfoConfig `json:"config,omitempty" yaml:"config,omitempty"`
}

type DeaStagingInfoConfig struct {
	EntrypointPrefix string `json:"entrypoint_prefix,omitempty" yaml:"entrypoint_prefix,omitempty"`
}

func (stagingInfo DeaStagingInfo) GetEntrypointPrefix() string {
	if stagingInfo.Config != nil {
		return stagingInfo.Config.EntrypointPrefix
	}

	return ""
}
