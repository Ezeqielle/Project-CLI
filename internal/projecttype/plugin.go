package projecttype

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Plugin interface {
	ID() string
	DisplayName() string
	Description() string

	NewWizard() tea.Model
}
