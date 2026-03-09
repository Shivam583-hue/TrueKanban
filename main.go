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
	m := tui.New()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
