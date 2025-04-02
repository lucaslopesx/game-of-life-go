package main

import (
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 600

	Rows     = 40
	Cols     = 60
	CellSize = 12

	UpdateInterval = 100
	Ratio          = 0.3
)

const (
	Paused GameState = iota
	Running
)

var (
	ColorBackground = color.RGBA{0, 0, 0, 255}
	ColorGridLine   = color.RGBA{40, 40, 40, 255}
	ColorLiveCell   = color.RGBA{240, 240, 240, 255}
)

type GameState int

type Game struct {
	grid             *Grid
	gridOffsetX      int
	gridOffsetY      int
	touchIds         []ebiten.TouchID
	lastTouchedCellX int
	lastTouchedCellY int
	state            GameState
	lastUpdateTime   time.Time
	rng              *rand.Rand
}

type Cell struct {
	Alive bool
}

type Grid struct {
	Cells [][]Cell
}

var directions = []struct{ dx, dy int }{
	{-1, -1}, {0, -1}, {1, -1},
	{-1, 0}, {1, 0},
	{-1, 1}, {0, 1}, {1, 1},
}

func (g *Grid) Randomize(rng *rand.Rand, ratio float64) {
	for x := range g.Cells {
		for y := range g.Cells[x] {
			if rng.Float64() < ratio {
				g.Cells[x][y].Alive = true
			}
		}
	}
}

func NewGrid() *Grid {
	cells := make([][]Cell, Cols)
	for x := range cells {
		cells[x] = make([]Cell, Rows)
	}

	return &Grid{
		Cells: cells,
	}
}

func (g *Game) handleKeyboardInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if g.state == Running {
			g.state = Paused
		} else {
			g.state = Running
		}
	}
}

func (g *Game) handleMouseInput() {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()

		gridX := (mx - g.gridOffsetX) / CellSize
		gridY := (my - g.gridOffsetY) / CellSize

		if gridX < 0 || gridX >= Cols || gridY < 0 || gridY >= Rows {
			return
		}

		if gridX != g.lastTouchedCellX || gridY != g.lastTouchedCellY {
			g.lastTouchedCellX = gridX
			g.lastTouchedCellY = gridY
			g.grid.Cells[gridX][gridY].Alive = !g.grid.Cells[gridX][gridY].Alive
		}
	} else {
		g.lastTouchedCellX = -1
		g.lastTouchedCellY = -1
	}
}

func (grid *Grid) CalculateNextGeneration() {
	newCellsGeneration := make(map[int]map[int]bool)
	for x := 0; x < Cols; x++ {
		newCellsGeneration[x] = make(map[int]bool)
		for y := 0; y < Rows; y++ {
			newCellsGeneration[x][y] = false
		}
	}

	for x, cells := range grid.Cells {
		for y, cell := range cells {
			liveNeighboors := grid.CountNeighboors(x, y)
			if cell.Alive && liveNeighboors < 2 {
				continue
			}

			if cell.Alive && (liveNeighboors == 2 || liveNeighboors == 3) {
				newCellsGeneration[x][y] = true
				continue
			}

			if cell.Alive && liveNeighboors > 3 {
				continue
			}

			if !cell.Alive && liveNeighboors == 3 {
				newCellsGeneration[x][y] = true
				continue
			}
		}
	}

	for x, cells := range grid.Cells {
		for y := range cells {
			grid.Cells[x][y].Alive = newCellsGeneration[x][y]
		}
	}
}

func (grid *Grid) CountNeighboors(x, y int) int {
	count := 0
	for _, v := range directions {
		nx, ny := v.dx+x, v.dy+y
		if nx < 0 || nx >= Cols || ny < 0 || ny >= Rows {
			continue
		}
		if grid.Cells[nx][ny].Alive {
			count++
		}
	}

	return count
}

func (g *Game) drawCells(screen *ebiten.Image) {
	for x, row := range g.grid.Cells {
		for y, cell := range row {

			posX := float32(g.gridOffsetX + x*CellSize)
			posY := float32(g.gridOffsetY + y*CellSize)

			if cell.Alive {
				vector.DrawFilledRect(
					screen,
					posX,
					posY,
					CellSize,
					CellSize,
					ColorLiveCell,
					false,
				)
			}
		}
	}
}

func (g *Game) drawGrid(screen *ebiten.Image) {
	for y := 0; y <= Rows; y++ {
		lineY := float32(g.gridOffsetY + y*CellSize)
		vector.StrokeLine(
			screen,
			float32(g.gridOffsetX),
			lineY,
			float32(g.gridOffsetX+Cols*CellSize),
			lineY,
			1.0,
			ColorGridLine,
			false,
		)
	}

	for x := 0; x <= Cols; x++ {
		lineX := float32(g.gridOffsetX + x*CellSize)
		vector.StrokeLine(
			screen,
			lineX,
			float32(g.gridOffsetY),
			lineX,
			float32(g.gridOffsetY+Rows*CellSize),
			1.0,
			ColorGridLine,
			false,
		)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(ColorBackground)

	g.drawCells(screen)
	g.drawGrid(screen)
}

func (g *Game) Update() error {
	g.handleMouseInput()
	g.handleKeyboardInput()

	if g.state == Paused {
		return nil
	}

	if time.Since(g.lastUpdateTime).Milliseconds() < UpdateInterval {
		return nil
	}

	g.grid.CalculateNextGeneration()

	g.lastUpdateTime = time.Now()

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func NewGame() *Game {
	grid := NewGrid()

	gridOffsetX := (ScreenWidth - Cols*CellSize) / 2
	gridOffsetY := (ScreenHeight - Rows*CellSize) / 2

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	grid.Randomize(rng, Ratio)

	return &Game{
		grid:             grid,
		gridOffsetX:      gridOffsetX,
		gridOffsetY:      gridOffsetY,
		lastTouchedCellX: -1,
		lastTouchedCellY: -1,
		lastUpdateTime:   time.Now(),
		state:            Running,
		rng:              rng,
	}
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Conway's Game of Life!")
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
