package models

import "encoding/json"

type DesiredLRP struct {
	ProcessGuid     string                `json:"process_guid"`
	Instances       int                   `json:"instances"`
	Stack           string                `json:"stack"`
	MemoryMB        int                   `json:"memory_mb"`
	DiskMB          int                   `json:"disk_mb"`
	FileDescriptors uint64                `json:"file_descriptors"`
	Source          string                `json:"source"`
	StartCommand    string                `json:"start_command"`
	Environment     []EnvironmentVariable `json:"environment"`
	Routes          []string              `json:"routes"`
	LogGuid         string                `json:"log_guid"`
}

type DesiredLRPChange struct {
	Before *DesiredLRP
	After  *DesiredLRP
}

func NewDesiredLRPFromJSON(payload []byte) (DesiredLRP, error) {
	var task DesiredLRP

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return DesiredLRP{}, err
	}

	return task, nil
}

func (desired DesiredLRP) ToJSON() []byte {
	bytes, err := json.Marshal(desired)
	if err != nil {
		panic(err)
	}

	return bytes
}
