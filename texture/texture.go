package texture // import "engo.io/engo/texture"

import (
	"engo.io/engo/image"
	internalengo "engo.io/engo/internal/engo"
	internalwindow "engo.io/engo/internal/window"
	"engo.io/gl"
)

type Texture struct {
	id     *gl.Texture
	width  float32
	height float32
}

func NewTexture(img image.Image) *Texture {
	var id *gl.Texture
	if !internalengo.Headless {
		id = internalwindow.Gl.CreateTexture()

		internalwindow.Gl.BindTexture(internalwindow.Gl.TEXTURE_2D, id)

		internalwindow.Gl.TexParameteri(internalwindow.Gl.TEXTURE_2D, internalwindow.Gl.TEXTURE_WRAP_S, internalwindow.Gl.CLAMP_TO_EDGE)
		internalwindow.Gl.TexParameteri(internalwindow.Gl.TEXTURE_2D, internalwindow.Gl.TEXTURE_WRAP_T, internalwindow.Gl.CLAMP_TO_EDGE)
		internalwindow.Gl.TexParameteri(internalwindow.Gl.TEXTURE_2D, internalwindow.Gl.TEXTURE_MIN_FILTER, internalwindow.Gl.LINEAR)
		internalwindow.Gl.TexParameteri(internalwindow.Gl.TEXTURE_2D, internalwindow.Gl.TEXTURE_MAG_FILTER, internalwindow.Gl.NEAREST)

		if img.Data() == nil {
			panic("Texture image data is nil.")
		}

		internalwindow.Gl.TexImage2D(internalwindow.Gl.TEXTURE_2D, 0, internalwindow.Gl.RGBA, internalwindow.Gl.RGBA, internalwindow.Gl.UNSIGNED_BYTE, img.Data())
	}

	return &Texture{id, float32(img.Width()), float32(img.Height())}
}

// Width returns the width of the texture.
func (t *Texture) Width() float32 {
	return t.width
}

// Height returns the height of the texture.
func (t *Texture) Height() float32 {
	return t.height
}

func (t *Texture) Texture() *gl.Texture {
	return t.id
}

func (r *Texture) View() (float32, float32, float32, float32) {
	return 0.0, 0.0, 1.0, 1.0
}
