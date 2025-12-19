package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	CELL_SIZE = 1
)

func main() {
	width, height := ebiten.Monitor().Size()
	ebiten.SetWindowSize(width, height)
	ebiten.SetFullscreen(true)
	game := NewGame(width, height, CELL_SIZE)
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
