package models

import "encoding/json"

type StopLRPInstance struct {
	ProcessGuid  string `json:"process_guid"`
	InstanceGuid string `json:"instance_guid"`
	Index        int    `json:"index"`
}

func NewStopLRPInstanceFromJSON(payload []byte) (StopLRPInstance, error) {
	var stopInstance StopLRPInstance

	err := json.Unmarshal(payload, &stopInstance)
	if err != nil {
		return StopLRPInstance{}, err
	}

	return stopInstance, nil
}

func (stop StopLRPInstance) ToJSON() []byte {
	bytes, err := json.Marshal(stop)
	if err != nil {
		panic(err)
	}

	return bytes
}
