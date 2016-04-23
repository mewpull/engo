package render // import "engo.io/engo/render"

import (
	"fmt"
	"image/color"
	"sort"

	"engo.io/ecs"
	internalengo "engo.io/engo/internal/engo"
	internalwindow "engo.io/engo/internal/window"
	engomath "engo.io/engo/math"
	"engo.io/engo/message"
	"engo.io/engo/shaders"
	"engo.io/engo/space"
	"engo.io/engo/window"
	"engo.io/gl"
	"github.com/luxengine/math"
)

const (
	RenderSystemPriority = -1000
)

type renderChangeMessage struct{}

func (renderChangeMessage) Type() string {
	return "renderChangeMessage"
}

type Drawable interface {
	Texture() *gl.Texture
	Width() float32
	Height() float32
	View() (float32, float32, float32, float32)
}

type RenderComponent struct {
	// Hidden is used to prevent drawing by OpenGL
	Hidden bool

	// Transparency is the level of transparency that is used to draw the texture
	Transparency float32

	scale engomath.Point
	Color color.Color
	// TODO(u): Unexport Shader.
	Shader shaders.Shader
	zIndex float32

	drawable      Drawable
	buffer        *gl.Buffer
	bufferContent []float32
}

func NewRenderComponent(d Drawable, scale engomath.Point, label string) RenderComponent {
	rc := RenderComponent{
		Transparency: 1,
		Color:        color.White,

		scale: scale,
	}
	rc.SetDrawable(d)

	return rc
}

func (r *RenderComponent) SetDrawable(d Drawable) {
	r.drawable = d
	r.preloadTexture()
}

func (r *RenderComponent) Drawable() Drawable {
	return r.drawable
}

func (r *RenderComponent) SetScale(scale engomath.Point) {
	r.scale = scale
	r.preloadTexture()
}

func (r *RenderComponent) Scale() engomath.Point {
	return r.scale
}

func (r *RenderComponent) SetShader(s shaders.Shader) {
	r.Shader = s
	message.Mailbox.Dispatch(&renderChangeMessage{})
}

func (r *RenderComponent) SetZIndex(index float32) {
	r.zIndex = index
	message.Mailbox.Dispatch(&renderChangeMessage{})
}

// Init is called to initialize the RenderElement
func (ren *RenderComponent) preloadTexture() {
	if ren.drawable == nil || internalengo.Headless {
		return
	}

	ren.bufferContent = ren.generateBufferContent()

	ren.buffer = internalwindow.Gl.CreateBuffer()
	internalwindow.Gl.BindBuffer(internalwindow.Gl.ARRAY_BUFFER, ren.buffer)
	internalwindow.Gl.BufferData(internalwindow.Gl.ARRAY_BUFFER, ren.bufferContent, internalwindow.Gl.STATIC_DRAW)
}

// generateBufferContent computes information about the 4 vertices needed to draw the texture, which should
// be stored in the buffer
func (ren *RenderComponent) generateBufferContent() []float32 {
	scaleX := ren.scale.X
	scaleY := ren.scale.Y
	rotation := float32(0.0)
	transparency := float32(1.0)
	c := ren.Color

	fx := float32(0)
	fy := float32(0)
	fx2 := ren.drawable.Width()
	fy2 := ren.drawable.Height()

	if scaleX != 1 || scaleY != 1 {
		//fx *= scaleX
		//fy *= scaleY
		fx2 *= scaleX
		fy2 *= scaleY
	}

	p1x := fx
	p1y := fy
	p2x := fx
	p2y := fy2
	p3x := fx2
	p3y := fy2
	p4x := fx2
	p4y := fy

	var x1 float32
	var y1 float32
	var x2 float32
	var y2 float32
	var x3 float32
	var y3 float32
	var x4 float32
	var y4 float32

	if rotation != 0 {
		rot := rotation * (math.Pi / 180.0)

		cos := math.Cos(rot)
		sin := math.Sin(rot)

		x1 = cos*p1x - sin*p1y
		y1 = sin*p1x + cos*p1y

		x2 = cos*p2x - sin*p2y
		y2 = sin*p2x + cos*p2y

		x3 = cos*p3x - sin*p3y
		y3 = sin*p3x + cos*p3y

		x4 = x1 + (x3 - x2)
		y4 = y3 - (y2 - y1)
	} else {
		x1 = p1x
		y1 = p1y

		x2 = p2x
		y2 = p2y

		x3 = p3x
		y3 = p3y

		x4 = p4x
		y4 = p4y
	}

	colorR, colorG, colorB, _ := c.RGBA()

	red := colorR
	green := colorG << 8
	blue := colorB << 16
	alpha := uint32(transparency*255.0) << 24

	tint := math.Float32frombits((alpha | blue | green | red) & 0xfeffffff)

	u, v, u2, v2 := ren.drawable.View()

	return []float32{x1, y1, u, v, tint, x4, y4, u2, v, tint, x3, y3, u2, v2, tint, x2, y2, u, v2, tint}
}

