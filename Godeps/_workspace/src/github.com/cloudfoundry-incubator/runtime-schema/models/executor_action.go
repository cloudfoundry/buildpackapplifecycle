package models

import (
	"encoding/json"
	"errors"
	"time"
)

var InvalidActionConversion = errors.New("Invalid Action Conversion")

type DownloadAction struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Extract bool   `json:"extract"`
}

type UploadAction struct {
	To   string `json:"to"`
	From string `json:"from"`
}

type RunAction struct {
	Script  string        `json:"script"`
	Env     [][]string    `json:"env"`
	Timeout time.Duration `json:"timeout"`
}

type FetchResultAction struct {
	File string `json:"file"`
}

type executorActionEnvelope struct {
	Name          string           `json:"action"`
	ActionPayload *json.RawMessage `json:"args"`
}

type ExecutorAction struct {
	Action interface{} `json:"-"`
}

func (a ExecutorAction) MarshalJSON() ([]byte, error) {
	var envelope executorActionEnvelope

	payload, err := json.Marshal(a.Action)

	if err != nil {
		return nil, err
	}

	switch a.Action.(type) {
	case DownloadAction:
		envelope.Name = "download"
	case RunAction:
		envelope.Name = "run"
	case UploadAction:
		envelope.Name = "upload"
	case FetchResultAction:
		envelope.Name = "fetch_result"
	default:
		return nil, InvalidActionConversion
	}

	envelope.ActionPayload = (*json.RawMessage)(&payload)

	return json.Marshal(envelope)
}

func (a *ExecutorAction) UnmarshalJSON(bytes []byte) error {
	var envelope executorActionEnvelope

	err := json.Unmarshal(bytes, &envelope)
	if err != nil {
		return err
	}

	switch envelope.Name {
	case "download":
		downloadAction := DownloadAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &downloadAction)
		a.Action = downloadAction
	case "run":
		runAction := RunAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &runAction)
		a.Action = runAction
	case "upload":
		uploadAction := UploadAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &uploadAction)
		a.Action = uploadAction
	case "fetch_result":
		fetchResultAction := FetchResultAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &fetchResultAction)
		a.Action = fetchResultAction
	default:
		err = InvalidActionConversion
	}

	return err
}
