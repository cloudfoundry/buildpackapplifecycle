package models

import "encoding/json"

type PortMapping struct {
	ContainerPort uint32 `json:"container_port"`
	HostPort      uint32 `json:"host_port,omitempty"`
}

type DesiredLRPChange struct {
	Before *DesiredLRP
	After  *DesiredLRP
}

type ActualLRPChange struct {
	Before *LRP
	After  *LRP
}

///

type LRPStartAuctionState int

const (
	LRPStartAuctionStateInvalid LRPStartAuctionState = iota
	LRPStartAuctionStatePending
	LRPStartAuctionStateClaimed
)

type LRPStartAuction struct {
	ProcessGuid  string `json:"process_guid"`
	InstanceGuid string `json:"instance_guid"`

	DiskMB   int `json:"disk_mb"`
	MemoryMB int `json:"memory_mb"`

	Stack   string           `json:"stack"`
	Actions []ExecutorAction `json:"actions"`
	Log     LogConfig        `json:"log"`
	Ports   []PortMapping    `json:"ports"`

	Index int `json:"index"`

	State LRPStartAuctionState `json:"state"`
}

func NewLRPStartAuctionFromJSON(payload []byte) (LRPStartAuction, error) {
	var task LRPStartAuction

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return LRPStartAuction{}, err
	}

	return task, nil
}

func (self LRPStartAuction) ToJSON() []byte {
	bytes, err := json.Marshal(self)
	if err != nil {
		panic(err)
	}

	return bytes
}

///

type LRPState int

const (
	LRPStateInvalid LRPState = iota
	LRPStateStarting
	LRPStateRunning
)

type LRP struct {
	ProcessGuid  string `json:"process_guid"`
	InstanceGuid string `json:"instance_guid"`

	Index int `json:"index"`

	Host  string        `json:"host"`
	Ports []PortMapping `json:"ports"`

	State LRPState `json:"state"`
}

func NewLRPFromJSON(payload []byte) (LRP, error) {
	var task LRP

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return LRP{}, err
	}

	return task, nil
}

func (self LRP) ToJSON() []byte {
	bytes, err := json.Marshal(self)
	if err != nil {
		panic(err)
	}

	return bytes
}

///

type DesiredLRP struct {
	ProcessGuid string   `json:"process_guid"`
	Instances   int      `json:"instances"`
	Stack       string   `json:"stack"`
	MemoryMB    int      `json:"memory_mb"`
	DiskMB      int      `json:"disk_mb"`
	Routes      []string `json:"routes"`
}

func NewDesiredLRPFromJSON(payload []byte) (DesiredLRP, error) {
	var task DesiredLRP

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return DesiredLRP{}, err
	}

	return task, nil
}

func (self DesiredLRP) ToJSON() []byte {
	bytes, err := json.Marshal(self)
	if err != nil {
		panic(err)
	}

	return bytes
}
