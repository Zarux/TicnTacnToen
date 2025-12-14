package game

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Zarux/ticntacntoen/pkg/tictactoe"
)

type botPlayer interface {
	GetNextMove(context.Context, *tictactoe.Board, tictactoe.Player) int
}

type model struct {
	board         *tictactoe.Board
	cursor        int
	currentPlayer tictactoe.Player
	botPlayer     tictactoe.Player
	bot           botPlayer
	spinner       spinner.Model
	sub           chan botDoneMsg

	gameOver bool
	winner   tictactoe.Player
}

func (m model) Init() tea.Cmd {
	if m.bot != nil && m.botPlayer == m.currentPlayer {
		return tea.Batch(m.beginTick(), waitForBot(m.sub), m.botMove(context.Background(), m.sub))
	}

	return nil
}

var (
	p1Style      = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#007e50ff", Dark: "#04B575"}).Render
	p2Style      = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0004ffff", Dark: "#0483b5ff"}).Render
	cursorStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#ff0000ff", Dark: "#ec6b6bff"}).Render
	bracketStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9e9e9eff", Dark: "#9e9e9eff"}).Render
)

func InitialModel(b *tictactoe.Board, bot botPlayer, playerStone tictactoe.Player) model {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		board:         b,
		currentPlayer: tictactoe.P1,
		botPlayer:     -playerStone,
		bot:           bot,
		spinner:       s,
		sub:           make(chan botDoneMsg),
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case botDoneMsg:
		if msg.draw {
			m.gameOver = true
			m.winner = tictactoe.Empty
			return m, nil
		}

		if msg.winner != tictactoe.Empty {
			m.gameOver = true
			m.winner = msg.winner
			return m, nil
		}

		m.cursor = msg.cursor
		m.currentPlayer = -m.currentPlayer
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "right":
			cursor, _ := m.moveRight()
			m.cursor = cursor
		case "left":
			cursor, _ := m.moveLeft()
			m.cursor = cursor
		case "up":
			if m.cursor > m.board.N-1 {
				oCursor := m.cursor
				m.cursor -= m.board.N
				for {
					if m.cursor < 0 {
						m.cursor = oCursor
						return m, nil
					}

					if m.board.Cells[m.cursor] == tictactoe.Empty {
						break
					}

					m.cursor -= m.board.N
				}
			}
		case "down":
			if m.cursor < m.board.N*(m.board.N-1) {
				oCursor := m.cursor
				m.cursor += m.board.N
				for {
					if m.cursor > len(m.board.Cells)-1 {
						m.cursor = oCursor
						return m, nil
					}

					if m.board.Cells[m.cursor] == tictactoe.Empty {
						break
					}

					m.cursor += m.board.N
				}
			}
		case "enter":
			if m.gameOver {
				return m, tea.Quit
			}

			newCursor, winner := m.playerMove(m.cursor, m.currentPlayer)
			if !m.board.AnyLegalMoves() && winner == tictactoe.Empty {
				m.gameOver = true
				m.winner = tictactoe.Empty
				return m, nil
			}

			if winner == m.currentPlayer {
				m.gameOver = true
				m.winner = m.currentPlayer
				return m, nil
			}

			m.cursor = newCursor
			m.currentPlayer = -m.currentPlayer

			return m, tea.Batch(m.beginTick(), waitForBot(m.sub), m.botMove(context.Background(), m.sub))
		}

	default:
		if m.gameOver {
			m.cursor = -1
			return m, nil
		}

		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) playerMove(move int, p tictactoe.Player) (int, tictactoe.Player) {
	err := m.board.ApplyMove(move, p)
	if err != nil {
		panic(err)
	}

	winner := m.board.CheckWinner()

	if m.cursor != move {
		return m.cursor, winner
	}

	cursor, ok := m.moveRight()
	if !ok {
		cursor, ok = m.moveLeft()
	}

	if !ok {
		cursor = -1
	}

	return cursor, winner
}

type botDoneMsg struct {
	cursor int
	winner tictactoe.Player
	draw   bool
}

func waitForBot(sub chan botDoneMsg) tea.Cmd {
	return func() tea.Msg {
		return botDoneMsg(<-sub)
	}
}

func (m model) beginTick() tea.Cmd {
	return func() tea.Msg {
		return m.spinner.Tick()
	}
}

func (m model) botMove(ctx context.Context, sub chan botDoneMsg) tea.Cmd {
	return func() tea.Msg {
		nextMove := m.bot.GetNextMove(ctx, m.board, m.botPlayer)
		cursor, winner := m.playerMove(nextMove, m.botPlayer)
		sub <- botDoneMsg{
			cursor: cursor,
			winner: winner,
			draw:   !m.board.AnyLegalMoves() && winner == tictactoe.Empty,
		}

		return nil
	}
}

func (m model) moveRight() (int, bool) {
	if m.cursor >= 0 && m.cursor < len(m.board.Cells)-1 {
		oCursor := m.cursor
		m.cursor++
		for {
			if m.cursor > len(m.board.Cells)-1 {
				return oCursor, false
			}

			if m.board.Cells[m.cursor] == tictactoe.Empty {
				break
			}

			m.cursor++
		}
	}

	return m.cursor, true
}

func (m model) moveLeft() (int, bool) {
	if m.cursor > 0 {
		oCursor := m.cursor
		m.cursor--
		for {
			if m.cursor < 0 {
				return oCursor, false
			}

			if m.board.Cells[m.cursor] == tictactoe.Empty {
				break
			}

			m.cursor--
		}
	}

	return m.cursor, true
}

func (m model) View() string {
	botTurn := m.bot != nil && m.currentPlayer == m.botPlayer && !m.gameOver

	s := "Current player: "
	switch m.currentPlayer {
	case tictactoe.P1:
		s += p1Style(m.currentPlayer.Mark())
	case tictactoe.P2:
		s += p2Style(m.currentPlayer.Mark())
	}

	if botTurn {
		s += " (bot) " + m.spinner.View()
	}

	s += "\n"

	for i, p := range m.board.Cells {
		mark := p.Mark()
		if m.cursor == i {
			mark = cursorStyle("*")
		}

		if botTurn && p == tictactoe.Empty {
			mark = []string{"o", "x"}[rand.N(2)]
		}

		switch p {
		case tictactoe.P1:
			mark = p1Style(p.Mark())
		case tictactoe.P2:
			mark = p2Style(p.Mark())
		}

		s += fmt.Sprintf("%s%s%s", bracketStyle("["), mark, bracketStyle("]"))
		if (i+1)%m.board.N == 0 {
			s += "\n"
		}
	}

	if m.gameOver {
		s += "\n" + gameOverText

		s += "\nTHE WINNER IS: "
		if m.winner == tictactoe.Empty {
			s += "NO ONE"
			return s
		}

		switch m.winner {
		case tictactoe.P1:
			s += p1Style(m.winner.Mark())
		case tictactoe.P2:
			s += p2Style(m.winner.Mark())
		}

		return s
	}

	return s
}

const gameOverText = `ＧＡＭＥ ＯＶＥＲ`
