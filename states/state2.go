package states

import (
	"chess/game"
	"chess/ws"

	"github.com/gorilla/websocket"
)

func State2(g *game.Context, conn *websocket.Conn, message game.Message) {
	if message.Type == "click" {
		playerColor := g.PlayerColor[conn]
		selectTilePiece := g.Board[message.Position.Row][message.Position.Col].Piece
		if selectTilePiece == "whiteKing" || selectTilePiece == "blackKing" {
			return
		}
		selectTileColor := g.Board[message.Position.Row][message.Position.Col].Color
		if g.Turn == playerColor {
			if len(g.PossibleMoves) == 0 { // 첫 클릭일 때
				if selectTilePiece != "" && selectTileColor == playerColor {
					g.PossibleMoves = calculatePossibleMoves(g, selectTilePiece, message.Position.Row, message.Position.Col)
					g.SelectedPiece = message.Position

					// 킹의 이동 가능 위치 추가
					kingMove := calculateKingMove(g, selectTilePiece, conn)
					if kingMove != nil {
						g.PossibleMoves = append(g.PossibleMoves, *kingMove)
					}

					conn.WriteJSON(&game.Message{
						Type:      "click",
						Positions: g.PossibleMoves,
					})
				}
			} else { // 두번째 클릭일 때
				for _, move := range g.PossibleMoves {
					if move.Row == message.Position.Row && move.Col == message.Position.Col {
						paintPath(g, g.SelectedPiece.Row, g.SelectedPiece.Col, message.Position.Row, message.Position.Col, g.Turn)          // 경로 색칠
						g.Board[message.Position.Row][message.Position.Col].Piece = g.Board[g.SelectedPiece.Row][g.SelectedPiece.Col].Piece // 보드에서 기물 이동
						g.Board[g.SelectedPiece.Row][g.SelectedPiece.Col].Piece = ""
						g.PossibleMoves = []game.Position{}
						if moveKing(g, conn, g.Board[message.Position.Row][message.Position.Col].Piece) {
							g.GameState = 3
							for conn := range g.PlayerColor {
								conn.WriteJSON(&game.Message{
									Type:        "gameOver",
									PlayerColor: g.Turn,
									Piece:       "Pawn",
								})
							}
						}
						g.Turn = (g.Turn + 1) % 2

						// 3가지를 체크해야함
						// 1. 둘러 싸인 기물이 있는지
						// 2. 색칠을 완료했는지
						// 3. 폰이 움직이지 못하는지
						checkGameOver(g)

						break
					}
				}
				if len(g.PossibleMoves) != 0 {
					g.PossibleMoves = []game.Position{}
					ws.BroadcastBoard(g, false)
				}
			}
		}
	}
}

func calculatePossibleMoves(g *game.Context, piece string, row, col int) []game.Position {
	pieceType := piece[5:]
	possibleMoves := []game.Position{}

	for _, direction := range game.Directions[pieceType] {
		if pieceType == "Rook" || pieceType == "Bishop" {
			for i := 1; i < 8; i++ {
				newRow := row + direction.Row*i
				newCol := col + direction.Col*i
				if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 && g.Board[newRow][newCol].Piece == "" {
					possibleMoves = append(possibleMoves, game.Position{Row: newRow, Col: newCol, Piece: piece})
				} else {
					break
				}
			}
		} else {
			newRow := row + direction.Row
			newCol := col + direction.Col
			if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 && g.Board[newRow][newCol].Piece == "" {
				possibleMoves = append(possibleMoves, game.Position{Row: newRow, Col: newCol, Piece: piece})
			}
		}
	}
	return possibleMoves
}

