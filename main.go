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

// var allPlayer = make(map[*websocket.Conn]int)
var playerColor = make(map[*websocket.Conn]int)      // 0: 백, 1: 흑
var playerPiece = make(map[*websocket.Conn][]string) // 백, 흑 체스말
var playerPawn = make(map[*websocket.Conn]Position)  // 백, 흑 폰 위치
var playerReady = []bool{}                           // 준비 완료하면 append
var gameState = 0                                    // 1: 세팅, 2: 게임, 3: 결과
var board = [8][8]Tile{}                             // 체스판(기물 포함)
var turn = 0                                         // 0: 백, 1: 흑
var possibleMoves = []Position{}                     // 비어있을 때: 클릭, 채워져 있을 때: 이동
var selectedPiece = Position{}
var goal = make(map[int][]Position)
var pawnCheck = make(map[*websocket.Conn]int)
var result = 0

func init() {
	goal[0] = initGoal(0)
	goal[1] = initGoal(1)
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
	log.Println("서버 시작: http://localhost:30")
	initBoard()
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
	log.Println("클라이언트 연결", conn.RemoteAddr())

	for {
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println(err)
			break
		}
		if message.Type == "join" && gameState == 0 { // 게임 참가 버튼 눌렀을 때
			log.Println("게임 참가", conn.RemoteAddr())
			playerColor[conn] = len(playerColor)
			playerPiece[conn] = []string{"Knight", "Bishop", "Rook", "King"}
			conn.WriteJSON(&Message{
				Type:        "color",
				PlayerColor: playerColor[conn],
			})
			if len(playerColor) == 2 {
				conn.WriteJSON(&Board{
					Type:  "board",
					Board: board,
					Goal: func() []Position {
						if playerColor[conn] == 0 {
							return goal[0]
						}
						return goal[1]
					}(),
				})
				gameState = 1
				broadcastBoard()
			}
		}
		if gameState == 1 {
			setupState(conn, message)
		} else if gameState == 2 {
			playState(conn, message)
		}
	}
}

func getPiece(conn *websocket.Conn) string {
	pieces := playerPiece[conn]
	piece := ""
	lastIdx := len(pieces) - 1
	if playerColor[conn] == 0 {
		piece = "white" + pieces[lastIdx]
	} else {
		piece = "black" + pieces[lastIdx]
	}
	playerPiece[conn] = pieces[:lastIdx]
	if len(playerPiece[conn]) == 0 {
		playerReady = append(playerReady, false)
	}
	return piece
}

// 체스말은 숫자종류 형식으로 저장

func placePiece(conn *websocket.Conn, message Message) {
	if canPlace(conn, message) {
		piece := getPiece(conn)

		// 킹을 배치할 때 폰도 함께 배치
		if piece == "whiteKing" {
			if message.Position.Row > 0 { // 경계 체크 추가
				board[message.Position.Row-1][message.Position.Col].Piece = "whitePawn"
				board[message.Position.Row-1][message.Position.Col].Color = playerColor[conn]
				playerPawn[conn] = Position{Row: message.Position.Row - 1, Col: message.Position.Col}
			}
		} else if piece == "blackKing" {
			if message.Position.Row < 7 { // 경계 체크 추가
				board[message.Position.Row+1][message.Position.Col].Piece = "blackPawn"
				board[message.Position.Row+1][message.Position.Col].Color = playerColor[conn]
				playerPawn[conn] = Position{Row: message.Position.Row + 1, Col: message.Position.Col}
			}
		}

		setupMessage := &Message{
			Type:     "spawn",
			Position: message.Position,
			Piece:    piece,
		}
		conn.WriteJSON(setupMessage)
		board[message.Position.Row][message.Position.Col].Piece = piece
		board[message.Position.Row][message.Position.Col].Color = playerColor[conn]
	} else {
		log.Println("잘못된 위치 혹은 기물 없음")
	}
}

