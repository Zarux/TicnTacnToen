package game

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Zarux/ticntacntoen/pkg/mcts"
	"github.com/Zarux/ticntacntoen/pkg/tictactoe"
)

type botPlayer interface {
	GetNextMove(context.Context, *tictactoe.Board, tictactoe.Player) (int, error)
	Stats() *mcts.LastMoveStats
}

type model struct {
	board         *tictactoe.Board
	cursor        int
	currentPlayer tictactoe.Player
	botPlayer     tictactoe.Player
	bot           botPlayer
	spinner       spinner.Model
	sub           chan botDoneMsg
	header        string

	gameOver bool
	winner   tictactoe.Player
	Replay   bool
}

func (m model) Init() tea.Cmd {
	if m.bot != nil && m.botPlayer == m.currentPlayer {
		return tea.Batch(m.beginTick(), waitForBot(m.sub), m.botMove(context.Background(), m.sub))
	}

	return nil
}

var (
	p1Style              = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#007e50ff", Dark: "#6afd76ff"}).Render
	p2Style              = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0003adff", Dark: "#5f61fcff"}).Render
	cursorStyle          = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#960000ff", Dark: "#fc7e7eff"}).Render
	winningRowStyle      = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#bb0000ff", Dark: "#df1010ff"}).Render
	lastWinningRowStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#f80000ff", Dark: "#f18787ff"}).Render
	bracketStyle         = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#414141ff", Dark: "#8f8f8fff"}).Render
	lastMoveBracketStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000ff", Dark: "#ffffffff"}).Render
	statStyle1           = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#8a880fff", Dark: "#ddda1dff"}).Render
	statStyle2           = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#138a0fff", Dark: "#1ddd37ff"}).Render
)

var thinkingColors = []func(strs ...string) string{
	bracketStyle,
	lastMoveBracketStyle,
}

func InitialModel(header string, b *tictactoe.Board, bot botPlayer, playerStone tictactoe.Player) *model {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &model{
		board:         b,
		currentPlayer: tictactoe.P1,
		botPlayer:     -playerStone,
		bot:           bot,
		spinner:       s,
		sub:           make(chan botDoneMsg),
		Replay:        false,
		header:        header,
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.Replay = true
				return m, tea.Quit
			}

			if m.bot != nil && m.currentPlayer == m.botPlayer {
				return m, nil
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

	cursor, rOk := m.moveRight()
	if rOk {
		return cursor, winner
	}

	cursor, lOk := m.moveLeft()
	if lOk {
		return cursor, winner
	}

	return -1, winner
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
		nextMove, err := m.bot.GetNextMove(ctx, m.board, m.botPlayer)
		if err != nil {
			panic(err)
		}

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
	oCursor := m.cursor

	if m.cursor >= 0 && m.cursor < len(m.board.Cells)-1 {
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

	return m.cursor, m.cursor != oCursor
}

func (m model) moveLeft() (int, bool) {
	oCursor := m.cursor

	if m.cursor > 0 {
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

	return m.cursor, m.cursor != oCursor
}

func (m model) View() string {
	if m.gameOver && m.Replay {
		return ""
	}

	var highlights []int
	if m.gameOver && m.winner != tictactoe.Empty {
		highlights = m.board.GetKRow(m.winner)
	}

	s := m.header

	botTurn := m.bot != nil && m.currentPlayer == m.botPlayer && !m.gameOver

	s += "Current player: "
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
			mark = []string{"o", "x", " ", " "}[rand.N(4)]
			mark = thinkingColors[rand.IntN(len(thinkingColors))](mark)
		}

		switch p {
		case tictactoe.P1:
			mark = p1Style(p.Mark())
		case tictactoe.P2:
			mark = p2Style(p.Mark())
		}

		bStyle := bracketStyle
		winningRow := slices.Contains(highlights, i)

		if winningRow {
			bStyle = winningRowStyle
		}

		if m.board.Turn > 0 && m.board.LastMove == i && p != tictactoe.Empty {
			bStyle = lastMoveBracketStyle
			if winningRow {
				bStyle = lastWinningRowStyle
			}
		}

		s += fmt.Sprintf("%s%s%s", bStyle("["), mark, bStyle("]"))
		if (i+1)%m.board.N == 0 {
			s += "\n"
		}
	}

	stats := m.bot.Stats()
	if m.currentPlayer != m.botPlayer && stats != nil {
		s += "\n"
		m := m.board.GetMove(stats.BestMove)
		s += fmt.Sprintf(
			"Found move: %s\nDid %s iterations over %s (Total: %s)\nMost visited node for move across workers: Visits: %s - Score: %s\n",
			statStyle1(fmt.Sprintf("(%d, %d)", m.X+1, m.Y+1)),
			statStyle2(strconv.Itoa(stats.NumIterations)),
			statStyle2(stats.RealThinkTime.Round(time.Millisecond).String()),
			statStyle2(stats.ActualThinkTime.Round(time.Millisecond).String()),
			statStyle1(strconv.Itoa(stats.MoveVisits)),
			statStyle1(fmt.Sprintf("%f", stats.MoveWins/float64(stats.MoveVisits))),
		)

		if stats.TacticalOverride {
			if stats.NumIterations == 0 {
				s += statStyle1("FAST ")
			}

			s += cursorStyle("TACTICAL OVERRIDE\n")
		}
	}

	if m.gameOver {
		s += "\n" + gameOverText

		s += "\nTHE WINNER IS: "
		if m.winner == tictactoe.Empty {
			s += cursorStyle("NO ONE\n")
			return s
		}

		switch m.winner {
		case tictactoe.P1:
			s += p1Style(m.winner.Mark())
		case tictactoe.P2:
			s += p2Style(m.winner.Mark())
		}

		s += "\n"
		return s
	}

	return s
}

const gameOverText = `ＧＡＭＥ ＯＶＥＲ`
