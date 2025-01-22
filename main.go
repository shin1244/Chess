package main

import (
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/websocket"
)

type Position struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type Message struct {
	Type        string   `json:"type"`
	PlayerColor int      `json:"player_color"`
	Position    Position `json:"position"`
	Piece       string   `json:"piece"`
}

type Board struct {
	Type  string     `json:"type"`
	Board [8][8]Tile `json:"board"`
	Goal  []Position `json:"goal"`
}

type Tile struct {
	Color int
	Piece string
}

type click struct {
	Type     string     `json:"type"`
	Position []Position `json:"position"`
}

var playerColor = make(map[*websocket.Conn]int)      // 0: 백, 1: 흑
var playerPiece = make(map[*websocket.Conn][]string) // 백, 흑 체스말
var playerReady = []bool{}                           // 준비 완료하면 append
var count = 0                                        // 유저 수
var gameState = 0                                    // 1: 세팅, 2: 게임
var board = [8][8]Tile{}                             // 체스판(기물 포함)
var turn = 0                                         // 0: 백, 1: 흑
var possibleMoves = []Position{}                     // 비어있을 때: 클릭, 채워져 있을 때: 이동
var selectedPiece = Position{}
var whiteGoal = []Position{}
var blackGoal = []Position{}

func init() {
	whiteGoal = initGoal(0)
	blackGoal = initGoal(1)
}

var directions = map[string][]Position{
	"Knight": {
		{Row: -2, Col: -1}, {Row: -2, Col: 1},
		{Row: -1, Col: -2}, {Row: -1, Col: 2},
		{Row: 1, Col: -2}, {Row: 1, Col: 2},
		{Row: 2, Col: -1}, {Row: 2, Col: 1},
	},
	"Bishop": {
		{Row: -1, Col: -1}, {Row: -1, Col: 1},
		{Row: 1, Col: -1}, {Row: 1, Col: 1},
	},
	"Rook": {
		{Row: -1, Col: 0}, {Row: 1, Col: 0},
		{Row: 0, Col: -1}, {Row: 0, Col: 1},
	},
	"King": {
		{Row: -1, Col: -1}, {Row: -1, Col: 0}, {Row: -1, Col: 1},
		{Row: 0, Col: -1}, {Row: 0, Col: 1},
		{Row: 1, Col: -1}, {Row: 1, Col: 0}, {Row: 1, Col: 1},
	},
}

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

func HandleWebSocket(w http.ResponseWriter, r *http.Request) { // 웹소켓 연결
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

	conn.WriteJSON(&Message{
		Type:        "color",
		PlayerColor: playerColor[conn],
	})

	conn.WriteJSON(&Board{
		Type:  "board",
		Board: board,
		Goal: func() []Position {
			if playerColor[conn] == 0 {
				return whiteGoal
			}
			return blackGoal
		}(),
	})
	// 2명이 들어오기 전까지 기물을 놓을 수 없게 해야함

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
			playerReady = append(playerReady, false)
		}
		return piece
	}
	// 체스말은 색깔종류 형식으로 저장
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
	if message.Type == "click" {
		if turn == playerColor[conn] {
			if len(possibleMoves) == 0 { // 첫 클릭일 때
				if board[message.Position.Row][message.Position.Col].Piece != "" {
					possibleMoves = calculatePossibleMoves(board[message.Position.Row][message.Position.Col].Piece, message.Position.Row, message.Position.Col)
					selectedPiece = message.Position
					conn.WriteJSON(&click{
						Type:     "click",
						Position: possibleMoves,
					})
				}
			} else { // 두번째 클릭일 때
				for _, move := range possibleMoves {
					if move.Row == message.Position.Row && move.Col == message.Position.Col {
						board[message.Position.Row][message.Position.Col].Piece = board[selectedPiece.Row][selectedPiece.Col].Piece
						board[selectedPiece.Row][selectedPiece.Col].Piece = ""
						possibleMoves = []Position{}
						turn = (turn + 1) % 2
						broadcastBoard()
						// 3가지를 체크해야함
						// 1. 둘러 싸인 기물이 있는지
						// 2. 색칠을 완료했는지
						// 3. 폰이 움직이지 못하는지
						break
					}
				}
				if len(possibleMoves) != 0 {
					possibleMoves = []Position{}
					conn.WriteJSON(&Board{
						Type:  "board",
						Board: board,
					})
				}
			}
		}
	}
}

func broadcastBoard() {
	for conn := range playerColor {
		conn.WriteJSON(&Board{
			Type:  "board",
			Board: board,
			Goal: func() []Position {
				if playerColor[conn] == 0 {
					return whiteGoal
				}
				return blackGoal
			}(),
		})
	}
}
func calculatePossibleMoves(piece string, row, col int) []Position {
	pieceType := piece[5:]
	possibleMoves := []Position{}

	for _, direction := range directions[pieceType] {
		if pieceType == "Rook" || pieceType == "Bishop" {
			for i := 1; i < 8; i++ {
				newRow := row + direction.Row*i
				newCol := col + direction.Col*i
				if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 && board[newRow][newCol].Piece == "" {
					possibleMoves = append(possibleMoves, Position{Row: newRow, Col: newCol})
				} else {
					break
				}
			}
		} else {
			newRow := row + direction.Row
			newCol := col + direction.Col
			if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 && board[newRow][newCol].Piece == "" {
				possibleMoves = append(possibleMoves, Position{Row: newRow, Col: newCol})
			}
		}
	}
	return possibleMoves
}

func initGoal(color int) []Position {
	goal := []Position{}
	if color == 0 {
		for i := 5; i > 0; i-- {
			goal = append(goal, Position{Row: i, Col: rand.Intn(8)})
		}
	} else {
		for i := 2; i < 6; i++ {
			goal = append(goal, Position{Row: i, Col: rand.Intn(8)})
		}
	}
	return goal
}
