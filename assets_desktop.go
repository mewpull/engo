// +build !netgo,!android

package engo

import (
	"image"
	"image/draw"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func loadImage(r Resource) (Image, error) {
	file, err := os.Open(r.url)
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

func loadJSON(r Resource) (string, error) {
	file, err := ioutil.ReadFile(r.url)
	if err != nil {
		return "", err
	}
	return string(file), nil
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

func NewImageObjectFromNRGBA(i *image.NRGBA) *ImageObject {
	return &ImageObject{i}
}

func LoadImage(data interface{}) Image {
	var m image.Image

	switch data := data.(type) {
	default:
		log.Fatal("NewTexture needs a string or io.Reader")
	case string:
		file, err := os.Open(data)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		img, _, err := image.Decode(file)
		if err != nil {
			log.Fatal(err)
		}
		m = img
	case io.Reader:
		img, _, err := image.Decode(data)
		if err != nil {
			log.Fatal(err)
		}
		m = img
	case image.Image:
		m = data
	}

	b := m.Bounds()
	newm := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(newm, newm.Bounds(), m, b.Min, draw.Src)

	return &ImageObject{newm}
}
