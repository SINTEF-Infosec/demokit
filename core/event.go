package core

const (
	InternalMediaStarted = "I_MEDIA_STARTED"
	InternalMediaPaused  = "I_MEDIA_PAUSED"
	InternalMediaEnded   = "I_MEDIA_ENDED"
)

type Event struct {
	Name     string
	Emitter  string
	Receiver string
	Payload  string
}
