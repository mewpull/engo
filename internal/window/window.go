package window // import "engo.io/engo/internal/window"

import (
	"engo.io/engo/space"
	"engo.io/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

var (
	Window        *glfw.Window
	WorldBounds   space.AABB
	Gl            *gl.Context
	Vsync         = true
	ScaleOnResize = false
)
