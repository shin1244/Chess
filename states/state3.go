package states

import (
	"chess/game"
	"chess/ws"
)

func State3(g *game.Context, message game.Message) {
	if message.Type != "restart" {
		return
	}
	*g = *game.InitGame()
	ws.BroadcastBoard(g, false)
}
