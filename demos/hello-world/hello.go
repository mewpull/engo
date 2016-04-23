package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/assets"
	"engo.io/engo/math"
	"engo.io/engo/render"
	"engo.io/engo/space"
	"engo.io/engo/window"
)

type DefaultScene struct{}

type Guy struct {
	ecs.BasicEntity
	render.RenderComponent
	space.SpaceComponent
}

func (*DefaultScene) Preload() {
	// Load all files from the data directory. `false` means: do not do it recursively.
	assets.Files.AddFromDir("data", false)
}

func (*DefaultScene) Setup(w *ecs.World) {
	window.SetBackground(color.White)

	w.AddSystem(&render.RenderSystem{})

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

	// Add it to appropriate systems
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *render.RenderSystem:
			sys.Add(&guy.BasicEntity, &guy.RenderComponent, &guy.SpaceComponent)
		}
	}
}

func (*DefaultScene) Type() string { return "GameWorld" }

func main() {
	opts := engo.RunOptions{
		Title:  "Hello World Demo",
		Width:  1024,
		Height: 640,
	}
	engo.Run(opts, &DefaultScene{})
}
