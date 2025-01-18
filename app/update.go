package app

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) Update() error {
	// 게임 기물 준비 상태
	if g.gameState == 0 {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			// 마우스 클릭 좌표 얻기
			mouseX, mouseY := ebiten.CursorPosition()

			// 타일 크기 (Draw 함수에서 사용한 값과 동일하게)
			tileSize := 80

			// 보드 좌표 계산
			boardX := mouseX / tileSize
			boardY := mouseY / tileSize

			// SetupGame 호출
			SetupGame(&ready, g.playerTurn, boardX, boardY)
		}
	}
	return nil
}
