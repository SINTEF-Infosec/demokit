// +build hardware

package raspberrypi

import (
	log "github.com/sirupsen/logrus"
)

// SenseHatRaspberry is a pre-defined Hardware Layer for a RaspberryPi equipped with a SenseHat
type SenseHatRaspberry struct {
	logger *log.Entry
	*SenseHat
	eventHandler func(interface{})
}

func NewRaspberryPiWithSenseHat() *SenseHatRaspberry {
	senseHat, err := NewSenseHat()
	if err != nil {
		log.Fatalf("Could not instantiate new sense hat: %v", err)
	}
	return &SenseHatRaspberry{
		SenseHat: senseHat,
		eventHandler: func(_ interface{}) {},
	}
}

func (r *SenseHatRaspberry) Init() {
	r.SenseHat.StartListeningForJoystickEvents(func(event InputEvent) {
		r.eventHandler(event)
	}, false)
	r.logger.Info("Listening for joystick events")
}

func (r *SenseHatRaspberry) SetLogger(logger *log.Entry) {
	r.logger = logger
}

func (r *SenseHatRaspberry) SetEventHandler(handler func(interface{})) {
	r.eventHandler = handler
}

func (r *SenseHatRaspberry) IsAvailable() bool {
	return true
}
