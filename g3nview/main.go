// Copyright 2016 The G3N Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/camera/control"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/util/logger"
	//"github.com/g3n/engine/loader/obj"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
	"runtime"
)

// Application context
type Context struct {
	win      window.IWindow
	gs       *gls.GLS
	scene    *core.Node
	ambLight *light.Ambient
	dirLight *light.Directional
	cam      *camera.Perspective
	root     *gui.Root
	tb       *toolBar
	axis     *graphic.AxisHelper
	grid     *graphic.GridHelper
}

var log *logger.Logger

func main() {

	// Creates window and OpenGL context
	win, err := window.New("glfw", 800, 600, "Viewer", false)
	if err != nil {
		panic(err)
	}

	// Creates independent logger for the application
	log = logger.New("G3NVIEW", nil)
	log.AddWriter(logger.NewConsole(false))
	log.SetFormat(logger.FTIME | logger.FMICROS)
	log.SetLevel(logger.DEBUG)

	// Starts building app context
	ctx := new(Context)
	ctx.win = win

	// OpenGL functions must be executed in the same thread where
	// the context was created (by window.New())
	runtime.LockOSThread()

	// Create OpenGL state
	gs, err := gls.New()
	if err != nil {
		panic(err)
	}
	ctx.gs = gs

	// Sets the initial OpenGL viewport size the same as the window size
	// This will be updated when the window is resized
	width, height := win.GetSize()
	gs.Viewport(0, 0, int32(width), int32(height))

	// Creates scene for 3D objects
	ctx.scene = core.NewNode()

	// Creates root panel for GUI
	ctx.root = gui.NewRoot(gs, win)
	buildGui(ctx)

	// Adds white ambient light to the scene
	ctx.ambLight = light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.5)
	ctx.scene.Add(ctx.ambLight)

	// Add directional white light from right
	ctx.dirLight = light.NewDirectional(&math32.Color{1, 1, 1}, 1.0)
	ctx.dirLight.SetPosition(1, 0, 0)
	ctx.scene.Add(ctx.dirLight)

	// Adds a perspective camera to the scene
	// The camera aspect ratio will be updated when the window is resized.
	aspect := float32(width) / float32(height)
	ctx.cam = camera.NewPerspective(65, aspect, 0.01, 1000)

	// Creates orbit camera control and position the camera
	_ = control.NewOrbitControl(ctx.cam, ctx.win)
	ctx.cam.SetPosition(8.3, 4.7, 3.7)

	// Subscribe to window resize events
	win.Subscribe(window.OnWindowSize, func(evname string, ev interface{}) {
		onWinResize(ctx)
	})

	// Add an axis helper to the scene initially not visible
	ctx.axis = graphic.NewAxisHelper(2)
	ctx.axis.SetVisible(false)
	ctx.scene.Add(ctx.axis)

	// Adds a grid helper to the scene initially not visible
	ctx.grid = graphic.NewGridHelper(50, 1, &math32.Color{0.4, 0.4, 0.4})
	ctx.grid.SetVisible(false)
	ctx.scene.Add(ctx.grid)

	// Creates a renderer and adds default shaders
	rend := renderer.NewRenderer(gs)
	err = rend.AddDefaultShaders()
	if err != nil {
		panic(err)
	}

	// Sets window background color
	gs.ClearColor(0.6, 0.6, 0.6, 1.0)

	onWinResize(ctx)

	// Render loop
	for !win.ShouldClose() {

		// Clear buffers
		gs.Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)

		// Render the scene using the specified camera
		err := rend.Render(ctx.scene, ctx.cam)
		if err != nil {
			log.Fatal("Render error: %s\n", err)
		}

		// Render GUI over everything
		gs.Clear(gls.DEPTH_BUFFER_BIT)
		err = rend.Render(ctx.root, ctx.cam)
		if err != nil {
			log.Fatal("Render error: %s\n", err)
		}

		// Update window and checks for I/O events
		win.SwapBuffers()
		win.PollEvents()
	}
}

// winResizeEvent is called when the window resize event is received
func onWinResize(ctx *Context) {

	// Sets view port
	width, height := ctx.win.GetSize()
	ctx.gs.Viewport(0, 0, int32(width), int32(height))

	// Sets camera aspect ratio
	aspect := float32(width) / float32(height)
	ctx.cam.SetAspect(aspect)

	// Sets GUI root panel size
	ctx.root.SetSize(float32(width), float32(height))
}

type toolBar struct {
	*gui.Panel
	ctx *Context
	mb  *gui.Menu
}

func NewToolbar(ctx *Context) *toolBar {

	// Creates toolbar container panel
	tb := new(toolBar)
	tb.Panel = gui.NewPanel(400, 32)
	tb.Panel.SetColor(&math32.White)
	tb.ctx = ctx
	ctx.root.Add(tb.Panel)
	ctx.root.Subscribe(gui.OnResize, func(evname string, ev interface{}) {
		tb.onResize()
	})

	// Set the toolbar layout
	hbl := gui.NewHBoxLayout()
	tb.Panel.SetLayout(hbl)

	// Create menu bar and adds it to the toolbar
	tb.mb = gui.NewMenuBar()
	tb.mb.SetLayoutParams(&gui.HBoxLayoutParams{Expand: 0, AlignV: gui.AlignCenter})
	//styles := gui.StyleDefault.Menu
	//tb.mb.SetStyles(styles)
	tb.Add(tb.mb)

	// Create "File" menu and adds it to the menu bar
	m1 := gui.NewMenu()
	m1.AddOption("Open model").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		tb.openModel()
	})
	m1.AddSeparator()
	m1.AddOption("Quit").SetId("quit").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		tb.ctx.win.SetShouldClose(true)
	})
	tb.mb.AddMenu("File", m1)

	// Adds spacer
	sp := gui.NewPanel(8, 0)
	sp.SetLayoutParams(&gui.HBoxLayoutParams{Expand: 0, AlignV: gui.AlignCenter})
	sp.SetRenderable(false)
	tb.Add(sp)

	// Create grid checkbox and adds it to the menu bar
	cbGrid := gui.NewCheckBox("Show grid helper")
	cbGrid.SetLayoutParams(&gui.HBoxLayoutParams{Expand: 0, AlignV: gui.AlignCenter})
	cbGrid.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		tb.ctx.grid.SetVisible(cbGrid.Value())
	})
	tb.Add(cbGrid)

	// Adds spacer
	sp = gui.NewPanel(8, 0)
	sp.SetLayoutParams(&gui.HBoxLayoutParams{Expand: 0, AlignV: gui.AlignCenter})
	tb.Add(sp)

	// Create axis checkbox and adds it to the menu bar
	cbAxis := gui.NewCheckBox("Show axis helper")
	cbAxis.SetLayoutParams(&gui.HBoxLayoutParams{Expand: 0, AlignV: gui.AlignCenter})
	cbAxis.Subscribe(gui.OnChange, func(name string, ev interface{}) {
		tb.ctx.axis.SetVisible(cbAxis.Value())
	})
	tb.Add(cbAxis)

	return tb
}

func (tb *toolBar) onResize() {

	width, _ := tb.ctx.win.GetSize()
	tb.SetWidth(float32(width))
}

func (tb *toolBar) openModel() {

	log.Debug("openModel")
}

func buildGui(ctx *Context) {

	ctx.tb = NewToolbar(ctx)
}
