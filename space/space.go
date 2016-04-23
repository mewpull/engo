package space // import "engo.io/engo/space"

import (
	engomath "engo.io/engo/math"
)

type AABB struct {
	Min, Max engomath.Point
}

type SpaceComponent struct {
	Position engomath.Point
	Width    float32
	Height   float32
}

// Center positions the space component according to its center instead of its
// top-left point (this avoids doing the same math each time in your systems)
func (sc *SpaceComponent) Center(p engomath.Point) {
	xDelta := sc.Width / 2
	yDelta := sc.Height / 2
	// update position according to point being used as our center
	sc.Position.X = p.X - xDelta
	sc.Position.Y = p.Y - yDelta
}

func (sc SpaceComponent) AABB() AABB {
	return AABB{Min: sc.Position, Max: engomath.Point{sc.Position.X + sc.Width, sc.Position.Y + sc.Height}}
}
