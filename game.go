package main

import (
	"math"
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
	LastMousePos 	struct{ X, Y int }
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
		LastMousePos: struct{ X, Y int }{ X: -1, Y: -1 },
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

func (g *Game) Reset() {
	for i := range(g.Width * g.Height) {
		g.Cells[i] = uint8(rand.IntN(2))
	}
}

func (g *Game) Clear() {
	g.Cells = make([]uint8, g.Width * g.Height)
}

func (g *Game) AddCell(x, y int) {
	i := y * g.Width + x
	g.Cells[i] = 1
}

func (g *Game) AddRandomCells(n int) {
	for _ = range(n) {
		g.AddCell(rand.IntN(g.Width), rand.IntN(g.Height))
	}
}

func (g *Game) KeyboardInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.Running = !g.Running
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.Clear()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.Reset()
	}
}

func (g *Game) DrawLine(x0, y0, x1, y1 int) {
	dx := int(math.Abs(float64(x1 - x0)))
	dy := int(math.Abs(float64(y1 - y0)))
	sx := -1
	sy := -1
	if x0 < x1 { sx = 1 }
	if y0 < y1 { sy = 1 }
	err := dx - dy
	for {
		g.AddCell(x0, y0)
		if x0 == x1 && y0 == y1 { break }
		err2 := err << 1
		if err2 > -dy {
			err -= dy
			x0 += sx
		}
		if err2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func (g *Game) MouseInput() {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x2, y2 := ebiten.CursorPosition()
		x1 := func() int {
			if g.LastMousePos.X == -1 {
				g.LastMousePos.X = x2
			}
			return g.LastMousePos.X
		}()
		y1 := func() int {
			if g.LastMousePos.Y == -1 {
				g.LastMousePos.Y = y2
			}
			return g.LastMousePos.Y
		}()
		g.DrawLine(x1, y1, x2, y2)
		g.LastMousePos.X = x2
		g.LastMousePos.Y = y2
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.LastMousePos.X = -1
		g.LastMousePos.Y = -1
	}
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if g.Running {
		g.NextCycle()
		g.AddRandomCells(10)
	}
	g.KeyboardInput()
	g.MouseInput()
	return nil
}

func (g *Game) renderRange(startY, endY int) {
	for y := startY; y < endY; y++ {
		rowOffset := y * g.Width
		for x := range(g.Width) {
			idx := rowOffset + x
			color := uint32(
				y * 0xFF / g.Height << 16 |
				x * 0xFF / g.Width << 8 |
				0xFF,
			)
			g.Pixel32[idx] = color * uint32(g.Cells[idx])
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
