package ws

import (
	"chess/game"

	"github.com/gorilla/websocket"
)

func BroadcastBoard(g *game.Context, soundType int) {
	for conn := range g.PlayerColor {
		conn.WriteJSON(&game.Message{
			Type:          "board",
			Board:         g.Board,
			SoundType:     soundType,
			PlayerColor:   g.PlayerColor[conn],
			Turn:          g.Turn,
			PrintingTiles: g.PrintingTiles,
		})
	}
}

func BroadcastGameOver(g *game.Context, conn *websocket.Conn, piece string) {
	for conn := range g.PlayerColor {
		conn.WriteJSON(&game.Message{
			Type:          "gameOver",
			PlayerColor:   g.Turn,
			Piece:         piece,
			PrintingTiles: g.PrintingTiles,
		})
	}
	g.GameState = 3
}
