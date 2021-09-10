package core

import log "github.com/sirupsen/logrus"

type EventHandler func(event *Event)

type EventNetwork interface {
	BroadcastEvent(event *Event)
	SendEventTo(receiver string, event *Event)
	SetReceivedEventCallback(handler EventHandler)
	StartListeningForEvents()
	SetLogger(entry *log.Entry)
}
