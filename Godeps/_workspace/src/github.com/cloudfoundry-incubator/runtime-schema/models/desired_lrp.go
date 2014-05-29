package models

import "encoding/json"

type DesiredLRP struct {
	ProcessGuid string   `json:"process_guid"`
	Instances   int      `json:"instances"`
	Stack       string   `json:"stack"`
	MemoryMB    int      `json:"memory_mb"`
	DiskMB      int      `json:"disk_mb"`
	Routes      []string `json:"routes"`
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

func (self DesiredLRP) ToJSON() []byte {
	bytes, err := json.Marshal(self)
	if err != nil {
		panic(err)
	}

	return bytes
}
