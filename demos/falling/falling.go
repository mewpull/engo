package main

import (
	"image/color"
	"log"
	"math/rand"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/assets"
	"engo.io/engo/collision"
	"engo.io/engo/input"
	"engo.io/engo/math"
	"engo.io/engo/message"
	"engo.io/engo/render"
	"engo.io/engo/space"
	"engo.io/engo/window"
)

type Guy struct {
	ecs.BasicEntity
	collision.CollisionComponent
	render.RenderComponent
	space.SpaceComponent
}

type Rock struct {
	ecs.BasicEntity
	collision.CollisionComponent
	render.RenderComponent
	space.SpaceComponent
}

type DefaultScene struct{}

func (*DefaultScene) Preload() {
	// Add all the files in the data directory non recursively
	assets.Files.AddFromDir("data", false)
}

func (*DefaultScene) Setup(w *ecs.World) {
	window.SetBackground(color.White)

	// Add all of the systems
	w.AddSystem(&render.RenderSystem{})
	w.AddSystem(&collision.CollisionSystem{})
	w.AddSystem(&DeathSystem{})
	w.AddSystem(&FallingSystem{})
	w.AddSystem(&ControlSystem{})
	w.AddSystem(&RockSpawnSystem{})

	texture := assets.Files.Image("icon.png")

	// Create an entity
	guy := Guy{BasicEntity: ecs.NewBasic()}

	// Initialize the components, set scale to 4x
	guy.RenderComponent = render.NewRenderComponent(texture, math.Point{4, 4}, "guy")
	guy.SpaceComponent = space.SpaceComponent{
		Position: math.Point{0, 0},
		Width:    texture.Width() * guy.RenderComponent.Scale().X,
		Height:   texture.Height() * guy.RenderComponent.Scale().Y,
	}
	guy.CollisionComponent = collision.CollisionComponent{Solid: true, Main: true}

	// Add it to appropriate systems
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *render.RenderSystem:
			sys.Add(&guy.BasicEntity, &guy.RenderComponent, &guy.SpaceComponent)
		case *collision.CollisionSystem:
			sys.Add(&guy.BasicEntity, &guy.CollisionComponent, &guy.SpaceComponent)
		case *ControlSystem:
			sys.Add(&guy.BasicEntity, &guy.SpaceComponent)
		}
	}
}

func (*DefaultScene) Type() string { return "Game" }

type controlEntity struct {
	*ecs.BasicEntity
	*space.SpaceComponent
}

type ControlSystem struct {
	entities []controlEntity
}

func (c *ControlSystem) Add(basic *ecs.BasicEntity, space *space.SpaceComponent) {
	c.entities = append(c.entities, controlEntity{basic, space})
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
	speed := 400 * dt

	for _, e := range c.entities {
		if input.Keys.Get(input.A).Down() {
			e.SpaceComponent.Position.X -= speed
		}

		if input.Keys.Get(input.D).Down() {
			e.SpaceComponent.Position.X += speed
		}

		if input.Keys.Get(input.W).Down() {
			e.SpaceComponent.Position.Y -= speed
		}

		if input.Keys.Get(input.S).Down() {
			e.SpaceComponent.Position.Y += speed
		}
	}
}

type RockSpawnSystem struct {
	world *ecs.World
}

func (rock *RockSpawnSystem) New(w *ecs.World) {
	rock.world = w
}

func (*RockSpawnSystem) Remove(ecs.BasicEntity) {}

func (rock *RockSpawnSystem) Update(dt float32) {
	// 4% change of spawning a rock each frame
	if rand.Float32() < .96 {
		return
	}

	position := math.Point{
		X: rand.Float32() * window.Width(),
		Y: -32,
	}
	NewRock(rock.world, position)
}

func NewRock(world *ecs.World, position math.Point) {
	texture := assets.Files.Image("rock.png")

	rock := Rock{BasicEntity: ecs.NewBasic()}
	rock.RenderComponent = render.NewRenderComponent(texture, math.Point{4, 4}, "rock")
	rock.SpaceComponent = space.SpaceComponent{
		Position: position,
		Width:    texture.Width() * rock.RenderComponent.Scale().X,
		Height:   texture.Height() * rock.RenderComponent.Scale().Y,
	}
	rock.CollisionComponent = collision.CollisionComponent{Solid: true}

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *render.RenderSystem:
			sys.Add(&rock.BasicEntity, &rock.RenderComponent, &rock.SpaceComponent)
		case *collision.CollisionSystem:
			sys.Add(&rock.BasicEntity, &rock.CollisionComponent, &rock.SpaceComponent)
		case *FallingSystem:
			sys.Add(&rock.BasicEntity, &rock.SpaceComponent)
		}
	}
}

type fallingEntity struct {
	*ecs.BasicEntity
	*space.SpaceComponent
}

type FallingSystem struct {
	entities []fallingEntity
}

func (f *FallingSystem) Add(basic *ecs.BasicEntity, space *space.SpaceComponent) {
	f.entities = append(f.entities, fallingEntity{basic, space})
}

func (f *FallingSystem) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range f.entities {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		f.entities = append(f.entities[:delete], f.entities[delete+1:]...)
	}
}

func (f *FallingSystem) Update(dt float32) {
	for _, e := range f.entities {
		e.SpaceComponent.Position.Y += 200 * dt
	}
}

type DeathSystem struct{}

func (*DeathSystem) New(*ecs.World) {
	// Subscribe to ScoreMessage
	message.Mailbox.Listen("CollisionMessage", func(message message.Message) {
		_, isCollision := message.(collision.CollisionMessage)
		if isCollision {
			log.Println("DEAD")
		}
	})
}

func (*DeathSystem) Remove(ecs.BasicEntity) {}
func (*DeathSystem) Update(dt float32)      {}

func main() {
	opts := engo.RunOptions{
		Title:  "Falling Demo",
		Width:  1024,
		Height: 640,
	}
	engo.Run(opts, &DefaultScene{})
}
