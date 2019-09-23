// Copyright 2016 The G3N Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/gui/assets/icon"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/loader/collada"
	"github.com/g3n/engine/loader/obj"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
)

type g3nView struct {
	*app.Application                // Embedded application object
	fs               *FileSelect    // File selection dialog
	ed               *ErrorDialog   // Error dialog
	axes             *helper.Axes   // Axis helper
	grid             *helper.Grid   // Grid helper
	viewAxes         bool           // Axis helper visible flag
	viewGrid         bool           // Grid helper visible flag
	camPos           math32.Vector3 // Initial camera position
	models           []*core.Node   // Models being shown
	scene            *core.Node
	cam              *camera.Camera
	orbit            *camera.OrbitControl
}

const (
	checkON  = icon.CheckBox
	checkOFF = icon.CheckBoxOutlineBlank
)

func main() {

	// Parse command line parameters
	flag.Usage = usage

	// Creates G3N application
	gv := new(g3nView)
	a := app.App()
	gv.Application = a
	gv.scene = core.NewNode()

	// Adds ambient light
	ambLight := light.NewAmbient(math32.NewColor("white"), 0.5)
	gv.scene.Add(ambLight)

	// Add directional white light from right
	dirLight := light.NewDirectional(math32.NewColor("white"), 1.0)
	dirLight.SetPosition(1, 0, 0)
	gv.scene.Add(dirLight)

	// Add an axis helper to the scene initially not visible
	gv.axes = helper.NewAxes(2)
	gv.viewAxes = true
	gv.axes.SetVisible(gv.viewAxes)
	gv.scene.Add(gv.axes)

	// Adds a grid helper to the scene initially not visible
	gv.grid = helper.NewGrid(50, 1, &math32.Color{0.4, 0.4, 0.4})
	gv.viewGrid = true
	gv.grid.SetVisible(gv.viewGrid)
	gv.scene.Add(gv.grid)

	// Sets the initial camera position
	gv.camPos = math32.Vector3{8.3, 4.7, 3.7}
	gv.cam = camera.New(1)
	gv.cam.SetPositionVec(&gv.camPos)
	gv.cam.LookAt(&math32.Vector3{0, 0, 0}, &math32.Vector3{0, 1, 0})
	gv.orbit = camera.NewOrbitControl(gv.cam)

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := a.GetSize()
		a.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		gv.cam.SetAspect(float32(width) / float32(height))
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	// Build the user interface
	gv.buildGui()

	// Try to load models specified in the command line
	for _, m := range flag.Args() {
		err := gv.openModel(m)
		if err != nil {
			log.Printf("error: %s", err)
			return
		}
	}

	// Set background color to gray
	a.Gls().ClearColor(0.5, 0.5, 0.5, 1.0)

	// Run application main render loop
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(gv.scene, gv.cam)
	})

}

// setupGui builds the GUI
func (gv *g3nView) buildGui() error {

	gui.Manager().Set(gv.scene)

	// Adds menu bar
	mb := gui.NewMenuBar()
	mb.SetLayoutParams(&gui.VBoxLayoutParams{Expand: 0, AlignH: gui.AlignWidth})
	gv.scene.Add(mb)

	// Create "File" menu and adds it to the menu bar
	m1 := gui.NewMenu()
	m1.AddOption("Open model").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		gv.fs.Show(true)
	})
	m1.AddOption("Remove models").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		gv.removeModels()
	})
	m1.AddOption("Reset camera").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		gv.cam.SetPositionVec(&gv.camPos)
		gv.cam.LookAt(&math32.Vector3{0, 0, 0}, &math32.Vector3{0, 1, 0})
		gv.orbit.Reset()
	})
	m1.AddSeparator()
	m1.AddOption("Quit").SetId("quit").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		gv.Exit()
	})
	mb.AddMenu("File", m1)

	// Create "View" menu and adds it to the menu bar
	m2 := gui.NewMenu()
	vAxis := m2.AddOption("View axis helper").SetIcon(checkOFF)
	vAxis.SetIcon(getIcon(gv.viewAxes))
	vAxis.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		gv.viewAxes = !gv.viewAxes
		vAxis.SetIcon(getIcon(gv.viewAxes))
		gv.axes.SetVisible(gv.viewAxes)
	})

	vGrid := m2.AddOption("View grid helper").SetIcon(checkOFF)
	vGrid.SetIcon(getIcon(gv.viewGrid))
	vGrid.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		gv.viewGrid = !gv.viewGrid
		vGrid.SetIcon(getIcon(gv.viewGrid))
		gv.grid.SetVisible(gv.viewGrid)
	})
	mb.AddMenu("View", m2)

	// Creates file selection dialog
	fs, err := NewFileSelect(400, 300)
	if err != nil {
		return err
	}
	gv.fs = fs
	gv.fs.SetVisible(false)
	gv.fs.Subscribe("OnOK", func(evname string, ev interface{}) {
		fpath := gv.fs.Selected()
		if fpath == "" {
			gv.ed.Show("File not selected")
			return
		}
		err := gv.openModel(fpath)
		if err != nil {
			gv.ed.Show(err.Error())
			return
		}
		gv.fs.SetVisible(false)

	})
	gv.fs.Subscribe("OnCancel", func(evname string, ev interface{}) {
		gv.fs.Show(false)
	})
	gv.scene.Add(gv.fs)

	// Creates error dialog
	gv.ed = NewErrorDialog(600, 100)
	gv.scene.Add(gv.ed)

	return nil
}

// openModel try to open the specified model and add it to the scene
func (gv *g3nView) openModel(fpath string) error {

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
		gv.scene.Add(group)
		gv.models = append(gv.models, group)
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
		gv.scene.Add(s)
		gv.models = append(gv.models, s.GetNode())
		return nil
	}
	return fmt.Errorf("Unrecognized model file extension:[%s]", ext)
}

// removeModels removes and disposes of all loaded models in the scene
func (gv *g3nView) removeModels() {

	for i := 0; i < len(gv.models); i++ {
		model := gv.models[i]
		gv.scene.Remove(model)
		model.Dispose()
	}
}

func getIcon(state bool) string {

	if state {
		return checkON
	} else {
		return checkOFF
	}
}

// usage shows the application usage
func usage() {

	fmt.Fprintf(os.Stderr, "usage: g3nview [model1 model2   modelN]\n")
	flag.PrintDefaults()
	os.Exit(2)
}
