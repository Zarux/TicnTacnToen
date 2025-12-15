package main

import (
	"runtime"

	"github.com/Zarux/ticntacntoen/pkg/mcts"
	"github.com/Zarux/ticntacntoen/services/game"
)

func main() {
	bot := mcts.New(runtime.NumCPU()/2, 250_000)

	gameService := game.New(bot)
	gameService.Play()
}
