package models

import (
	"encoding/json"
	"errors"
	"time"
)

var InvalidActionConversion = errors.New("Invalid Action Conversion")

type DownloadAction struct {
	Name    string `json:"name"`
	From    string `json:"from"`
	To      string `json:"to"`
	Extract bool   `json:"extract"`
}

type UploadAction struct {
	Name     string `json:"name"`
	To       string `json:"to"`
	From     string `json:"from"`
	Compress bool   `json:"compress"`
}

type RunAction struct {
	Name    string        `json:"name"`
	Script  string        `json:"script"`
	Env     [][]string    `json:"env"`
	Timeout time.Duration `json:"timeout"`
}

type TryAction struct {
	Action ExecutorAction `json:"action"`
}

type FetchResultAction struct {
	Name string `json:"name"`
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
	case TryAction:
		envelope.Name = "try"
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
	case "try":
		tryAction := TryAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &tryAction)
		a.Action = tryAction
	default:
		err = InvalidActionConversion
	}

	return err
}
