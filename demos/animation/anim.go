package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/animation"
	"engo.io/engo/assets"
	"engo.io/engo/camera"
	"engo.io/engo/input"
	"engo.io/engo/math"
	"engo.io/engo/render"
	"engo.io/engo/space"
	"engo.io/engo/spritesheet"
	"engo.io/engo/window"
)

var (
	zoomSpeed   float32 = -0.125
	RunAction   *animation.AnimationAction
	WalkAction  *animation.AnimationAction
	StopAction  *animation.AnimationAction
	SkillAction *animation.AnimationAction
	DieAction   *animation.AnimationAction
	actions     []*animation.AnimationAction
)

type DefaultScene struct{}

type Animation struct {
	ecs.BasicEntity
	animation.AnimationComponent
	render.RenderComponent
	space.SpaceComponent
}

func (*DefaultScene) Preload() {
	assets.Files.Add("assets/hero.png")
	StopAction = &animation.AnimationAction{Name: "stop", Frames: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
	RunAction = &animation.AnimationAction{Name: "run", Frames: []int{16, 17, 18, 19, 20, 21}}
	WalkAction = &animation.AnimationAction{Name: "move", Frames: []int{11, 12, 13, 14, 15}}
	SkillAction = &animation.AnimationAction{Name: "skill", Frames: []int{44, 45, 46, 47, 48, 49, 50, 51, 52, 53}}
	DieAction = &animation.AnimationAction{Name: "die", Frames: []int{28, 29, 30}}
	actions = []*animation.AnimationAction{DieAction, StopAction, WalkAction, RunAction, SkillAction}
}

func (scene *DefaultScene) Setup(w *ecs.World) {
	window.SetBackground(color.White)

	w.AddSystem(&render.RenderSystem{})
	w.AddSystem(&animation.AnimationSystem{})
	w.AddSystem(&ControlSystem{})
	w.AddSystem(&camera.MouseZoomer{zoomSpeed})

	spriteSheet := spritesheet.NewSpritesheetFromFile("hero.png", 150, 150)

	hero := scene.CreateEntity(&math.Point{0, 0}, spriteSheet, StopAction)

	// Add our hero to the appropriate systems
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *render.RenderSystem:
			sys.Add(&hero.BasicEntity, &hero.RenderComponent, &hero.SpaceComponent)
		case *animation.AnimationSystem:
			sys.Add(&hero.BasicEntity, &hero.AnimationComponent, &hero.RenderComponent)
		case *ControlSystem:
			sys.Add(&hero.BasicEntity, &hero.AnimationComponent)
		}
	}
}

func (*DefaultScene) Type() string { return "GameWorld" }

func (*DefaultScene) CreateEntity(point *math.Point, spriteSheet *spritesheet.Spritesheet, action *animation.AnimationAction) *Animation {
	entity := &Animation{BasicEntity: ecs.NewBasic()}

	entity.SpaceComponent = space.SpaceComponent{*point, 150, 150}
	entity.RenderComponent = render.NewRenderComponent(spriteSheet.Cell(action.Frames[0]), math.Point{3, 3}, "hero")
	entity.AnimationComponent = animation.NewAnimationComponent(spriteSheet.Drawables(), 0.1)
	entity.AnimationComponent.AddAnimationActions(actions)
	entity.AnimationComponent.SelectAnimationByAction(action)

	return entity
}

type controlEntity struct {
	*ecs.BasicEntity
	*animation.AnimationComponent
}

type ControlSystem struct {
	entities []controlEntity
}

func (c *ControlSystem) Add(basic *ecs.BasicEntity, anim *animation.AnimationComponent) {
	c.entities = append(c.entities, controlEntity{basic, anim})
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
		if input.Keys.Get(input.ArrowRight).Down() {
			e.AnimationComponent.SelectAnimationByAction(WalkAction)
		} else if input.Keys.Get(input.Space).Down() {
			e.AnimationComponent.SelectAnimationByAction(SkillAction)
		} else {
			e.AnimationComponent.SelectAnimationByAction(StopAction)
		}
	}
}

func main() {
	opts := engo.RunOptions{
		Title:  "Animation Demo",
		Width:  1024,
		Height: 640,
	}
	engo.Run(opts, &DefaultScene{})
}
