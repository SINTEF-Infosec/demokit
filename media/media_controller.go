package media

import log "github.com/sirupsen/logrus"

type MediaEventCallback func()

type MediaController interface {
	IsAvailable() bool
	SetLogger(entry *log.Entry)

	LoadMediaFromPath(path string) error
	LoadMediaFromURL(url string) error

	Play() error
	Pause() error
	Mute() error
	Stop() error

	SetOnMediaStartedCallback(cb MediaEventCallback)
	SetOnMediaPausedCallback(cb MediaEventCallback)
	SetOnMediaEndedCallback(cb MediaEventCallback)

	GetCurrentMediaPosition() (float32, error)
	SetCurrentMediaPosition(float32) error
}
