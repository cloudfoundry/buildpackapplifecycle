package models

import "encoding/json"

type DesireAppRequestFromCC struct {
	ProcessGuid     string                `json:"process_guid"`
	DropletUri      string                `json:"droplet_uri"`
	Stack           string                `json:"stack"`
	StartCommand    string                `json:"start_command"`
	Environment     []EnvironmentVariable `json:"environment"`
	MemoryMB        int                   `json:"memory_mb"`
	DiskMB          int                   `json:"disk_mb"`
	FileDescriptors uint64                `json:"file_descriptors"`
	NumInstances    int                   `json:"num_instances"`
	Routes          []string              `json:"routes"`
	LogGuid         string                `json:"log_guid"`
}

func (d DesireAppRequestFromCC) ToJSON() []byte {
	encoded, _ := json.Marshal(d)
	return encoded
}

type CCDesiredStateServerResponse struct {
	Apps        []CCBulkDesiredApp `json:"apps"`
	CCBulkToken *json.RawMessage   `json:"token"`
}

type CCBulkDesiredApp struct {
	DiskMB          uint64                `json:"disk_mb"`
	Environment     []EnvironmentVariable `json:"environment"`
	FileDescriptors uint64                `json:"file_descriptors"`
	Instances       uint                  `json:"instances"`
	LogGuid         string                `json:"log_guid"`
	MemoryMB        uint64                `json:"memory_mb"`
	ProcessGuid     string                `json:"process_guid"`
	Routes          []string              `json:"routes"`
	SourceURL       string                `json:"source_url"`
	Stack           string                `json:"stack"`
	StartCommand    string                `json:"start_command"`
}

type CCBulkToken struct {
	Id int `json:"id"`
}
