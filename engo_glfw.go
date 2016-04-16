// +build !netgo,!android

package engo

import (
	"image"
	"image/draw"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"engo.io/webgl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var (
	window *glfw.Window

	Arrow     *glfw.Cursor
	IBeam     *glfw.Cursor
	Crosshair *glfw.Cursor
	Hand      *glfw.Cursor
	HResize   *glfw.Cursor
	VResize   *glfw.Cursor

	headlessWidth             = 800
	headlessHeight            = 800
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

	window, err = glfw.CreateWindow(width, height, title, nil, nil)
	fatalErr(err)

	window.MakeContextCurrent()

	if !fullscreen {
		window.SetPos((mode.Width-width)/2, (mode.Height-height)/2)
	}

	width, height = window.GetFramebufferSize()
	windowWidth, windowHeight = float32(width), float32(height)

	SetVSync(vsync)

	Gl = webgl.NewContext()
	Gl.Viewport(0, 0, width, height)

	window.SetFramebufferSizeCallback(func(window *glfw.Window, w, h int) {
		width, height = window.GetFramebufferSize()
		Gl.Viewport(0, 0, width, height)

		// TODO: when do we want to handle resizing? and who should deal with it?
		// responder.Resize(w, h)
	})

	window.SetCursorPosCallback(func(window *glfw.Window, x, y float64) {
		Mouse.X, Mouse.Y = float32(x), float32(y)
		Mouse.Action = MOVE
	})

	window.SetMouseButtonCallback(func(window *glfw.Window, b glfw.MouseButton, a glfw.Action, m glfw.ModifierKey) {
		x, y := window.GetCursorPos()
		Mouse.X, Mouse.Y = float32(x), float32(y)
		// this is only valid because we use an internal structure that is
		// 100% compatible with glfw3.h
		Mouse.Button = MouseButton(b)
		Mouse.Modifer = Modifier(m)

		if a == glfw.Press {
			Mouse.Action = PRESS
		} else {
			Mouse.Action = RELEASE
		}
	})

	window.SetScrollCallback(func(window *glfw.Window, xoff, yoff float64) {
		Mouse.ScrollX = float32(xoff)
		Mouse.ScrollY = float32(yoff)
	})

	window.SetKeyCallback(func(window *glfw.Window, k glfw.Key, s int, a glfw.Action, m glfw.ModifierKey) {
		key := Key(k)
		if a == glfw.Press {
			keyStates[key] = true
		} else if a == glfw.Release {
			keyStates[key] = false
		}
	})

	window.SetSizeCallback(func(w *glfw.Window, widthInt int, heightInt int) {
		windowWidth = float32(widthInt)
		windowHeight = float32(heightInt)

		if !scaleOnResize {
			gameWidth, gameHeight = float32(widthInt), float32(heightInt)

			// Update default batch
			for _, scene := range scenes {
				if scene.world == nil {
					continue // with other scenes
				}

				for _, s := range scene.world.Systems() {
					if _, ok := s.(*RenderSystem); ok {
						DefaultShader.SetProjection(gameWidth, gameHeight)
					}
				}
			}
		}

		// Update HUD batch
		for _, scene := range scenes {
			if scene.world == nil {
				continue // with other scenes
			}

			for _, s := range scene.world.Systems() {
				if _, ok := s.(*RenderSystem); ok {
					// TODO: don't call it directly, but let HUD listen for it
					//Shaders.HUD.SetProjection(windowWidth, windowHeight)
				}
			}
		}
	})

	window.SetCharCallback(func(window *glfw.Window, char rune) {
		// TODO: what does this do, when can we use it?
		// it's like KeyCallback, but for specific characters instead of keys...?
		// responder.Type(char)
	})
}

func DestroyWindow() {
	glfw.Terminate()
}

func SetTitle(title string) {
	if headless {
		log.Println("Title set to:", title)
	} else {
		window.SetTitle(title)
	}
}

// RunIteration runs one iteration / frame
func RunIteration() {
	// First check for new keypresses
	if !headless {
		glfw.PollEvents()
		keysUpdate()
	}

	// Then update the world and all Systems
	currentWorld.Update(Time.Delta())

	// Lastly, forget keypresses and swap buffers
	if !headless {
		Mouse.ScrollX, Mouse.ScrollY = 0, 0
		window.SwapBuffers()
		if window.ShouldClose() {
			closeEvent()
		}
	}

	Time.Tick()
}

func WindowWidth() float32 {
	return windowWidth
}

func WindowHeight() float32 {
	return windowHeight
}

func SetCursor(c *glfw.Cursor) {
	window.SetCursor(c)
}

func SetVSync(enabled bool) {
	vsync = enabled
	if vsync {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}
}

