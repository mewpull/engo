package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/assets"
	"engo.io/engo/camera"
	"engo.io/engo/math"
	"engo.io/engo/mouse"
	"engo.io/engo/render"
	"engo.io/engo/space"
	"engo.io/engo/window"
)

type DefaultScene struct{}

type Guy struct {
	ecs.BasicEntity
	mouse.MouseComponent
	render.RenderComponent
	space.SpaceComponent
}

func (*DefaultScene) Preload() {
	// Load all files from the data directory. `false` means: do not do it recursively.
	assets.Files.AddFromDir("data", false)
}

func (*DefaultScene) Setup(w *ecs.World) {
	window.SetBackground(color.White)

	w.AddSystem(&mouse.MouseSystem{})
	w.AddSystem(&render.RenderSystem{})
	w.AddSystem(&ControlSystem{})
	w.AddSystem(&camera.MouseZoomer{-0.125})

	// Retrieve a texture
	texture := assets.Files.Image("icon.png")

	// Create an entity
	guy := Guy{BasicEntity: ecs.NewBasic()}

	// Initialize the components, set scale to 8x
	guy.RenderComponent = render.NewRenderComponent(texture, math.Point{8, 8}, "guy")
	guy.SpaceComponent = space.SpaceComponent{
		Position: math.Point{0, 0},
		Width:    texture.Width() * guy.RenderComponent.Scale().X,
		Height:   texture.Height() * guy.RenderComponent.Scale().Y,
	}
	// guy.MouseComponent doesn't have to be set, because its default values will do

	// Add our guy to appropriate systems
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *render.RenderSystem:
			sys.Add(&guy.BasicEntity, &guy.RenderComponent, &guy.SpaceComponent)
		case *mouse.MouseSystem:
			sys.Add(&guy.BasicEntity, &guy.MouseComponent, &guy.SpaceComponent, &guy.RenderComponent)
		case *ControlSystem:
			sys.Add(&guy.BasicEntity, &guy.MouseComponent)
		}
	}
}

func (*DefaultScene) Type() string { return "GameWorld" }

type controlEntity struct {
	*ecs.BasicEntity
	*mouse.MouseComponent
}

type ControlSystem struct {
	entities []controlEntity
}

func (c *ControlSystem) Add(basic *ecs.BasicEntity, mouse *mouse.MouseComponent) {
	c.entities = append(c.entities, controlEntity{basic, mouse})
}

func (c *ControlSystem) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range c.entities {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		c.entities = append(c.entities[:delete], c.entities[delete+1:]...)
	}
}

func (c *ControlSystem) Update(dt float32) {
	for _, e := range c.entities {
		if e.MouseComponent.Enter {
			window.SetCursor(window.Hand)
		} else if e.MouseComponent.Leave {
			window.SetCursor(nil)
		}
	}
}

func main() {
	opts := engo.RunOptions{
		Title:  "Mouse Demo",
		Width:  1024,
		Height: 640,
	}
	engo.Run(opts, &DefaultScene{})
}
