package screen

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type tui struct {
	cursor   int
	choices  []string
	selected string
}

func initialModel() tui {
	return tui{
		choices: []string{"🔌 Connect to VPN", "❌ Disconnect VPN", "📊 VPN Status", "🧾 View Logs", "🚪 Exit"},
	}
}

func (m tui) Init() tea.Cmd {
	return nil
}

func (m tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.choices[m.cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m tui) View() string {
	s := headerStyle.Render("\n🔧 vpnctl: Choose an action\n\n")

	for i, choice := range m.choices {
		cursor := "  "
		lineStyle := normalStyle
		if m.cursor == i {
			cursor = cursorStyle.Render("❯ ")
			lineStyle = selectedStyle
		}
		s += fmt.Sprintf("%s%s\n", cursor, lineStyle.Render(choice))
	}

	if m.selected != "" {
		s += fmt.Sprintf("\n👉 Selected: %s\n", selectedStyle.Render(m.selected))
	}

	return s
}

