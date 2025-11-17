package postplugin

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Plugin interface {
	ID() string
	DisplayName() string

	NewWizard(projectPath, projectType string) tea.Model
}
