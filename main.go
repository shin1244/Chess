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

type Board struct {
	Type  string     `json:"type"`
	Board [8][8]Tile `json:"board"`
}

type Tile struct {
	Color int
	Piece string
}

var playerColor = make(map[*websocket.Conn]int)      // 0: 백, 1: 흑
var playerPiece = make(map[*websocket.Conn][]string) // 백, 흑 체스말
var playerReady = []bool{}                           // 준비 완료하면 append
var count = 0                                        // 유저 수
var gameState = 0                                    // 1: 세팅, 2: 게임
var board = [8][8]Tile{}                             // 체스판(기물 포함)

func initBoard() {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if (i+j)%2 == 0 {
				board[i][j] = Tile{Color: 2, Piece: ""}
			} else {
				board[i][j] = Tile{Color: 3, Piece: ""}
			}
		}
	}
}

func main() {
	log.Println("서버 시작: http://localhost:3000")
	initBoard()
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
	conn.WriteJSON(&Board{
		Type:  "board",
		Board: board,
	})

	for {
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println(err)
			break
		}
		if gameState == 0 {
			setupState(conn, message)
		} else if gameState == 1 {
			playState(conn, message)
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
			playerReady = append(playerReady, true)
		}
		return piece
	} else {
		piece := "black" + pieces[lastIdx]
		playerPiece[conn] = pieces[:lastIdx]
		if len(playerPiece[conn]) == 0 {
			playerReady = append(playerReady, true)
		}
		return piece
	}
}

func placePiece(conn *websocket.Conn, message Message) {
	if canPlace(conn, message) {
		piece := getPiece(conn)
		setupMessage := &Message{
			Type:     "spawn",
			Position: message.Position,
			Piece:    piece,
		}
		conn.WriteJSON(setupMessage)
		board[message.Position.Row][message.Position.Col].Piece = piece
		log.Println(board)
	} else {
		log.Println("잘못된 위치 혹은 기물 없음")
	}
}

func setupState(conn *websocket.Conn, message Message) {
	placePiece(conn, message)

	if len(playerReady) == 2 {
		broadcastBoard()
		gameState = 1
	}
}

func canPlace(conn *websocket.Conn, message Message) bool {
	if len(playerPiece[conn]) > 0 {
		if (playerColor[conn] == 0 && message.Position.Row == 7) ||
			(playerColor[conn] == 1 && message.Position.Row == 0) {
			if board[message.Position.Row][message.Position.Col].Piece == "" {
				return true
			}
		}
	}
	return false
}

func playState(conn *websocket.Conn, message Message) {

}

func broadcastBoard() {
	for conn := range playerColor {
		conn.WriteJSON(&Board{
			Type:  "board",
			Board: board,
		})
	}
}
