package tui

import (
	"strings"
	"testing"

	"github.com/Shivam583-hue/TrueKanban/db"
	"github.com/Shivam583-hue/TrueKanban/types"
	tea "github.com/charmbracelet/bubbletea"
)

// ── TestMain: global setup for the whole package ────────────────
//
// The TUI tests need a real database because initLists() calls
// db.Fetch(), and Update() calls db.Insert/Delete/Update.
// We initialize it once here for all tests.

func TestMain(m *testing.M) {
	db.Init()
	m.Run()
	db.Close()
}

// ── Helpers ─────────────────────────────────────────────────────
//
// These set up common state so individual tests stay short.
// t.Helper() ensures failures point to the test, not the helper.

// cleanDB wipes all tasks so each test starts fresh.
func cleanDB(t *testing.T) {
	t.Helper()
	// Delete all tasks across all statuses by fetching and deleting each.
	for _, s := range []types.Status{types.Todo, types.InProgress, types.Done} {
		items := db.Fetch(s)
		for _, item := range items {
			task := item.(types.Task)
			db.Delete(task.Id)
		}
	}
}

// newLoadedModel returns a model that has received a WindowSizeMsg,
// so its lists are initialized and it's ready for interaction.
func newLoadedModel(t *testing.T) *model {
	t.Helper()
	m := New()
	// Simulate the terminal reporting its size — this triggers initLists().
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if !m.loaded {
		t.Fatal("model should be loaded after WindowSizeMsg")
	}
	return m
}

// setupModelsSlice initializes the package-level `models` slice,
// which is needed by the "n" key handler and Form.
func setupModelsSlice(t *testing.T, m *model) {
	t.Helper()
	SetModels([]tea.Model{m, NewForm(0)})
}

// sendKey simulates a key press and returns the resulting model and cmd.
func sendKey(m *model, keyStr string) (tea.Model, tea.Cmd) {
	return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyStr)})
}

// sendSpecialKey simulates a special key (enter, esc, left, right, etc.).
func sendSpecialKey(m *model, keyType tea.KeyType) (tea.Model, tea.Cmd) {
	return m.Update(tea.KeyMsg{Type: keyType})
}

// ── Board: New() and Init() ─────────────────────────────────────

func TestNewReturnsUnloadedModel(t *testing.T) {
	m := New()

	if m.loaded {
		t.Error("new model should not be loaded")
	}
	if m.quitting {
		t.Error("new model should not be quitting")
	}
	if m.focused != todo {
		t.Errorf("new model focused = %d, want %d (todo)", m.focused, todo)
	}
}

func TestInitReturnsNil(t *testing.T) {
	m := New()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

// ── Board: WindowSizeMsg (lazy loading) ─────────────────────────

func TestWindowSizeMsgInitializesLists(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)

	if len(m.lists) != 3 {
		t.Fatalf("expected 3 lists, got %d", len(m.lists))
	}
	if m.lists[todo].Title != "To Do" {
		t.Errorf("lists[0].Title = %q, want %q", m.lists[todo].Title, "To Do")
	}
	if m.lists[inProgress].Title != "In Progress" {
		t.Errorf("lists[1].Title = %q, want %q", m.lists[inProgress].Title, "In Progress")
	}
	if m.lists[done].Title != "Done" {
		t.Errorf("lists[2].Title = %q, want %q", m.lists[done].Title, "Done")
	}
}

func TestWindowSizeMsgOnlyLoadOnce(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)

	// Send a second WindowSizeMsg — should be ignored (loaded is already true).
	m.Update(tea.WindowSizeMsg{Width: 200, Height: 80})

	// If it re-initialized, the list width would change. But since initLists
	// only runs once, we just verify loaded is still true and no panic.
	if !m.loaded {
		t.Error("model should still be loaded after second WindowSizeMsg")
	}
}

// ── Board: Focus navigation (Next / Prev) ───────────────────────
//
// These test the column cycling logic WITHOUT going through Update,
// proving the methods work in isolation.

func TestNextCyclesFocus(t *testing.T) {
	m := New()

	tests := []struct {
		name string
		want types.Status
	}{
		{"todo -> inProgress", inProgress},
		{"inProgress -> done", done},
		{"done -> todo (wrap)", todo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.Next()
			if m.focused != tt.want {
				t.Errorf("focused = %d, want %d", m.focused, tt.want)
			}
		})
	}
}

