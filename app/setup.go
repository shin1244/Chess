package app

type PlayerReady struct {
	Player0 chan bool
	Player1 chan bool
	Piece0  []Piece
	Piece1  []Piece
}

// 초기 기물 설정
func (g *Game) SetupGame(player, x, y int) {
	if player == 0 && len(g.ready.Piece0) > 0 {
		if y == 0 {
			g.ready.Piece0 = g.ready.Piece0[:len(g.ready.Piece0)]
			if len(g.ready.Piece0) == 0 {
				g.ready.Player0 <- true
			}
		}
	} else if player == 1 && len(g.ready.Piece1) > 0 {
		if y == 7 {
			g.ready.Piece1 = g.ready.Piece1[:len(g.ready.Piece1)]
			if len(g.ready.Piece1) == 0 {
				g.ready.Player1 <- true
			}
		}
	}
}

// 두 플레이어 모두 초기 기물 설정 완료
func (g *Game) SetupGameDone() int {
	if <-g.ready.Player0 && <-g.ready.Player1 {
		return 1
	}
	return 0
}
