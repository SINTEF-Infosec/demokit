// +build libvlc_available

package media

import (
	"fmt"
	"github.com/adrg/libvlc-go/v3"
	log "github.com/sirupsen/logrus"
)

type VLCMediaController struct {
	logger       *log.Entry
	player       *vlc.Player
	currentMedia *vlc.Media
	eventManager *vlc.EventManager

	onMediaStartedCallback MediaEventCallback
	onMediaPausedCallback  MediaEventCallback
	onMediaEndedCallback   MediaEventCallback
}

func NewVLCMediaController(logger *log.Entry) (*VLCMediaController, error) {
	mediaControllerLogger := logger.WithField("component", "media-controller")

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
		func(event vlc.Event, data interface{}) { mediaController.onMediaStartedCallback() }, nil)
	if err != nil {
		mediaControllerLogger.Errorf("could not attach event: %v", err)
		return nil, err
	}

	_, err = manager.Attach(vlc.MediaPlayerPaused,
		func(event vlc.Event, data interface{}) { mediaController.onMediaPausedCallback() }, nil)
	if err != nil {
		mediaControllerLogger.Errorf("could not attach event: %v", err)
		return nil, err
	}

	_, err = manager.Attach(vlc.MediaPlayerEndReached,
		func(event vlc.Event, data interface{}) { mediaController.onMediaEndedCallback() }, nil)
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
	if mc.currentMedia != nil {
		if err := mc.currentMedia.Release(); err != nil {
			mc.logger.Errorf("could not release previous media: %v", err)
		}
	}

	media, err := mc.player.LoadMediaFromPath(path)
	if err != nil {
		mc.logger.Fatalf("could not load media from file: %v", err)
		return err
	}

	mc.currentMedia = media
	return nil
}

func (mc *VLCMediaController) LoadMediaFromURL(url string) error {
	if mc.currentMedia != nil {
		if err := mc.currentMedia.Release(); err != nil {
			mc.logger.Errorf("could not release previous media: %v", err)
		}
	}

	media, err := mc.player.LoadMediaFromURL(url)
	if err != nil {
		mc.logger.Fatalf("could not load media from url: %v", err)
		return err
	}

	mc.currentMedia = media
	return nil
}

func (mc *VLCMediaController) Play() error {
	if mc.currentMedia != nil {
		return mc.player.Play()
	}
	return fmt.Errorf("no media loaded")
}

func (mc *VLCMediaController) Pause() error {
	if mc.currentMedia != nil {
		return mc.player.TogglePause()
	}
	return fmt.Errorf("no media loaded")
}

func (mc *VLCMediaController) Mute() error {
	if mc.currentMedia != nil {
		return mc.player.ToggleMute()
	}
	return fmt.Errorf("no media loaded")
}

func (mc *VLCMediaController) Stop() error {
	if mc.currentMedia != nil {
		mc.currentMedia.Release()
		mc.currentMedia = nil
		return mc.player.Stop()
	}
	return fmt.Errorf("no media loaded")
}

func (mc *VLCMediaController) SetOnMediaStartedCallback(cb MediaEventCallback) {
	mc.onMediaStartedCallback = cb
}

func (mc *VLCMediaController) SetOnMediaPausedCallback(cb MediaEventCallback) {
	mc.onMediaPausedCallback = cb
}

func (mc *VLCMediaController) SetOnMediaEndedCallback(cb MediaEventCallback) {
	mc.onMediaEndedCallback = cb
}
