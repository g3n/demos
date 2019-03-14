# G3N Standalone demos

This repository contains standalone demo programs
for the [G3N](https://github.com/g3n/engine) Go 3D Game Engine.

## Install with Go Modules (Go 1.11 or higher)

    git clone https://github.com/g3n/demos
    cd demos
    go install ./...

## Install without Go Modules (Before Go 1.11)

    go get -u github.com/g3n/demos
    cd ~/go/src/github.com/g3n/demos
    go install ./...

## Run

The gopher3d demo requires artifacts from gopher3d directory:

    cd ~/go/src/github.com/g3n/demos/gopher3d
    gopher3d

You can run the other demos directly from any location:

    g3nplay
    g3nview
    gui_button
    hellog3n
    hellog3n-no-app

