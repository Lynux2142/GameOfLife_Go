package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	WINDOW_WIDTH  	= 1260
	WINDOW_HEIGHT 	= 960
	CELL_SIZE	 	= 1
)

func main() {
	game := NewGame(WINDOW_WIDTH, WINDOW_HEIGHT, CELL_SIZE)
	ebiten.SetWindowSize(game.Width, game.Height)
	if err := ebiten.RunGame(&game); err != nil {
		log.Fatal(err)
	}
}
