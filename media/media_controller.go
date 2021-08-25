package media

type MediaEventCallback func()

type MediaController interface {
	LoadMediaFromPath(path string) error
	LoadMediaFromURL(url string) error

	Play() error
	Pause() error
	Mute() error
	Stop() error

	SetOnMediaStartedCallback(cb MediaEventCallback)
	SetOnMediaPausedCallback(cb MediaEventCallback)
	SetOnMediaEndedCallback(cb MediaEventCallback)
}
