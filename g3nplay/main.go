// Copyright 2016 The G3N Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is a minimum G3N command line audio player (no 3D)
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/g3n/engine/audio"
	"github.com/g3n/engine/audio/al"
	"github.com/g3n/engine/audio/vorbis"
)

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

	// Open default audio device
	err := openAudio()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Creates player
	player, err := audio.NewPlayer(fpath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get total play time
	total := player.TotalTime()
	fmt.Printf("Playing:[%s] (%3.1f seconds)\n", fpath, total)

	// Starts player and waits for it to stop
	player.Play()
	for {
		if player.State() == al.Stopped {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// usage shows the application usage
func usage() {

	fmt.Fprintf(os.Stderr, "usage: g3nplay <soundfile>\n")
}

// openAudio opens the defaul audio device
func openAudio() error {

	// Opens default audio device
	dev, err := al.OpenDevice("")
	if dev == nil {
		return fmt.Errorf("Error: %s opening OpenAL default device", err)
	}

	// Creates audio context
	acx, err := al.CreateContext(dev, nil)
	if err != nil {
		return fmt.Errorf("Error creating audio context:%s", err)
	}

	// Makes the context the current one
	err = al.MakeContextCurrent(acx)
	if err != nil {
		return fmt.Errorf("Error setting audio context current:%s", err)
	}

	// Logs library versions
	fmt.Printf("%s version: %s\n", al.GetString(al.Vendor), al.GetString(al.Version))
	fmt.Printf("%s\n", vorbis.VersionString())
	return nil
}
