package models

import "encoding/json"

type TransitionalLRPState int

const (
	TransitionalLRPStateInvalid TransitionalLRPState = iota
	TransitionalLRPStateDesired
	TransitionalLRPStateRunning
)

type TransitionalLongRunningProcess struct {
	Guid     string               `json:"guid"`
	Stack    string               `json:"stack"`
	Actions  []ExecutorAction     `json:"actions"`
	Log      LogConfig            `json:"log"`
	State    TransitionalLRPState `json:"state"`
	MemoryMB int                  `json:"memory_mb"`
	DiskMB   int                  `json:"disk_mb"`
}

func NewTransitionalLongRunningProcessFromJSON(payload []byte) (TransitionalLongRunningProcess, error) {
	var task TransitionalLongRunningProcess

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return TransitionalLongRunningProcess{}, err
	}

	return task, nil
}

func (self TransitionalLongRunningProcess) ToJSON() []byte {
	bytes, err := json.Marshal(self)
	if err != nil {
		panic(err)
	}

	return bytes
}
