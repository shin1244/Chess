package app

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) Draw(screen *ebiten.Image) {
	tileSize := float32(80)

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			x := float32(col) * tileSize
			y := float32(row) * tileSize

			piece := g.board[row][col]

			switch piece.Color {
			case 0:
				vector.DrawFilledRect(screen, x, y, tileSize, tileSize, color.RGBA{204, 132, 0, 255}, false)
			case 1:
				vector.DrawFilledRect(screen, x, y, tileSize, tileSize, color.RGBA{0, 204, 0, 255}, false)
			case 2:
				vector.DrawFilledRect(screen, x, y, tileSize, tileSize, color.RGBA{181, 136, 99, 255}, false)
			case 3:
				vector.DrawFilledRect(screen, x, y, tileSize, tileSize, color.RGBA{40, 40, 40, 255}, false)
			}
		}
	}
}
