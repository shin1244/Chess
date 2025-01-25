package ws

import "chess/game"

func BroadcastBoard(g *game.Context, start bool) {
	for conn := range g.PlayerColor {
		conn.WriteJSON(&game.Message{
			Type:        "board",
			Board:       g.Board,
			Start:       start,
			PlayerColor: g.PlayerColor[conn],
			Turn:        g.Turn,
		})
	}
}
