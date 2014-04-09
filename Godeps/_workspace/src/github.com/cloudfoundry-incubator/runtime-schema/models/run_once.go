package models

import (
	"encoding/json"
)

type RunOnceState int

const (
	RunOnceStateInvalid RunOnceState = iota
	RunOnceStatePending
	RunOnceStateClaimed
	RunOnceStateRunning
	RunOnceStateCompleted
	RunOnceStateResolving
)

type RunOnce struct {
	Guid            string           `json:"guid"`
	Actions         []ExecutorAction `json:"actions"`
	Stack           string           `json:"stack"`
	FileDescriptors int              `json:"file_descriptors"`
	MemoryMB        int              `json:"memory_mb"`
	DiskMB          int              `json:"disk_mb"`
	Log             LogConfig        `json:"log"`
	CreatedAt       int64            `json:"created_at"` //  the number of nanoseconds elapsed since January 1, 1970 UTC
	UpdatedAt       int64            `json:"updated_at"`

	State RunOnceState `json:"state"`

	// this is so that any stager can process a complete event,
	// because the CC <-> Stager interaction is a one-to-one request-response
	//
	// ideally staging completion is a "broadcast" event instead and this goes away
	ReplyTo string `json:"reply_to"`

	ExecutorID string `json:"executor_id"`

	ContainerHandle string `json:"container_handle"`

	Result        string `json:"result"`
	Failed        bool   `json:"failed"`
	FailureReason string `json:"failure_reason"`
}

type LogConfig struct {
	Guid       string `json:"guid"`
	SourceName string `json:"source_name"`
	Index      *int   `json:"index"`
}

func NewRunOnceFromJSON(payload []byte) (RunOnce, error) {
	var runOnce RunOnce

	err := json.Unmarshal(payload, &runOnce)
	if err != nil {
		return RunOnce{}, err
	}

	return runOnce, nil
}

func (self RunOnce) ToJSON() []byte {
	bytes, err := json.Marshal(self)
	if err != nil {
		panic(err)
	}

	return bytes
}
