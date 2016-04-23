package assets // import "engo.io/engo/assets"

import (
	"image"
	"image/draw"
	"io/ioutil"
	"log"
	"os"
	"path"

	"engo.io/engo/assets/audio"
	engoimage "engo.io/engo/image"
	"engo.io/engo/level"
	"engo.io/engo/resource"
	"engo.io/engo/texture"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var (
	Files *Loader
)

type Resource struct {
	kind string
	name string
	url  string
}

// TODO(u): Consider removing resource.Resource interface.
func (res Resource) Kind() string { return res.kind }
func (res Resource) Name() string { return res.name }
func (res Resource) URL() string  { return res.url }

type Loader struct {
	resources []Resource
	images    map[string]*texture.Texture
	jsons     map[string]string
	levels    map[string]*level.Level
	sounds    map[string]string
	fonts     map[string]*truetype.Font
}

func NewLoader() *Loader {
	return &Loader{
		resources: make([]Resource, 1),
		images:    make(map[string]*texture.Texture),
		jsons:     make(map[string]string),
		//levels:    make(map[string]*Level),
		sounds: make(map[string]string),
		fonts:  make(map[string]*truetype.Font),
	}
}

func NewResource(url string) Resource {
	kind := path.Ext(url)
	name := path.Base(url)

	if len(kind) == 0 {
		log.Println("WARNING: Cannot laod extensionless resource.")
		return Resource{}
	}

	return Resource{name: name, url: url, kind: kind[1:]}
}

func (l *Loader) AddFromDir(url string, recurse bool) {
	files, err := ioutil.ReadDir(url)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		furl := url + "/" + f.Name()
		if !f.IsDir() {
			Files.Add(furl)
		} else if recurse {
			Files.AddFromDir(furl, recurse)
		}
	}
}

func (l *Loader) Add(urls ...string) {
	for _, u := range urls {
		r := NewResource(u)
		l.resources = append(l.resources, r)
		log.Println(r)
	}
}

func (l *Loader) Image(name string) *texture.Texture {
	return l.images[name]
}

func (l *Loader) Json(name string) string {
	return l.jsons[name]
}

func (l *Loader) Level(name string) *level.Level {
	return l.levels[name]
}

func (l *Loader) Sound(name string) audio.ReadSeekCloser {
	f, err := os.Open(l.sounds[name])
	if err != nil {
		return nil
	}
	return f
}

func (l *Loader) Font(name string) *truetype.Font {
	return l.fonts[name]
}

func (l *Loader) Load(onFinish func()) {
	for _, r := range l.resources {
		switch r.kind {
		case "png":
			if _, ok := l.images[r.name]; ok {
				continue // with other resources
			}

			data, err := engoimage.LoadImage(r)
			if err != nil {
				log.Println("Error loading resource:", err)
				continue // with other resources
			}

			l.images[r.name] = texture.NewTexture(data)
		case "jpg":
			if _, ok := l.images[r.name]; ok {
				continue // with other resources
			}

			data, err := engoimage.LoadImage(r)
			if err != nil {
				log.Println("Error loading resource:", err)
				continue // with other resources
			}

			l.images[r.name] = texture.NewTexture(data)
		case "json":
			if _, ok := l.jsons[r.name]; ok {
				continue // with other resources
			}

			data, err := loadJSON(r)
			if err != nil {
				log.Println("Error loading resource:", err)
				continue // with other resources
			}

			l.jsons[r.name] = data
		// TODO: Figure out how to handle tmx files without import cycles. Define
		// LoadResource interface?

		/*
			case "tmx":
				if _, ok := l.levels[r.name]; ok {
					continue // with other resources
				}

				data, err := tmx.CreateLevelFromTmx(r)
				if err != nil {
					log.Println("Error loading resource:", err)
					continue // with other resources
				}

				l.levels[r.name] = data
		*/
		case "wav":
			l.sounds[r.name] = r.url

		case "ttf":
			if _, ok := l.fonts[r.name]; ok {
				continue // with other resources
			}

			f, err := loadFont(r)
			if err != nil {
				log.Println("Error loading resource:", err)
				continue // with other resources
			}

			l.fonts[r.name] = f
		}
	}
	onFinish()
}

func ImageToNRGBA(img image.Image, width, height int) *image.NRGBA {
	newm := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.Draw(newm, newm.Bounds(), img, image.Point{0, 0}, draw.Src)

	return newm
}

func loadJSON(r Resource) (string, error) {
	file, err := ioutil.ReadFile(r.url)
	if err != nil {
		return "", err
	}
	return string(file), nil
}

func loadFont(r resource.Resource) (*truetype.Font, error) {
	ttfBytes, err := ioutil.ReadFile(r.URL())
	if err != nil {
		return nil, err
	}

	return freetype.ParseFont(ttfBytes)
}

// TODO(u): Figure out if needed, and if so, whether this code belong here or in
// the image pckage.

/*

type Assets struct {
	queue  []string
	cache  map[string]image.Image
	loads  int
	errors int
}

func NewAssets() *Assets {
	return &Assets{make([]string, 0), make(map[string]image.Image), 0, 0}
}

func (a *Assets) Image(path string) {
	a.queue = append(a.queue, path)
}

func (a *Assets) Get(path string) image.Image {
	return a.cache[path]
}

func (a *Assets) Load(onFinish func()) {
	if len(a.queue) == 0 {
		onFinish()
	} else {
		for _, path := range a.queue {
			img := LoadImage(path)
			a.cache[path] = img
		}
	}
}

func LoadImage(data interface{}) image.Image {
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
*/