func setupState(conn *websocket.Conn, message Message) {
	// 클릭 이벤트인 경우에만 기물 배치
	if message.Type == "click" {
		placePiece(conn, message)

		// 모든 플레이어가 준비되었는지 확인
		if len(playerReady) == 2 {
			// 각 플레이어의 기물이 모두 배치되었는지 확인
			allPiecesPlaced := true
			for _, pieces := range playerPiece {
				if len(pieces) > 0 {
					allPiecesPlaced = false
					break
				}
			}

			// 모든 기물이 배치된 경우에만 게임 시작
			if allPiecesPlaced {
				broadcastBoard()
				gameState = 2
			}
		}
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
				if board[message.Position.Row][message.Position.Col].Piece != "" && checkColor(board[message.Position.Row][message.Position.Col].Piece) == playerColor[conn] {
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
						paintPath(selectedPiece.Row, selectedPiece.Col, message.Position.Row, message.Position.Col, turn)           // 경로 색칠
						board[message.Position.Row][message.Position.Col].Piece = board[selectedPiece.Row][selectedPiece.Col].Piece // 보드에서 기물 이동
						board[selectedPiece.Row][selectedPiece.Col].Piece = ""
						possibleMoves = []Position{}
						turn = (turn + 1) % 2
						movePawn(conn)

						// 3가지를 체크해야함
						// 1. 둘러 싸인 기물이 있는지
						// 2. 색칠을 완료했는지

						checkGameOver()
						// 3. 폰이 움직이지 못하는지
						broadcastBoard()
						if pawnCheck[conn] == 3 {
							gameState = 3
							countResult()
						}
						break
					}
				}
				if len(possibleMoves) != 0 {
					possibleMoves = []Position{}
					broadcastBoard()
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
					return goal[0]
				}
				return goal[1]
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

// 이동한 경로 색칠하는 함수 만들어야 함
func paintPath(row, col, endRow, endCol, color int) {
	// 이동 방향 계산
	rowDir := 0
	if endRow-row > 0 {
		rowDir = 1
	} else if endRow-row < 0 {
		rowDir = -1
	}

	colDir := 0
	if endCol-col > 0 {
		colDir = 1
	} else if endCol-col < 0 {
		colDir = -1
	}

	// 현재 위치의 기물 타입 확인
	piece := board[row][col].Piece
	pieceType := piece[5:] // "white" 또는 "black" 제거

	// 나이트의 경우 'ㄱ' 모양으로 경로 색칠
	if pieceType == "Knight" {
		board[row][col].Color = color // 시작점

		// 2칸 이동 먼저 (수직 또는 수평)
		if abs(endRow-row) == 2 {
			// 수직으로 2칸 이동
			intermediateRow := row + rowDir // 중간 칸
			if board[intermediateRow][col].Piece == "" {
				board[intermediateRow][col].Color = color
			}

			intermediateRow = row + rowDir*2 // 2칸 이동 후
			if board[intermediateRow][col].Piece == "" {
				board[intermediateRow][col].Color = color
			}

			// 그 다음 수평으로 1칸 이동
			board[intermediateRow][endCol].Color = color
		} else {
			// 수평으로 2칸 이동
			intermediateCol := col + colDir // 중간 칸
			if board[row][intermediateCol].Piece == "" {
				board[row][intermediateCol].Color = color
			}

			intermediateCol = col + colDir*2 // 2칸 이동 후
			if board[row][intermediateCol].Piece == "" {
				board[row][intermediateCol].Color = color
			}

			// 그 다음 수직으로 1칸 이동
			board[endRow][intermediateCol].Color = color
		}

		return
	}

	// 룩, 비숍, 킹의 경우 경로 색칠
	currentRow := row
	currentCol := col

	for currentRow != endRow || currentCol != endCol {
		board[currentRow][currentCol].Color = color

		// 대각선 이동 (비숍)
		if rowDir != 0 && colDir != 0 {
			currentRow += rowDir
			currentCol += colDir
			// 수직 이동 (룩)
		} else if rowDir != 0 {
			currentRow += rowDir
			// 수평 이동 (룩)
		} else if colDir != 0 {
			currentCol += colDir
		}
	}
	// 도착 지점 색칠
	board[endRow][endCol].Color = color
}

func checkColor(piece string) int {
	if piece[0:5] == "white" {
		return 0
	} else {
		return 1
	}
}

// 절대값 계산을 위한 헬퍼 함수 추가
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// 목표를 모두 색칠했는지 확인
func paintCheck() int {
	for _, color := range playerColor {
		for _, position := range goal[color] {
			if board[position.Row][position.Col].Color != color {
				return -1
			}
		}
		return color
	}
	return -1
}

// 죽은 기물 확인 폰은 안죽음
func dieCheck() int {
	var checkPos = []Position{
		{Row: 1, Col: 0}, {Row: -1, Col: 0}, {Row: 0, Col: 1}, {Row: 0, Col: -1},
	}
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if board[i][j].Piece != "" {
				c := true
				if board[i][j].Piece[0:5] == "white" && board[i][j].Piece[5:] != "Pawn" {
					for _, position := range checkPos {
						if i+position.Row >= 0 && i+position.Row < 8 && j+position.Col >= 0 && j+position.Col < 8 && board[i+position.Row][j+position.Col].Color != 1 {
							c = false
						}
					}
					if c {
						if board[i][j].Piece == "whiteKing" {
							log.Println("흰군 죽음")
							board[i][j].Piece = ""
							return 1
						}
						board[i][j].Piece = ""
					}
				} else if board[i][j].Piece[0:5] == "black" && board[i][j].Piece[5:] != "Pawn" {
					for _, position := range checkPos {
						if i+position.Row >= 0 && i+position.Row < 8 && j+position.Col >= 0 && j+position.Col < 8 && board[i+position.Row][j+position.Col].Color != 0 {
							c = false
						}
					}
					if c {
						if board[i][j].Piece == "blackKing" {
							log.Println("흑군 죽음")
							board[i][j].Piece = ""
							return 0
						}
						board[i][j].Piece = ""
					}
				}
			}
		}
	}
	return -1
}

