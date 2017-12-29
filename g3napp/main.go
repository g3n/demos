package main

import (
	"github.com/g3n/engine/util/app"
	"github.com/g3n/engine/util/logger"
)

func main() {

	// Creates application
	app, err := app.Create("G3NApp", app.Options{
		WinWidth:     800,
		WinHeight:    600,
		VersionMajor: 0,
		VersionMinor: 1,
		LogLevel:     logger.DEBUG,
		EnableFlags:  true,
	})
	if err != nil {
		panic(err)
	}

	// Runs main loop
	app.Run()

}
