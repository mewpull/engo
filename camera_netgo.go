//+build netgo

package engo

import (
	"log"

	"engo.io/ecs"
)

// EdgeScroller is a System that allows for scrolling when the cursor is near the edges of
// the window
type EdgeScroller struct {
	ScrollSpeed float32
	EdgeMargin  float64
}

func (*EdgeScroller) Type() string             { return "EdgeScroller" }
func (*EdgeScroller) Priority() int            { return 10 }
func (*EdgeScroller) AddEntity(*ecs.Entity)    {}
func (*EdgeScroller) RemoveEntity(*ecs.Entity) {}
func (*EdgeScroller) New(*ecs.World) {
	log.Println("EdgeScroller is not yet supported with gopherjs")
}

func (c *EdgeScroller) Update(dt float32) {}
