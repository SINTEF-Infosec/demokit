// +build libvlc_available

package core

import (
	"fmt"
	"github.com/SINTEF-Infosec/demokit/media"
	"github.com/goombaio/namegenerator"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

// NewDefaultNodeWithVideo returns a Node with a default configuration for a Node having video capabilities.
// In addition to the mandatory components of a Node, the Media Controller is available.
//
// This Node is only available if libvlc is available on the system it is build on (libvlc_available tag when building).
func NewDefaultNodeWithVideo() *Node {
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

	mediaController, err := media.NewVLCMediaController(logger)
	if err != nil {
		log.Fatalf("could not create media controller: %v", err)
	}
	n := newNode(info, logger, rabbitMQEventNetwork, router, mediaController, nil)

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