func init() {
	runtime.LockOSThread()

	Dash = Key(glfw.KeyMinus)
	Apostrophe = Key(glfw.KeyApostrophe)
	Semicolon = Key(glfw.KeySemicolon)
	Equals = Key(glfw.KeyEqual)
	Comma = Key(glfw.KeyComma)
	Period = Key(glfw.KeyPeriod)
	Slash = Key(glfw.KeySlash)
	Backslash = Key(glfw.KeyBackslash)
	Backspace = Key(glfw.KeyBackspace)
	Tab = Key(glfw.KeyTab)
	CapsLock = Key(glfw.KeyCapsLock)
	Space = Key(glfw.KeySpace)
	Enter = Key(glfw.KeyEnter)
	Escape = Key(glfw.KeyEscape)
	Insert = Key(glfw.KeyInsert)
	PrintScreen = Key(glfw.KeyPrintScreen)
	Delete = Key(glfw.KeyDelete)
	PageUp = Key(glfw.KeyPageUp)
	PageDown = Key(glfw.KeyPageDown)
	Home = Key(glfw.KeyHome)
	End = Key(glfw.KeyEnd)
	Pause = Key(glfw.KeyPause)
	ScrollLock = Key(glfw.KeyScrollLock)
	ArrowLeft = Key(glfw.KeyLeft)
	ArrowRight = Key(glfw.KeyRight)
	ArrowDown = Key(glfw.KeyDown)
	ArrowUp = Key(glfw.KeyUp)
	LeftBracket = Key(glfw.KeyLeftBracket)
	LeftShift = Key(glfw.KeyLeftShift)
	LeftControl = Key(glfw.KeyLeftControl)
	LeftSuper = Key(glfw.KeyLeftSuper)
	LeftAlt = Key(glfw.KeyLeftAlt)
	RightBracket = Key(glfw.KeyRightBracket)
	RightShift = Key(glfw.KeyRightShift)
	RightControl = Key(glfw.KeyRightControl)
	RightSuper = Key(glfw.KeyRightSuper)
	RightAlt = Key(glfw.KeyRightAlt)
	Zero = Key(glfw.Key0)
	One = Key(glfw.Key1)
	Two = Key(glfw.Key2)
	Three = Key(glfw.Key3)
	Four = Key(glfw.Key4)
	Five = Key(glfw.Key5)
	Six = Key(glfw.Key6)
	Seven = Key(glfw.Key7)
	Eight = Key(glfw.Key8)
	Nine = Key(glfw.Key9)
	F1 = Key(glfw.KeyF1)
	F2 = Key(glfw.KeyF2)
	F3 = Key(glfw.KeyF3)
	F4 = Key(glfw.KeyF4)
	F5 = Key(glfw.KeyF5)
	F6 = Key(glfw.KeyF6)
	F7 = Key(glfw.KeyF7)
	F8 = Key(glfw.KeyF8)
	F9 = Key(glfw.KeyF9)
	F10 = Key(glfw.KeyF10)
	F11 = Key(glfw.KeyF11)
	F12 = Key(glfw.KeyF12)
	A = Key(glfw.KeyA)
	B = Key(glfw.KeyB)
	C = Key(glfw.KeyC)
	D = Key(glfw.KeyD)
	E = Key(glfw.KeyE)
	F = Key(glfw.KeyF)
	G = Key(glfw.KeyG)
	H = Key(glfw.KeyH)
	I = Key(glfw.KeyI)
	J = Key(glfw.KeyJ)
	K = Key(glfw.KeyK)
	L = Key(glfw.KeyL)
	M = Key(glfw.KeyM)
	N = Key(glfw.KeyN)
	O = Key(glfw.KeyO)
	P = Key(glfw.KeyP)
	Q = Key(glfw.KeyQ)
	R = Key(glfw.KeyR)
	S = Key(glfw.KeyS)
	T = Key(glfw.KeyT)
	U = Key(glfw.KeyU)
	V = Key(glfw.KeyV)
	W = Key(glfw.KeyW)
	X = Key(glfw.KeyX)
	Y = Key(glfw.KeyY)
	Z = Key(glfw.KeyZ)
	NumLock = Key(glfw.KeyNumLock)
	NumMultiply = Key(glfw.KeyKPMultiply)
	NumDivide = Key(glfw.KeyKPDivide)
	NumAdd = Key(glfw.KeyKPAdd)
	NumSubtract = Key(glfw.KeyKPSubtract)
	NumZero = Key(glfw.KeyKP0)
	NumOne = Key(glfw.KeyKP1)
	NumTwo = Key(glfw.KeyKP2)
	NumThree = Key(glfw.KeyKP3)
	NumFour = Key(glfw.KeyKP4)
	NumFive = Key(glfw.KeyKP5)
	NumSix = Key(glfw.KeyKP6)
	NumSeven = Key(glfw.KeyKP7)
	NumEight = Key(glfw.KeyKP8)
	NumNine = Key(glfw.KeyKP9)
	NumDecimal = Key(glfw.KeyKPDecimal)
	NumEnter = Key(glfw.KeyKPEnter)
}
