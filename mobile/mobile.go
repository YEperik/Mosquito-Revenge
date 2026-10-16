package mobile

import (
	ebitenmobile "github.com/hajimehoshi/ebiten/v2/mobile"

	"example.com/mosquito-revenge/game"
)

func init() {
	ebitenmobile.SetGame(game.New())
}

// Dummy is exported so ebitenmobile bind has a symbol to expose to Java.
func Dummy() {}
