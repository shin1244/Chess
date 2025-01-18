package app

type Game struct {
	turn      bool
	board     [8][8]Tile
	ready     PlayerReady
	gameState int
}

type Piece struct {
	Color     int
	Position  Position
	PieceType int
}

type Position struct {
	X int
	Y int
}

type Tile struct {
	Color     int
	Position  Position
	Available bool
}

func NewGame() *Game {
	return &Game{
		turn:  true,
		board: newBoard(),
		ready: PlayerReady{},
	}
}

func newBoard() [8][8]Tile {
	board := [8][8]Tile{}
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			color := (x+y)%2 + 2 // 2, 3이 기본 타일 색(갈색, 검은색)
			board[y][x] = Tile{
				Color:     color,
				Position:  Position{X: x, Y: y},
				Available: true,
			}
		}
	}
	return board
}
