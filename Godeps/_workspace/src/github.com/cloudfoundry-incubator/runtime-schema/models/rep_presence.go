package models

import "encoding/json"

type RepPresence struct {
	RepID string `json:"rep_id"`
	Stack string `json:"stack"`
}

func NewRepPresenceFromJSON(payload []byte) (RepPresence, error) {
	var task RepPresence

	err := json.Unmarshal(payload, &task)
	if err != nil {
		return RepPresence{}, err
	}

	return task, nil
}

func (presence RepPresence) ToJSON() []byte {
	bytes, err := json.Marshal(presence)
	if err != nil {
		panic(err)
	}

	return bytes
}
