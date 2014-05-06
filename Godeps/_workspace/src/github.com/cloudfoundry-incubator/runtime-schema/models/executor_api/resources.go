package executor_api

import "github.com/cloudfoundry-incubator/runtime-schema/models"

const (
	StateReserved = "reserved"
	StateCreated  = "created"
)

type Container struct {
	Guid string `json:"guid"`

	ExecutorGuid    string  `json:"executor_guid"`
	MemoryMB        int     `json:"memory_mb"`
	DiskMB          int     `json:"disk_mb"`
	CpuPercent      float64 `json:"cpu_percent"`
	FileDescriptors int     `json:"file_descriptors"`
	State           string  `json:"state"`
	ContainerHandle string  `json:"container_handle"`
}

type ContainerAllocationRequest struct {
	MemoryMB        int     `json:"memory_mb"`
	DiskMB          int     `json:"disk_mb"`
	CpuPercent      float64 `json:"cpu_percent"`
	FileDescriptors int     `json:"file_descriptors"`
}

type ContainerRunRequest struct {
	Actions     []models.ExecutorAction `json:"actions"`
	Metadata    []byte                  `json:"metadata"`
	CompleteURL string                  `json:"complete_url"`
}

type ContainerRunResult struct {
	Failed        bool   `json:"failed"`
	FailureReason string `json:"failure_reason"`
	Result        string `json:"result"`
	Metadata      []byte `json:"metadata"`
}
