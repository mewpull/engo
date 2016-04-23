package region // import "engo.io/engo/region"

import (
	"engo.io/engo/texture"
	"engo.io/gl"
	"github.com/luxengine/math"
)

type Region struct {
	texture       *texture.Texture
	u, v          float32
	u2, v2        float32
	width, height float32
}

func NewRegion(tex *texture.Texture, x, y, w, h float32) *Region {
	invTexWidth := 1.0 / tex.Width()
	invTexHeight := 1.0 / tex.Height()

	u := x * invTexWidth
	v := y * invTexHeight
	u2 := (x + w) * invTexWidth
	v2 := (y + h) * invTexHeight

	width := math.Abs(w)
	height := math.Abs(h)

	return &Region{tex, u, v, u2, v2, width, height}
}

func (r *Region) Width() float32 {
	return float32(r.width)
}

func (r *Region) Height() float32 {
	return float32(r.height)
}

func (r *Region) Texture() *gl.Texture {
	return r.texture.Texture()
}

func (r *Region) View() (float32, float32, float32, float32) {
	return r.u, r.v, r.u2, r.v2
}

// Works for tiles rendered right-down
func RegionFromSheet(sheet *texture.Texture, tw, th int, index int) *Region {
	setWidth := int(sheet.Width()) / tw
	x := (index % setWidth) * tw
	y := (index / setWidth) * th
	return NewRegion(sheet, float32(x), float32(y), float32(tw), float32(th))
}
