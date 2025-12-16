package main

import (
	"github.com/Zarux/ticntacntoen/pkg/mcts"
	"github.com/Zarux/ticntacntoen/services/game"
)

func main() {
	bot := mcts.New(1, 250_000)

	gameService := game.New(bot)
	gameService.Play()
}
