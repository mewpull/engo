package main

import (
	"log"

	"engo.io/ecs"
	"engo.io/engo"
)

type DefaultScene struct{}

func (*DefaultScene) Preload()           {}
func (*DefaultScene) Setup(w *ecs.World) {}

func (*DefaultScene) Exit() {
	log.Println("Exit event called; we can do whatever we want now")
	// Here if you want you can prompt the user if they're sure they want to close
	log.Println("Manually closing")
	engo.Exit()
}

func (*DefaultScene) Type() string { return "Game" }

func main() {
	opts := engo.RunOptions{
		Title:               "Exit Demo",
		Width:               1024,
		Height:              640,
		OverrideCloseAction: true,
	}
	engo.Run(opts, &DefaultScene{})
}
