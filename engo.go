package engo // import "engo.io/engo"

import (
	internalengo "engo.io/engo/internal/engo"
	internalwindow "engo.io/engo/internal/window"
	"engo.io/engo/scene"
	"engo.io/engo/window"
)

type RunOptions struct {
	// NoRun indicates the Open function should return immediately, without looping
	NoRun bool

	// Title is the Window title
	Title string

	// HeadlessMode indicates whether or not OpenGL calls should be made
	HeadlessMode bool

	Fullscreen bool

	Width, Height int

	// VSync indicates whether or not OpenGL should wait for the monitor to swp the buffers
	VSync bool

	// ScaleOnResize indicates whether or not engo should make things larger/smaller whenever the screen resizes
	ScaleOnResize bool

	// FPSLimit indicates the maximum number of frames per second
	FPSLimit int

	// OverrideCloseAction indicates that (when true) engo will never close whenever the gamer wants to close the
	// game - that will be your responsibility
	OverrideCloseAction bool
}

func Run(opts RunOptions, defaultScene scene.Scene) {
	// Save settings
	window.SetScaleOnResize(opts.ScaleOnResize)
	window.SetFPSLimit(opts.FPSLimit)
	internalwindow.Vsync = opts.VSync
	window.SetOverrideCloseAction(opts.OverrideCloseAction)

	if opts.HeadlessMode {
		internalengo.Headless = true

		if !opts.NoRun {
			window.RunHeadless(defaultScene)
		}
	} else {
		window.CreateWindow(opts.Title, opts.Width, opts.Height, opts.Fullscreen)
		defer window.DestroyWindow()

		if !opts.NoRun {
			window.RunLoop(defaultScene, false)
		}
	}
}
