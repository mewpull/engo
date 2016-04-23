// +build !netgo,!android

package window

import (
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"engo.io/engo/assets"
	"engo.io/engo/clock"
	"engo.io/engo/input"
	internalengo "engo.io/engo/internal/engo"
	internalinput "engo.io/engo/internal/input"
	internalwindow "engo.io/engo/internal/window"
	"engo.io/engo/math"
	"engo.io/engo/message"
	"engo.io/engo/scene"
	"engo.io/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

var (
	Arrow     *glfw.Cursor
	IBeam     *glfw.Cursor
	Crosshair *glfw.Cursor
	Hand      *glfw.Cursor
	HResize   *glfw.Cursor
	VResize   *glfw.Cursor

	close                     bool
	resetLoopTicker           = make(chan bool, 1)
	fpsLimit                  = 60
	defaultCloseAction        bool
	headlessWidth             = 800
	headlessHeight            = 800
	gameWidth, gameHeight     float32
	windowWidth, windowHeight float32
)

// fatalErr calls log.Fatal with the given error if it is non-nil.
func fatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func CreateWindow(title string, width, height int, fullscreen bool) {
	err := glfw.Init()
	fatalErr(err)

	Arrow = glfw.CreateStandardCursor(int(glfw.ArrowCursor))
	Hand = glfw.CreateStandardCursor(int(glfw.HandCursor))
	IBeam = glfw.CreateStandardCursor(int(glfw.IBeamCursor))
	Crosshair = glfw.CreateStandardCursor(int(glfw.CrosshairCursor))

	monitor := glfw.GetPrimaryMonitor()
	mode := monitor.GetVideoMode()

	gameWidth = float32(width)
	gameHeight = float32(height)

	if fullscreen {
		width = mode.Width
		height = mode.Height
		glfw.WindowHint(glfw.Decorated, 0)
	} else {
		monitor = nil
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	internalwindow.Window, err = glfw.CreateWindow(width, height, title, nil, nil)
	fatalErr(err)

	internalwindow.Window.MakeContextCurrent()

	if !fullscreen {
		internalwindow.Window.SetPos((mode.Width-width)/2, (mode.Height-height)/2)
	}

	width, height = internalwindow.Window.GetFramebufferSize()
	windowWidth, windowHeight = float32(width), float32(height)

	SetVSync(internalwindow.Vsync)

	internalwindow.Gl = gl.NewContext()
	internalwindow.Gl.Viewport(0, 0, width, height)

	internalwindow.Window.SetFramebufferSizeCallback(func(window *glfw.Window, w, h int) {
		width, height = internalwindow.Window.GetFramebufferSize()
		internalwindow.Gl.Viewport(0, 0, width, height)

		// TODO: when do we want to handle resizing? and who should deal with it?
		// responder.Resize(w, h)
	})

	internalwindow.Window.SetCursorPosCallback(func(window *glfw.Window, x, y float64) {
		internalinput.Mouse.X, internalinput.Mouse.Y = float32(x), float32(y)
		internalinput.Mouse.Action = input.MOVE
	})

	internalwindow.Window.SetMouseButtonCallback(func(window *glfw.Window, b glfw.MouseButton, a glfw.Action, m glfw.ModifierKey) {
		x, y := internalwindow.Window.GetCursorPos()
		internalinput.Mouse.X, internalinput.Mouse.Y = float32(x), float32(y)
		// this is only valid because we use an internal structure that is
		// 100% compatible with glfw3.h
		internalinput.Mouse.Button = input.MouseButton(b)
		internalinput.Mouse.Modifer = input.Modifier(m)

		if a == glfw.Press {
			internalinput.Mouse.Action = input.PRESS
		} else {
			internalinput.Mouse.Action = input.RELEASE
		}
	})

	internalwindow.Window.SetScrollCallback(func(window *glfw.Window, xoff, yoff float64) {
		internalinput.Mouse.ScrollX = float32(xoff)
		internalinput.Mouse.ScrollY = float32(yoff)
	})

	internalwindow.Window.SetKeyCallback(func(window *glfw.Window, k glfw.Key, s int, a glfw.Action, m glfw.ModifierKey) {
		key := input.Key(k)
		if a == glfw.Press {
			input.KeyStates[key] = true
		} else if a == glfw.Release {
			input.KeyStates[key] = false
		}
	})

	internalwindow.Window.SetSizeCallback(func(w *glfw.Window, widthInt int, heightInt int) {
		msg := WindowResizeMessage{
			OldWidth:  int(windowWidth),
			OldHeight: int(windowHeight),
			NewWidth:  widthInt,
			NewHeight: heightInt,
		}

		windowWidth = float32(widthInt)
		windowHeight = float32(heightInt)

		if !internalwindow.ScaleOnResize {
			gameWidth, gameHeight = float32(widthInt), float32(heightInt)
		}

		message.Mailbox.Dispatch(msg)
	})

	internalwindow.Window.SetCharCallback(func(window *glfw.Window, char rune) {
		// TODO: what does this do, when can we use it?
		// it's like KeyCallback, but for specific characters instead of keys...?
		// responder.Type(char)
	})
}

func DestroyWindow() {
	glfw.Terminate()
}

func SetTitle(title string) {
	if internalengo.Headless {
		log.Println("Title set to:", title)
	} else {
		internalwindow.Window.SetTitle(title)
	}
}

// TODO(u): Unexport RunHeadless.
func RunHeadless(defaultScene scene.Scene) {
	RunLoop(defaultScene, true)
}

// RunIteration runs one iteration / frame
func RunIteration() {
	// First check for new keypresses
	if !internalengo.Headless {
		glfw.PollEvents()
		input.KeysUpdate()
	}

	// Then update the world and all Systems
	internalengo.CurrentWorld.Update(clock.Time.Delta())

	// Lastly, forget keypresses and swap buffers
	if !internalengo.Headless {
		internalinput.Mouse.ScrollX, internalinput.Mouse.ScrollY = 0, 0
		internalwindow.Window.SwapBuffers()
	}

	clock.Time.Tick()
}

// RunPreparation is called only once, and is called automatically when calling Open
// It is only here for benchmarking in combination with OpenHeadlessNoRun
func RunPreparation(defaultScene scene.Scene) {
	input.KeyStates = make(map[input.Key]bool)
	clock.Time = clock.NewClock()
	assets.Files = assets.NewLoader()

	// Default WorldBounds values
	internalwindow.WorldBounds.Max = math.Point{Width(), Height()}

	scene.SetScene(defaultScene, false)
}

func closeEvent() {
	for _, scenes := range scene.Scenes {
		if exiter, ok := scenes.Scene.(scene.Exiter); ok {
			exiter.Exit()
		}
	}

	if defaultCloseAction {
		Exit()
	} else {
		log.Println("Warning: default close action set to false, please make sure you manually handle this")
	}
}

// TODO(u): Unexport RunLoop.
func RunLoop(defaultScene scene.Scene, headless bool) {
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
			if !headless && internalwindow.Window.ShouldClose() {
				closeEvent()
			}
		case <-resetLoopTicker:
			ticker.Stop()
			ticker = time.NewTicker(time.Duration(int(time.Second) / fpsLimit))
		}
	}
	ticker.Stop()
}

