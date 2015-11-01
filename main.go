// This program is a simple pluggable suscriber for Marathon event
// bus (i.e.
// https://mesosphere.github.io/marathon/docs/event-bus.html) which in
// turn forward events to Riemann servers. (i.e. http://riemann.io)

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/amir/raidman"
	log "github.com/golang/glog"
	"github.com/gorilla/mux"
)

var (
	host    = flag.String("host", "localhost", "Server's host.")
	port    = flag.Int64("port", 1337, "Server's port.")
	riemann = flag.String("riemann", "localhost:4242", "Riemann address.")
)

func main() {
	flag.Parse()

	log.Infof("Starting Marathon pluggable subscriber...")

	r := mux.NewRouter()
	r.HandleFunc("/callbacks", eventHandler).Methods("POST")

	http.Handle("/", r)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Infof("Start listening at %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Unable to listen and serve: %v", err)
	}
}

func eventHandler(rw http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Errorf("Unable to read HTTP request body: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	event := make(map[string]interface{})
	if err := json.Unmarshal(b, &event); err != nil {
		log.Errorf("Unable to unmarshal request body: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("New Marathon event received: %s", event["eventType"].(string))

	revent, err := renderEvent(event)
	if err != nil {
		rw.WriteHeader(http.StatusOK) // Ignore it
		return
	}

	c, err := raidman.Dial("tcp", *riemann)
	if err != nil {
		log.Errorf("Unable to connect riemann server: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := c.Send(revent); err != nil {
		log.Errorf("Couldn't send event: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

var (
	updateEventStates = map[string]string{
		"TASK_STAGING":  "ok",
		"TASK_STARTING": "ok",
		"TASK_RUNNING":  "ok",
		"TASK_FINISHED": "ok",
		"TASK_FAILED":   "critical",
		"TASK_KILLED":   "warning",
		"TASK_LOST":     "critical",
	}
)

func renderEvent(event map[string]interface{}) (*raidman.Event, error) {
	switch event["eventType"].(string) {
	case "api_post_event":
		// TODO(rzagabe): Support API request event.
		return nil, nil
	case "status_update_event":
		return &raidman.Event{
			State:   updateEventStates[event["taskStatus"].(string)],
			Host:    event["host"].(string),
			Service: event["appId"].(string),
			Tags:    []string{event["taskStatus"].(string)},
			Attributes: map[string]string{
				"timestamp": event["timestamp"].(string),
				"slaveId":   event["slaveId"].(string),
				"taskId":    event["taskId"].(string),
				"version":   event["version"].(string),
			},
		}, nil
	default:
		return nil, fmt.Errorf("Not supported")
	}
}
