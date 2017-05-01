// Copyright 2016 The G3N Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/camera/control"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/gui/assets"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/loader/collada"
	"github.com/g3n/engine/loader/obj"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/logger"
	"github.com/g3n/engine/window"
	"io"
	"os"
	"path/filepath"
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
	camPos   math32.Vector3
	oc       *control.OrbitControl
	root     *gui.Root
	axis     *graphic.AxisHelper
	grid     *graphic.GridHelper
	models   []*core.Node
	ui       *guiState
}

const (
	checkON  = assets.CheckBox
	checkOFF = assets.CheckBoxOutlineBlank
)

// Package logger
var log *logger.Logger

func main() {

	// Parse command line parameters
	flag.Usage = usage
	flag.Parse()

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

	// Creates root panel for GUI and sets the GUI
	ctx.root = gui.NewRoot(gs, win)
	setupGui(ctx)

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
	ctx.oc = control.NewOrbitControl(ctx.cam, ctx.win)
	ctx.camPos = math32.Vector3{8.3, 4.7, 3.7}
	ctx.cam.SetPositionVec(&ctx.camPos)

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

	// Subscribe to window resize events
	win.Subscribe(window.OnWindowSize, func(evname string, ev interface{}) {
		onWinResize(ctx)
	})
	onWinResize(ctx)

	// Sets window background color
	gs.ClearColor(0.6, 0.6, 0.6, 1.0)

	// Try to load models specified in the command line
	for _, m := range flag.Args() {
		log.Debug("m:%v", m)
		err = openModel(ctx, m)
		if err != nil {
			log.Error("%s", err)
		}
	}

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

// usage shows the application usage
func usage() {

	fmt.Fprintf(os.Stderr, "usage: g3nview [model1 model2   modelN]\n")
	flag.PrintDefaults()
	os.Exit(2)
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

// guiState contains all gui elements and states
type guiState struct {
	ctx      *Context
	mb       *gui.Menu
	fs       *FileSelect
	ed       *ErrorDialog
	viewAxis bool
	viewGrid bool
}

// setupGui builds the gui
func setupGui(ctx *Context) *guiState {

	ui := new(guiState)
	ui.ctx = ctx
	ctx.ui = ui

	// Create menu bar and adds it to the toolbar
	ui.mb = gui.NewMenuBar()
	ui.mb.SetLayoutParams(&gui.HBoxLayoutParams{Expand: 0, AlignV: gui.AlignCenter})
	ui.ctx.root.Add(ui.mb)
	ctx.root.Subscribe(gui.OnResize, func(evname string, ev interface{}) {
		ui.onResize()
	})

	// Create "File" menu and adds it to the menu bar
	m1 := gui.NewMenu()
	m1.AddOption("Open model").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		ui.fs.SetVisible(true)
	})
	m1.AddOption("Remove models").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		removeModels(ctx)
	})
	m1.AddOption("Reset camera").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		ctx.cam.SetPositionVec(&ctx.camPos)
		ctx.cam.LookAt(&math32.Vector3{0, 0, 0})
	})
	m1.AddSeparator()
	m1.AddOption("Quit").SetId("quit").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		ui.ctx.win.SetShouldClose(true)
	})
	ui.mb.AddMenu("File", m1)

	// Create "View" menu and adds it to the menu bar
	m2 := gui.NewMenu()
	vAxis := m2.AddOption("View axis helper").SetIcon(checkOFF)
	vAxis.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		if ui.viewAxis {
			vAxis.SetIcon(checkOFF)
			ui.viewAxis = false
		} else {
			vAxis.SetIcon(checkON)
			ui.viewAxis = true
		}
		ctx.axis.SetVisible(ui.viewAxis)
	})
	vGrid := m2.AddOption("View grid helper").SetIcon(checkOFF)
	vGrid.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		if ui.viewGrid {
			vGrid.SetIcon(checkOFF)
			ui.viewGrid = false
		} else {
			vGrid.SetIcon(checkON)
			ui.viewGrid = true
		}
		ctx.grid.SetVisible(ui.viewGrid)
	})
	ui.mb.AddMenu("View", m2)

	// Creates file select
	ui.fs = NewFileSelect(400, 300)
	ui.fs.SetVisible(false)
	ui.fs.Subscribe("OnOK", func(evname string, ev interface{}) {
		fpath := ui.fs.Selected()
		if fpath == "" {
			ui.ed.Show("File not selected")
			return
		}
		err := openModel(ui.ctx, fpath)
		if err != nil {
			ui.ed.Show(err.Error())
			return
		}
		ui.fs.SetVisible(false)

	})
	ui.fs.Subscribe("OnCancel", func(evname string, ev interface{}) {
		ui.fs.SetVisible(false)
	})
	ui.ctx.root.Add(ui.fs)

	// Creates error dialog
	ui.ed = NewErrorDialog(440, 100)
	ui.ctx.root.Add(ui.ed)

	return ui
}

// onResize is called when the root panel is resized and sets
// the size and position of the gui elements
func (ui *guiState) onResize() {

	// Sets menu width
	width, height := ui.ctx.win.GetSize()
	ui.mb.SetWidth(float32(width))

	// Center file select
	w := ui.fs.Width()
	h := ui.fs.Height()
	px := (float32(width) - w) / 2
	py := (float32(height) - h) / 2
	ui.fs.SetPosition(px, py)

	// Center error dialog
	w = ui.ed.Width()
	h = ui.ed.Height()
	px = (float32(width) - w) / 2
	py = (float32(height) - h) / 2
	ui.ed.SetPosition(px, py)
}

// openModel try to open the specified model and add it to the scene
func openModel(ctx *Context, fpath string) error {

	dir, file := filepath.Split(fpath)
	ext := filepath.Ext(file)

	// Loads OBJ model
	if ext == ".obj" {
		// Checks for material file in the same dir
		matfile := file[:len(file)-len(ext)]
		matpath := filepath.Join(dir, matfile)
		_, err := os.Stat(matpath)
		if err != nil {
			matpath = ""
		}

		// Decodes model in in OBJ format
		dec, err := obj.Decode(fpath, matpath)
		if err != nil {
			return err
		}

		// Creates a new node with all the objects in the decoded file and adds it to the scene
		group, err := dec.NewGroup()
		if err != nil {
			return err
		}
		ctx.scene.Add(group)
		ctx.models = append(ctx.models, group)
		return nil
	}

	// Loads COLLADA model
	if ext == ".dae" {
		dec, err := collada.Decode(fpath)
		if err != nil && err != io.EOF {
			return err
		}
		dec.SetDirImages(dir)

		// Loads collada scene
		s, err := dec.NewScene()
		if err != nil {
			return err
		}
		ctx.scene.Add(s)
		ctx.models = append(ctx.models, s.GetNode())
		return nil
	}
	return fmt.Errorf("Unrecognized model file extension:[%s]", ext)
}

// removeModels removes and disposes of all loaded models in the scene
func removeModels(ctx *Context) {

	for i := 0; i < len(ctx.models); i++ {
		model := ctx.models[i]
		ctx.scene.Remove(model)
		model.Dispose()
	}
}
