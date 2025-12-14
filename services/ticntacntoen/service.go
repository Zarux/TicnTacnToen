package ticntacntoen

import (
	"context"
	"fmt"
	"time"

	"github.com/Zarux/ticntacntoen/pkg/mcts"
	"github.com/Zarux/ticntacntoen/pkg/tictactoe"
)

type botPlayer interface {
	GetNextMove(context.Context, *tictactoe.Board, tictactoe.Player) int
}

type Service struct {
	bot botPlayer
}

func New(bot botPlayer) *Service {
	return &Service{
		bot: bot,
	}
}

func (s *Service) NewMove(ctx context.Context, move tictactoe.Move, player tictactoe.Player) error {
	return nil
}

func (s *Service) NewGame(ctx context.Context) error {
	return nil
}

func (s *Service) Play(ctx context.Context) error {
	bot := mcts.New(4, 100_000)

	game, err := tictactoe.New(4, 3)
	if err != nil {
		return err
	}

	board := game.Board

	turnNumber := 0
	player := tictactoe.P1
	for {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		t := time.Now()
		iterations := mcts.Iterations(board.N, board.K, turnNumber)
		exploration := mcts.ExplorationParameter(board.N, board.K, turnNumber)

		bot.UpdateExploationParam(exploration)

		nextMove := s.bot.GetNextMove(ctx, board, player)
		move := board.GetMove(nextMove)
		board.Play(player, move)
		fmt.Println("Current player:", player.Mark(), "move:", move, "thinking for:", time.Since(t), "iterations", iterations, "c", exploration)
		board.Print()

		winner := board.CheckWinner()
		if winner != tictactoe.Empty {
			fmt.Println("WINNER IS:", winner.Mark())
			break
		} else if winner == tictactoe.Empty && !board.AnyLegalMoves() {
			fmt.Println("DRAW")
			break
		}

		player = -player
		turnNumber++
	}
	return nil
}
