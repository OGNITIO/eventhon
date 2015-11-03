// This program is a simple and unfinished pluggable suscriber for
// Marathon event bus which captures task related events and send them
// to Sentry. (i.e. getsentry.com)
//
// - https://mesosphere.github.io/marathon/docs/event-bus.html
//
// Author: Lucien R. Zagabe <rz@ognitio.com>

package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"

	log "github.com/golang/glog"
	"github.com/gorilla/mux"
)

var (
	addr          = flag.String("addr", "", "IP address and port of eventhon (e.g. localhost:1337)")
	sentryDsn     = flag.String("sentry_dsn", "", "")
	sentryProject = flag.String("sentry_project", "", "Sentry project ID")
)

func main() {
	flag.Parse()

	log.Infof("Starting Marathon subscriber eventhon...")

	r := mux.NewRouter()
	r.HandleFunc("/callbacks", eventHandler).Methods("POST")

	http.Handle("/", r)

	log.Infof("Start listening at %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
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

	err = checkAndSendEvent(event)
	if err != nil {
		log.Warning("Couldn't send event: %v", err)
		rw.WriteHeader(http.StatusOK)
		return
	}

	rw.WriteHeader(http.StatusOK)
}
