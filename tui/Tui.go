package tui

import (
	"github.com/Shivam583-hue/TrueKanban/db"
	"github.com/Shivam583-hue/TrueKanban/types"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type refreshMsg struct{}

// refreshMsg used to refetch after every action
func (m *model) refresh() tea.Msg {
	return refreshMsg{}
}

// -- indices into the shared models slice --
const (
	mainModel = 0
	formModel = 1
)

// -- column indices --
const (
	todo       types.Status = 0
	inProgress types.Status = 1
	done       types.Status = 2
)

const divisor = 3

var models []tea.Model

var (
	columnStyle  = lipgloss.NewStyle().Padding(1, 2)
	focusedStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
)

// ── Main board ────────────────────────────────────────────────

type model struct {
	quitting bool
	help     help.Model
	focused  types.Status
	lists    []list.Model
	loaded   bool
}

func New() *model {
	return &model{help: help.New()}
}

func SetModels(m []tea.Model) {
	models = m
}

func (m *model) initLists(width, height int) {
	defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), width/divisor, height/2)
	defaultList.SetShowHelp(false)
	m.lists = []list.Model{defaultList, defaultList, defaultList}

	m.lists[todo].Title = "To Do"
	m.lists[todo].SetItems(db.Fetch(types.Todo))
	m.lists[inProgress].Title = "In Progress"
	m.lists[inProgress].SetItems(db.Fetch(types.InProgress))
	m.lists[done].Title = "Done"
	m.lists[done].SetItems(db.Fetch(types.Done))
}

func (m model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.loaded {
			columnStyle = columnStyle.Width(msg.Width / divisor)
			focusedStyle = focusedStyle.Width(msg.Width / divisor)
			m.initLists(msg.Width, msg.Height)
			m.loaded = true
		}
	case refreshMsg:
		m.lists[todo].SetItems(db.Fetch(types.Todo))
		m.lists[inProgress].SetItems(db.Fetch(types.InProgress))
		m.lists[done].SetItems(db.Fetch(types.Done))
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "left", "h":
			m.Prev()
		case "right", "l":
			m.Next()
		case "enter":
			if len(m.lists[m.focused].Items()) == 0 {
				return m, nil
			}
			if m.focused == done {
				selected := m.lists[m.focused].SelectedItem().(types.Task)
				db.Delete(selected.Id)
				return m, m.refresh
				// selected := m.lists[m.focused].SelectedItem().(types.Task)
				// db.Delete(selected.Id)
				// m.lists[m.focused].RemoveItem(m.lists[m.focused].Index())
				// return m, nil
			}
			return m, m.MoveToNext
		case "x":
			if len(m.lists[m.focused].Items()) > 0 {
				selected := m.lists[m.focused].SelectedItem().(types.Task)
				db.Delete(selected.Id)
				return m, m.refresh
				// selected := m.lists[m.focused].SelectedItem().(types.Task)
				// db.Delete(selected.Id)
				// m.lists[m.focused].RemoveItem(m.lists[m.focused].Index())
			}
			return m, nil
		case "n":
			models[mainModel] = m
			models[formModel] = NewForm(m.focused)
			return models[formModel], nil
		}
	case types.Task:
		return m, m.refresh
	}
	var cmd tea.Cmd
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	if !m.loaded {
		return "Loading..."
	}

	todoView := m.lists[todo].View()
	inProgressView := m.lists[inProgress].View()
	doneView := m.lists[done].View()

	var board string

	switch m.focused {
	case inProgress:
		board = lipgloss.JoinHorizontal(
			lipgloss.Left,
			columnStyle.Render(todoView),
			focusedStyle.Render(inProgressView),
			columnStyle.Render(doneView),
		)

	case done:
		board = lipgloss.JoinHorizontal(
			lipgloss.Left,
			columnStyle.Render(todoView),
			columnStyle.Render(inProgressView),
			focusedStyle.Render(doneView),
		)

	default:
		board = lipgloss.JoinHorizontal(
			lipgloss.Left,
			focusedStyle.Render(todoView),
			columnStyle.Render(inProgressView),
			columnStyle.Render(doneView),
		)
	}

	helpView := m.help.View(keys)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		board,
		"",
		helpView,
	)
}

func (m *model) MoveToNext() tea.Msg {
	selected := m.lists[m.focused].SelectedItem().(types.Task)
	selected.Next()
	db.Update(selected.Id, selected.Status)
	return refreshMsg{}
	// selected := m.lists[m.focused].SelectedItem().(types.Task)
	// m.lists[selected.Status].RemoveItem(m.lists[m.focused].Index())
	// selected.Next()
	// db.Update(selected.Id, selected.Status)
	// m.lists[selected.Status].InsertItem(len(m.lists[selected.Status].Items()), list.Item(selected))
	// return nil
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

// ── Form ──────────────────────────────────────────────────────

type Form struct {
	focused types.Status
	title   textinput.Model
}

func NewForm(focused types.Status) Form {
	ti := textinput.New()
	ti.Placeholder = "Enter title"
	ti.Focus()
	return Form{focused: focused, title: ti}
}

func (f Form) Init() tea.Cmd { return nil }

func (f Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return f, tea.Quit
		case "esc":
			return models[mainModel], nil
		case "enter":
			return models[mainModel], f.createTask
		}
	}
	f.title, cmd = f.title.Update(msg)
	return f, cmd
}

func (f Form) createTask() tea.Msg {
	db.Insert(f.title.Value(), f.focused)
	return types.Task{
		TaskTitle: f.title.Value(),
		Status:    f.focused,
	}
}

func (f Form) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		"Create New Task",
		"",
		f.title.View(),
		"",
		"Enter to save • Esc to cancel",
	)
}

// ── Help ──────────────────────────────────────────────────────
type keyMap struct {
	Left   key.Binding
	Right  key.Binding
	Move   key.Binding
	Delete key.Binding
	New    key.Binding
	Quit   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Move, k.Delete, k.New, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right},
		{k.Move, k.Delete},
		{k.New, k.Quit},
	}
}

var keys = keyMap{
	Left:   key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "move left")),
	Right:  key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "move right")),
	Move:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "move task forward")),
	Delete: key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete task")),
	New:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new task")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}
