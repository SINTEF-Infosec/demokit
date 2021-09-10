package core

import (
	"fmt"
	"github.com/SINTEF-Infosec/demokit/hardware/raspberrypi"
	log "github.com/sirupsen/logrus"
	"strings"
)

// NewDefaultRaspberryPiNode returns a Node with a default configuration for a RaspberryPi Node.
// In addition to the mandatory components of a Node, the Hardware Layer is available.
func NewDefaultRaspberryPiNode() *Node {
	info := NodeInfo{} // Will default to a NODE_NAME or to a random name
	logger := log.NewEntry(log.New())

	rabbitMQEventNetwork := NewRabbitMQEventNetwork(ConnexionDetails{
		Username: getFromEnvOrFail("RABBIT_MQ_USERNAME", info.Name),
		Password: getFromEnvOrFail("RABBIT_MQ_PASSWORD", info.Name),
		Host:     getFromEnvOrFail("RABBIT_MQ_HOST", info.Name),
		Port:     getFromEnvOrFail("RABBIT_MQ_PORT", info.Name),
	})

	rpi := raspberrypi.NewRaspberryPiWithSenseHat()

	n := NewNode(info, DefaultNodeConfig(), logger, NewDefaultRegistrationServer(), rabbitMQEventNetwork, nil, rpi)

	hardwareEventHandler := func(e interface{}) {
		inputEvent, ok := e.(raspberrypi.InputEvent)
		if !ok {
			n.Logger.Errorf("could not get event")
		}
		event := &Event{
			Name:     fmt.Sprintf("I_%s_%s", strings.ToUpper(inputEvent.Direction), strings.ToUpper(inputEvent.Action)),
			Emitter:  fmt.Sprintf("%s-hardware", n.Info.Name),
			Receiver: "*",
			Payload:  fmt.Sprintf("{\"timestamp\": %d }", inputEvent.Timestamp.Unix()),
		}
		n.handleEvent(event)
	}

	rpi.SetEventHandler(hardwareEventHandler)

	return n
}
