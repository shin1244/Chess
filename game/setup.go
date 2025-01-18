package game

import "log"

type PlayerReady struct {
	Player0 chan bool
	Player1 chan bool
	Piece0  []Piece
	Piece1  []Piece
}

var Ready = PlayerReady{
	Player0: make(chan bool, 1),
	Player1: make(chan bool, 1),
	Piece0:  []Piece{},
	Piece1:  []Piece{},
}

// 초기 기물 설정
func SetupGame(ready *PlayerReady, player, x, y int) {
	if player == 0 && len(ready.Piece0) < 4 {
		if y == 0 {
			ready.Piece0 = append(ready.Piece0, Piece{Color: 0, Position: Position{X: x, Y: y}, PieceType: len(ready.Piece0)})
			log.Println("Player 0 기물 추가:", ready.Piece0)
			if len(ready.Piece0) == 4 {
				log.Println("Player 0 준비 완료")
				select {
				case ready.Player0 <- true:
					log.Println("Player 0 채널에 값 전송 완료")
				default:
					log.Println("Player 0 채널이 가득 참")
				}
			}
		}
	} else if player == 1 && len(ready.Piece1) < 4 {
		if y == 7 {
			ready.Piece1 = append(ready.Piece1, Piece{Color: 1, Position: Position{X: x, Y: y}, PieceType: len(ready.Piece1)})
			log.Println("Player 1 기물 추가:", ready.Piece1)
			if len(ready.Piece1) == 4 {
				log.Println("Player 1 준비 완료")
				select {
				case ready.Player1 <- true:
					log.Println("Player 1 채널에 값 전송 완료")
				default:
					log.Println("Player 1 채널이 가득 참")
				}
			}
		}
	}
}

// 두 플레이어 모두 초기 기물 설정 완료
func SetupGameDone(ready *PlayerReady) int {
	if <-ready.Player0 && <-ready.Player1 {
		log.Println("두 플레이어 모두 초기 기물 설정 완료")
		return 1
	}
	return 0
}
