/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

const (
	appPort                 = 3000
	daprV1URL               = "http://localhost:3500/v1.0"
	actorMethodURLFormat    = daprV1URL + "/actors/%s/%s/%s/%s"
	actorSaveStateURLFormat = daprV1URL + "/actors/%s/%s/state/"
	actorGetStateURLFormat  = daprV1URL + "/actors/%s/%s/state/%s/"
	defaultActorType        = "testActorPubsubTypePerf" // Actor type must be unique per test app.
	actorTypeEnvName        = "TEST_APP_ACTOR_TYPE"     // Env variable tests can set to change actor type.
	actorIdleTimeout        = "1h"
	actorScanInterval       = "30s"
	drainOngoingCallTimeout = "30s"
	drainRebalancedActors   = true
	secondsToWaitInMethod   = 5
	defaultpubsubName       = "messagebus"
	defaultTopicName        = "topicActorPubsubPerf"
)

type PubSubConfig struct {
	PubSubName           string `json:"pubsubName"`
	Topic                string `json:"topic"`
	ActorType            string `json:"actorType"`
	Method               string `json:"method"`
	ActorIDDataAttribute string `json:"actorIDDataAttribute,omitempty"`
}

type daprConfig struct {
	Entities                []string       `json:"entities,omitempty"`
	ActorIdleTimeout        string         `json:"actorIdleTimeout,omitempty"`
	ActorScanInterval       string         `json:"actorScanInterval,omitempty"`
	DrainOngoingCallTimeout string         `json:"drainOngoingCallTimeout,omitempty"`
	DrainRebalancedActors   bool           `json:"drainRebalancedActors,omitempty"`
	Pubsub                  []PubSubConfig `json:"pubsub,omitempty"`
}

var registeredActorType = getActorType()

var daprConfigResponse = daprConfig{
	Entities:                []string{defaultActorType},
	ActorIdleTimeout:        actorIdleTimeout,
	ActorScanInterval:       actorScanInterval,
	DrainOngoingCallTimeout: drainOngoingCallTimeout,
	DrainRebalancedActors:   drainRebalancedActors,
	Pubsub: []PubSubConfig{
		{
			PubSubName: defaultpubsubName,
			Topic:      defaultTopicName,
			ActorType:  defaultActorType,
			Method:     "mymethod",
		},
	},
}

func getActorType() string {
	actorType := os.Getenv(actorTypeEnvName)
	if actorType == "" {
		return defaultActorType
	}

	return actorType
}

// indexHandler is the handler for root path
func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("indexHandler is called")
	w.WriteHeader(http.StatusOK)
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Processing dapr request for %s, responding with %v", r.URL.RequestURI(), daprConfigResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(daprConfigResponse)
}

func actorMethodHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Processing actor method request for %s", r.URL.RequestURI())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func deactivateActorHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Processing %s actor request for %s", r.Method, r.URL.RequestURI())
	actorType := mux.Vars(r)["actorType"]

	if actorType != registeredActorType {
		log.Printf("Unknown actor type: %s", actorType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

// appRouter initializes restful api router
func appRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", indexHandler).Methods("GET")
	router.HandleFunc("/dapr/config", configHandler).Methods("GET")

	router.HandleFunc("/actors/{actorType}/{id}/method/{method}", actorMethodHandler).Methods("PUT")
	router.HandleFunc("actors/{actorType}/{actorId}/publish/{pubsubname}/{topic}", actorMethodHandler).Methods("POST")
	router.HandleFunc("/actors/{actorType}/{id}", deactivateActorHandler).Methods("POST", "DELETE")

	router.HandleFunc("/healthz", healthzHandler).Methods("GET")

	router.Use(mux.CORSMethodMiddleware(router))

	return router
}

func main() {
	// initGRPCClient()
	log.Printf("Actor App - listening on http://localhost:%d", appPort)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", appPort), appRouter()))
}
