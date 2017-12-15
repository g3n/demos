package main

import (
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
)

type ErrorDialog struct {
	gui.Panel
	msg *gui.ImageLabel
	bok *gui.Button
}

func NewErrorDialog(width, height float32) *ErrorDialog {

	e := new(ErrorDialog)
	e.Initialize(width, height)
	e.SetVisible(false)
	e.SetBorders(2, 2, 2, 2)
	e.SetPaddings(4, 4, 4, 4)
	e.SetColor(math32.NewColor("White"))

	// Set vertical box layout for the whole panel
	l := gui.NewVBoxLayout()
	l.SetSpacing(4)
	e.SetLayout(l)

	// Creates error message label
	e.msg = gui.NewImageLabel("")
	e.msg.SetFontSize(18)
	e.msg.SetLayoutParams(&gui.VBoxLayoutParams{Expand: 2, AlignH: gui.AlignWidth})
	e.Add(e.msg)

	// Creates button
	e.bok = gui.NewButton("OK")
	e.bok.SetLayoutParams(&gui.VBoxLayoutParams{Expand: 1, AlignH: gui.AlignCenter})
	e.bok.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		e.SetVisible(false)
	})
	e.Add(e.bok)

	return e
}

func (e *ErrorDialog) Show(msg string) {

	e.msg.SetText(msg)
	e.SetVisible(true)
}
