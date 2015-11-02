// This program is a simple pluggable suscriber for Marathon event
// bus. (i.e.
// https://mesosphere.github.io/marathon/docs/event-bus.html)
//
// Author: Lucien R. Zagabe <rz@ognitio.com>

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/getsentry/raven-go"
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
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	event := make(map[string]interface{})
	if err := json.Unmarshal(b, &event); err != nil {
		log.Errorf("Unable to unmarshal request body: %v", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Infof("New Marathon event received: %s", event["eventType"].(string))

	err = renderEvent(event)
	if err != nil {
		rw.WriteHeader(http.StatusOK) // Ignore it
		return
	}

	rw.WriteHeader(http.StatusOK)
}

var (
	updateEventLevel = map[string]raven.Severity{
		"TASK_STAGING":  raven.INFO,
		"TASK_STARTING": raven.INFO,
		"TASK_RUNNING":  raven.INFO,
		"TASK_FINISHED": raven.INFO,
		"TASK_FAILED":   raven.ERROR,
		"TASK_KILLED":   raven.WARNING,
		"TASK_LOST":     raven.ERROR,
	}
)

func renderEvent(event map[string]interface{}) error {
	switch event["eventType"].(string) {
	case "api_post_event":
		// TODO(rzagabe): Support API request event.
		return nil
	case "status_update_event":
		return captureMessage(updateEventLevel[event["taskStatus"].(string)], event)
	}
	return fmt.Errorf("Not supported")
}

const (
	taskTemplate = `%s: %s (host: %s)

Timestamp: %s
Host: %s
Version: %s
`
)

func captureMessage(level raven.Severity, event map[string]interface{}) error {
	packet := raven.NewPacket(fmt.Sprintf(taskTemplate,
		event["taskStatus"].(string),
		event["appId"].(string), event["host"].(string),
		event["timestamp"].(string), event["host"].(string),
		event["version"].(string)))
	packet.Init(os.Getenv("SENTRY_PROJECT"))
	packet.Level = level
	packet.Interfaces = []raven.Interface{&raven.Message{
		Message: fmt.Sprintf("%s %s",
			event["taskStatus"].(string), event["appId"].(string)),
	}}
	raven.Capture(packet, map[string]string{
		"appId":  event["appId"].(string),
		"taskId": event["taskId"].(string),
	})
	return nil
}
