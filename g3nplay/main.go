package main

import (
	"flag"
	"fmt"
	"github.com/g3n/engine/app"
	"github.com/g3n/engine/audio"
	"github.com/g3n/engine/renderer"
	"os"
	"time"
)

// usage shows the application usage
func usage() {

	fmt.Fprintf(os.Stderr, "usage: g3nplay <soundfile>\n")
}

func main() {

	// Parse command line parameters
	flag.Usage = usage
	flag.Parse()

	// Get file to play
	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}
	fpath := args[0]

	// Create application
	app.App()

	// Create player
	player, err := audio.NewPlayer(fpath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get total play time
	total := player.TotalTime()
	fmt.Printf("Playing:[%s] (%3.1f seconds)\n", fpath, total)

	// Start player
	player.Play()

	// Run the application
	app.App().Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {})
}
