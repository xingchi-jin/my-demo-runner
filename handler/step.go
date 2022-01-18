package handler

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/harness/lite-engine/api"
	"github.com/harness/lite-engine/engine"
	"github.com/harness/lite-engine/engine/spec"
	"github.com/harness/lite-engine/logger"
	"github.com/harness/lite-engine/pipeline"
	pruntime "github.com/harness/lite-engine/pipeline/runtime"
)

// HandleExecuteStep returns an http.HandlerFunc that executes a step
func HandleStartStep(e *pruntime.StepExecutor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()

		var s api.StartStepRequest
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			WriteBadRequest(w, err)
			return
		}

		s.Volumes = append(s.Volumes, getSharedVolumeMount(), getDockerSockVolumeMount())

		if err := e.StartStep(r.Context(), &s); err != nil {
			WriteError(w, err)
		} else {
			WriteJSON(w, api.StartStepResponse{}, http.StatusOK)
		}

		logger.FromRequest(r).
			WithField("latency", time.Since(st)).
			WithField("time", time.Now().Format(time.RFC3339)).
			Infoln("api: successfully started the step")
	}
}

func HandlePollStep(e *pruntime.StepExecutor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()

		var s api.PollStepRequest
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			WriteBadRequest(w, err)
			return
		}

		if response, err := e.PollStep(r.Context(), &s); err != nil {
			WriteError(w, err)
		} else {
			WriteJSON(w, response, http.StatusOK)
		}

		logger.FromRequest(r).
			WithField("latency", time.Since(st)).
			WithField("time", time.Now().Format(time.RFC3339)).
			Infoln("api: successfully polled the step response")
	}
}

func HandleStreamOutput(e *pruntime.StepExecutor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()

		var s api.StreamOutputRequest
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			WriteBadRequest(w, err)
			return
		}

		var count int

		oldData, newData, err := e.StreamOutput(r.Context(), &s)
		if err != nil {
			WriteError(w, err)
			return
		}

		flusher, _ := w.(http.Flusher)

		_, _ = w.Write(oldData)
		count += len(oldData)
		if flusher != nil {
			flusher.Flush()
		}

	out:
		for {
			select {
			case <-r.Context().Done():
				break out
			case data, ok := <-newData:
				if !ok {
					break out
				}
				_, _ = w.Write(data)
				count += len(data)
				if flusher != nil {
					flusher.Flush()
				}
			}
		}

		logger.FromRequest(r).
			WithField("latency", time.Since(st)).
			WithField("time", time.Now().Format(time.RFC3339)).
			WithField("count", count).
			Infoln("api: successfully streamed the step log")
	}
}

func getSharedVolumeMount() *spec.VolumeMount {
	return &spec.VolumeMount{
		Name: pipeline.SharedVolName,
		Path: pipeline.SharedVolPath,
	}
}

func getDockerSockVolumeMount() *spec.VolumeMount {
	path := engine.DockerSockUnixPath
	if runtime.GOOS == "windows" {
		path = engine.DockerSockWinPath
	}
	return &spec.VolumeMount{
		Name: engine.DockerSockVolName,
		Path: path,
	}
}
