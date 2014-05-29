package models

import "encoding/json"

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
