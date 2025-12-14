package settings

import (
	"fmt"
	"strings"
	"time"

	"github.com/Zarux/ticntacntoen/pkg/tictactoe"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	listSelectorStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).Render
)

var nChoiceRange = []int{3, 15}
var kChoiceRange = []int{3, 6}
var timeChoiceRange = []int{1, 60}
var pChoiceRange = []int{0, 1}

type settings struct {
	N         int
	K         int
	ThinkTime time.Duration
	P         tictactoe.Player
}

type choiceLevel int

const (
	choiceLevelP choiceLevel = iota
	choiceLevelN
	choiceLevelK
	choiceLevelThink
)

type model struct {
	cursor      int
	choiceLevel choiceLevel
	header      string

	settings settings

	clear bool
}

func (m model) GetSettings() settings {
	return m.settings
}

func InitialModel(header string) *model {
	return &model{
		header: header,
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var choices []int
	if m.choiceLevel == choiceLevelP {
		for i := pChoiceRange[0]; i <= pChoiceRange[1]; i++ {
			choices = append(choices, i)
		}
	}

	if m.choiceLevel == choiceLevelN {
		for i := nChoiceRange[0]; i <= nChoiceRange[1]; i++ {
			choices = append(choices, i)
		}
	}

	if m.choiceLevel == choiceLevelK {
		for i := kChoiceRange[0]; i <= kChoiceRange[1] && i <= m.settings.N; i++ {
			choices = append(choices, i)
		}
	}

	if m.choiceLevel == choiceLevelThink {
		for i := timeChoiceRange[0]; i <= timeChoiceRange[1]; i++ {
			choices = append(choices, i)
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.clear = true
			return m, tea.Quit

		case "enter":
			if m.choiceLevel == choiceLevelP {
				p := tictactoe.P1
				if choices[m.cursor] == 1 {
					p = tictactoe.P2
				}

				m.settings.P = p
			}

			if m.choiceLevel == choiceLevelN {
				m.settings.N = choices[m.cursor]
			}

			if m.choiceLevel == choiceLevelK {
				m.settings.K = choices[m.cursor]
			}

			if m.choiceLevel == choiceLevelThink {
				m.settings.ThinkTime = time.Duration(choices[m.cursor]) * time.Second
			}

			m.choiceLevel++
			if m.choiceLevel > choiceLevelThink {
				m.clear = true
				return m, tea.Quit
			}

			m.cursor = 0
			return m, nil

		case "down", "j":
			m.cursor++
			if m.cursor >= len(choices) {
				m.cursor = 0
			}

		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(choices) - 1
			}
		}
	}

	return m, nil
}

func (m *model) View() string {
	if m.clear {
		return ""
	}

	var choices []int

	s := strings.Builder{}
	s.WriteString(m.header)

	if m.choiceLevel == choiceLevelP {
		s.WriteString("Choose stone:\n")
		for i := pChoiceRange[0]; i <= pChoiceRange[1]; i++ {
			choices = append(choices, i)
		}
	}

	if m.choiceLevel == choiceLevelN {
		s.WriteString("Choose board size:\n")
		for i := nChoiceRange[0]; i <= nChoiceRange[1]; i++ {
			choices = append(choices, i)
		}
	}

	if m.choiceLevel == choiceLevelK {
		s.WriteString("Choose win condition:\n")
		for i := kChoiceRange[0]; i <= kChoiceRange[1] && i <= m.settings.N; i++ {
			choices = append(choices, i)
		}
	}

	if m.choiceLevel == choiceLevelThink {
		s.WriteString("Choose bot think time:\n")
		for i := timeChoiceRange[0]; i <= timeChoiceRange[1]; i++ {
			choices = append(choices, i)
		}
	}

	aroundCursor := 3

	minItem := m.cursor - aroundCursor
	maxItem := m.cursor + aroundCursor
	if minItem < 0 {
		maxItem = aroundCursor * 2
		minItem = 0
	}

	if maxItem > len(choices) {
		minItem = len(choices) - aroundCursor*2
		maxItem = len(choices)
	}

	for i, v := range choices {
		if i < minItem || i > maxItem {
			continue
		}

		if m.cursor == i {
			s.WriteString(listSelectorStyle("(•) "))
		} else {
			s.WriteString(listSelectorStyle("( ) "))
		}

		if m.choiceLevel == choiceLevelP {
			switch v {
			case 0:
				s.WriteString("X (first)")
			case 1:
				s.WriteString("O")
			}
		}

		if m.choiceLevel == choiceLevelN {
			s.WriteString(fmt.Sprintf("%dx%d", v, v))
		}

		if m.choiceLevel == choiceLevelK {
			s.WriteString(fmt.Sprintf("%d in row", v))
		}

		if m.choiceLevel == choiceLevelThink {
			s.WriteString(fmt.Sprintf("%ds of thinking", v))
		}

		s.WriteString("\n")
	}

	return s.String()
}
