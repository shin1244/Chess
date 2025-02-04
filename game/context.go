package game

import (
	cryptorand "crypto/rand"
	"math/rand"

	"github.com/gorilla/websocket"
)

// 좌표
type Position struct {
	Row   int    `json:"row"`
	Col   int    `json:"col"`
	Piece string `json:"piece"`
}

type Message struct {
	Type          string     `json:"type"`
	Board         [8][8]Tile `json:"board"`
	PlayerColor   int        `json:"player_color"`
	Position      Position   `json:"position"`
	Positions     []Position `json:"positions"`
	Piece         string     `json:"piece"`
	SoundType     int        `json:"sound_type"`
	Turn          int        `json:"turn"`
	PrintingTiles [2]int     `json:"printing_tiles"`
	KingMove      Position   `json:"king_move"`
}

type Tile struct {
	Color int    `json:"color"`
	Piece string `json:"piece"`
	Goal  int    `json:"goal"`
}

type Context struct {
	PlayerColor   map[*websocket.Conn]int
	GameState     int
	Board         [8][8]Tile
	Turn          int
	PossibleMoves []Position
	SelectedPiece Position
	Pieces        [][]string
	PrintingTiles [2]int
}

var Directions = map[string][]Position{
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
}

func GenerateRoomCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 6

	b := make([]byte, codeLength)
	if _, err := cryptorand.Read(b); err != nil {
		panic(err)
	}

	code := make([]byte, codeLength)
	for i := range b {
		code[i] = charset[int(b[i])%len(charset)]
	}
	return string(code)
}

var GameRooms = make(map[string]*Context)
var PlayerRooms = make(map[*websocket.Conn]string)
var EmptyRoom string = ""

func InitGame() *Context {
	game := &Context{}
	initBoard(game)                                                                                      // 체스판 초기화
	initGoal(game)                                                                                       // 목표 초기화
	game.PlayerColor = make(map[*websocket.Conn]int)                                                     // 플레이어 색
	game.GameState = 0                                                                                   // 게임 상태
	game.Turn = 0                                                                                        // 턴
	game.PossibleMoves = []Position{}                                                                    // 가능한 이동
	game.SelectedPiece = Position{}                                                                      // 선택한 기물
	game.Pieces = [][]string{{"King", "Rook", "Bishop", "Knight"}, {"King", "Rook", "Bishop", "Knight"}} // 체스말 종류
	game.PrintingTiles = [2]int{0, 0}                                                                    // 카운트 초기화

	return game
}

// 체스판 초기화
func initBoard(game *Context) {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if (i+j)%2 == 0 {
				game.Board[i][j] = Tile{Color: 2, Piece: "", Goal: -1}
			} else {
				game.Board[i][j] = Tile{Color: 3, Piece: "", Goal: -1}
			}
		}
	}
}

// 목표 초기화
func initGoal(game *Context) {
	for i := 3; i < 6; i++ {
		game.Board[i][rand.Intn(8)].Goal = 0
	}
	for i := 2; i < 5; i++ {
		col := rand.Intn(8)
		for game.Board[i][col].Goal == 0 {
			col = rand.Intn(8)
		}
		game.Board[i][col].Goal = 1
	}
}

func InitRooms() map[int]*Context {
	AllRooms := make(map[int]*Context)
	return AllRooms
}
