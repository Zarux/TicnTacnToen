package ticntacntoen

import (
	"context"
	"fmt"
	"time"

	"github.com/Zarux/ticntacntoen/pkg/mcts"
	"github.com/Zarux/ticntacntoen/pkg/tictactoe"
)

type botPlayer interface {
	GetNextMove(context.Context, *tictactoe.Board, tictactoe.Player) (int, error)
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
	bot := mcts.New(2, 1_000_000)
	bot.UpdateThinkTime(5 * time.Second)

	game, err := tictactoe.New(7, 4)
	if err != nil {
		return err
	}

	board := game.Board

	turnNumber := 0
	player := tictactoe.P1
	for {
		t := time.Now()

		nextMove, err := bot.GetNextMove(ctx, board, player)
		if err != nil {
			return err
		}

		move := board.GetMove(nextMove)
		board.Play(player, move)
		stats := bot.Stats()
		iterations := 0
		if stats != nil {
			iterations = stats.NumIterations
		}

		fmt.Println("Current player:", player.Mark(), "move:", move, "thinking for:", time.Since(t), "iterations", iterations)
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
