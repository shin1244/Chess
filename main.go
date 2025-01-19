package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type     string `json:"type"`
	Position struct {
		Row int `json:"row"`
		Col int `json:"col"`
	} `json:"position"`
	Piece string `json:"piece"`
}

var playerColor = make(map[*websocket.Conn]int)
var playerPiece = make(map[*websocket.Conn][]string)
var playerReady = make(map[*websocket.Conn]bool)
var count = 0

func main() {
	log.Println("서버 시작: http://localhost:3000")
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/ws", HandleWebSocket)

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

var upgrader = websocket.Upgrader{}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if count < 2 {
		playerColor[conn] = count
		playerPiece[conn] = []string{"Pawn", "Knight", "Bishop", "Rook", "King"}
		count++
	}

	for {
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println(err)
			break
		}

		if playerColor[conn] == 0 && len(playerPiece[conn]) > 0 && message.Position.Row == 7 {
			conn.WriteJSON(&Message{
				Type:     "spawn",
				Position: message.Position,
				Piece:    getPiece(conn),
			})
			log.Println(playerPiece[conn])
		} else if playerColor[conn] == 1 && len(playerPiece[conn]) > 0 && message.Position.Row == 0 {
			conn.WriteJSON(&Message{
				Type:     "spawn",
				Position: message.Position,
				Piece:    getPiece(conn),
			})
			log.Println(playerPiece[conn])
		} else {
			log.Println("잘못된 위치 혹은 기물 없음")
		}

		readyCount := 0
		for _, ready := range playerReady {
			if !ready {
				break
			}
			readyCount++
		}
		if readyCount == 2 {
			log.Println("시작")
		}
	}
}

func getPiece(conn *websocket.Conn) string {
	pieces := playerPiece[conn]
	lastIdx := len(pieces) - 1

	if playerColor[conn] == 0 {
		piece := "white" + pieces[lastIdx]
		playerPiece[conn] = pieces[:lastIdx]
		if len(playerPiece[conn]) == 0 {
			playerReady[conn] = true
		}
		return piece
	} else {
		piece := "black" + pieces[lastIdx]
		playerPiece[conn] = pieces[:lastIdx]
		if len(playerPiece[conn]) == 0 {
			playerReady[conn] = true
		}
		return piece
	}
}
