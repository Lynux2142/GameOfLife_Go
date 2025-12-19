package main

import (
	"image/color"
	"math/rand/v2"
	"runtime"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct{
	Width 		int
	Height 		int
	Cells 		[][]uint8
	NextCells 	[][]uint8
	NumCPU 		int
	Running 	bool
}

func NewGame(width, height int) Game {
	game := &Game{
		Width:		width,
		Height: 	height,
		NumCPU: 	runtime.NumCPU(),
		Running: 	true,
	}
	game.Cells = make([][]uint8, height)
	game.NextCells = make([][]uint8, height)
	for j := range(height) {
		game.Cells[j] = make([]uint8, width)
		game.NextCells[j] = make([]uint8, width)
		for i := range(width) {
			game.Cells[j][i] = uint8(rand.IntN(2))
			game.NextCells[j][i] = 0
		}
	}
	return *game
}

func (g *Game) CountNeighbors(y, x int) uint8 {
	yUp := y - 1
	if yUp < 0 {
		yUp = g.Height - 1
	}
	yDown := y + 1
	if yDown >= g.Height {
		yDown = 0
	}

	xLeft := x - 1
	if xLeft < 0 {
		xLeft = g.Width - 1
	}
	xRight := x + 1
	if xRight >= g.Width {
		xRight = 0
	}

	return g.Cells[yUp][xLeft] + g.Cells[yUp][x] + g.Cells[yUp][xRight] +
	g.Cells[y][xLeft] + g.Cells[y][xRight] +
	g.Cells[yDown][xLeft] + g.Cells[yDown][x] + g.Cells[yDown][xRight]
}

func (g *Game) GetNextCycleState(y, x int) {
	neighbors := g.CountNeighbors(y, x)
	if g.Cells[y][x] == 1 && (neighbors < 2 || neighbors > 3) {
		g.NextCells[y][x] = 0
		return
	}
	if (g.Cells[y][x] == 0 && neighbors == 3) {
		g.NextCells[y][x] = 1
		return
	}
	g.NextCells[y][x] = g.Cells[y][x]
}

func (g *Game) NextCycle(num_workers int) {
	var wg sync.WaitGroup
	rows_per_worker := g.Height / num_workers
	for i := range(num_workers) {
		startY := i * rows_per_worker
		endY := startY + rows_per_worker
		if i == num_workers - 1 {
			endY = g.Height
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for y := start; y < end; y++ {
				for x := range(g.Width) {
					g.GetNextCycleState(y, x)
				}
			}
		}(startY, endY)
	}
	wg.Wait()
	g.Cells, g.NextCells = g.NextCells, g.Cells
}

func (g *Game) KeyboardInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.Running = !g.Running
	}
}

func (g *Game) MouseInput() {
	x, y := ebiten.CursorPosition()
	
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.Cells[y][x] = 1
	}
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if g.Running {
		g.NextCycle(g.NumCPU)
	}
	g.KeyboardInput()
	g.MouseInput()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for y := range(g.Height) {
		for x := range(g.Width) {
			if g.Cells[y][x] == 1 {
				vector.FillRect(
					screen,
					float32(x),
					float32(y),
					1,
					1,
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
