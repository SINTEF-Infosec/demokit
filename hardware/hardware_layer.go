// Package hardware provides a basic interface for interacting with hardware
// Due to the wide range of hardware available,
package hardware

import (
	log "github.com/sirupsen/logrus"
)

// Hal (Hardware Abstraction Layer) provides access to hardware functionalities
// SetEventHandler is used to configure the handler for events emitted by the hardware layer. The type of events received
// varies depending on the hardware layer, so the handler will have to take care of checking the type of the received event.
// Init should be used to perform any required initialisations of the Hardware.
// Both SetEventHandler and Init will be called during the node initialisation.
type Hal interface {
	Init()
	SetLogger(entry *log.Entry)
	SetEventHandler(handler func(interface{}))
	IsAvailable() bool
}
