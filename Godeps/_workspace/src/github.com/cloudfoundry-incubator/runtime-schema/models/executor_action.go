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
	To       string `json:"to"`
	From     string `json:"from"`
	Compress bool   `json:"compress"`
}

type RunAction struct {
	Script  string                `json:"script"`
	Env     []EnvironmentVariable `json:"env"`
	Timeout time.Duration         `json:"timeout"`
}

type FetchResultAction struct {
	File string `json:"file"`
}

type TryAction struct {
	Action ExecutorAction `json:"action"`
}

type EmitProgressAction struct {
	Action         ExecutorAction `json:"action"`
	StartMessage   string         `json:"start_message"`
	SuccessMessage string         `json:"success_message"`
	FailureMessage string         `json:"failure_message"`
}

func EmitProgressFor(action ExecutorAction, startMessage string, successMessage string, failureMessage string) ExecutorAction {
	return ExecutorAction{
		EmitProgressAction{
			Action:         action,
			StartMessage:   startMessage,
			SuccessMessage: successMessage,
			FailureMessage: failureMessage,
		},
	}
}

func Try(action ExecutorAction) ExecutorAction {
	return ExecutorAction{
		TryAction{
			Action: action,
		},
	}
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
	case EmitProgressAction:
		envelope.Name = "emit_progress"
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
	case "emit_progress":
		emitProgressAction := EmitProgressAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &emitProgressAction)
		a.Action = emitProgressAction
	case "try":
		tryAction := TryAction{}
		err = json.Unmarshal(*envelope.ActionPayload, &tryAction)
		a.Action = tryAction
	default:
		err = InvalidActionConversion
	}

	return err
}