type renderEntity struct {
	*ecs.BasicEntity
	*RenderComponent
	*space.SpaceComponent
}

type renderEntityList []renderEntity

func (r renderEntityList) Len() int {
	return len(r)
}

func (r renderEntityList) Less(i, j int) bool {
	// Sort by shader-pointer if they have the same zIndex
	if r[i].RenderComponent.zIndex == r[j].RenderComponent.zIndex {
		// TODO: optimize this for performance
		return fmt.Sprintf("%p", r[i].RenderComponent.Shader) < fmt.Sprintf("%p", r[j].RenderComponent.Shader)
	}

	return r[i].RenderComponent.zIndex < r[j].RenderComponent.zIndex
}

func (r renderEntityList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

type RenderSystem struct {
	entities renderEntityList
	world    *ecs.World

	sortingNeeded bool
	currentShader shaders.Shader
}

func (*RenderSystem) Priority() int { return RenderSystemPriority }

func (rs *RenderSystem) New(w *ecs.World) {
	rs.world = w

	if !internalengo.Headless {
		shaders.InitShaders(window.Width(), window.Height())
	}

	message.Mailbox.Listen("renderChangeMessage", func(message.Message) {
		rs.sortingNeeded = true
	})
}

func (rs *RenderSystem) Add(basic *ecs.BasicEntity, render *RenderComponent, space *space.SpaceComponent) {
	rs.entities = append(rs.entities, renderEntity{basic, render, space})
	rs.sortingNeeded = true
}

func (rs *RenderSystem) Remove(basic ecs.BasicEntity) {
	var delete int = -1
	for index, entity := range rs.entities {
		if entity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		rs.entities = append(rs.entities[:delete], rs.entities[delete+1:]...)
		rs.sortingNeeded = true
	}
}

func (rs *RenderSystem) Update(dt float32) {
	if internalengo.Headless {
		return
	}

	if rs.sortingNeeded {
		sort.Sort(rs.entities)
		rs.sortingNeeded = false
	}

	internalwindow.Gl.Clear(internalwindow.Gl.COLOR_BUFFER_BIT)

	// TODO: it's linear for now, but that might very well be a bad idea
	for _, e := range rs.entities {
		if e.RenderComponent.Hidden {
			continue // with other entities
		}

		// Retrieve a shader, may be the default one -- then use it if we aren't already using it
		shader := e.RenderComponent.Shader
		if shader == nil {
			shader = shaders.DefaultShader
		}

		// Change Shader if we have to
		if shader != rs.currentShader {
			if rs.currentShader != nil {
				rs.currentShader.Post()
			}
			shader.Pre()
			rs.currentShader = shader
		}

		rs.currentShader.Draw(e.RenderComponent.drawable.Texture(), e.RenderComponent.buffer, e.SpaceComponent.Position.X, e.SpaceComponent.Position.Y, 0) // TODO: add rotation
	}

	if rs.currentShader != nil {
		rs.currentShader.Post()
		rs.currentShader = nil
	}
}
