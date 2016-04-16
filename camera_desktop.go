// +build !netgo,!android

package engo

import "engo.io/ecs"

// EdgeScroller is a System that allows for scrolling when the cursor is near the edges of
// the window
type EdgeScroller struct {
	ScrollSpeed float32
	EdgeMargin  float64
}

func (*EdgeScroller) Type() string             { return "EdgeScroller" }
func (*EdgeScroller) Priority() int            { return 10 }
func (*EdgeScroller) AddEntity(*ecs.BasicEntity)    {}
func (*EdgeScroller) RemoveEntity(*ecs.BasicEntity) {}
func (*EdgeScroller) New(*ecs.World)           {}

func (c *EdgeScroller) Update(dt float32) {
	curX, curY := window.GetCursorPos()
	maxX, maxY := window.GetSize()

	if curX < c.EdgeMargin {
		Mailbox.Dispatch(CameraMessage{XAxis, -c.ScrollSpeed * dt, true})
	} else if curX > float64(maxX)-c.EdgeMargin {
		Mailbox.Dispatch(CameraMessage{XAxis, c.ScrollSpeed * dt, true})
	}

	if curY < c.EdgeMargin {
		Mailbox.Dispatch(CameraMessage{YAxis, -c.ScrollSpeed * dt, true})
	} else if curY > float64(maxY)-c.EdgeMargin {
		Mailbox.Dispatch(CameraMessage{YAxis, c.ScrollSpeed * dt, true})
	}
}
