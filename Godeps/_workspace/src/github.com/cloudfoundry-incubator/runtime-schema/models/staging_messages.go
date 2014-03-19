package models

type StagingRequestFromCC struct {
	AppId           string           `json:"app_id"`
	TaskId          string           `json:"task_id"`
	Stack           string           `json:"stack"`
	DownloadUri     string           `json:"download_uri"`
	FileDescriptors int              `json:"file_descriptors"`
	MemoryMB        int              `json:"memory_mb"`
	DiskMB          int              `json:"disk_mb"`
	AdminBuildpacks []AdminBuildpack `json:"admin_buildpacks"`
	Environment     [][]string       `json:"environment"`
}

type AdminBuildpack struct {
	Key string `json:"key"`
	Url string `json:"url"`
}

type StagingInfo struct {
	DetectedBuildpack string `yaml:"detected_buildpack" json:"detected_buildpack"`
	StartCommand      string `yaml:"start_command" json:"-"`
}

type StagingResponseForCC struct {
	DetectedBuildpack string `json:"detected_buildpack,omitempty"`
	Error             string `json:"error,omitempty"`
}
