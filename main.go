package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"example.com/mosquito-revenge/game"
)

func main() {
	ebiten.SetWindowTitle("Mosquito Revenge")
	ebiten.SetWindowSize(540, 960)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(game.New()); err != nil {
		log.Fatal(err)
	}
}
