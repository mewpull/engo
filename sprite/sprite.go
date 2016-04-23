package sprite // import "engo.io/engo/sprite"

import (
	"image/color"

	"engo.io/engo/math"
	"engo.io/engo/region"
)

type Sprite struct {
	Position *math.Point
	Scale    *math.Point
	Anchor   *math.Point
	Rotation float32
	Color    color.Color
	Alpha    float32
	Region   *region.Region
}

func NewSprite(region *region.Region, x, y float32) *Sprite {
	return &Sprite{
		Position: &math.Point{x, y},
		Scale:    &math.Point{1, 1},
		Anchor:   &math.Point{0, 0},
		Rotation: 0,
		Color:    color.White,
		Alpha:    1,
		Region:   region,
	}
}
