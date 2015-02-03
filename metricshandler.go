package main

import (
	"errors"
	"time"

	"github.com/chuckpreslar/emission"
	"github.com/inconshreveable/go-keen"
)

// A MetricsEventHandler reporting to keen.io.
type MetricsEventHandler struct {
	keen  *keen.Client
	start time.Time
}

// MetricsApplicationPayload is the app data we're reporting
// to keen. Part of MetricsPayload.
type MetricsApplicationPayload struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OwnerName string `json:"ownerName"`
}

// MetricsPayload is the data we're sending to keen.
// Identical to legacy pool but we've renamed
// `sentinel` -> `core`.
type MetricsPayload struct {
	MetricsApplicationPayload *MetricsApplicationPayload `json:"application"`
	ApplicationStartedByName  string                     `json:"startedBy"`
	BuildID                   string                     `json:"buildID"`
	DeployID                  string                     `json:"deployID"`
	Event                     string                     `json:"event"`
	Step                      *Step                      `json:"step"`
	GitVersion                string                     `json:"GitVersion"`
	StepOrder                 int                        `json:"stepOrder"`
	Successful                bool                       `json:"succesful,omitempty"`
	TimeElapsed               float64                    `json:"timeElapsed,omitempty"`
	Timestamp                 int32                      `json:"timestamp"`
	VCS                       string                     `json:"versionControl"`
	// Box                        string `json:"box"`       // todo
	// Core                       string `json:"core"`      // todo
	// JobID                      string `json:"jobId"`     // todo
}

// NewMetricsHandler will create a new NewMetricsHandler.
func NewMetricsHandler(opts *PipelineOptions) (*MetricsEventHandler, error) {

	if "" == opts.KeenProjectWriteKey {
		return nil, errors.New("No KeenProjectWriteKey specified")
	}

	if "" == opts.KeenProjectID {
		return nil, errors.New("No KeenProjectID specified")
	}

	// todo(yoshuawuyts): replace with `keen batch client` + regular flushes
	keenInstance := &keen.Client{
		WriteKey:  opts.KeenProjectWriteKey,
		ProjectID: opts.KeenProjectID,
	}

	return &MetricsEventHandler{keen: keenInstance}, nil
}

// BuildStepStarted responds to the BuildStepStarted event.
func (h *MetricsEventHandler) BuildStepStarted(event *BuildStepStartedArgs) {
	h.start = time.Now()

	h.keen.AddEvent("build-events-ewok", &MetricsPayload{
		MetricsApplicationPayload: &MetricsApplicationPayload{
			ID:        event.Options.ApplicationID,
			Name:      event.Options.ApplicationName,
			OwnerName: event.Options.ApplicationOwnerName,
		},
		ApplicationStartedByName: event.Options.ApplicationStartedByName,
		BuildID:                  event.Options.BuildID,
		DeployID:                 event.Options.DeployID,
		Event:                    "buildStepStarted",
		Step:                     event.Step,
		GitVersion:               GitCommit,
		StepOrder:                event.Order,
		Timestamp:                int32(time.Now().Unix()),
		VCS:                      "git",
		// Box:     event.Box,
		// Core:      "",
		// JobID:     "",
	})
}

// BuildStepFinished responds to the BuildStepFinished event.
func (h *MetricsEventHandler) BuildStepFinished(event *BuildStepFinishedArgs) {

	elapsed := time.Since(h.start)
	h.start = time.Now()

	h.keen.AddEvent("build-events-ewok", &MetricsPayload{
		MetricsApplicationPayload: &MetricsApplicationPayload{
			ID:        event.Options.ApplicationID,
			Name:      event.Options.ApplicationName,
			OwnerName: event.Options.ApplicationOwnerName,
		},
		ApplicationStartedByName: event.Options.ApplicationStartedByName,
		BuildID:                  event.Options.BuildID,
		DeployID:                 event.Options.DeployID,
		Event:                    "buildStepFinished",
		GitVersion:               GitCommit,
		Step:                     event.Step,
		StepOrder:                event.Order,
		TimeElapsed:              elapsed.Seconds(),
		Timestamp:                int32(time.Now().Unix()),
		VCS:                      "git",
		Successful:               event.Successful,
		// Box:     event.Box,
		// Core:      "",
		// JobID:     "",
	})
}

// ListenTo will add eventhandlers to e.
func (h *MetricsEventHandler) ListenTo(e *emission.Emitter) {
	e.AddListener(BuildStepStarted, h.BuildStepStarted)
	e.AddListener(BuildStepFinished, h.BuildStepFinished)
}
