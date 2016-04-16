//+build netgo

package engo

import (
	"math/rand"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
	"image"
)

func loadImage(r Resource) (Image, error) {
	ch := make(chan error, 1)

	img := js.Global.Get("Image").New()
	img.Call("addEventListener", "load", func(*js.Object) {
		go func() { ch <- nil }()
	}, false)
	img.Call("addEventListener", "error", func(o *js.Object) {
		go func() { ch <- &js.Error{Object: o} }()
	}, false)
	img.Set("src", r.url+"?"+strconv.FormatInt(rand.Int63(), 10))

	// TODO: I don't see how this can work. For starters, we're not closing the channel. And, ch is always nil the first time
	err := <-ch
	if err != nil {
		return nil, err
	}

	return &HTMLImageObject{img}, nil
}

func loadJSON(r Resource) (string, error) {
	ch := make(chan error, 1)

	req := js.Global.Get("XMLHttpRequest").New()
	req.Call("open", "GET", r.url, true)
	req.Call("addEventListener", "load", func(*js.Object) {
		go func() { ch <- nil }()
	}, false)
	req.Call("addEventListener", "error", func(o *js.Object) {
		go func() { ch <- &js.Error{Object: o} }()
	}, false)
	req.Call("send", nil)

	// TODO: I don't see how this can work. For starters, we're not closing the channel. And, ch is always nil the first time
	err := <-ch
	if err != nil {
		return "", err
	}

	return req.Get("responseText").String(), nil
}

type HTMLImageObject struct {
	data *js.Object
}

func (i *HTMLImageObject) Data() interface{} {
	return i.data
}

func (i *HTMLImageObject) Width() int {
	return i.data.Get("width").Int()
}

func (i *HTMLImageObject) Height() int {
	return i.data.Get("height").Int()
}

type ImageObject struct {
	data   []uint8
	width  int
	height int
}

func (i *ImageObject) Data() interface{} {
	return i.data
}

func (i *ImageObject) Width() int {
	return i.width
}

func (i *ImageObject) Height() int {
	return i.height
}

func NewImageObjectFromNRGBA(i *image.NRGBA) *ImageObject {
	/*
		// Create a PNG of the NRGBA
		var rawPNG bytes.Buffer
		err := png.Encode(&rawPNG, i)
		if err != nil {
			log.Println("Unable to encode png:", err)
			return nil
		}

		// Base64 encode it
		var basePNG bytes.Buffer
		encoder := base64.NewEncoder(base64.RawStdEncoding, &basePNG)
		_, err = encoder.Write(rawPNG.Bytes())
		if err != nil {
			log.Println("Unable to base64 encode png:", err)
			return nil
		}

		// Save it as an image
		img := js.Global.Get("Image").New()
		img.Set("src", "data:image/png;base64,"+basePNG.String())

		return &ImageObject{img}*/
	return &ImageObject{i.Pix, i.Bounds().Dx(), i.Bounds().Dy()}
}
