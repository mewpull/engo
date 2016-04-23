package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"sync"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/assets"
	"engo.io/engo/collision"
	"engo.io/engo/font"
	"engo.io/engo/input"
	"engo.io/engo/math"
	"engo.io/engo/message"
	"engo.io/engo/render"
	"engo.io/engo/space"
	"engo.io/engo/window"
)

type PongGame struct{}

var (
	basicFont *font.Font
)

type Ball struct {
	ecs.BasicEntity
	render.RenderComponent
	space.SpaceComponent
	collision.CollisionComponent
	SpeedComponent
}

type Score struct {
	ecs.BasicEntity
	render.RenderComponent
	space.SpaceComponent
}

type Paddle struct {
	ecs.BasicEntity
	ControlComponent
	collision.CollisionComponent
	render.RenderComponent
	space.SpaceComponent
}

func (pong *PongGame) Preload() {
	assets.Files.AddFromDir("assets", true)
}

func (pong *PongGame) Setup(w *ecs.World) {
	window.SetBackground(color.Black)
	w.AddSystem(&render.RenderSystem{})
	w.AddSystem(&collision.CollisionSystem{})
	w.AddSystem(&SpeedSystem{})
	w.AddSystem(&ControlSystem{})
	w.AddSystem(&BallSystem{})
	w.AddSystem(&ScoreSystem{})

	basicFont = (&font.Font{URL: "Roboto-Regular.ttf", Size: 32, FG: color.NRGBA{255, 255, 255, 255}})
	basicFont.TTF = assets.Files.Font(basicFont.URL)

	ballTexture := assets.Files.Image("ball.png")

	ball := Ball{BasicEntity: ecs.NewBasic()}
	ball.RenderComponent = render.NewRenderComponent(ballTexture, math.Point{2, 2}, "ball")
	ball.SpaceComponent = space.SpaceComponent{
		Position: math.Point{(window.Width() - ballTexture.Width()) / 2, (window.Height() - ballTexture.Height()) / 2},
		Width:    ballTexture.Width() * ball.RenderComponent.Scale().X,
		Height:   ballTexture.Height() * ball.RenderComponent.Scale().Y,
	}
	ball.CollisionComponent = collision.CollisionComponent{Main: true, Solid: true}
	ball.SpeedComponent = SpeedComponent{Point: math.Point{300, 1000}}

	// Add our entity to the appropriate systems
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *render.RenderSystem:
			sys.Add(&ball.BasicEntity, &ball.RenderComponent, &ball.SpaceComponent)
		case *collision.CollisionSystem:
			sys.Add(&ball.BasicEntity, &ball.CollisionComponent, &ball.SpaceComponent)
		case *SpeedSystem:
			sys.Add(&ball.BasicEntity, &ball.SpeedComponent, &ball.SpaceComponent)
		case *BallSystem:
			sys.Add(&ball.BasicEntity, &ball.SpeedComponent, &ball.SpaceComponent)
		}
	}

	score := Score{BasicEntity: ecs.NewBasic()}
	score.RenderComponent = render.NewRenderComponent(basicFont.Render(" "), math.Point{1, 1}, "YOLO <3")
	score.SpaceComponent = space.SpaceComponent{math.Point{100, 100}, 100, 100}

	// Add our entity to the appropriate systems
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *render.RenderSystem:
			sys.Add(&score.BasicEntity, &score.RenderComponent, &score.SpaceComponent)
		case *ScoreSystem:
			sys.Add(&score.BasicEntity, &score.RenderComponent, &score.SpaceComponent)
		}
	}

	schemes := []string{"WASD", ""}
	paddleTexture := assets.Files.Image("paddle.png")

	for i := 0; i < 2; i++ {
		paddle := Paddle{BasicEntity: ecs.NewBasic()}
		paddle.RenderComponent = render.NewRenderComponent(paddleTexture, math.Point{2, 2}, "paddle")

		x := float32(0)
		if i != 0 {
			x = 800 - 16
		}

		paddle.SpaceComponent = space.SpaceComponent{
			Position: math.Point{x, (window.Height() - paddleTexture.Height()) / 2},
			Width:    paddle.RenderComponent.Scale().X * paddleTexture.Width(),
			Height:   paddle.RenderComponent.Scale().Y * paddleTexture.Height(),
		}
		paddle.ControlComponent = ControlComponent{schemes[i]}
		paddle.CollisionComponent = collision.CollisionComponent{Main: false, Solid: true}

		// Add our entity to the appropriate systems
		for _, system := range w.Systems() {
			switch sys := system.(type) {
			case *render.RenderSystem:
				sys.Add(&paddle.BasicEntity, &paddle.RenderComponent, &paddle.SpaceComponent)
			case *collision.CollisionSystem:
				sys.Add(&paddle.BasicEntity, &paddle.CollisionComponent, &paddle.SpaceComponent)
			case *ControlSystem:
				sys.Add(&paddle.BasicEntity, &paddle.ControlComponent, &paddle.SpaceComponent)
			}
		}
	}
}

func (*PongGame) Type() string { return "PongGame" }

type SpeedComponent struct {
	math.Point
}

type ControlComponent struct {
	Scheme string
}

type speedEntity struct {
	*ecs.BasicEntity
	*SpeedComponent
	*space.SpaceComponent
}

type SpeedSystem struct {
	entities []speedEntity
}

func (s *SpeedSystem) New(*ecs.World) {
	message.Mailbox.Listen("CollisionMessage", func(message message.Message) {
		log.Println("collision")

		collision, isCollision := message.(collision.CollisionMessage)
		if isCollision {
			// See if we also have that Entity, and if so, change the speed
			for _, e := range s.entities {
				if e.ID() == collision.Entity.BasicEntity.ID() {
					e.SpeedComponent.X *= -1
				}
			}
		}
	})
}

