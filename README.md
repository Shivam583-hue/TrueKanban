# TrueKanban

A terminal-based Kanban board built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and SQLite. Manage your tasks without leaving the terminal.

> 📹 Demo coming soon

---

## Features

- Three-column Kanban board: **Todo → In Progress → Done**
- Persistent storage via SQLite — tasks survive restarts
- Keyboard-driven, no mouse required
- Minimal, distraction-free UI powered by [Lip Gloss](https://github.com/charmbracelet/lipgloss)

---

## Installation

### Linux (x64)
```bash
curl -L https://github.com/Shivam583-hue/TrueKanban/releases/latest/download/truekanban-linux-amd64 -o truekanban
chmod +x truekanban
sudo mv truekanban /usr/local/bin/
truekanban
```

### macOS (Apple Silicon)
```bash
curl -L https://github.com/Shivam583-hue/TrueKanban/releases/latest/download/truekanban-macos-arm64 -o truekanban
chmod +x truekanban
sudo mv truekanban /usr/local/bin/
truekanban
```

### macOS (Intel)
```bash
curl -L https://github.com/Shivam583-hue/TrueKanban/releases/latest/download/truekanban-macos-amd64 -o truekanban
chmod +x truekanban
sudo mv truekanban /usr/local/bin/
truekanban
```

### Windows (PowerShell)
```powershell
curl -L https://github.com/Shivam583-hue/TrueKanban/releases/latest/download/truekanban-windows-amd64.exe -o truekanban.exe
move truekanban.exe C:\Windows\System32\truekanban.exe
truekanban
```

> **Note:** Requires CGO (for SQLite). Make sure you have `gcc` installed.
> On Ubuntu/Debian: `sudo apt install gcc`
> On macOS: `xcode-select --install`

---

## Keybindings

| Key | Action |
|-----|--------|
| `n` | Create a new task in the focused column |
| `enter` | Move selected task to the next column |
| `x` | Delete selected task |
| `←` / `h` | Focus previous column |
| `→` / `l` | Focus next column |
| `esc` | Cancel (in form) |
| `q` / `ctrl+c` | Quit |


## Project Structure

```
TrueKanban/
├── main.go           # Entry point — wires db, tui, and bubbletea together
├── db/
│   └── db.go         # SQLite operations: Init, Insert, Fetch, Update, Delete
├── tui/
│   └── tui.go        # All UI logic: board model, form model, keybindings, rendering
├── types/
│   ├── task.go       # Task struct and list.Item implementation
│   └── status.go     # Status type (Todo=0, InProgress=1, Done=2)
└── task.db           # Auto-created SQLite database 
```

---

## Requirements

- Go 1.21+
- GCC (for `go-sqlite3` CGO compilation)
