package engo

import (
	"math/rand"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
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

	err := <-ch
	if err != nil {
		return nil, err
	}

	return &ImageObject{img}, nil
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

	err := <-ch
	if err != nil {
		return "", err
	}

	return req.Get("responseText").String(), nil
}

type ImageObject struct {
	data *js.Object
}

func (i *ImageObject) Data() interface{} {
	return i.data
}

func (i *ImageObject) Width() int {
	return i.data.Get("width").Int()
}

func (i *ImageObject) Height() int {
	return i.data.Get("height").Int()
}
