# TrueKanban — Code Documentation

This document explains the architecture and code flow for anyone wanting to understand, modify, or extend TrueKanban.

---

## Architecture Overview

The app is split into four packages, each with a single responsibility:

```
main.go     → wiring only (start db, build models, launch bubbletea)
db/         → SQLite: all reads and writes
types/      → shared data types (Task, Status) with zero UI dependencies
tui/        → everything visual: the board, the form, keybindings, rendering
```

The strict rule is: **`types` and `db` never import `tui`**. Data flows one way — `tui` calls `db`, `db` returns `types`.

---

## The Bubbletea Pattern

Bubbletea is an Elm-inspired TUI framework. Every screen in the app is a **Model** that implements three methods:

```go
Init()              // runs once on startup, returns an initial Cmd
Update(msg) Model   // receives an event, returns updated model + next Cmd
View() string       // returns the string to render this frame
```

**Messages (`tea.Msg`)** are how everything communicates. Key presses, window resizes, and your own custom types (like `types.Task`) can all be messages. When `Update` returns a `Cmd`, Bubbletea runs it and feeds the result back as the next message.

---

## Package: `types`

### `status.go`

```go
type Status int
const (
    Todo       Status = 0
    InProgress Status = 1
    Done       Status = 2
)
```

`Status` is just an integer that maps to a column. It's used as an array index into `model.lists`, which is why it must stay as `int` and not `string`.

### `task.go`

```go
type Task struct {
    Id        int
    TaskTitle string
    Status    Status
}
```

`Task` implements `list.Item` (from `charmbracelet/bubbles/list`) by providing three methods:

```go
func (t Task) Title() string       { return t.TaskTitle }
func (t Task) Description() string { ... }
func (t Task) FilterValue() string { return t.TaskTitle }
```

`Next()` advances a task's status forward by one column, wrapping from Done back to Todo.

> **Why `TaskTitle` and not `Title`?** The field and the method can't share the same name in Go. Renaming the field avoids the collision.

---

## Package: `db`

All database interaction lives here. The package-level `var db *sql.DB` is the single shared connection, opened once in `Init()` and closed via `Close()` (deferred in `main`).

### Functions

**`Init()`** — opens the SQLite file and creates the `tasks` table if it doesn't exist. The schema stores `status` as an `INTEGER` (matching the `Status` int type).

**`Fetch(status Status) []list.Item`** — queries all tasks for a given column and returns them as `[]list.Item`, ready to be passed directly into a `list.Model`.

**`Insert(title string, status Status)`** — inserts a new task. Called by the form when the user submits.

**`Update(id int, newStatus Status)`** — updates a task's column after it's moved. Called by `MoveToNext`.

**`Delete(id int)`** — removes a task by ID. Called when the user presses `x` or `enter` on the Done column.

---

## Package: `tui`

This is the largest package and contains all UI logic.

### Shared state: `models` slice

```go
var models []tea.Model  // [0] = board, [1] = form
```

This slice is how the two screens hand control to each other. When the user presses `n` on the board, the board saves itself into `models[0]`, creates a fresh form in `models[1]`, and returns the form as the active model. When the form submits, it returns `models[0]` (the saved board) as the active model.

`SetModels` is called from `main.go` to inject the initial slice so both models share the same reference.

### The Board (`model`)

```go
type model struct {
    quitting bool
    focused  types.Status   // which column is currently selected (0, 1, or 2)
    lists    []list.Model   // the three column lists
    loaded   bool           // has the terminal size been received yet?
}
```

**`initLists(width, height)`** — sets up three `list.Model` columns and populates each from the database. Called once inside `Update` when the first `tea.WindowSizeMsg` arrives (we can't set sizes before knowing the terminal dimensions).

**`Update` key handling:**

| Key | What happens |
|-----|-------------|
| `←/h`, `→/l` | calls `Prev()` / `Next()` to shift `m.focused` |
| `enter` | calls `MoveToNext` as a Cmd, which returns a `refreshMsg` |
| `x` | deletes from DB, fires `refreshMsg` |
| `n` | saves board to `models[0]`, switches to form |

**`refreshMsg`** — a custom message type. Whenever the database changes, `MoveToNext`, `x`, and the `case types.Task` handler all return `m.refresh` as a command, which sends a `refreshMsg`. The `case refreshMsg` branch in `Update` calls `db.Fetch` for all three columns and reloads the lists. This keeps the UI always in sync with the database.

**`View`** — renders three columns side by side using `lipgloss.JoinHorizontal`. The focused column gets `focusedStyle` (rounded border), the others get plain `columnStyle`.

### The Form (`Form`)

```go
type Form struct {
    focused types.Status   // which column the new task will go into
    title   textinput.Model
}
```

A single-field form. The `focused` field is set when the form is created (from `m.focused` on the board), so the task lands in the right column on submit.

**`Update`:**
- `esc` → return to board without saving
- `enter` → call `createTask`, which inserts into DB and returns a `types.Task` as a message
- Anything else → forward to the `textinput`

**`createTask`** returns a `types.Task` value. Back on the board, `case types.Task:` in `Update` catches this message and triggers a refresh.

---

## Data Flow: Creating a New Task

```
User presses "n" on the board
  → board saves itself to models[0]
  → NewForm(m.focused) created, stored in models[1]
  → form becomes the active model

User types a title and presses Enter
  → form calls db.Insert(title, focused)
  → form returns types.Task{} as a tea.Msg
  → models[0] (the board) becomes the active model

Board receives case types.Task:
  → fires m.refresh command
  → refreshMsg triggers db.Fetch for all three columns
  → lists updated, new task appears on screen
```

## Data Flow: Moving a Task

```
User presses Enter on a task
  → MoveToNext() called as a Cmd
  → task removed from current column in memory
  → task.Next() increments its Status
  → db.Update(task.Id, task.Status) persists the change
  → returns refreshMsg
  → all three columns refetched from DB
```

---

## Adding a New Feature

**New keybinding on the board:** add a `case` inside the `tea.KeyMsg` switch in `model.Update`.

**New DB operation:** add a function to `db/db.go`. Keep it focused — one function per SQL statement.

**New field on a task:** add it to `types.Task`, update the `CREATE TABLE` SQL in `db.Init`, update `db.Fetch` to scan it, and update `db.Insert` to write it.

**New screen:** create a new model type in `tui/tui.go` (or a new file in the `tui` package), give it `Init/Update/View`, add it to the `models` slice in `main.go`, and wire up the switch using the same pattern as the form.
