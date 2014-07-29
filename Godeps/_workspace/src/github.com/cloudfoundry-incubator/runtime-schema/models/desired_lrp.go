package models

import "encoding/json"

type DesiredLRP struct {
	ProcessGuid string `json:"process_guid"`

	Instances int              `json:"instances"`
	Stack     string           `json:"stack"`
	Actions   []ExecutorAction `json:"actions"`
	DiskMB    int              `json:"disk_mb"`
	MemoryMB  int              `json:"memory_mb"`
	Ports     []PortMapping    `json:"ports"`
	Routes    []string         `json:"routes"`
	Log       LogConfig        `json:"log"`
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

	if task.ProcessGuid == "" {
		return DesiredLRP{}, ErrInvalidJSONMessage{"process_guid"}
	}

	if task.Stack == "" {
		return DesiredLRP{}, ErrInvalidJSONMessage{"stack"}
	}

	if len(task.Actions) == 0 {
		return DesiredLRP{}, ErrInvalidJSONMessage{"actions"}
	}

	return task, nil
}

func (desired DesiredLRP) ToJSON() []byte {
	bytes, err := json.Marshal(desired)
	if err != nil {
		panic(err)
	}

	return bytes
}