// 이동한 경로를 색칠하는 함수
func paintPath(g *game.Context, row, col, endRow, endCol, color int) {
	// 이동 방향 계산
	rowDir := 0
	if endRow-row > 0 {
		rowDir = 1
	} else if endRow-row < 0 {
		rowDir = -1
	}

	colDir := 0
	if endCol-col > 0 {
		colDir = 1
	} else if endCol-col < 0 {
		colDir = -1
	}

	// 현재 위치의 기물 타입 확인
	piece := g.Board[row][col].Piece
	pieceType := piece[5:] // "white" 또는 "black" 제거

	// 나이트의 경우 'ㄱ' 모양으로 경로 색칠
	if pieceType == "Knight" {
		g.Board[row][col].Color = color // 시작점

		// 2칸 이동 먼저 (수직 또는 수평)
		if abs(endRow-row) == 2 {
			// 수직으로 2칸 이동
			intermediateRow := row + rowDir // 중간 칸
			if g.Board[intermediateRow][col].Piece == "" {
				g.Board[intermediateRow][col].Color = color
			}

			intermediateRow = row + rowDir*2 // 2칸 이동 후
			if g.Board[intermediateRow][col].Piece == "" {
				g.Board[intermediateRow][col].Color = color
			}

			// 그 다음 수평으로 1칸 이동
			g.Board[intermediateRow][endCol].Color = color
		} else {
			// 수평으로 2칸 이동
			intermediateCol := col + colDir // 중간 칸
			if g.Board[row][intermediateCol].Piece == "" {
				g.Board[row][intermediateCol].Color = color
			}

			intermediateCol = col + colDir*2 // 2칸 이동 후
			if g.Board[row][intermediateCol].Piece == "" {
				g.Board[row][intermediateCol].Color = color
			}

			// 그 다음 수직으로 1칸 이동
			g.Board[endRow][intermediateCol].Color = color
		}

		return
	}

	// 룩, 비숍, 킹의 경우 경로 색칠
	currentRow := row
	currentCol := col

	for currentRow != endRow || currentCol != endCol {
		g.Board[currentRow][currentCol].Color = color

		// 대각선 이동 (비숍)
		if rowDir != 0 && colDir != 0 {
			currentRow += rowDir
			currentCol += colDir
			// 수직 이동 (룩)
		} else if rowDir != 0 {
			currentRow += rowDir
			// 수평 이동 (룩)
		} else if colDir != 0 {
			currentCol += colDir
		}
	}
	// 도착 지점 색칠
	g.Board[endRow][endCol].Color = color
}

// 절대값 계산을 위한 헬퍼 함수 추가
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func checkGameOver(g *game.Context) {
	if result := dieCheck(g); result != -1 {
		g.GameState = 3
		for conn := range g.PlayerColor {
			conn.WriteJSON(&game.Message{
				Type:        "gameOver",
				PlayerColor: result,
				Piece:       "King",
			})
		}
	} else if result := paintCheck(g); result != -1 {
		g.GameState = 3
		for conn := range g.PlayerColor {
			conn.WriteJSON(&game.Message{
				Type:        "gameOver",
				PlayerColor: result,
				Piece:       "Rook",
			})
		}
	}
	ws.BroadcastBoard(g, true)
}

// 죽은 기물 확인 폰은 안죽음
func dieCheck(g *game.Context) int {
	var checkPos = []game.Position{
		{Row: 1, Col: 0}, {Row: -1, Col: 0}, {Row: 0, Col: 1}, {Row: 0, Col: -1},
	}
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if g.Board[i][j].Piece != "" {
				pieceColor := g.Board[i][j].Piece[0:5]
				pieceType := g.Board[i][j].Piece[5:]
				targetColor := 1
				if pieceColor == "black" {
					targetColor = 0
				}

				// 상하좌우에 빈 공간이나 다른 색이 있는지 확인
				isTrapped := true
				for _, position := range checkPos {
					newRow := i + position.Row
					newCol := j + position.Col
					if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 {
						// 주변 타일이 적의 색이 아니면 포위되지 않은 것
						if g.Board[newRow][newCol].Color != targetColor {
							isTrapped = false
							break
						}
					}
				}

				if isTrapped {
					if pieceType == "King" {
						g.Board[i][j].Piece = ""
						return (targetColor + 1) % 2
					}
					g.Board[i][j].Piece = ""
				}
			}
		}
	}
	return -1
}

