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
	rooms := game.InitRooms()
	states.State3(game.Message{Type: "restart"})
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(rooms, w, r)
	})

	if err := http.ListenAndServe(":30", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 모든 도메인에서의 접근을 허용
	},
}

func HandleWebSocket(rooms map[int]*game.Context, w http.ResponseWriter, r *http.Request) { // 웹소켓 연결
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("클라이언트 연결", conn.RemoteAddr())

	for {
		var message game.Message
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println(err)
			break
		}
		states.State0(rooms, conn, message)
		switch rooms[roomId].GameState {
		case 1:
			states.State1(rooms[roomId], conn, message)
		case 2:
			states.State2(rooms[roomId], conn, message)
		case 3:
			states.State3(rooms[roomId], message)
		}
	}
}
