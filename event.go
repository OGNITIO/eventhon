package main

import (
	"errors"
	"fmt"

	"github.com/getsentry/raven-go"
	log "github.com/golang/glog"
)

var (
	eventLevel = map[string]raven.Severity{
		"TASK_STAGING":  raven.INFO,
		"TASK_STARTING": raven.INFO,
		"TASK_RUNNING":  raven.INFO,
		"TASK_FINISHED": raven.INFO,
		"TASK_FAILED":   raven.ERROR,
		"TASK_KILLED":   raven.WARNING,
		"TASK_LOST":     raven.ERROR,
	}
)

const (
	taskEventTemplate = `%s: %s [host: %s]

Timestamp: %s
Host: %s
Version: %s
`
)

var (
	ErrUnsupportedEvent = errors.New("eventhon: unsupported event")
)

func checkAndSendEvent(event map[string]interface{}) error {
	// TODO(rzagabe): Support framework related events.

	switch event["eventType"].(string) {
	case "status_update_event":
		switch event["taskStatus"].(string) {
		case "TASK_FINISHED", "TASK_RUNNING", "TASK_FAILED", "TASK_LOST":
			log.Infof("New status update event handled: %s")
			return captureEvent(event)
		}
	}

	return ErrUnsupportedEvent
}

func captureEvent(event map[string]interface{}) error {
	packet := raven.NewPacket(fmt.Sprintf(taskEventTemplate,
		event["taskStatus"].(string),
		event["appId"].(string), event["host"].(string),
		event["timestamp"].(string), event["host"].(string),
		event["version"].(string)))

	packet.Level = eventLevel[event["taskStatus"].(string)]

	// Sentry interfaces for events aggregation.
	packet.Interfaces = []raven.Interface{&raven.Message{
		Message: event["taskStatus"].(string) + event["appId"].(string),
	}}

	// TODO(rzagabe): Add link to mesos sandbox.
	raven.Capture(packet, map[string]string{
		"appId":  event["appId"].(string),
		"taskId": event["taskId"].(string),
	})

	return nil
}
