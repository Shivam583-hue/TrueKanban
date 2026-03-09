package tui

import (
	"github.com/Shivam583-hue/TrueKanban/db"
	"github.com/Shivam583-hue/TrueKanban/types"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const divisor = 3

var (
	columnStyle  = lipgloss.NewStyle().Padding(1, 2)
	focusedStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
)

type status int

const (
	todo status = iota
	inProgress
	done
)

type model struct {
	quitting bool
	focused  status
	lists    []list.Model
	loaded   bool
}

func (m *model) initLists(width, height int) {
	defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), width/divisor, height/2)
	defaultList.SetShowHelp(false)
	m.lists = []list.Model{defaultList, defaultList, defaultList}

	m.lists[todo].Title = "To Do"
	m.lists[todo].SetItems(db.Fetch("todo"))

	m.lists[inProgress].Title = "In Progress"
	m.lists[inProgress].SetItems(db.Fetch("inProgress"))

	m.lists[done].Title = "Done"
	m.lists[done].SetItems(db.Fetch("done"))
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	if m.loaded {
		todoView := m.lists[todo].View()
		inProgressView := m.lists[inProgress].View()
		doneView := m.lists[done].View()
		switch m.focused {
		case inProgress:
			return lipgloss.JoinHorizontal(
				lipgloss.Left,
				columnStyle.Render(todoView),
				focusedStyle.Render(inProgressView),
				columnStyle.Render(doneView),
			)
		case done:
			return lipgloss.JoinHorizontal(
				lipgloss.Left,
				columnStyle.Render(todoView),
				columnStyle.Render(inProgressView),
				focusedStyle.Render(doneView),
			)
		default:
			return lipgloss.JoinHorizontal(
				lipgloss.Left,
				focusedStyle.Render(todoView),
				columnStyle.Render(inProgressView),
				columnStyle.Render(doneView),
			)
		}
	} else {
		return "Loading..."
	}
}

func New() *model {
	return &model{}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.loaded {
			columnStyle = columnStyle.Width(msg.Width / divisor)
			focusedStyle = focusedStyle.Width(msg.Width / divisor)
			m.initLists(msg.Width, msg.Height)
			m.loaded = true
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			m.Prev()
		case "right", "l":
			m.Next()
		case "enter":
			return m, m.MoveToNext
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
	return m, cmd
}

func (m *model) MoveToNext() tea.Msg {
	selectedItem := m.lists[m.focused].SelectedItem()
	selectedTask := selectedItem.(types.Task)
	m.lists[selectedTask.Status].RemoveItem(m.lists[m.focused].Index())
	selectedTask.Next()
	// m.lists[selectedTask.status].InsertItem(len(m.lists[selectedTask.status].Items())-1, list.Item(selectedTask))
	m.lists[selectedTask.Status].InsertItem(
		len(m.lists[selectedTask.Status].Items()), // Correct index to append
		list.Item(selectedTask),
	)
	return nil
}

func (m *model) Next() {
	if m.focused == done {
		m.focused = todo
	} else {
		m.focused++
	}
}

func (m *model) Prev() {
	if m.focused == todo {
		m.focused = done
	} else {
		m.focused--
	}
}
