package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	WINDOW_WIDTH  	= 1600
	WINDOW_HEIGHT 	= 1200
	CELL_SIZE	 	= 2
)

func main() {
	game := NewGame(WINDOW_WIDTH, WINDOW_HEIGHT, CELL_SIZE)
	ebiten.SetWindowSize(game.Width, game.Height)
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
