package main

import (
	"fmt"
	"os"

	"github.com/Shivam583-hue/TrueKanban/db"
	"github.com/Shivam583-hue/TrueKanban/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	db.Init()
	defer db.Close()
	mainModel := tui.New()
	models := []tea.Model{mainModel, tui.NewForm(0)}
	tui.SetModels(models)
	p := tea.NewProgram(mainModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
