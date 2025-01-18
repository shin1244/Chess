package ws

import (
	"chess/game"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS 허용
	},
}

type TileClick struct {
	Type int `json:"type"`
	Row  int `json:"row"`
	Col  int `json:"col"`
}

var (
	// 모든 웹소켓 연결을 저장
	Players = make(map[*websocket.Conn]int)
	mu      sync.Mutex
	G       *game.Game
	count   int
)

// 웹소켓 연결 핸들러
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("웹소켓 연결 에러:", err)
		return
	}
	defer conn.Close()

	Players[conn] = count
	count++

	BroadcastCount(Players)
	if len(Players) == 2 {
		G = ConnectPlayer(Players)
	}
	log.Println(Players)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			mu.Lock()
			delete(Players, conn)
			mu.Unlock()
			log.Println("연결 끊김")

			BroadcastCount(Players)
			break
		}
		var tileClick TileClick
		json.Unmarshal(message, &tileClick)
		if tileClick.Type == 1 {
			game.SetupGame(&game.Ready, Players[conn], tileClick.Col, tileClick.Row)
		}
	}
}

// 플레이어 연결 수 표시
func BroadcastCount(connections map[*websocket.Conn]int) {
	if len(connections) > 2 {
		return
	}
	message := map[string]interface{}{
		"type":  "count",
		"count": len(connections),
	}

	for conn := range connections {
		conn.WriteJSON(message)
	}
}

// 플레이어 연결 시 게임 초기화
func ConnectPlayer(connections map[*websocket.Conn]int) *game.Game {
	game := game.NewGame()

	message := map[string]interface{}{
		"type": "newGame",
	}

	for conn := range connections {
		conn.WriteJSON(message)
	}
	return game
}
