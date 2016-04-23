package main

import (
	"image/color"
	"time"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/camera"
	"engo.io/engo/demos/demoutils"
	"engo.io/engo/message"
	"engo.io/engo/render"
	"engo.io/engo/window"
)

type DefaultScene struct{}

var (
	worldWidth  int = 800
	worldHeight int = 800
)

func (*DefaultScene) Preload() {}

// Setup is called before the main loop is started
func (*DefaultScene) Setup(w *ecs.World) {
	window.SetBackground(color.White)
	w.AddSystem(&render.RenderSystem{})

	demoutils.NewBackground(w, worldWidth, worldHeight, color.RGBA{102, 153, 0, 255}, color.RGBA{102, 173, 0, 255})

	// We issue one camera zoom command at the start, but it takes a while to process because we set a duration
	message.Mailbox.Dispatch(camera.CameraMessage{
		Axis:        camera.ZAxis,
		Value:       3, // so zooming out a lot
		Incremental: true,
		Duration:    time.Second * 5,
	})
}

func (*DefaultScene) Type() string { return "Game" }

func main() {
	opts := engo.RunOptions{
		Title:  "IncrementalCamera Demo",
		Width:  worldWidth,
		Height: worldHeight,
	}
	engo.Run(opts, &DefaultScene{})
}