func TestPrevCyclesFocus(t *testing.T) {
	m := New()

	tests := []struct {
		name string
		want types.Status
	}{
		{"todo -> done (wrap)", done},
		{"done -> inProgress", inProgress},
		{"inProgress -> todo", todo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.Prev()
			if m.focused != tt.want {
				t.Errorf("focused = %d, want %d", m.focused, tt.want)
			}
		})
	}
}

// ── Board: Key handling via Update() ────────────────────────────

func TestRightKeyMovesFocus(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)

	sendKey(m, "l") // vim-style right
	if m.focused != inProgress {
		t.Errorf("after 'l': focused = %d, want %d", m.focused, inProgress)
	}

	sendSpecialKey(m, tea.KeyRight)
	if m.focused != done {
		t.Errorf("after right arrow: focused = %d, want %d", m.focused, done)
	}
}

func TestLeftKeyMovesFocus(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)

	sendKey(m, "h") // wraps from todo -> done
	if m.focused != done {
		t.Errorf("after 'h': focused = %d, want %d", m.focused, done)
	}

	sendSpecialKey(m, tea.KeyLeft)
	if m.focused != inProgress {
		t.Errorf("after left arrow: focused = %d, want %d", m.focused, inProgress)
	}
}

func TestQuitKey(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)

	_, cmd := sendKey(m, "q")
	if !m.quitting {
		t.Error("model should be quitting after 'q'")
	}
	if cmd == nil {
		t.Error("quit should return a non-nil cmd (tea.Quit)")
	}
}

func TestCtrlCQuits(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)

	_, cmd := sendSpecialKey(m, tea.KeyCtrlC)
	if !m.quitting {
		t.Error("model should be quitting after ctrl+c")
	}
	if cmd == nil {
		t.Error("ctrl+c should return a non-nil cmd")
	}
}

func TestEnterOnEmptyListDoesNothing(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)

	// The list is empty, so enter should be a no-op (no panic).
	_, cmd := sendSpecialKey(m, tea.KeyEnter)
	if cmd != nil {
		t.Error("enter on empty list should return nil cmd")
	}
}

func TestEnterOnTodoMovesTask(t *testing.T) {
	cleanDB(t)
	db.Insert("Move me", types.Todo)
	m := newLoadedModel(t)

	// Verify the task is in the Todo list
	if len(m.lists[todo].Items()) != 1 {
		t.Fatalf("expected 1 todo item, got %d", len(m.lists[todo].Items()))
	}

	// Press enter — this returns MoveToNext as a cmd
	_, cmd := sendSpecialKey(m, tea.KeyEnter)
	if cmd == nil {
		t.Fatal("enter on a todo item should return a cmd (MoveToNext)")
	}

	// Execute the cmd — it calls db.Update and returns refreshMsg
	msg := cmd()
	if _, ok := msg.(refreshMsg); !ok {
		t.Errorf("MoveToNext should return refreshMsg, got %T", msg)
	}

	// Send the refreshMsg to re-fetch from DB
	m.Update(msg)

	// Task should have moved to InProgress
	if len(m.lists[todo].Items()) != 0 {
		t.Errorf("todo list should be empty after move, has %d items", len(m.lists[todo].Items()))
	}
	if len(m.lists[inProgress].Items()) != 1 {
		t.Errorf("inProgress list should have 1 item after move, has %d", len(m.lists[inProgress].Items()))
	}
}

func TestEnterOnDoneDeletesTask(t *testing.T) {
	cleanDB(t)
	db.Insert("Finish me", types.Done)
	m := newLoadedModel(t)

	// Focus the Done column
	m.focused = done

	if len(m.lists[done].Items()) != 1 {
		t.Fatalf("expected 1 done item, got %d", len(m.lists[done].Items()))
	}

	// Press enter on a Done item — this deletes it
	_, cmd := sendSpecialKey(m, tea.KeyEnter)
	if cmd == nil {
		t.Fatal("enter on done item should return a cmd (refresh)")
	}

	// Execute the cmd and send the resulting message
	msg := cmd()
	m.Update(msg)

	if len(m.lists[done].Items()) != 0 {
		t.Errorf("done list should be empty after delete, has %d items", len(m.lists[done].Items()))
	}
}