// 목표를 모두 색칠했는지 확인
func paintCheck(g *game.Context) int {
	for i := 0; i < 2; i++ {
		allPainted := true
		for j := 0; j < 8; j++ {
			for k := 0; k < 8; k++ {
				if g.Board[j][k].Goal == i && g.Board[j][k].Color != i {
					allPainted = false
					break
				}
			}
		}
		if allPainted {
			return i
		}
	}
	return -1
}

func countResult(g *game.Context) {
	result := 0
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if g.Board[i][j].Color == 0 {
				result += 1
			} else if g.Board[i][j].Color == 1 {
				result -= 1
			}
		}
	}
	if result > 0 {
		for conn := range g.PlayerColor {
			conn.WriteJSON(&game.Message{
				Type:        "gameOver",
				PlayerColor: 0,
				Piece:       "Pawn",
			})
		}
	} else {
		for conn := range g.PlayerColor {
			conn.WriteJSON(&game.Message{
				Type:        "gameOver",
				PlayerColor: 1,
				Piece:       "Pawn",
			})
		}
	}
}

// 턴 종료 시 킹 이동
func moveKing(g *game.Context, conn *websocket.Conn, piece string) bool {
	color := g.PlayerColor[conn]
	moveDirection := getMoveDirection(piece)

	if color == 0 {
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				if g.Board[i][j].Piece == "whiteKing" {
					newRow := i + moveDirection[0]
					newCol := j + moveDirection[1]
					// 보드 경계 체크
					if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 && g.Board[newRow][newCol].Piece == "" {
						g.Board[i][j].Piece = ""
						g.Board[newRow][newCol].Piece = "whiteKing"
						g.Board[newRow][newCol].Color = 0
						return newRow == 0
					}
				}
			}
		}
	} else {
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				if g.Board[i][j].Piece == "blackKing" {
					newRow := i - moveDirection[0]
					newCol := j - moveDirection[1]
					// 보드 경계 체크
					if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 && g.Board[newRow][newCol].Piece == "" {
						g.Board[i][j].Piece = ""
						g.Board[newRow][newCol].Piece = "blackKing"
						g.Board[newRow][newCol].Color = 1
						return newRow == 7
					}
				}
			}
		}
	}
	return false
}

func getMoveDirection(piece string) []int {
	pieceType := piece[5:]
	if pieceType == "Rook" {
		return []int{-1, 0}
	} else if pieceType == "Bishop" {
		return []int{0, -1}
	} else if pieceType == "Knight" {
		return []int{0, 1}
	}
	return []int{0, 0}
}

// 킹의 이동 가능 위치를 계산하는 함수
func calculateKingMove(g *game.Context, piece string, conn *websocket.Conn) *game.Position {
	color := g.PlayerColor[conn]
	moveDirection := getMoveDirection(piece)

	if color == 0 {
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				if g.Board[i][j].Piece == "whiteKing" {
					newRow := i + moveDirection[0]
					newCol := j + moveDirection[1]
					// 보드 경계 체크
					if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 && g.Board[newRow][newCol].Piece == "" {
						return &game.Position{Row: newRow, Col: newCol, Piece: "whiteKing"}
					}
					return nil
				}
			}
		}
	} else {
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				if g.Board[i][j].Piece == "blackKing" {
					newRow := i - moveDirection[0]
					newCol := j - moveDirection[1]
					// 보드 경계 체크
					if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 && g.Board[newRow][newCol].Piece == "" {
						return &game.Position{Row: newRow, Col: newCol, Piece: "blackKing"}
					}
					return nil
				}
			}
		}
	}
	return nil
}
