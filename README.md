# Engo
[![Join the chat at https://gitter.im/EngoEngine/engo](https://badges.gitter.im/EngoEngine/engo.svg)](https://gitter.im/EngoEngine/engo?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) ![License](https://img.shields.io/badge/License-MIT-blue.svg) [![Build Status](https://travis-ci.org/EngoEngine/engo.svg?branch=master)](https://travis-ci.org/EngoEngine/engo)

A cross-platform game engine written in Go following an interpretation of the Entity Component System paradigm. Engo is
currently compilable for Mac OSX, Linux and Windows. With the release of Go 1.4, sporting Android and the inception of
iOS compatibility, mobile will [soon](https://github.com/EngoEngine/engo/issues/63) be added as a release target. Web
support  ([gopherjs](https://github.com/gopherjs/gopherjs)) is also [planned](https://github.com/EngoEngine/engo/issues/71).

Currently documentation is pretty scarce, this is because we have not *completely* finalized the API and are about to 
go through a "prettification" process in order to increase elegance and usability. For a basic up-to-date example of 
most features, look at the demos.

## Getting in touch / Contributing

We have a [gitter](https://gitter.im/EngoEngine/engo) chat for people to join who want to further discuss `engo`. We are happy to discuss bugs, feature requests and would love to hear about the projects you are building!

## Getting Started

1. First, you have to install some dependencies:
  1. If you're running on Debian/Ubuntu:
    `sudo apt-get install libopenal-dev libglu1-mesa-dev freeglut3-dev mesa-common-dev xorg-dev libgl1-mesa-dev`
  2. If you're running on Windows: download all above packages using [win-builds](http://win-builds.org/doku.php) (Open an issue to let us know about other methods)
  3. If you're on OSX, you should be OK. [Open an issue if you are not](https://github.com/EngoEngine/engo/issues/new)
2. Then, you can go get it:
`go get -u engo.io/engo`
3. Now, you have two choices:
  1. Visit [our website](https://engo.io/), which hosts a full-blown toturial series on how to create your own game, and on top of that, has some conceptual explanations;
  2. Check out some demos in our [demos folder](https://github.com/EngoEngine/engo/tree/master/demos). 
4. Finally, if you run into problems, if you've encountered a bug, or want to request a feature, feel free to shoot 
us a DM or [create an issue](https://github.com/EngoEngine/engo/issues/new). 

## Breaking Changes
Engo is currently undergoing a lot of optimizations and constantly gets new features. However, this sometimes means things break. In order to make transitioning easier for you, 
we have a list of those changes, with the most recent being at the top. If you run into any problems, please contact us at [gitter](https://gitter.im/EngoEngine/engo). 

* `ecs.Entity` changed to `ecs.BasicEntity`, `world.AddEntity` is gone - **a lot** has changed here. The entire issue is described [here](https://github.com/EngoEngine/ecs/issues/13), while [this comment](https://github.com/EngoEngine/ecs/issues/13#issuecomment-210887914) in particular, should help you migrate your code. 
* Renamed `engo.io/webgl` to `engo.io/gl`, because the package handles more than only *web*gl. 
* `scene.Exit()` - a `Scene` now also requires an `Exit()` function, alongside the `Hide()` and `Show()` it already required. 
* `github.com/EngoEngine/engo` -> `engo.io/engo` - Our packages `engo`, `ecs` and `webgl` should now be imported using the `engo.io` path. 
* `engi.XXX` -> `engo.XXX` - We renamed our package `engi` to `engo`. 

## History

Engo, originally known as `Engi` was written by [ajhager](https://github.com/ajhager) as a general purpose Go game engine. With a desire to build it into an "ECS" game engine, it was forked to `github.com/paked/engi`. After passing through several iterations, it was decided that the project would be rebranded and rereleased as Engo on its own GitHub organisation.

## Credits

Thank you to everyone who has worked on, or with `Engo`. Non of this would be possible without you, and your help has been truly amazing.

- [ajhager](https://github.com/ajhager): Building the original `engi`, which engo was based off of
- [paked](https://github.com/paked): Adding ECS element, project maintenance and management
- [EtienneBruines](https://github.com/EtienneBruines): Rewriting the OpenGL code, maintenance and helping redesign the API
- [Everyone else who has submitted PRs over the years, to any iteration of the project](https://github.com/EngoEngine/engo/graphs/contributors)