func Width() float32 {
	return gameWidth
}

func Height() float32 {
	return gameHeight
}

func WindowWidth() float32 {
	return windowWidth
}

func WindowHeight() float32 {
	return windowHeight
}

func Exit() {
	close = true
}

func SetCursor(c *glfw.Cursor) {
	internalwindow.Window.SetCursor(c)
}

func SetVSync(enabled bool) {
	internalwindow.Vsync = enabled
	if internalwindow.Vsync {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}
}

func init() {
	runtime.LockOSThread()

	input.Dash = input.Key(glfw.KeyMinus)
	input.Apostrophe = input.Key(glfw.KeyApostrophe)
	input.Semicolon = input.Key(glfw.KeySemicolon)
	input.Equals = input.Key(glfw.KeyEqual)
	input.Comma = input.Key(glfw.KeyComma)
	input.Period = input.Key(glfw.KeyPeriod)
	input.Slash = input.Key(glfw.KeySlash)
	input.Backslash = input.Key(glfw.KeyBackslash)
	input.Backspace = input.Key(glfw.KeyBackspace)
	input.Tab = input.Key(glfw.KeyTab)
	input.CapsLock = input.Key(glfw.KeyCapsLock)
	input.Space = input.Key(glfw.KeySpace)
	input.Enter = input.Key(glfw.KeyEnter)
	input.Escape = input.Key(glfw.KeyEscape)
	input.Insert = input.Key(glfw.KeyInsert)
	input.PrintScreen = input.Key(glfw.KeyPrintScreen)
	input.Delete = input.Key(glfw.KeyDelete)
	input.PageUp = input.Key(glfw.KeyPageUp)
	input.PageDown = input.Key(glfw.KeyPageDown)
	input.Home = input.Key(glfw.KeyHome)
	input.End = input.Key(glfw.KeyEnd)
	input.Pause = input.Key(glfw.KeyPause)
	input.ScrollLock = input.Key(glfw.KeyScrollLock)
	input.ArrowLeft = input.Key(glfw.KeyLeft)
	input.ArrowRight = input.Key(glfw.KeyRight)
	input.ArrowDown = input.Key(glfw.KeyDown)
	input.ArrowUp = input.Key(glfw.KeyUp)
	input.LeftBracket = input.Key(glfw.KeyLeftBracket)
	input.LeftShift = input.Key(glfw.KeyLeftShift)
	input.LeftControl = input.Key(glfw.KeyLeftControl)
	input.LeftSuper = input.Key(glfw.KeyLeftSuper)
	input.LeftAlt = input.Key(glfw.KeyLeftAlt)
	input.RightBracket = input.Key(glfw.KeyRightBracket)
	input.RightShift = input.Key(glfw.KeyRightShift)
	input.RightControl = input.Key(glfw.KeyRightControl)
	input.RightSuper = input.Key(glfw.KeyRightSuper)
	input.RightAlt = input.Key(glfw.KeyRightAlt)
	input.Zero = input.Key(glfw.Key0)
	input.One = input.Key(glfw.Key1)
	input.Two = input.Key(glfw.Key2)
	input.Three = input.Key(glfw.Key3)
	input.Four = input.Key(glfw.Key4)
	input.Five = input.Key(glfw.Key5)
	input.Six = input.Key(glfw.Key6)
	input.Seven = input.Key(glfw.Key7)
	input.Eight = input.Key(glfw.Key8)
	input.Nine = input.Key(glfw.Key9)
	input.F1 = input.Key(glfw.KeyF1)
	input.F2 = input.Key(glfw.KeyF2)
	input.F3 = input.Key(glfw.KeyF3)
	input.F4 = input.Key(glfw.KeyF4)
	input.F5 = input.Key(glfw.KeyF5)
	input.F6 = input.Key(glfw.KeyF6)
	input.F7 = input.Key(glfw.KeyF7)
	input.F8 = input.Key(glfw.KeyF8)
	input.F9 = input.Key(glfw.KeyF9)
	input.F10 = input.Key(glfw.KeyF10)
	input.F11 = input.Key(glfw.KeyF11)
	input.F12 = input.Key(glfw.KeyF12)
	input.A = input.Key(glfw.KeyA)
	input.B = input.Key(glfw.KeyB)
	input.C = input.Key(glfw.KeyC)
	input.D = input.Key(glfw.KeyD)
	input.E = input.Key(glfw.KeyE)
	input.F = input.Key(glfw.KeyF)
	input.G = input.Key(glfw.KeyG)
	input.H = input.Key(glfw.KeyH)
	input.I = input.Key(glfw.KeyI)
	input.J = input.Key(glfw.KeyJ)
	input.K = input.Key(glfw.KeyK)
	input.L = input.Key(glfw.KeyL)
	input.M = input.Key(glfw.KeyM)
	input.N = input.Key(glfw.KeyN)
	input.O = input.Key(glfw.KeyO)
	input.P = input.Key(glfw.KeyP)
	input.Q = input.Key(glfw.KeyQ)
	input.R = input.Key(glfw.KeyR)
	input.S = input.Key(glfw.KeyS)
	input.T = input.Key(glfw.KeyT)
	input.U = input.Key(glfw.KeyU)
	input.V = input.Key(glfw.KeyV)
	input.W = input.Key(glfw.KeyW)
	input.X = input.Key(glfw.KeyX)
	input.Y = input.Key(glfw.KeyY)
	input.Z = input.Key(glfw.KeyZ)
	input.NumLock = input.Key(glfw.KeyNumLock)
	input.NumMultiply = input.Key(glfw.KeyKPMultiply)
	input.NumDivide = input.Key(glfw.KeyKPDivide)
	input.NumAdd = input.Key(glfw.KeyKPAdd)
	input.NumSubtract = input.Key(glfw.KeyKPSubtract)
	input.NumZero = input.Key(glfw.KeyKP0)
	input.NumOne = input.Key(glfw.KeyKP1)
	input.NumTwo = input.Key(glfw.KeyKP2)
	input.NumThree = input.Key(glfw.KeyKP3)
	input.NumFour = input.Key(glfw.KeyKP4)
	input.NumFive = input.Key(glfw.KeyKP5)
	input.NumSix = input.Key(glfw.KeyKP6)
	input.NumSeven = input.Key(glfw.KeyKP7)
	input.NumEight = input.Key(glfw.KeyKP8)
	input.NumNine = input.Key(glfw.KeyKP9)
	input.NumDecimal = input.Key(glfw.KeyKPDecimal)
	input.NumEnter = input.Key(glfw.KeyKPEnter)
}

func SetBackground(c color.Color) {
	if !internalengo.Headless {
		r, g, b, a := c.RGBA()

		internalwindow.Gl.ClearColor(float32(r), float32(g), float32(b), float32(a))
	}
}

func SetScaleOnResize(b bool) {
	internalwindow.ScaleOnResize = b
}

func SetOverrideCloseAction(value bool) {
	defaultCloseAction = !value
}

func SetFPSLimit(limit int) error {
	if limit <= 0 {
		return fmt.Errorf("FPS Limit out of bounds. Requires > 0")
	}
	fpsLimit = limit
	resetLoopTicker <- true
	return nil
}
