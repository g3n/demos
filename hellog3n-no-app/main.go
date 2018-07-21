package main

import (
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/camera/control"
	"runtime"
)

func main() {

	// Creates window and OpenGL context
	wmgr, err := window.Manager("glfw")
	if err != nil {
		panic(err)
	}
	win, err := wmgr.CreateWindow(800, 600, "Hello G3N", false)
	if err != nil {
		panic(err)
	}

	// OpenGL functions must be executed in the same thread where
	// the context was created (by wmgr.CreateWindow())
	runtime.LockOSThread()

	// Create OpenGL state
	gs, err := gls.New()
	if err != nil {
		panic(err)
	}

	// Set the OpenGL viewport size the same as the window size
	// We later set up a callback to update this if the window is resized
	width, height := win.Size()
	gs.Viewport(0, 0, int32(width), int32(height))

	// Add a perspective camera to the scene
	// We later set up a callback to update the camera aspect ratio if the window is resized
	aspect := float32(width) / float32(height)
	camera := camera.NewPerspective(65, aspect, 0.01, 1000)
	camera.SetPosition(0, 0, 3)

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	win.Subscribe(window.OnWindowSize, func(evname string, ev interface{}) {

		// Get framebuffer size and update viewport accordingly
		width, height := win.FramebufferSize()
		gs.Viewport(0, 0, int32(width), int32(height))

		// Update the camera's aspect ratio
		aspect := float32(width) / float32(height)
		camera.SetAspect(aspect)
	})

	// Set up orbit control for the camera
	control.NewOrbitControl(camera, win)

	// Create scene for 3D objects
	scene := core.NewNode()
	scene.Add(camera)

	// Create a blue torus and add it to the scene
	geom := geometry.NewTorus(1, .4, 12, 32, math32.Pi*2)
	mat := material.NewPhong(math32.NewColor("DarkBlue"))
	torusMesh := graphic.NewMesh(geom, mat)
	scene.Add(torusMesh)

	// Add lights to the scene
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	scene.Add(ambientLight)
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	scene.Add(pointLight)

	// Add an axis helper to the scene
	axis := graphic.NewAxisHelper(0.5)
	scene.Add(axis)

	// Create a renderer and add default shaders
	rend := renderer.NewRenderer(gs)
	err = rend.AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	rend.SetScene(scene)

	// Set window background color to gray
	gs.ClearColor(0.5, 0.5, 0.5, 1.0)

	// Render loop
	for !win.ShouldClose() {

		// Render the scene using the specified camera
		rend.Render(camera)

		// Update window and check for I/O events
		win.SwapBuffers()
		wmgr.PollEvents()
	}
}
