package states

import (
	"chess/game"
	"chess/ws"
	"log"

	"github.com/gorilla/websocket"
)

// 유저가 게임에 참여하기를 기다립니다. 2명이 참여하면 state1로 이동합니다.
func State0(conn *websocket.Conn) {
	log.Println("게임 참가", conn.RemoteAddr())
	var g *game.Context

	if game.EmptyRoom == "" {
		roomCode := game.GenerateRoomCode()
		g = game.InitGame()
		g.PlayerColor[conn] = 0
		game.GameRooms[roomCode] = g
		game.PlayerRooms[conn] = roomCode
		game.EmptyRoom = roomCode
		log.Println("방 생성", roomCode)
	} else {
		game.PlayerRooms[conn] = game.EmptyRoom
		g = game.GameRooms[game.EmptyRoom]
		g.PlayerColor[conn] = 1

		log.Println("방 입장", game.EmptyRoom)
		game.EmptyRoom = ""
	}

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
		ws.BroadcastBoard(g, 1)
	}
}
