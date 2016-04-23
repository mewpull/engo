package camera // import "engo.io/engo/camera"

import (
	"log"
	"sync"
	"time"

	"engo.io/ecs"
	"engo.io/engo/input"
	internalinput "engo.io/engo/internal/input"
	internalwindow "engo.io/engo/internal/window"
	"engo.io/engo/message"
	"engo.io/engo/space"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	MinZoom float32 = 0.25
	MaxZoom float32 = 3
)

type cameraEntity struct {
	*ecs.BasicEntity
	*space.SpaceComponent
}

// CameraSystem is a System that manages the state of the Camera
type CameraSystem struct {
	x, y, z  float32
	tracking cameraEntity // The entity that is currently being followed

	longTasks map[CameraAxis]*CameraMessage
}

func (cam *CameraSystem) New(*ecs.World) {
	cam.x = internalwindow.WorldBounds.Max.X / 2
	cam.y = internalwindow.WorldBounds.Max.Y / 2
	cam.z = 1
	cam.longTasks = make(map[CameraAxis]*CameraMessage)

	message.Mailbox.Listen("CameraMessage", func(msg message.Message) {
		cammsg, ok := msg.(CameraMessage)
		if !ok {
			return
		}

		// Stop with whatever we're doing now
		if _, ok := cam.longTasks[cammsg.Axis]; ok {
			delete(cam.longTasks, cammsg.Axis)
		}

		if cammsg.Duration > time.Duration(0) {
			cam.longTasks[cammsg.Axis] = &cammsg
			return // because it's handled incrementally
		}

		if cammsg.Incremental {
			switch cammsg.Axis {
			case XAxis:
				cam.moveX(cammsg.Value)
			case YAxis:
				cam.moveY(cammsg.Value)
			case ZAxis:
				cam.zoom(cammsg.Value)
			}
		} else {
			switch cammsg.Axis {
			case XAxis:
				cam.moveToX(cammsg.Value)
			case YAxis:
				cam.moveToY(cammsg.Value)
			case ZAxis:
				cam.zoomTo(cammsg.Value)
			}
		}
	})
}

func (cam *CameraSystem) Remove(basic ecs.BasicEntity) {}

func (cam *CameraSystem) Update(dt float32) {
	for axis, longTask := range cam.longTasks {
		if !longTask.Incremental {
			longTask.Incremental = true

			switch axis {
			case XAxis:
				longTask.Value -= cam.X()
			case YAxis:
				longTask.Value -= cam.Y()
			case ZAxis:
				longTask.Value -= cam.Z()
			}
		}

		// Set speed if needed
		if longTask.speed == 0 {
			longTask.speed = longTask.Value / float32(longTask.Duration.Seconds())
		}

		dAxis := longTask.speed * dt
		switch axis {
		case XAxis:
			cam.moveX(dAxis)
		case YAxis:
			cam.moveY(dAxis)
		case ZAxis:
			cam.zoom(dAxis)
		}

		longTask.Duration -= time.Duration(dt)
		if longTask.Duration <= time.Duration(0) {
			delete(cam.longTasks, axis)
		}
	}

	if cam.tracking.BasicEntity == nil {
		return
	}

	if cam.tracking.SpaceComponent == nil {
		log.Println("Should be tracking", cam.tracking.BasicEntity.ID(), "but SpaceComponent is nil")
		cam.tracking.BasicEntity = nil
		return
	}

	cam.centerCam(cam.tracking.SpaceComponent.Position.X+cam.tracking.SpaceComponent.Width/2,
		cam.tracking.SpaceComponent.Position.Y+cam.tracking.SpaceComponent.Height/2,
		cam.z,
	)
}

func (cam *CameraSystem) FollowEntity(basic *ecs.BasicEntity, space *space.SpaceComponent) {
	cam.tracking = cameraEntity{basic, space}
}

func (cam *CameraSystem) X() float32 {
	return cam.x
}

func (cam *CameraSystem) Y() float32 {
	return cam.y
}

func (cam *CameraSystem) Z() float32 {
	return cam.z
}

func (cam *CameraSystem) moveX(value float32) {
	cam.moveToX(cam.x + value)
}

func (cam *CameraSystem) moveY(value float32) {
	cam.moveToY(cam.y + value)
}

func (cam *CameraSystem) zoom(value float32) {
	cam.zoomTo(cam.z + value)
}

func (cam *CameraSystem) moveToX(location float32) {
	cam.x = mgl32.Clamp(location, internalwindow.WorldBounds.Min.X, internalwindow.WorldBounds.Max.X)
}

func (cam *CameraSystem) moveToY(location float32) {
	cam.y = mgl32.Clamp(location, internalwindow.WorldBounds.Min.Y, internalwindow.WorldBounds.Max.Y)
}

func (cam *CameraSystem) zoomTo(zoomLevel float32) {
	cam.z = mgl32.Clamp(zoomLevel, MinZoom, MaxZoom)
}

