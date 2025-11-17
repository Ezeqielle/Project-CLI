package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ezeqielle/pcli/internal/plugins"
	"github.com/ezeqielle/pcli/internal/postplugin"
	"github.com/ezeqielle/pcli/internal/ui"
)

func main() {
	if err := run(); err != nil {
		log.Println("pcli exited with error:", err)
		os.Exit(1)
	}
}

func run() error {
	plugins.RegisterAll()

	postplugin.RegisterAll()

	m := ui.NewTypeChooserModel()
	p := tea.NewProgram(m)

	_, err := p.Run()
	return err
}
