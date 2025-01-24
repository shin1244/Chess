package main

import (
	"log"

	"github.com/gorilla/websocket"
)

func waitState(conn *websocket.Conn, message Message) {
	if message.Type != "join" {
		return
	}
	log.Println("게임 참가", conn.RemoteAddr())
	playerColor[conn] = len(playerColor)
	conn.WriteJSON(&Message{
		Type:        "color",
		PlayerColor: playerColor[conn],
	})
	if len(playerColor) == 2 {
		conn.WriteJSON(&Board{
			Type:  "board",
			Board: board,
			Goal: func() []Position {
				if playerColor[conn] == 0 {
					return goal[0]
				}
				return goal[1]
			}(),
		})
		gameState = 1
		broadcastBoard(true)
	}
}
