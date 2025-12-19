package main

import (
	"math/rand/v2"
	"runtime"
	"sync"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct{
	Width 		int
	Height 		int
	Cells 		[]uint8
	NextCells 	[]uint8
	Pixel 		[]byte
	Pixel32 	[]uint32
	NumCPU 		int
	Running 	bool
}

func NewGame(width, height int) Game {
	game := &Game{
		Width:		width,
		Height: 	height,
		Cells:		make([]uint8, width * height),
		NextCells:	make([]uint8, width * height),
		Pixel:		make([]byte, width * height << 2),
		Pixel32:	nil,
		NumCPU: 	runtime.NumCPU(),
		Running: 	true,
	}
	game.Pixel32 = unsafe.Slice((*uint32)(unsafe.Pointer(&game.Pixel[0])), width * height)
	for i := range(width * height) {
		game.Cells[i] = uint8(rand.IntN(2))
	}
	return *game
}

func (g *Game) updateRange(startY, endY int) {
	for y := startY; y < endY; y++ {
		yUp    := ((y - 1 + g.Height) % g.Height) * g.Width
		yMid   := y * g.Width
		yDown  := ((y + 1) % g.Height) * g.Width
		for x := 0; x < g.Width; x++ {
			xLeft  := (x - 1 + g.Width) % g.Width
			xRight := (x + 1) % g.Width
			neighbors := g.Cells[yUp+xLeft] + g.Cells[yUp+x] + g.Cells[yUp+xRight] +
			g.Cells[yMid+xLeft] + g.Cells[yMid+xRight] +
			g.Cells[yDown+xLeft] + g.Cells[yDown+x] + g.Cells[yDown+xRight]
			idx := yMid + x
			if g.Cells[idx] == 1 && (neighbors < 2 || neighbors > 3) {
				g.NextCells[idx] = 0
				continue
			}
			if g.Cells[idx] == 0 && neighbors == 3 {
				g.NextCells[idx] = 1
				continue
			}
			g.NextCells[idx] = g.Cells[idx]
		}
	}
}

func (g *Game) NextCycle() {
	var wg sync.WaitGroup
	rowsPerWorker := g.Height / g.NumCPU
	for i := 0; i < g.NumCPU; i++ {
		startY := i * rowsPerWorker
		endY := startY + rowsPerWorker
		if i == g.NumCPU - 1 {
			endY = g.Height
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			g.updateRange(s, e)
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
	i := y * g.Width + x
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.Cells[i] = 1
	}
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if g.Running { g.NextCycle() }
	g.KeyboardInput()
	g.MouseInput()
	return nil
}

func (g *Game) renderRange(startY, endY int) {
	for y := startY; y < endY; y++ {
		rowOffset := y * g.Width
		for x := range(g.Width) {
			idx := rowOffset + x
			g.Pixel32[idx] = 0xFFFFFFFF * uint32(g.Cells[idx])
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	var wg sync.WaitGroup
	rowsPerWorker := g.Height / g.NumCPU
	for i := 0; i < g.NumCPU; i++ {
		startY := i * rowsPerWorker
		endY := startY + rowsPerWorker
		if i == g.NumCPU - 1 {
			endY = g.Height
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			g.renderRange(s, e)
		}(startY, endY)
	}
	wg.Wait()
	screen.WritePixels(g.Pixel)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.Width, g.Height
}
