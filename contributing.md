# Contributing to TrueKanban

Thanks for taking the time to contribute! Here's everything you need to get started.

---

## Getting Started

1. Fork the repo and clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/TrueKanban.git
   cd TrueKanban
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the app:
   ```bash
   go run .
   ```

---

## Making Changes

- **Branch off `main`** for all changes:
  ```bash
  git checkout -b feat/your-feature-name
  ```

- **Keep commits small and focused.** One logical change per commit.

- **Test manually** by running `go run .` and verifying your change works end-to-end.

- **Run `go vet ./...`** before submitting to catch common issues.

---

## Pull Requests

- Give your PR a clear title describing what it does, e.g. `feat: add due dates to tasks`
- Include a short description of why the change is needed
- If it fixes a bug, reference the issue: `Fixes #42`
- Keep PRs focused — avoid bundling unrelated changes together

---

## Reporting Bugs

Open an issue with:
- What you expected to happen
- What actually happened
- Steps to reproduce
- Your OS and Go version (`go version`)

---

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep the package structure clean:
  - `types/` — pure data, no UI imports
  - `db/` — only database logic, no UI imports
  - `tui/` — all UI code lives here
  - `main.go` — wiring only, minimal logic

---

## Ideas for Contribution

- [ ] Task editing (rename a task in-place)
- [ ] Due dates
- [ ] Multiple boards
- [ ] Task priorities
- [ ] Search / filter
- [ ] Export to markdown
