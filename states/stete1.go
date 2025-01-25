package states

import (
	"chess/game"
	"chess/ws"
	"log"

	"github.com/gorilla/websocket"
)

// 게임이 시작되면 플레이어는 체스말을 배치합니다. 체스말을 모두 배치하면 state2로 이동합니다.
func State1(g *game.Context, conn *websocket.Conn, message game.Message) {
	if message.Type == "click" && message.PlayerColor == g.Turn {
		if placePiece(g, conn, message) {
			g.Turn = (g.Turn + 1) % 2
			ws.BroadcastBoard(g, false)
		}
		if len(g.Pieces[0]) == 0 && len(g.Pieces[1]) == 0 {
			g.GameState = 2
			log.Println("게임 시작")
		}
	}
}

// 체스말을 배치하는 함수입니다. 내부에서 체스 말을 배치할 수 있는지 확인하고 가능하다면 체스 말을 배치합니다.
func placePiece(g *game.Context, conn *websocket.Conn, message game.Message) bool {
	if canPlace(g, conn, message) {
		piece := getPiece(g, conn)

		g.Board[message.Position.Row][message.Position.Col].Piece = piece
		g.Board[message.Position.Row][message.Position.Col].Color = g.PlayerColor[conn]

		setupMessage := &game.Message{
			Type:     "board",
			Board:    g.Board,
			Position: message.Position,
			Piece:    piece,
		}
		conn.WriteJSON(setupMessage)
		return true
	}
	log.Println("잘못된 위치 혹은 기물 없음")
	return false
}

// 해당 위치에 체스말을 배치할 수 있는지 확인하는 함수입니다.
func canPlace(g *game.Context, conn *websocket.Conn, message game.Message) bool {
	if len(g.Pieces[g.PlayerColor[conn]]) > 0 {
		if (g.PlayerColor[conn] == 0 && message.Position.Row == 7) ||
			(g.PlayerColor[conn] == 1 && message.Position.Row == 0) {
			if g.Board[message.Position.Row][message.Position.Col].Piece == "" {
				return true
			}
		}
	}
	return false
}

// 배치할 순서의 체스말을 가져오는 함수입니다.
func getPiece(g *game.Context, conn *websocket.Conn) string {
	piece := ""
	if g.PlayerColor[conn] == 0 {
		piece = "white" + g.Pieces[g.PlayerColor[conn]][0]
	} else {
		piece = "black" + g.Pieces[g.PlayerColor[conn]][0]
	}
	g.Pieces[g.PlayerColor[conn]] = g.Pieces[g.PlayerColor[conn]][1:]

	return piece
}
