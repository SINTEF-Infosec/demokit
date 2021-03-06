// +build libvlc_available

package vlc

import (
	"fmt"
	"github.com/SINTEF-Infosec/demokit/media"
	"github.com/adrg/libvlc-go/v3"
	log "github.com/sirupsen/logrus"
)

type VLCMediaController struct {
	logger       *log.Entry
	player       *vlc.Player
	eventManager *vlc.EventManager

	onMediaStartedCallback media.MediaEventCallback
	onMediaPausedCallback  media.MediaEventCallback
	onMediaEndedCallback   media.MediaEventCallback
}

func (mc *VLCMediaController) SetLogger(logger *log.Entry) {
	mc.logger = logger
}

func NewVLCMediaController() (*VLCMediaController, error) {
	mediaControllerLogger := log.WithField("node", "na-media-controller-setup")

	if err := vlc.Init("--fullscreen", "--quiet"); err != nil {
		mediaControllerLogger.Errorf("could not init VLC: %v", err)
		return nil, err
	}

	player, err := vlc.NewPlayer()
	if err != nil {
		mediaControllerLogger.Errorf("could not create player: %v", err)
		return nil, err
	}

	manager, err := player.EventManager()
	if err != nil {
		mediaControllerLogger.Fatalf("could not retrieve event manager: %v", err)
		return nil, err
	}

	mediaController := &VLCMediaController{
		player:       player,
		eventManager: manager,
		logger:       mediaControllerLogger,
	}

	// Default media event callbacks
	mediaController.onMediaStartedCallback = func() { mediaControllerLogger.Debug("Media started") }
	mediaController.onMediaPausedCallback = func() { mediaControllerLogger.Debug("Media paused") }
	mediaController.onMediaEndedCallback = func() { mediaControllerLogger.Debug("Media ended") }

	// Registering default eventCallback against the event manager
	_, err = manager.Attach(vlc.MediaPlayerPlaying,
		func(event vlc.Event, data interface{}) { go mediaController.onMediaStartedCallback() }, nil)
	if err != nil {
		mediaControllerLogger.Errorf("could not attach event: %v", err)
		return nil, err
	}

	_, err = manager.Attach(vlc.MediaPlayerPaused,
		func(event vlc.Event, data interface{}) { go mediaController.onMediaPausedCallback() }, nil)
	if err != nil {
		mediaControllerLogger.Errorf("could not attach event: %v", err)
		return nil, err
	}

	_, err = manager.Attach(vlc.MediaPlayerEndReached,
		func(event vlc.Event, data interface{}) { go mediaController.onMediaEndedCallback() }, nil)
	if err != nil {
		mediaControllerLogger.Errorf("could not attach event: %v", err)
		return nil, err
	}

	return mediaController, nil
}

func (mc *VLCMediaController) IsAvailable() bool {
	return true
}

func (mc *VLCMediaController) LoadMediaFromPath(path string) error {
	if err := mc.releaseCurrentMedia(); err != nil {
		mc.logger.Errorf("could not release previous media: %v", err)
	}

	mc.logger.Debugf("loading media from file: %s", path)
	_, err := mc.player.LoadMediaFromPath(path)
	if err != nil {
		mc.logger.Fatalf("could not load media from file: %v", err)
		return err
	}

	return nil
}

func (mc *VLCMediaController) LoadMediaFromURL(url string) error {
	if err := mc.releaseCurrentMedia(); err != nil {
		mc.logger.Errorf("could not release previous media: %v", err)
	}

	mc.logger.Debugf("loading media from url: %s", url)
	_, err := mc.player.LoadMediaFromURL(url)
	if err != nil {
		mc.logger.Fatalf("could not load media from url: %v", err)
		return err
	}

	return nil
}

func (mc *VLCMediaController) Play() error {
	if mc.isMediaAvailable() {
		return mc.player.Play()
	}
	return fmt.Errorf("no media loaded")
}

func (mc *VLCMediaController) Pause() error {
	if mc.isMediaAvailable() {
		return mc.player.TogglePause()
	}
	return fmt.Errorf("no media loaded")
}

func (mc *VLCMediaController) Mute() error {
	if mc.isMediaAvailable() {
		return mc.player.ToggleMute()
	}
	return fmt.Errorf("no media loaded")
}

func (mc *VLCMediaController) Stop() error {
	if mc.isMediaAvailable() {
		if err := mc.releaseCurrentMedia(); err != nil {
			return err
		}
		return mc.player.Stop()
	}
	return fmt.Errorf("no media loaded")
}

func (mc *VLCMediaController) SetOnMediaStartedCallback(cb media.MediaEventCallback) {
	mc.onMediaStartedCallback = cb
}

func (mc *VLCMediaController) SetOnMediaPausedCallback(cb media.MediaEventCallback) {
	mc.onMediaPausedCallback = cb
}

func (mc *VLCMediaController) SetOnMediaEndedCallback(cb media.MediaEventCallback) {
	mc.onMediaEndedCallback = cb
}

func (mc *VLCMediaController) GetCurrentMediaPosition() (float32, error) {
	if mc.isMediaAvailable() {
		mediaPosition, err := mc.player.MediaPosition()
		if err != nil {
			return 0.0, fmt.Errorf("could not get media position: %v", err)
		}
		return mediaPosition, nil
	}
	return 0, fmt.Errorf("media unvailable")
}

func (mc *VLCMediaController) SetCurrentMediaPosition(position float32) error {
	if mc.isMediaAvailable() {
		err := mc.player.SetMediaPosition(position)
		if err != nil {
			return fmt.Errorf("could not set media position: %v", err)
		}
		return nil
	}
	return fmt.Errorf("media unvailable")
}

func (mc *VLCMediaController) isMediaAvailable() bool {
	if mc.player != nil {
		currentMedia, err := mc.player.Media()
		if err != nil {
			mc.logger.Errorf("could not get current media: %v", err)
			return false
		}
		return currentMedia != nil
	}
	return false
}

func (mc *VLCMediaController) releaseCurrentMedia() error {
	if mc.player != nil {
		currentMedia, err := mc.player.Media()
		if err != nil {
			mc.logger.Errorf("could not get current media: %v", err)
			return fmt.Errorf("could not get current media: %v", err)
		}

		err = currentMedia.Release()
		if err != nil {
			return fmt.Errorf("could not release current media: %v", err)
		}
		return nil
	}
	return fmt.Errorf("player is nil")
}
