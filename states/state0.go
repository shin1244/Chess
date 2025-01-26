package states

import (
	"chess/game"
	"chess/ws"

	"github.com/gorilla/websocket"
)

// 유저가 게임에 참여하기를 기다립니다. 2명이 참여하면 state1로 이동합니다.
func State0(rooms map[int]*game.Context, conn *websocket.Conn, message game.Message) {
	if message.Type != "join" {
		return
	}
	roomId := joinRoom(rooms, conn)
	room := rooms[roomId]

	conn.WriteJSON(&game.Message{
		Type:        "color",
		PlayerColor: room.PlayerColor[conn],
	})
	if len(room.PlayerColor) == 2 {
		conn.WriteJSON(&game.Message{
			Type:  "board",
			Board: room.Board,
		})
		room.GameState = 1
		ws.BroadcastBoard(room, true)
	}
}

func joinRoom(rooms map[int]*game.Context, conn *websocket.Conn) int {
	joinGame := false
	// 빈 방이 있는지 확인
	for _, room := range rooms {
		if len(room.PlayerColor) < 2 {
			room.PlayerColor[conn] = len(room.PlayerColor)
			joinGame = true
			return len(rooms)
		}
	}
	// 빈 방이 없으면 새 방 생성
	if !joinGame {
		rooms[len(rooms)] = game.InitGame()
		rooms[len(rooms)].PlayerColor[conn] = len(rooms[len(rooms)].PlayerColor)
		return len(rooms)
	}
	return -1
}
