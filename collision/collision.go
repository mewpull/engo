package collision // import "engo.io/engo/collision"

import (
	"log"

	"engo.io/ecs"
	engomath "engo.io/engo/math"
	"engo.io/engo/message"
	"engo.io/engo/space"
	"github.com/luxengine/math"
)

type CollisionComponent struct {
	Solid, Main bool
	Extra       engomath.Point
}

type CollisionMessage struct {
	Entity collisionEntity
	To     collisionEntity
}

func (CollisionMessage) Type() string { return "CollisionMessage" }

type collisionEntity struct {
	*ecs.BasicEntity
	*CollisionComponent
	*space.SpaceComponent
}

type CollisionSystem struct {
	entities []collisionEntity
}

func (c *CollisionSystem) Add(basic *ecs.BasicEntity, collision *CollisionComponent, space *space.SpaceComponent) {
	c.entities = append(c.entities, collisionEntity{basic, collision, space})
}

func (c *CollisionSystem) Remove(basic ecs.BasicEntity) {
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

func (cs *CollisionSystem) Update(dt float32) {
	for i1, e1 := range cs.entities {
		if !e1.CollisionComponent.Main {
			continue // with other entities
		}

		entityAABB := e1.SpaceComponent.AABB()
		offset := engomath.Point{e1.CollisionComponent.Extra.X / 2, e1.CollisionComponent.Extra.Y / 2}
		entityAABB.Min.X -= offset.X
		entityAABB.Min.Y -= offset.Y
		entityAABB.Max.X += offset.X
		entityAABB.Max.Y += offset.Y

		for i2, e2 := range cs.entities {
			if i1 == i2 {
				continue // with other entities, because we won't collide with ourselves
			}

			otherAABB := e2.SpaceComponent.AABB()
			offset = engomath.Point{e2.CollisionComponent.Extra.X / 2, e2.CollisionComponent.Extra.Y / 2}
			otherAABB.Min.X -= offset.X
			otherAABB.Min.Y -= offset.Y
			otherAABB.Max.X += offset.X
			otherAABB.Max.Y += offset.Y

			if IsIntersecting(entityAABB, otherAABB) {
				if e1.CollisionComponent.Solid && e2.CollisionComponent.Solid {
					mtd := MinimumTranslation(entityAABB, otherAABB)
					e1.SpaceComponent.Position.X += mtd.X
					e1.SpaceComponent.Position.Y += mtd.Y
				}

				message.Mailbox.Dispatch(CollisionMessage{Entity: e1, To: e2})
			}
		}
	}
}

func IsIntersecting(rect1 space.AABB, rect2 space.AABB) bool {
	if rect1.Max.X > rect2.Min.X && rect1.Min.X < rect2.Max.X && rect1.Max.Y > rect2.Min.Y && rect1.Min.Y < rect2.Max.Y {
		return true
	}

	return false
}

func MinimumTranslation(rect1 space.AABB, rect2 space.AABB) engomath.Point {
	mtd := engomath.Point{}

	left := rect2.Min.X - rect1.Max.X
	right := rect2.Max.X - rect1.Min.X
	top := rect2.Min.Y - rect1.Max.Y
	bottom := rect2.Max.Y - rect1.Min.Y

	if left > 0 || right < 0 {
		log.Println("Box aint intercepting")
		return mtd
		//box doesnt intercept
	}

	if top > 0 || bottom < 0 {
		log.Println("Box aint intercepting")
		return mtd
		//box doesnt intercept
	}
	if math.Abs(left) < right {
		mtd.X = left
	} else {
		mtd.X = right
	}

	if math.Abs(top) < bottom {
		mtd.Y = top
	} else {
		mtd.Y = bottom
	}

	if math.Abs(mtd.X) < math.Abs(mtd.Y) {
		mtd.Y = 0
	} else {
		mtd.X = 0
	}

	return mtd
}