func (s *SpeedSystem) Add(basic *ecs.BasicEntity, speed *SpeedComponent, space *space.SpaceComponent) {
	s.entities = append(s.entities, speedEntity{basic, speed, space})
}

func (s *SpeedSystem) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range s.entities {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		s.entities = append(s.entities[:delete], s.entities[delete+1:]...)
	}
}

func (s *SpeedSystem) Update(dt float32) {
	for _, e := range s.entities {
		e.SpaceComponent.Position.X += e.SpeedComponent.X * dt
		e.SpaceComponent.Position.Y += e.SpeedComponent.Y * dt
	}
}

type ballEntity struct {
	*ecs.BasicEntity
	*SpeedComponent
	*space.SpaceComponent
}

type BallSystem struct {
	entities []ballEntity
}

func (b *BallSystem) Add(basic *ecs.BasicEntity, speed *SpeedComponent, space *space.SpaceComponent) {
	b.entities = append(b.entities, ballEntity{basic, speed, space})
}

func (b *BallSystem) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range b.entities {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		b.entities = append(b.entities[:delete], b.entities[delete+1:]...)
	}
}

func (b *BallSystem) Update(dt float32) {
	for _, e := range b.entities {
		if e.SpaceComponent.Position.X < 0 {
			message.Mailbox.Dispatch(ScoreMessage{1})

			e.SpaceComponent.Position.X = 400 - 16
			e.SpaceComponent.Position.Y = 400 - 16
			e.SpeedComponent.X = 800 * rand.Float32()
			e.SpeedComponent.Y = 800 * rand.Float32()
		}

		if e.SpaceComponent.Position.Y < 0 {
			e.SpaceComponent.Position.Y = 0
			e.SpeedComponent.Y *= -1
		}

		if e.SpaceComponent.Position.X > (800 - 16) {
			message.Mailbox.Dispatch(ScoreMessage{2})

			e.SpaceComponent.Position.X = 400 - 16
			e.SpaceComponent.Position.Y = 400 - 16
			e.SpeedComponent.X = 800 * rand.Float32()
			e.SpeedComponent.Y = 800 * rand.Float32()
		}

		if e.SpaceComponent.Position.Y > (800 - 16) {
			e.SpaceComponent.Position.Y = 800 - 16
			e.SpeedComponent.Y *= -1
		}
	}
}

type controlEntity struct {
	*ecs.BasicEntity
	*ControlComponent
	*space.SpaceComponent
}

type ControlSystem struct {
	entities []controlEntity
}

func (c *ControlSystem) Add(basic *ecs.BasicEntity, control *ControlComponent, space *space.SpaceComponent) {
	c.entities = append(c.entities, controlEntity{basic, control, space})
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
		up := false
		down := false
		if e.ControlComponent.Scheme == "WASD" {
			up = input.Keys.Get(input.W).Down()
			down = input.Keys.Get(input.S).Down()
		} else {
			up = input.Keys.Get(input.ArrowUp).Down()
			down = input.Keys.Get(input.ArrowDown).Down()
		}

		if up {
			if e.SpaceComponent.Position.Y > 0 {
				e.SpaceComponent.Position.Y -= 800 * dt
			}
		}

		if down {
			if (e.SpaceComponent.Height + e.SpaceComponent.Position.Y) < 800 {
				e.SpaceComponent.Position.Y += 800 * dt
			}
		}
	}
}

type scoreEntity struct {
	*ecs.BasicEntity
	*render.RenderComponent
	*space.SpaceComponent
}

type ScoreSystem struct {
	entities []scoreEntity

	PlayerOneScore, PlayerTwoScore int
	upToDate                       bool
	scoreLock                      sync.RWMutex
}

func (s *ScoreSystem) New(*ecs.World) {
	s.upToDate = true
	message.Mailbox.Listen("ScoreMessage", func(message message.Message) {
		scoreMessage, isScore := message.(ScoreMessage)
		if !isScore {
			return
		}

		s.scoreLock.Lock()
		if scoreMessage.Player != 1 {
			s.PlayerOneScore += 1
		} else {
			s.PlayerTwoScore += 1
		}
		log.Println("The score is now", s.PlayerOneScore, "vs", s.PlayerTwoScore)
		s.upToDate = false
		s.scoreLock.Unlock()
	})
}

func (c *ScoreSystem) Add(basic *ecs.BasicEntity, render *render.RenderComponent, space *space.SpaceComponent) {
	c.entities = append(c.entities, scoreEntity{basic, render, space})
}

func (s *ScoreSystem) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range s.entities {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		s.entities = append(s.entities[:delete], s.entities[delete+1:]...)
	}
}

func (s *ScoreSystem) Update(dt float32) {
	for _, e := range s.entities {
		if !s.upToDate {
			s.scoreLock.RLock()
			label := fmt.Sprintf("%v vs %v", s.PlayerOneScore, s.PlayerTwoScore)
			s.upToDate = true
			s.scoreLock.RUnlock()

			e.RenderComponent.SetDrawable(basicFont.Render(label))
			width := len(label) * 20

			e.SpaceComponent.Position.X = float32(400 - (width / 2))
		}
	}
}

type ScoreMessage struct {
	Player int
}

func (ScoreMessage) Type() string {
	return "ScoreMessage"
}

func main() {
	opts := engo.RunOptions{
		HeadlessMode: true,
	}
	engo.Run(opts, &PongGame{})
}
