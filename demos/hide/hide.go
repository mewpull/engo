package main

import (
	"image/color"
	"math/rand"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/assets"
	"engo.io/engo/math"
	"engo.io/engo/render"
	"engo.io/engo/space"
	"engo.io/engo/window"
)

type DefaultScene struct{}

type Rock struct {
	ecs.BasicEntity
	render.RenderComponent
	space.SpaceComponent
}

func (*DefaultScene) Preload() {
	assets.Files.AddFromDir("assets", false)
}

func (*DefaultScene) Setup(w *ecs.World) {
	window.SetBackground(color.White)

	w.AddSystem(&render.RenderSystem{})
	w.AddSystem(&HideSystem{})

	// Retrieve a texture
	texture := assets.Files.Image("rock.png")

	// Create an entity
	rock := Rock{BasicEntity: ecs.NewBasic()}

	// Initialize the components, set scale to 8x
	rock.RenderComponent = render.NewRenderComponent(texture, math.Point{8, 8}, "rock")
	rock.SpaceComponent = space.SpaceComponent{
		Position: math.Point{0, 0},
		Width:    texture.Width() * rock.RenderComponent.Scale().X,
		Height:   texture.Height() * rock.RenderComponent.Scale().Y,
	}

	// Add it to appropriate systems
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *render.RenderSystem:
			sys.Add(&rock.BasicEntity, &rock.RenderComponent, &rock.SpaceComponent)
		case *HideSystem:
			sys.Add(&rock.BasicEntity, &rock.RenderComponent)
		}
	}
}

func (*DefaultScene) Type() string { return "GameWorld" }

type hideEntity struct {
	*ecs.BasicEntity
	*render.RenderComponent
}

type HideSystem struct {
	entities []hideEntity
}

func (h *HideSystem) Add(basic *ecs.BasicEntity, render *render.RenderComponent) {
	h.entities = append(h.entities, hideEntity{basic, render})
}

func (h *HideSystem) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range h.entities {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		h.entities = append(h.entities[:delete], h.entities[delete+1:]...)
	}
}

func (h *HideSystem) Update(dt float32) {
	for _, e := range h.entities {
		if rand.Int()%10 == 0 {
			e.RenderComponent.Hidden = true
		} else {
			e.RenderComponent.Hidden = false
		}
	}
}

func main() {
	opts := engo.RunOptions{
		Title:  "Show and Hide Demo",
		Width:  1024,
		Height: 640,
	}
	engo.Run(opts, &DefaultScene{})
}
