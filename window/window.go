package window // import "engo.io/engo/window"

// WindowResizeMessage is a message that's being dispatched whenever the game window is being resized by the gamer
type WindowResizeMessage struct {
	OldWidth, OldHeight int
	NewWidth, NewHeight int
}

func (WindowResizeMessage) Type() string { return "WindowResizeMessage" }