func TestXKeyDeletesTask(t *testing.T) {
	cleanDB(t)
	db.Insert("Delete me", types.Todo)
	m := newLoadedModel(t)

	if len(m.lists[todo].Items()) != 1 {
		t.Fatalf("expected 1 todo item, got %d", len(m.lists[todo].Items()))
	}

	_, cmd := sendKey(m, "x")
	if cmd == nil {
		t.Fatal("'x' on a task should return a cmd (refresh)")
	}

	msg := cmd()
	m.Update(msg)

	if len(m.lists[todo].Items()) != 0 {
		t.Errorf("todo list should be empty after 'x', has %d items", len(m.lists[todo].Items()))
	}
}

func TestXKeyOnEmptyListIsNoop(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)

	// Should not panic on empty list
	_, cmd := sendKey(m, "x")
	if cmd != nil {
		t.Error("'x' on empty list should return nil cmd")
	}
}

func TestNKeySwitchesToForm(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)
	setupModelsSlice(t, m)

	returned, _ := sendKey(m, "n")

	// The returned model should be a Form, not the board model
	if _, ok := returned.(Form); !ok {
		t.Errorf("'n' should switch to Form model, got %T", returned)
	}
}

func TestNKeyPassesFocusedColumnToForm(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)
	m.focused = inProgress
	setupModelsSlice(t, m)

	returned, _ := sendKey(m, "n")

	form, ok := returned.(Form)
	if !ok {
		t.Fatalf("expected Form, got %T", returned)
	}
	if form.focused != inProgress {
		t.Errorf("form.focused = %d, want %d (inProgress)", form.focused, inProgress)
	}
}

// ── Board: refreshMsg ───────────────────────────────────────────

func TestRefreshMsgReloadsFromDB(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)

	// Lists should be empty
	if len(m.lists[todo].Items()) != 0 {
		t.Fatalf("expected 0 todo items, got %d", len(m.lists[todo].Items()))
	}

	// Insert directly into DB (bypassing the UI)
	db.Insert("Sneaky task", types.Todo)

	// Send refreshMsg — should pick up the new task
	m.Update(refreshMsg{})

	if len(m.lists[todo].Items()) != 1 {
		t.Errorf("after refresh: expected 1 todo item, got %d", len(m.lists[todo].Items()))
	}
}

// ── Board: Task message triggers refresh ────────────────────────

func TestTaskMsgTriggersRefresh(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)

	// Sending a types.Task message should return a refresh cmd
	_, cmd := m.Update(types.Task{TaskTitle: "test", Status: types.Todo})
	if cmd == nil {
		t.Error("Task message should return a refresh cmd")
	}
}

// ── Board: View() ───────────────────────────────────────────────

func TestViewShowsLoadingWhenNotLoaded(t *testing.T) {
	m := New()
	view := m.View()

	if view != "Loading..." {
		t.Errorf("View() before load = %q, want %q", view, "Loading...")
	}
}

func TestViewShowsEmptyStringWhenQuitting(t *testing.T) {
	m := New()
	m.quitting = true

	view := m.View()
	if view != "" {
		t.Errorf("View() when quitting = %q, want empty string", view)
	}
}

func TestViewContainsColumnTitles(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)

	view := m.View()

	for _, title := range []string{"To Do", "In Progress", "Done"} {
		if !strings.Contains(view, title) {
			t.Errorf("View() should contain %q", title)
		}
	}
}

// ── Board: MoveToNext() ─────────────────────────────────────────

func TestMoveToNextReturnsRefreshMsg(t *testing.T) {
	cleanDB(t)
	db.Insert("Task", types.Todo)
	m := newLoadedModel(t)

	msg := m.MoveToNext()

	if _, ok := msg.(refreshMsg); !ok {
		t.Errorf("MoveToNext() returned %T, want refreshMsg", msg)
	}
}

// ── Board: refresh() method ─────────────────────────────────────

func TestRefreshMethodReturnsRefreshMsg(t *testing.T) {
	m := New()
	msg := m.refresh()

	if _, ok := msg.(refreshMsg); !ok {
		t.Errorf("refresh() returned %T, want refreshMsg", msg)
	}
}

// ── Form: NewForm ───────────────────────────────────────────────

