package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ezeqielle/pcli/internal/projecttype"
)

type TypeChooserModel struct {
	plugins       []projecttype.Plugin
	selectedIndex int
}

func NewTypeChooserModel() TypeChooserModel {
	plugins := projecttype.All()

	return TypeChooserModel{
		plugins:       plugins,
		selectedIndex: 0,
	}
}

func (m TypeChooserModel) Init() tea.Cmd {
	return nil
}

func (m TypeChooserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()

		switch s {
		case "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if len(m.plugins) == 0 {
				return m, nil
			}
			m.selectedIndex--
			if m.selectedIndex < 0 {
				m.selectedIndex = len(m.plugins) - 1
			}
			return m, nil

		case "down", "j":
			if len(m.plugins) == 0 {
				return m, nil
			}
			m.selectedIndex++
			if m.selectedIndex >= len(m.plugins) {
				m.selectedIndex = 0
			}
			return m, nil

		case "enter":
			if len(m.plugins) == 0 {
				return m, nil
			}
			plugin := m.plugins[m.selectedIndex]
			return plugin.NewWizard(), nil
		}
	}

	return m, nil
}

func (m TypeChooserModel) View() string {
	if len(m.plugins) == 0 {
		return "No project types registered.\n\n[ctrl+c] Quit\n"
	}

	var b strings.Builder

	b.WriteString("pcli – Create project\n\n")
	b.WriteString("Select project type (↑/↓, enter):\n\n")

	for i, p := range m.plugins {
		cursor := " "
		if i == m.selectedIndex {
			cursor = ">"
		}

		fmt.Fprintf(
			&b,
			"%s %s – %s\n",
			cursor,
			p.DisplayName(),
			p.Description(),
		)
	}

	b.WriteString("\n[ctrl+c] Quit\n")

	return b.String()
}
