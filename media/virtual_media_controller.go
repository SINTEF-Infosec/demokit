package media

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

const UnavailableController = "media controller unavailable, this is a virtual controller"

type VirtualMediaController struct{}

func NewVirtualMediaController() *VirtualMediaController {
	return &VirtualMediaController{}
}

func (v VirtualMediaController) IsAvailable() bool {
	return false
}

func (v VirtualMediaController) SetLogger(_ *log.Entry) {}

func (v VirtualMediaController) LoadMediaFromPath(path string) error {
	return fmt.Errorf(UnavailableController)
}

func (v VirtualMediaController) LoadMediaFromURL(url string) error {
	return fmt.Errorf(UnavailableController)
}

func (v VirtualMediaController) Play() error {
	return fmt.Errorf(UnavailableController)
}

func (v VirtualMediaController) Pause() error {
	return fmt.Errorf(UnavailableController)
}

func (v VirtualMediaController) Mute() error {
	return fmt.Errorf(UnavailableController)
}

func (v VirtualMediaController) Stop() error {
	return fmt.Errorf(UnavailableController)
}

func (v VirtualMediaController) SetOnMediaStartedCallback(cb MediaEventCallback) {}

func (v VirtualMediaController) SetOnMediaPausedCallback(cb MediaEventCallback) {}

func (v VirtualMediaController) SetOnMediaEndedCallback(cb MediaEventCallback) {}

func (v VirtualMediaController) GetCurrentMediaPosition() (float32, error) {
	return 0.0, fmt.Errorf(UnavailableController)
}

func (v VirtualMediaController) SetCurrentMediaPosition(float32) error {
	return fmt.Errorf(UnavailableController)
}
