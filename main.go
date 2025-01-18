package main

import (
	"chess/game"
	"chess/ws"
	"log"
	"net/http"
)

func main() {
	ws.G.GameState = game.SetupGameDone(&game.Ready)

	http.HandleFunc("/ws", ws.HandleWebSocket)
	http.Handle("/", http.FileServer(http.Dir(".")))

	log.Println("서버 시작: http://localhost:3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
