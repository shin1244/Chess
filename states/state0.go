package states

import (
	"chess/game"
	"chess/ws"
	"log"

	"github.com/gorilla/websocket"
)

// 유저가 게임에 참여하기를 기다립니다. 2명이 참여하면 state1로 이동합니다.
func State0(g *game.Context, conn *websocket.Conn, message game.Message) {
	if message.Type != "join" {
		return
	}
	log.Println("게임 참가", conn.RemoteAddr())
	g.PlayerColor[conn] = len(g.PlayerColor)
	conn.WriteJSON(&game.Message{
		Type:        "color",
		PlayerColor: g.PlayerColor[conn],
	})
	if len(g.PlayerColor) == 2 {
		conn.WriteJSON(&game.Message{
			Type:  "board",
			Board: g.Board,
		})
		g.GameState = 1
		ws.BroadcastBoard(g, true)
	}
}
