package main

import (
	"chess/game"
	"chess/states"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	log.Println("서버 시작: http://localhost:30")
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ws", HandleWebSocket)

	if err := http.ListenAndServe(":30", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 모든 도메인에서의 접근을 허용
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) { // 웹소켓 연결
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	defer delete(game.PlayerRooms, conn)
	defer delete(game.GameRooms, game.PlayerRooms[conn])
	log.Println("클라이언트 연결", conn.RemoteAddr())

	for {
		var message game.Message
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println(err)
			break
		}
		if roomId, exists := game.PlayerRooms[conn]; !exists {
			if message.Type == "join" {
				states.State0(conn)
			}
		} else {
			switch game.GameRooms[roomId].GameState {
			case 1:
				states.State1(game.GameRooms[roomId], conn, message)
			case 2:
				states.State2(game.GameRooms[roomId], conn, message)
			}
		}
	}
}
