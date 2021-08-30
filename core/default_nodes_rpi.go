package core

import (
	"fmt"
	"github.com/SINTEF-Infosec/demokit/hardware"
	"github.com/goombaio/namegenerator"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

// NewDefaultRaspberryPiNode returns a Node with a default configuration for a RaspberryPi Node.
// In addition to the mandatory components of a Node, the Hardware Layer is available.
func NewDefaultRaspberryPiNode() *Node {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		seed := time.Now().UTC().UnixNano()
		nameGenerator := namegenerator.NewNameGenerator(seed)
		nodeName = nameGenerator.Generate()
	}

	info := NodeInfo{
		Name: nodeName,
	}

	logger := log.WithField("node", info.Name)

	rabbitMQEventNetwork := NewRabbitMQEventNetwork(ConnexionDetails{
		Username: getFromEnvOrFail("RABBIT_MQ_USERNAME", info.Name),
		Password: getFromEnvOrFail("RABBIT_MQ_PASSWORD", info.Name),
		Host:     getFromEnvOrFail("RABBIT_MQ_HOST", info.Name),
		Port:     getFromEnvOrFail("RABBIT_MQ_PORT", info.Name),
	}, logger)

	router := NewNodeRouter(logger)
	rpi := hardware.NewRaspberryPiWithSenseHat()

	n := NewNode(info, DefaultNodeConfig(), logger, rabbitMQEventNetwork, router, nil, rpi)

	if n.MediaController != nil {
		// By default, we emit "internal" event when there is a media event
		n.MediaController.SetOnMediaStartedCallback(func() {
			n.handleEvent(&Event{
				InternalMediaStarted,
				fmt.Sprintf("%s.media-controller", n.Info.Name),
				n.Info.Name,
				"{}"})
		})

		n.MediaController.SetOnMediaPausedCallback(func() {
			n.handleEvent(&Event{
				InternalMediaPaused,
				fmt.Sprintf("%s.media-controller", n.Info.Name),
				n.Info.Name,
				"{}"})
		})

		n.MediaController.SetOnMediaEndedCallback(func() {
			n.handleEvent(&Event{
				InternalMediaEnded,
				fmt.Sprintf("%s.media-controller", n.Info.Name),
				n.Info.Name,
				"{}"})
		})
	}

	return n

}
