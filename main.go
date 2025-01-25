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
	g := game.InitGame()
	states.State3(g, game.Message{Type: "restart"})
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(g, w, r)
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

func HandleWebSocket(g *game.Context, w http.ResponseWriter, r *http.Request) { // 웹소켓 연결
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
		switch g.GameState {
		case 0:
			states.State0(g, conn, message)
		case 1:
			states.State1(g, conn, message)
		case 2:
			states.State2(g, conn, message)
		case 3:
			states.State3(g, message)
		}
	}
}
