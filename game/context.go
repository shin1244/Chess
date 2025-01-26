package game

import (
	"log"
	"math/rand"

	"github.com/gorilla/websocket"
)

// 좌표
type Position struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type Message struct {
	Type        string     `json:"type"`
	Board       [8][8]Tile `json:"board"`
	PlayerColor int        `json:"player_color"`
	Position    Position   `json:"position"`
	Positions   []Position `json:"positions"`
	Piece       string     `json:"piece"`
	Start       bool       `json:"start"`
	Turn        int        `json:"turn"`
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
	"King": {
		{Row: -1, Col: -1}, {Row: -1, Col: 0}, {Row: -1, Col: 1},
		{Row: 0, Col: -1}, {Row: 0, Col: 1},
		{Row: 1, Col: -1}, {Row: 1, Col: 0}, {Row: 1, Col: 1},
	},
}

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

	log.Println(game.Board)

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
	for i := 2; i < 6; i++ {
		game.Board[i][rand.Intn(8)].Goal = 0
	}
	for i := 2; i < 6; i++ {
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
