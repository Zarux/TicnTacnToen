package game

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Zarux/ticntacntoen/pkg/mcts"
	"github.com/Zarux/ticntacntoen/pkg/tictactoe"
	"github.com/Zarux/ticntacntoen/services/game/game"
	"github.com/Zarux/ticntacntoen/services/game/settings"
)

type botPlayer interface {
	GetNextMove(context.Context, *tictactoe.Board, tictactoe.Player) int
	Stats() *mcts.LastMoveStats
	UpdateThinkTime(t time.Duration)
	UpdateExplorationParam(ep float64)
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
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic recovered:", r)
			os.Exit(1)
		}
	}()

	settingsModel := settings.InitialModel(header())
	p := tea.NewProgram(settingsModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}

	settings := settingsModel.GetSettings()
	if settings == nil {
		return
	}

	s.bot.UpdateThinkTime(settings.ThinkTime)

	for {
		g, _ := tictactoe.New(settings.N, settings.K)
		gameModel := game.InitialModel(header(), g.Board, s.bot, settings.P)

		p = tea.NewProgram(gameModel, tea.WithAltScreen(), tea.WithoutCatchPanics())
		if _, err := p.Run(); err != nil {
			panic(err)
		}

		if !gameModel.Replay {
			break
		}
	}
}

var (
	headerStyle1 = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#4204b5ff", Dark: "#8a63d1ff"}).Render
	headerStyle2 = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#108101ff", Dark: "#2eac1dff"}).Render
	headerStyle3 = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#853d02ff", Dark: "#c0681fff"}).Render
	headerStyle4 = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#5f5f5fff", Dark: "#ada8a8ff"}).Render
)

func header() string {
	return fmt.Sprintf(
		"%s %s %s %s %s\n\n",
		headerStyle4("---"),
		headerStyle1("Ticᴺ"),
		headerStyle2("Tacᴺ"),
		headerStyle3("Toeᴺ"),
		headerStyle4("---"),
	)
}