// 턴 종료 시 폰 이동
func movePawn(conn *websocket.Conn) {
	color := playerColor[conn]
	if color == 0 {
		// 흰색 폰이 맨 위에 도달했거나, 앞이 막혀있을 때
		if playerPawn[conn].Row <= 0 || board[playerPawn[conn].Row-1][playerPawn[conn].Col].Piece != "" {
			pawnCheck[conn]++
			return
		}
		board[playerPawn[conn].Row][playerPawn[conn].Col].Piece = ""
		board[playerPawn[conn].Row-1][playerPawn[conn].Col].Piece = "whitePawn"
		board[playerPawn[conn].Row-1][playerPawn[conn].Col].Color = 0
		playerPawn[conn] = Position{Row: playerPawn[conn].Row - 1, Col: playerPawn[conn].Col}
		pawnCheck[conn] = 0 // 성공적으로 움직였으므로 카운터 초기화
	} else {
		// 검은색 폰이 맨 아래에 도달했거나, 앞이 막혀있을 때
		if playerPawn[conn].Row >= 7 || board[playerPawn[conn].Row+1][playerPawn[conn].Col].Piece != "" {
			pawnCheck[conn]++
			return
		}
		board[playerPawn[conn].Row][playerPawn[conn].Col].Piece = ""
		board[playerPawn[conn].Row+1][playerPawn[conn].Col].Piece = "blackPawn"
		board[playerPawn[conn].Row+1][playerPawn[conn].Col].Color = 1
		playerPawn[conn] = Position{Row: playerPawn[conn].Row + 1, Col: playerPawn[conn].Col}
		pawnCheck[conn] = 0 // 성공적으로 움직였으므로 카운터 초기화
	}
}

func checkGameOver() {
	if result := dieCheck(); result != -1 {
		gameState = 3
		for conn := range playerColor {
			conn.WriteJSON(&Message{
				Type:        "gameOver",
				PlayerColor: result,
			})
		}
	} else if result := paintCheck(); result != -1 {
		gameState = 3
		for conn := range playerColor {
			conn.WriteJSON(&Message{
				Type:        "gameOver",
				PlayerColor: result,
			})
		}
	}
}

// func resetGame() {
// 	initBoard()
// 	goal[0] = initGoal(0)
// 	goal[1] = initGoal(1)
// 	playerColor = make(map[*websocket.Conn]int)
// 	playerPiece = make(map[*websocket.Conn][]string)
// 	playerPawn = make(map[*websocket.Conn]Position)
// 	playerReady = []bool{}
// 	gameState = 0
// 	turn = 0
// 	possibleMoves = []Position{}
// 	selectedPiece = Position{}
// }

func countResult() {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if board[i][j].Color == 0 {
				result += 1
			} else if board[i][j].Color == 1 {
				result -= 1
			}
		}
	}
	if result > 0 {
		for conn := range playerColor {
			conn.WriteJSON(&Message{
				Type:        "gameOver",
				PlayerColor: 0,
			})
		}
	} else {
		for conn := range playerColor {
			conn.WriteJSON(&Message{
				Type:        "gameOver",
				PlayerColor: 1,
			})
		}
	}
}
