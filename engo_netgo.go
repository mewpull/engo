package engo

import (
	"log"
	"math"
	"strconv"

	"engo.io/webgl"
	"github.com/gopherjs/gopherjs/js"
)

func init() {
	rafPolyfill()
}

var canvas *js.Object

func CreateWindow(title string, width, height int, fullscreen bool) {
	document := js.Global.Get("document")
	canvas = document.Call("createElement", "canvas")

	target := document.Call("getElementById", title)
	if target == nil {
		target = document.Get("body")
	}
	target.Call("appendChild", canvas)

	attrs := webgl.DefaultAttributes()
	attrs.Alpha = false
	attrs.Depth = false
	attrs.PremultipliedAlpha = false
	attrs.PreserveDrawingBuffer = false
	attrs.Antialias = false

	var err error
	Gl, err = webgl.NewContext(canvas, attrs)
	if err != nil {
		log.Fatal(err)
	}

	js.Global.Set("onunload", func() {
		closeEvent()
	})

	canvas.Get("style").Set("display", "block")
	winWidth := js.Global.Get("innerWidth").Int()
	winHeight := js.Global.Get("innerHeight").Int()
	if fullscreen {
		canvas.Set("width", winWidth)
		canvas.Set("height", winHeight)
	} else {
		canvas.Set("width", width)
		canvas.Set("height", height)
		canvas.Get("style").Set("marginLeft", toPx((winWidth-width)/2))
		canvas.Get("style").Set("marginTop", toPx((winHeight-height)/2))
	}

	canvas.Call("addEventListener", "mousemove", func(ev *js.Object) {
		rect := canvas.Call("getBoundingClientRect")
		x := float32((ev.Get("clientX").Int() - rect.Get("left").Int()))
		y := float32((ev.Get("clientY").Int() - rect.Get("top").Int()))
		//responder.Mouse(x, y, MOVE)
		log.Println("Mouse:", x, y)
	}, false)

	canvas.Call("addEventListener", "mousedown", func(ev *js.Object) {
		rect := canvas.Call("getBoundingClientRect")
		x := float32((ev.Get("clientX").Int() - rect.Get("left").Int()))
		y := float32((ev.Get("clientY").Int() - rect.Get("top").Int()))
		//responder.Mouse(x, y, PRESS)
		log.Println("Mouse:", x, y)
	}, false)

	canvas.Call("addEventListener", "mouseup", func(ev *js.Object) {
		rect := canvas.Call("getBoundingClientRect")
		x := float32((ev.Get("clientX").Int() - rect.Get("left").Int()))
		y := float32((ev.Get("clientY").Int() - rect.Get("top").Int()))
		//responder.Mouse(x, y, RELEASE)
		log.Println("Mouse:", x, y)
	}, false)

	canvas.Call("addEventListener", "touchstart", func(ev *js.Object) {
		rect := canvas.Call("getBoundingClientRect")
		for i := 0; i < ev.Get("changedTouches").Get("length").Int(); i++ {
			touch := ev.Get("changedTouches").Index(i)
			x := float32((touch.Get("clientX").Int() - rect.Get("left").Int()))
			y := float32((touch.Get("clientY").Int() - rect.Get("top").Int()))
			//responder.Mouse(x, y, PRESS)
			log.Println("Mouse:", x, y)
		}
	}, false)

	canvas.Call("addEventListener", "touchcancel", func(ev *js.Object) {
		rect := canvas.Call("getBoundingClientRect")
		for i := 0; i < ev.Get("changedTouches").Get("length").Int(); i++ {
			touch := ev.Get("changedTouches").Index(i)
			x := float32((touch.Get("clientX").Int() - rect.Get("left").Int()))
			y := float32((touch.Get("clientY").Int() - rect.Get("top").Int()))
			//responder.Mouse(x, y, RELEASE)
			log.Println("Mouse:", x, y)
		}
	}, false)

	canvas.Call("addEventListener", "touchend", func(ev *js.Object) {
		rect := canvas.Call("getBoundingClientRect")
		for i := 0; i < ev.Get("changedTouches").Get("length").Int(); i++ {
			touch := ev.Get("changedTouches").Index(i)
			x := float32((touch.Get("clientX").Int() - rect.Get("left").Int()))
			y := float32((touch.Get("clientY").Int() - rect.Get("top").Int()))
			//responder.Mouse(x, y, PRESS)
			log.Println("Mouse:", x, y)
		}
	}, false)

	canvas.Call("addEventListener", "touchmove", func(ev *js.Object) {
		rect := canvas.Call("getBoundingClientRect")
		for i := 0; i < ev.Get("changedTouches").Get("length").Int(); i++ {
			touch := ev.Get("changedTouches").Index(i)
			x := float32((touch.Get("clientX").Int() - rect.Get("left").Int()))
			y := float32((touch.Get("clientY").Int() - rect.Get("top").Int()))
			//responder.Mouse(x, y, MOVE)
			log.Println("Mouse:", x, y)
		}
	}, false)

	js.Global.Call("addEventListener", "keypress", func(ev *js.Object) {
		//responder.Type(rune(ev.Get("charCode").Int()))
		log.Println("Keypresss:", rune(ev.Get("charCode").Int()))
	}, false)

	js.Global.Call("addEventListener", "keydown", func(ev *js.Object) {
		key := Key(ev.Get("keyCode").Int())
		keyStates[key] = true
	}, false)

	js.Global.Call("addEventListener", "keyup", func(ev *js.Object) {
		key := Key(ev.Get("keyCode").Int())
		keyStates[key] = false
		// responder.Key(Key(ev.Get("keyCode").Int()), 0, RELEASE)
	}, false)

	// TODO: add events for window resizing?

	Gl.Viewport(0, 0, width, height)
}

