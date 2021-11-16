// +build libvlc_available

package core

import (
	"fmt"
	"github.com/SINTEF-Infosec/demokit/media/vlc"
	log "github.com/sirupsen/logrus"
)

const (
	InternalMediaStarted = "I_MEDIA_STARTED"
	InternalMediaPaused  = "I_MEDIA_PAUSED"
	InternalMediaEnded   = "I_MEDIA_ENDED"
)

// NewDefaultNodeWithVideo returns a Node with a default configuration for a Node having video capabilities.
// In addition to the mandatory components of a Node, the Media Controller is available.
//
// This Node is only available if libvlc is available on the system it is build on (libvlc_available tag when building).
func NewDefaultNodeWithVideo() *Node {
	info := NodeInfo{} // Will default to NODE_NAME or to a random name
	logger := log.NewEntry(log.New())

	defaultHost := getFromEnvOrFail("RABBIT_MQ_HOST", info.Name)
	registrationServer := getFromEnvOrFail("REGISTRATION_SERVER", info.Name)

	rabbitMQEventNetwork := NewRabbitMQEventNetwork(ConnexionDetails{
		Username: getFromEnvOrFail("RABBIT_MQ_USERNAME", info.Name),
		Password: getFromEnvOrFail("RABBIT_MQ_PASSWORD", info.Name),
		Host:     defaultHost,
		Port:     getFromEnvOrFail("RABBIT_MQ_PORT", info.Name),
	})

	mediaController, err := vlc.NewVLCMediaController()
	if err != nil {
		log.Fatalf("could not create media controller: %v", err)
	}

	rs := NewDefaultRegistrationServer(fmt.Sprintf("%s:4000", registrationServer))
	n := NewNode(info, DefaultNodeConfig(), logger, rs, rabbitMQEventNetwork, mediaController, nil)

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
