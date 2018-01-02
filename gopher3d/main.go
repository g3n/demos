// Copyright 2016 The G3N Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is a minimum G3N application showing how to create a window
// and load an external 3D model in OBJ format.
// The gopher OBJ model was exported from the file "gopher.blend"
// at https://github.com/golang-samples/gopher-3d
// For more complete demos please see: https://github.com/g3n/g3nd
package main

import (
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/camera/control"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/loader/obj"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
	"runtime"
)

func main() {

	// Creates window and OpenGL context
	win, err := window.New("glfw", 800, 600, "Gopher3D", false)
	if err != nil {
		panic(err)
	}

	// OpenGL functions must be executed in the same thread where
	// the context was created (by window.New())
	runtime.LockOSThread()

	// Create OpenGL state
	gs, err := gls.New()
	if err != nil {
		panic(err)
	}

	// Sets the initial OpenGL viewport size the same as the window size
	// This will be updated when the window is resized
	width, height := win.GetSize()
	gs.Viewport(0, 0, int32(width), int32(height))

	// Creates scene for 3D objects
	scene := core.NewNode()

	// Adds white ambient light to the scene
	ambLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.5)
	scene.Add(ambLight)

	// Add directional white light from right
	dirLight := light.NewDirectional(&math32.Color{1, 1, 1}, 1.0)
	dirLight.SetPosition(1, 0, 0)
	scene.Add(dirLight)

	// Adds a perspective camera to the scene
	// The camera aspect ratio will be updated when the window is resized.
	aspect := float32(width) / float32(height)
	camera := camera.NewPerspective(65, aspect, 0.01, 1000)

	// Creates orbit camera control and position the camera
	_ = control.NewOrbitControl(camera, win)
	camera.SetPosition(8.3, 4.7, 3.7)

	// Subscribe to window resize events
	win.Subscribe(window.OnWindowSize, func(evname string, ev interface{}) {
		// Updates viewport
		width, height := win.GetSize()
		gs.Viewport(0, 0, int32(width), int32(height))
		// Updates camera aspect ratio
		aspect := float32(width) / float32(height)
		camera.SetAspect(aspect)
	})

	// Add an axis helper to the scene
	axis := graphic.NewAxisHelper(2)
	scene.Add(axis)

	// Decodes model in in OBJ format
	dec, err := obj.Decode("gopher.obj", "gopher.mtl")
	if err != nil {
		panic(err.Error())
	}

	// Creates a new node with all the objects in the decoded file and adds it to the scene
	group, err := dec.NewGroup()
	if err != nil {
		panic(err.Error())
	}
	scene.Add(group)

	// Creates a renderer and adds default shaders
	rend := renderer.NewRenderer(gs)
	err = rend.AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	rend.SetScene(scene)

	// Sets window background color
	gs.ClearColor(0.6, 0.6, 0.6, 1.0)

	// Render loop
	for !win.ShouldClose() {

		// Clear buffers
		gs.Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)

		// Render the scene using the specified camera
		rend.Render(camera)

		// Update window and checks for I/O events
		win.SwapBuffers()
		win.PollEvents()
	}
}
