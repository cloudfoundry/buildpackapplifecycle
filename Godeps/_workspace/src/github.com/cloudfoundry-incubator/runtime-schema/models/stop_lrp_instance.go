package models

import "encoding/json"

type StopLRPInstance struct {
	InstanceGuid string `json:"instance_guid"`
}

func NewStopLRPInstanceFromJSON(payload []byte) (StopLRPInstance, error) {
	var stopInstance StopLRPInstance

	err := json.Unmarshal(payload, &stopInstance)
	if err != nil {
		return StopLRPInstance{}, err
	}

	return stopInstance, nil
}

func (self StopLRPInstance) ToJSON() []byte {
	bytes, err := json.Marshal(self)
	if err != nil {
		panic(err)
	}

	return bytes
}