func TestNewFormSetsStatus(t *testing.T) {
	tests := []struct {
		name   string
		status types.Status
	}{
		{"todo form", types.Todo},
		{"in-progress form", types.InProgress},
		{"done form", types.Done},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewForm(tt.status)
			if f.focused != tt.status {
				t.Errorf("NewForm(%d).focused = %d", tt.status, f.focused)
			}
		})
	}
}

func TestNewFormHasPlaceholder(t *testing.T) {
	f := NewForm(types.Todo)

	if f.title.Placeholder != "Enter title" {
		t.Errorf("placeholder = %q, want %q", f.title.Placeholder, "Enter title")
	}
}

func TestFormInitReturnsNil(t *testing.T) {
	f := NewForm(types.Todo)
	cmd := f.Init()
	if cmd != nil {
		t.Error("Form.Init() should return nil")
	}
}

// ── Form: View ──────────────────────────────────────────────────

func TestFormViewContainsInstructions(t *testing.T) {
	f := NewForm(types.Todo)
	view := f.View()

	if !strings.Contains(view, "Create New Task") {
		t.Error("form View() should contain 'Create New Task'")
	}
	if !strings.Contains(view, "Enter to save") {
		t.Error("form View() should contain 'Enter to save'")
	}
	if !strings.Contains(view, "Esc to cancel") {
		t.Error("form View() should contain 'Esc to cancel'")
	}
}

// ── Form: Update (key handling) ─────────────────────────────────

func TestFormEscReturnsToBoard(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)
	setupModelsSlice(t, m)

	f := NewForm(types.Todo)
	returned, _ := f.Update(tea.KeyMsg{Type: tea.KeyEscape})

	// Should return the main board model, not the form
	if _, ok := returned.(*model); !ok {
		t.Errorf("Esc should return *model (board), got %T", returned)
	}
}

func TestFormCtrlCQuits(t *testing.T) {
	f := NewForm(types.Todo)
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("ctrl+c in form should return a quit cmd")
	}
}

func TestFormEnterCreatesTaskAndReturnsToBoard(t *testing.T) {
	cleanDB(t)
	m := newLoadedModel(t)
	setupModelsSlice(t, m)

	f := NewForm(types.Todo)

	// Type some text into the form's text input.
	// textinput processes regular key messages to build its value.
	for _, ch := range "My new task" {
		updated, _ := f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		f = updated.(Form)
	}

	// Press enter to submit
	returned, cmd := f.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should switch back to the board
	if _, ok := returned.(*model); !ok {
		t.Errorf("enter should return *model (board), got %T", returned)
	}

	// The cmd should be createTask, which inserts into DB
	if cmd == nil {
		t.Fatal("enter should return a cmd (createTask)")
	}

	// Execute the cmd
	msg := cmd()
	task, ok := msg.(types.Task)
	if !ok {
		t.Fatalf("createTask should return types.Task, got %T", msg)
	}
	if task.TaskTitle != "My new task" {
		t.Errorf("created task title = %q, want %q", task.TaskTitle, "My new task")
	}
	if task.Status != types.Todo {
		t.Errorf("created task status = %d, want %d (Todo)", task.Status, types.Todo)
	}

	// Verify it's actually in the database
	items := db.Fetch(types.Todo)
	if len(items) != 1 {
		t.Fatalf("expected 1 task in DB, got %d", len(items))
	}
	dbTask := items[0].(types.Task)
	if dbTask.TaskTitle != "My new task" {
		t.Errorf("DB task title = %q, want %q", dbTask.TaskTitle, "My new task")
	}
}

// ── Help: keyMap ────────────────────────────────────────────────

func TestShortHelpReturnsAllBindings(t *testing.T) {
	bindings := keys.ShortHelp()

	// Should have 6 bindings: left, right, move, delete, new, quit
	if len(bindings) != 6 {
		t.Errorf("ShortHelp() returned %d bindings, want 6", len(bindings))
	}
}

func TestFullHelpReturnsGroupedBindings(t *testing.T) {
	groups := keys.FullHelp()

	// Should have 3 groups of 2 bindings each
	if len(groups) != 3 {
		t.Errorf("FullHelp() returned %d groups, want 3", len(groups))
	}
	for i, group := range groups {
		if len(group) != 2 {
			t.Errorf("FullHelp()[%d] has %d bindings, want 2", i, len(group))
		}
	}
}
