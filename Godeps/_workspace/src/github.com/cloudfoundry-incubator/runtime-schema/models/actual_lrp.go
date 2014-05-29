package models

import "encoding/json"

type ActualLRPState int

const (
	ActualLRPStateInvalid ActualLRPState = iota
	ActualLRPStateStarting
	ActualLRPStateRunning
)

type ActualLRPChange struct {
	Before *ActualLRP
	After  *ActualLRP
}

type ActualLRP struct {
	ProcessGuid  string `json:"process_guid"`
	InstanceGuid string `json:"instance_guid"`

	Index int `json:"index"`

	Host  string        `json:"host"`
	Ports []PortMapping `json:"ports"`

	State ActualLRPState `json:"state"`
}

func NewActualLRPFromJSON(payload []byte) (ActualLRP, error) {
	var task ActualLRP

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return ActualLRP{}, err
	}

	return task, nil
}

func (self ActualLRP) ToJSON() []byte {
	bytes, err := json.Marshal(self)
	if err != nil {
		panic(err)
	}

	return bytes
}
