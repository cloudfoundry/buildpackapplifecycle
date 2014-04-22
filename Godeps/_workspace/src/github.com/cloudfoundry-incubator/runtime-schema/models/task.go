package models

import (
	"encoding/json"
)

type TaskState int

const (
	TaskStateInvalid TaskState = iota
	TaskStatePending
	TaskStateClaimed
	TaskStateRunning
	TaskStateCompleted
	TaskStateResolving
)

type Task struct {
	Guid            string           `json:"guid"`
	Actions         []ExecutorAction `json:"actions"`
	Stack           string           `json:"stack"`
	FileDescriptors int              `json:"file_descriptors"`
	MemoryMB        int              `json:"memory_mb"`
	DiskMB          int              `json:"disk_mb"`
	Log             LogConfig        `json:"log"`
	CreatedAt       int64            `json:"created_at"` //  the number of nanoseconds elapsed since January 1, 1970 UTC
	UpdatedAt       int64            `json:"updated_at"`

	State TaskState `json:"state"`

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

func NewTaskFromJSON(payload []byte) (Task, error) {
	var task Task

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func (self Task) ToJSON() []byte {
	bytes, err := json.Marshal(self)
	if err != nil {
		panic(err)
	}

	return bytes
}
