package main

import (
	"math/rand/v2"
	"sync"
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
	for j := range grid.Cells {
		grid.Cells[j] = make([]uint8, width)
		grid.NextCells[j] = make([]uint8, width)
		for i := range grid.Cells[j] {
			grid.Cells[j][i] = uint8(rand.IntN(2))
			grid.NextCells[j][i] = 0
		}
	}
	return *grid
}

func (g *Grid) CountNeighbors(y, x int) uint8 {
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

func (g *Grid) NextCycle(num_workers int) {
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
