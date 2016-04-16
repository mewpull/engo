package engo // import "engo.io/engo"

import (
	"fmt"
	"image/color"

	"engo.io/ecs"
	"engo.io/webgl"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	Time        *Clock
	Files       *Loader
	Gl          *webgl.Context
	WorldBounds AABB

	currentWorld *ecs.World
	currentScene Scene
	Mailbox      *MessageManager
	cam          *cameraSystem

	gameWidth, gameHeight float32
	defaultCloseAction    bool
	close                 bool
	scaleOnResize         = false
	fpsLimit              = 60
	headless              = false
	vsync                 = true
	resetLoopTicker       = make(chan bool, 1)
)

func Width() float32 {
	return gameWidth
}

func Height() float32 {
	return gameHeight
}

func SetBackground(c color.Color) {
	if !headless {
		r, g, b, a := c.RGBA()
		Gl.ClearColor(float32(r), float32(g), float32(b), float32(a))
	}
}

func SetScaleOnResize(b bool) {
	scaleOnResize = b
}

func OverrideCloseAction() {
	defaultCloseAction = false
}

func SetFPSLimit(limit int) error {
	if limit <= 0 {
		return fmt.Errorf("FPS Limit out of bounds. Requires > 0")
	}
	fpsLimit = limit
	resetLoopTicker <- true
	return nil
}

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
}

func Run(opts RunOptions, defaultScene Scene) {
	// Save settings
	SetScaleOnResize(opts.ScaleOnResize)
	SetFPSLimit(opts.FPSLimit)
	vsync = opts.VSync
	defaultCloseAction = true

	if opts.HeadlessMode {
		headless = true

		if !opts.NoRun {
			runHeadless(defaultScene)
		}
	} else {
		CreateWindow(opts.Title, opts.Width, opts.Height, opts.Fullscreen)
		defer DestroyWindow()

		if !opts.NoRun {
			runLoop(defaultScene, false)
		}
	}
}

// RunPreparation is called only once, and is called automatically when calling Open
// It is only here for benchmarking in combination with OpenHeadlessNoRun
func RunPreparation(defaultScene Scene) {
	keyStates = make(map[Key]bool)
	Time = NewClock()
	Files = NewLoader()

	// Default WorldBounds values
	WorldBounds.Max = Point{Width(), Height()}

	SetScene(defaultScene, false)
}

func runHeadless(defaultScene Scene) {
	runLoop(defaultScene, true)
}

func runLoop(defaultScene Scene, headless bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		closeEvent()
	}()

	RunPreparation(defaultScene)

	ticker := time.NewTicker(time.Duration(int(time.Second) / fpsLimit))
Outer:
	for {
		select {
		case <-ticker.C:
			RunIteration()
			if close {
				break Outer
			}
		case <-resetLoopTicker:
			ticker.Stop()
			ticker = time.NewTicker(time.Duration(int(time.Second) / fpsLimit))
		}
	}
	ticker.Stop()
}

func closeEvent() {
	for _, scenes := range scenes {
		scenes.scene.Exit()
	}

	if defaultCloseAction {
		Exit()
	} else {
		log.Println("Warning: default close action set to false, please make sure you manually handle this")
	}
}

func Exit() {
	close = true
}
