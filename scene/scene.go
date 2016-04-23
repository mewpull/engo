package scene // import "engo.io/engo/scene"

import (
	"fmt"

	"engo.io/ecs"
	"engo.io/engo/assets"
	"engo.io/engo/camera"
	internalcamera "engo.io/engo/internal/camera"
	internalengo "engo.io/engo/internal/engo"
	"engo.io/engo/message"
)

var (
	// TODO(u): Unexport Scenes.
	Scenes       = make(map[string]*SceneWrapper)
	currentScene Scene
)

// Scene represents a screen ingame.
// i.e.: main menu, settings, but also the game itself
type Scene interface {
	// Preload is called before loading resources
	Preload()

	// Setup is called before the main loop
	Setup(*ecs.World)

	// Type returns a unique string representation of the Scene, used to identify it
	Type() string
}

type Shower interface {
	// Show is called whenever the other Scene becomes inactive, and this one becomes the active one
	Show()
}

type Hider interface {
	// Hide is called when an other Scene becomes active
	Hide()
}

type Exiter interface {
	// Exit is called when the user or the system requests to close the game
	// This should be used to cleanup or prompt user if they're sure they want to close
	// To prevent the default action (close/exit) make sure to set OverrideCloseAction in
	// your RunOpts to `true`. You should then handle the exiting of the program by calling
	//    engo.Exit()
	Exit() bool
}

// TODO(u): Unexport SceneWrapper.
type SceneWrapper struct {
	// TODO(u): Unexport Scene.
	Scene   Scene
	world   *ecs.World
	mailbox *message.MessageManager
	camera  *camera.CameraSystem
}

// CurrentScene returns the SceneWorld that is currently active
func CurrentScene() Scene {
	return currentScene
}

// SetScene sets the currentScene to the given Scene, and
// optionally forcing to create a new ecs.World that goes with it.
func SetScene(s Scene, forceNewWorld bool) {
	// Break down currentScene
	if currentScene != nil {
		if hider, ok := currentScene.(Hider); ok {
			hider.Hide()
		}
	}

	// Register Scene if needed
	wrapper, registered := Scenes[s.Type()]
	if !registered {
		RegisterScene(s)
		wrapper = Scenes[s.Type()]
	}

	// Initialize new Scene / World if needed
	var doSetup bool

	if wrapper.world == nil || forceNewWorld {
		wrapper.world = &ecs.World{}
		wrapper.mailbox = &message.MessageManager{}
		wrapper.camera = &camera.CameraSystem{}

		doSetup = true
	}

	// Do the switch
	currentScene = s
	internalengo.CurrentWorld = wrapper.world
	message.Mailbox = wrapper.mailbox
	internalcamera.Cam = wrapper.camera

	// doSetup is true whenever we're (re)initializing the Scene
	if doSetup {
		s.Preload()
		assets.Files.Load(func() {})

		wrapper.mailbox.Listeners = make(map[string][]message.MessageHandler)

		wrapper.world.AddSystem(wrapper.camera)

		s.Setup(wrapper.world)
	} else {
		if shower, ok := currentScene.(Shower); ok {
			shower.Show()
		}
	}
}

// RegisterScene registers the `Scene`, so it can later be used by `SetSceneByName`
func RegisterScene(s Scene) {
	_, ok := Scenes[s.Type()]
	if !ok {
		Scenes[s.Type()] = &SceneWrapper{Scene: s}
	}
}

// SetSceneByName does a lookup for the `Scene` where its `Type()` equals `name`, and then sets it as current `Scene`
func SetSceneByName(name string, forceNewWorld bool) error {
	scene, ok := Scenes[name]
	if !ok {
		return fmt.Errorf("scene not registered: %s", name)
	}

	SetScene(scene.Scene, forceNewWorld)

	return nil
}
