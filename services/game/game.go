package game

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	settingsModel := settings.InitialModel(header())
	p := tea.NewProgram(settingsModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}

	settings := settingsModel.GetSettings()

	s.bot.UpdateThinkTime(settings.ThinkTime)

	for {
		g, _ := tictactoe.New(settings.N, settings.K)
		gameModel := game.InitialModel(header(), g.Board, s.bot, settings.P)

		p = tea.NewProgram(gameModel, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if !gameModel.Replay {
			break
		}
	}
}

var (
	headerStyle1 = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#4204b5ff", Dark: "#4204b5ff"}).Render
	headerStyle2 = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#19b504ff", Dark: "#19b504ff"}).Render
	headerStyle3 = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#b55404ff", Dark: "#b55404ff"}).Render
)

func header() string {
	return fmt.Sprintf(
		"%s %s %s %s %s\n\n",
		headerStyle2("---"),
		headerStyle1("Ticᴺ"),
		headerStyle2("Tacᴺ"),
		headerStyle3("Toeᴺ"),
		headerStyle2("---"),
	)
}