func (cam *CameraSystem) centerCam(x, y, z float32) {
	cam.moveToX(x)
	cam.moveToY(y)
	cam.zoomTo(z)
}

// CameraAxis is the axis at which the Camera can/has to move
type CameraAxis uint8

const (
	XAxis CameraAxis = iota
	YAxis
	ZAxis
)

// CameraMessage is a message that can be sent to the Camera (and other Systemers), to indicate movement
type CameraMessage struct {
	Axis        CameraAxis
	Value       float32
	Incremental bool
	Duration    time.Duration
	speed       float32
}

func (CameraMessage) Type() string {
	return "CameraMessage"
}

// KeyboardScroller is a System that allows for scrolling when certain keys are pressed
type KeyboardScroller struct {
	ScrollSpeed float32
	upKeys      []input.Key
	leftKeys    []input.Key
	downKeys    []input.Key
	rightKeys   []input.Key

	keysMu sync.RWMutex
}

func (*KeyboardScroller) Priority() int          { return 10 }
func (*KeyboardScroller) Remove(ecs.BasicEntity) {}

func (c *KeyboardScroller) Update(dt float32) {
	c.keysMu.RLock()
	defer c.keysMu.RUnlock()

	for _, upKey := range c.upKeys {
		if input.Keys.Get(upKey).Down() {
			message.Mailbox.Dispatch(CameraMessage{Axis: YAxis, Value: -c.ScrollSpeed * dt, Incremental: true})
			break
		}
	}

	for _, rightKey := range c.rightKeys {
		if input.Keys.Get(rightKey).Down() {
			message.Mailbox.Dispatch(CameraMessage{Axis: XAxis, Value: c.ScrollSpeed * dt, Incremental: true})
			break
		}
	}

	for _, downKey := range c.downKeys {
		if input.Keys.Get(downKey).Down() {
			message.Mailbox.Dispatch(CameraMessage{Axis: YAxis, Value: c.ScrollSpeed * dt, Incremental: true})
			break
		}
	}

	for _, leftKey := range c.leftKeys {
		if input.Keys.Get(leftKey).Down() {
			message.Mailbox.Dispatch(CameraMessage{Axis: XAxis, Value: -c.ScrollSpeed * dt, Incremental: true})
			break
		}
	}
}

func (c *KeyboardScroller) BindKeyboard(up, right, down, left input.Key) {
	c.keysMu.Lock()
	defer c.keysMu.Unlock()

	c.upKeys = append(c.upKeys, up)
	c.rightKeys = append(c.rightKeys, right)
	c.downKeys = append(c.downKeys, down)
	c.leftKeys = append(c.leftKeys, left)
}

func NewKeyboardScroller(scrollSpeed float32, up, right, down, left input.Key) *KeyboardScroller {
	kbs := &KeyboardScroller{
		ScrollSpeed: scrollSpeed,
	}
	kbs.BindKeyboard(up, right, down, left)
	return kbs
}

// EdgeScroller is a System that allows for scrolling when the cursor is near the edges of
// the window
type EdgeScroller struct {
	ScrollSpeed float32
	EdgeMargin  float64
}

func (*EdgeScroller) Priority() int          { return 10 }
func (*EdgeScroller) Remove(ecs.BasicEntity) {}

func (c *EdgeScroller) Update(dt float32) {
	curX, curY := internalwindow.Window.GetCursorPos()
	maxX, maxY := internalwindow.Window.GetSize()

	if curX < c.EdgeMargin {
		message.Mailbox.Dispatch(CameraMessage{Axis: XAxis, Value: -c.ScrollSpeed * dt, Incremental: true})
	} else if curX > float64(maxX)-c.EdgeMargin {
		message.Mailbox.Dispatch(CameraMessage{Axis: XAxis, Value: c.ScrollSpeed * dt, Incremental: true})
	}

	if curY < c.EdgeMargin {
		message.Mailbox.Dispatch(CameraMessage{Axis: YAxis, Value: -c.ScrollSpeed * dt, Incremental: true})
	} else if curY > float64(maxY)-c.EdgeMargin {
		message.Mailbox.Dispatch(CameraMessage{Axis: YAxis, Value: c.ScrollSpeed * dt, Incremental: true})
	}
}

// MouseZoomer is a System that allows for zooming when the scroll wheel is used
type MouseZoomer struct {
	ZoomSpeed float32
}

func (*MouseZoomer) Priority() int          { return 10 }
func (*MouseZoomer) Remove(ecs.BasicEntity) {}

func (c *MouseZoomer) Update(dt float32) {
	if internalinput.Mouse.ScrollY != 0 {
		message.Mailbox.Dispatch(CameraMessage{Axis: ZAxis, Value: internalinput.Mouse.ScrollY * c.ZoomSpeed, Incremental: true})
	}
}
