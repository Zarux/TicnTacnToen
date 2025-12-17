package main

import (
	"flag"

	"github.com/Zarux/ticntacntoen/pkg/mcts"
	"github.com/Zarux/ticntacntoen/services/game"
)

var (
	iterFlag = flag.Int("i", 250_000, "max iterations to run (default: 250_000)")
	concFlag = flag.Int("workers", 2, "concurrent workers (default: 2)")
)

func main() {
	flag.Parse()

	bot := mcts.New(*concFlag, *iterFlag)

	gameService := game.New(bot)
	gameService.Play()
}