func DestroyWindow() {
	// TODO: anything to do here?
}

func WindowWidth() float32 {
	return float32(canvas.Get("width").Int())
}

func WindowHeight() float32 {
	return float32(canvas.Get("height").Int())
}

func toPx(n int) string {
	return strconv.FormatInt(int64(n), 10) + "px"
}

func rafPolyfill() {
	window := js.Global
	vendors := []string{"ms", "moz", "webkit", "o"}
	if window.Get("requestAnimationFrame") == nil {
		for i := 0; i < len(vendors) && window.Get("requestAnimationFrame") == nil; i++ {
			vendor := vendors[i]
			window.Set("requestAnimationFrame", window.Get(vendor+"RequestAnimationFrame"))
			window.Set("cancelAnimationFrame", window.Get(vendor+"CancelAnimationFrame"))
			if window.Get("cancelAnimationFrame") == nil {
				window.Set("cancelAnimationFrame", window.Get(vendor+"CancelRequestAnimationFrame"))
			}
		}
	}

	lastTime := 0.0
	if window.Get("requestAnimationFrame") == nil {
		window.Set("requestAnimationFrame", func(callback func(float32)) int {
			currTime := js.Global.Get("Date").New().Call("getTime").Float()
			timeToCall := math.Max(0, 16-(currTime-lastTime))
			id := window.Call("setTimeout", func() { callback(float32(currTime + timeToCall)) }, timeToCall)
			lastTime = currTime + timeToCall
			return id.Int()
		})
	}

	if window.Get("cancelAnimationFrame") == nil {
		window.Set("cancelAnimationFrame", func(id int) {
			js.Global.Get("clearTimeout").Invoke(id)
		})
	}
}

func RunIteration() {
	// TODO: this may not work, and sky-rocket the FPS
	requestAnimationFrame(func(dt float32) {
		currentWorld.Update(Time.Delta())
		keysUpdate()
		if !headless {
			// TODO: does this require !headless?
			Mouse.ScrollX, Mouse.ScrollY = 0, 0
		}
		Time.Tick()
	})
}

func requestAnimationFrame(callback func(float32)) int {
	return js.Global.Call("requestAnimationFrame", callback).Int()
}

func cancelAnimationFrame(id int) {
	js.Global.Call("cancelAnimationFrame")
}
