package image // import "engo.io/engo/image"

import (
	"image"
	"image/draw"
	"os"

	"engo.io/engo/resource"
)

type Image interface {
	Data() interface{}
	Width() int
	Height() int
}

func NewImageRGBA(img *image.RGBA) *ImageRGBA {
	return &ImageRGBA{img}
}

type ImageRGBA struct {
	data *image.RGBA
}

func (i *ImageRGBA) Data() interface{} {
	return i.data
}

func (i *ImageRGBA) Width() int {
	return i.data.Rect.Max.X
}

func (i *ImageRGBA) Height() int {
	return i.data.Rect.Max.Y
}

func NewImageObject(img *image.NRGBA) *ImageObject {
	return &ImageObject{img}
}

type ImageObject struct {
	data *image.NRGBA
}

func (i *ImageObject) Data() interface{} {
	return i.data
}

func (i *ImageObject) Width() int {
	return i.data.Rect.Max.X
}

func (i *ImageObject) Height() int {
	return i.data.Rect.Max.Y
}

func LoadImage(r resource.Resource) (Image, error) {
	file, err := os.Open(r.URL())
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	b := img.Bounds()
	newm := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(newm, newm.Bounds(), img, b.Min, draw.Src)

	return &ImageObject{newm}, nil
}
