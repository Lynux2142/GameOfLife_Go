package main

import (
	"runtime"
	"sync"
	"image/color"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Grid struct {
	Width  		int
	Height 		int
	Cells  		[][]uint8
	NextCells 	[][]uint8
}

func NewGrid(width, height int) Grid {
	grid := &Grid{
		Width:  width,
		Height: height,
	}
	grid.Cells = make([][]uint8, height)
	grid.NextCells = make([][]uint8, height)
	for i := range grid.Cells {
		grid.Cells[i] = make([]uint8, width)
		grid.NextCells[i] = make([]uint8, width)
		for j := range grid.Cells[i] {
			grid.Cells[i][j] = uint8(rand.IntN(2))
			grid.NextCells[i][j] = 0
		}
	}
	return *grid
}

func (g *Grid) CountNeighbors(y, x int) uint8 {
	var count uint8 = 0
	for j := -1; j <= 1; j++ {
		for i := -1; i <= 1; i++ {
			if i == 0 && j == 0 { continue }
			ni, nj := x+i, y+j
			if ni >= 0 && ni < g.Width && nj >= 0 && nj < g.Height {
				count += g.Cells[nj][ni]
			}
		}
	}
	return count
}

func (g *Grid) GetNextCycleState(y, x int) {
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

func (g *Grid) NextCycle() {
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			g.GetNextCycleState(y, x)
		}
	}
	g.Cells, g.NextCells = g.NextCells, g.Cells
}

func (g *Grid) NextCycleMultiThreaded() {
	numCPU := runtime.NumCPU()
	var wg sync.WaitGroup
	rowsPerWorker := g.Height / numCPU
	for i := 0; i < numCPU; i++ {
		startY := i * rowsPerWorker
		endY := (i + 1) * rowsPerWorker
		if i == numCPU-1 {
			endY = g.Height
		}
		wg.Add(1)
		go func(sY, eY int) {
			defer wg.Done()
			for y := sY; y < eY; y++ {
				for x := 0; x < g.Width; x++ {
					g.GetNextCycleState(y, x)
				}
			}
		}(startY, endY)
	}
	wg.Wait()
	g.Cells, g.NextCells = g.NextCells, g.Cells
}

type Game struct{
	Width 	int
	Height 	int
	Size 	int
	Grid 	Grid
}

func NewGame(width, height, size int) Game {
	game := &Game{
		Width:  width,
		Height: height,
		Size: 	size,
	}
	game.Grid = NewGrid(width/size, height/size)
	return *game
}

func (g *Game) Update() error {
	//g.Grid.NextCycle()
	g.Grid.NextCycleMultiThreaded()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for y := int(0); y < g.Grid.Height; y++ {
		for x := int(0); x < g.Grid.Width; x++ {
			cell_color := color.Black
			if g.Grid.Cells[y][x] == 1 {
				cell_color = color.White
			}
			vector.FillRect(
				screen,
				float32(x*g.Size),
				float32(y*g.Size),
				float32(g.Size),
				float32(g.Size),
				cell_color,
				false,
			)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.Width, g.Height
}
