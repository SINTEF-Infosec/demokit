package core

type EventHandler func(event *Event)

type EventNetwork interface {
	BroadcastEvent(event *Event)
	SendEventTo(receiver string, event *Event)
	SetReceivedEventCallback(handler EventHandler)
	StartListeningForEvents()
}
