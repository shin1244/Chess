package main

import (
	"chess/app"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	game := app.NewGame()
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
