package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
)

type dockerHubWebhookRequest struct {
	CallbackUrl string `json:"callback_url"`
}

type dockerHubCallbackPayload struct {
	State       string `json:"state"`
	Description string `json:"description"`
	Context     string `json:"context"`
	TargetUrl   string `json:"target_url"`
}

func callbackDockerHub(w http.ResponseWriter, url string, rErr error) {
	var payload dockerHubCallbackPayload
	if rErr == nil {
		payload.State = "success"
	} else {
		payload.State = "error"
		payload.Description = rErr.Error()
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		log.Printf("error writing docker callback payload to buffer: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = http.Post(url, "application/json", &buf)
	if err != nil {
		log.Printf("error calling back docker hub: %s", err.Error())
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func (svc *server) handleDockerHubCallback(w http.ResponseWriter, r *http.Request) {
	var req dockerHubWebhookRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("error decoding webhook request: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body.Close()
	w.WriteHeader(http.StatusOK)

	go func(callbackUrl string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		resp, err := svc.docker.ImagePull(ctx, svc.cfg.RendererImage, types.ImagePullOptions{})
		if err != nil {
			callbackDockerHub(w, callbackUrl, err)
			return
		}
		defer resp.Close()
		_, err = io.Copy(ioutil.Discard, resp)
		if err != nil {
			log.Printf("error discarding docker image pull response: %s", err.Error())
		}
		callbackDockerHub(w, callbackUrl, nil)
	}(req.CallbackUrl)
}
