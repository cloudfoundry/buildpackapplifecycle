package models

import "encoding/json"

type DesireAppRequestFromCC struct {
	AppId           string                `json:"app_id"`
	AppVersion      string                `json:"app_version"`
	DropletUri      string                `json:"droplet_uri"`
	Stack           string                `json:"stack"`
	StartCommand    string                `json:"start_command"`
	Environment     []EnvironmentVariable `json:"environment"`
	MemoryMB        int                   `json:"memory_mb"`
	DiskMB          int                   `json:"disk_mb"`
	FileDescriptors uint64                `json:"file_descriptors"`
	NumInstances    int                   `json:"num_instances"`
	Routes          []string              `json:"routes"`
}

func (d DesireAppRequestFromCC) ToJSON() []byte {
	encoded, _ := json.Marshal(d)
	return encoded
}
