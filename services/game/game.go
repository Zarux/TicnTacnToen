package game

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Zarux/ticntacntoen/pkg/tictactoe"
	"github.com/Zarux/ticntacntoen/services/game/game"
	"github.com/Zarux/ticntacntoen/services/game/settings"
)

type botPlayer interface {
	GetNextMove(context.Context, *tictactoe.Board, tictactoe.Player) int
	UpdateThinkTime(t time.Duration)
}

type Service struct {
	bot botPlayer
}

func New(bot botPlayer) *Service {
	return &Service{
		bot: bot,
	}
}

func (s *Service) Play() {
	settingsModel := settings.InitialModel()
	p := tea.NewProgram(settingsModel)
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}

	settings := settingsModel.GetSettings()

	s.bot.UpdateThinkTime(settings.ThinkTime)

	for {
		g, _ := tictactoe.New(settings.N, settings.K)
		gameModel := game.InitialModel(g.Board, s.bot, settings.P)

		p = tea.NewProgram(gameModel)
		if _, err := p.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if !gameModel.Replay {
			break
		}
	}
}
