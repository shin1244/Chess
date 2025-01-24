package main

import "github.com/gorilla/websocket"

func endState(message Message) {
	if message.Type != "restart" {
		return
	}
	initBoard()
	goal[0] = initGoal(0)
	goal[1] = initGoal(1)
	playerColor = make(map[*websocket.Conn]int)
	playerPawn = make(map[*websocket.Conn]Position)
	pawnCount = []int{0, 0}
	playerReady = make(map[*websocket.Conn]int)
	gameState = 0
	turn = 0
	possibleMoves = []Position{}
	selectedPiece = Position{}
}
