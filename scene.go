package engo

import (
	"fmt"

	"engo.io/ecs"
)

var (
	scenes = make(map[string]*sceneWrapper)
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

type sceneWrapper struct {
	scene   Scene
	world   *ecs.World
	mailbox *MessageManager
	camera  *cameraSystem
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
	wrapper, registered := scenes[s.Type()]
	if !registered {
		RegisterScene(s)
		wrapper = scenes[s.Type()]
	}

	// Initialize new Scene / World if needed
	var doSetup bool

	if wrapper.world == nil || forceNewWorld {
		wrapper.world = &ecs.World{}
		wrapper.mailbox = &MessageManager{}
		wrapper.camera = &cameraSystem{}

		doSetup = true
	}

	// Do the switch
	currentScene = s
	currentWorld = wrapper.world
	Mailbox = wrapper.mailbox
	cam = wrapper.camera

	// doSetup is true whenever we're (re)initializing the Scene
	if doSetup {
		s.Preload()
		Files.Load(func() {})

		wrapper.mailbox.listeners = make(map[string][]MessageHandler)

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
	_, ok := scenes[s.Type()]
	if !ok {
		scenes[s.Type()] = &sceneWrapper{scene: s}
	}
}

// SetSceneByName does a lookup for the `Scene` where its `Type()` equals `name`, and then sets it as current `Scene`
func SetSceneByName(name string, forceNewWorld bool) error {
	scene, ok := scenes[name]
	if !ok {
		return fmt.Errorf("scene not registered: %s", name)
	}

	SetScene(scene.scene, forceNewWorld)

	return nil
}
