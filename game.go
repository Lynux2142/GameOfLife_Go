package main

import (
	"image/color"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct{
	Width 	int
	Height 	int
	Size 	int
	Grid 	Grid
	NumCPU 	int
}

func NewGame(width, height, size int) Game {
	game := &Game{
		Width:  width,
		Height: height,
		Size: 	size,
		NumCPU: runtime.NumCPU(),
	}
	game.Grid = NewGrid(width/size, height/size)
	return *game
}

func (g *Game) Update() error {
	g.Grid.NextCycle(g.NumCPU)
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for y := range(g.Grid.Height) {
		for x := range(g.Grid.Width) {
			if g.Grid.Cells[y][x] == 1 {
				vector.FillRect(
					screen,
					float32(x*g.Size),
					float32(y*g.Size),
					float32(g.Size),
					float32(g.Size),
					color.White,
					false,
				)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.Width, g.Height
}
