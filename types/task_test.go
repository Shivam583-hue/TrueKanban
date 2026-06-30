package types

import "testing"

// ── The simplest possible test ──────────────────────────────────
//
// Every test function:
//   1. Lives in a file ending with _test.go
//   2. Starts with "Test" (capital T)
//   3. Takes exactly one argument: *testing.T
//
// t.Errorf() marks the test as failed but keeps running.
// t.Fatalf() marks it as failed and stops immediately.

func TestTaskTitle(t *testing.T) {
	task := Task{Id: 1, TaskTitle: "Buy milk", Status: Todo}

	got := task.Title()
	want := "Buy milk"

	if got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
}

// ── Testing multiple cases with Table-Driven Tests ──────────────
//
// This is THE most common Go testing pattern. Instead of writing
// 3 separate test functions, you define a table of inputs/outputs
// and loop through them.

func TestTaskDescription(t *testing.T) {
	tests := []struct {
		name   string // name of this sub-test (shows up in output)
		status Status
		want   string
	}{
		{name: "todo task", status: Todo, want: "Todo"},
		{name: "in progress task", status: InProgress, want: "In Progress"},
		{name: "done task", status: Done, want: "Done"},
	}

	for _, tt := range tests {
		// t.Run creates a "subtest". Each one can pass/fail independently.
		// Run with: go test -v -run TestTaskDescription/todo_task
		t.Run(tt.name, func(t *testing.T) {
			task := Task{Status: tt.status}
			got := task.Description()
			if got != tt.want {
				t.Errorf("Description() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTaskFilterValue(t *testing.T) {
	task := Task{TaskTitle: "Deploy app"}

	got := task.FilterValue()
	want := "Deploy app"

	if got != want {
		t.Errorf("FilterValue() = %q, want %q", got, want)
	}
}

// ── Testing state mutations ─────────────────────────────────────
//
// Next() modifies the task. We test the full cycle.

func TestTaskNext(t *testing.T) {
	task := Task{Status: Todo}

	// Todo -> InProgress
	task.Next()
	if task.Status != InProgress {
		t.Errorf("after first Next(): Status = %d, want %d (InProgress)", task.Status, InProgress)
	}

	// InProgress -> Done
	task.Next()
	if task.Status != Done {
		t.Errorf("after second Next(): Status = %d, want %d (Done)", task.Status, Done)
	}

	// Done -> Todo (wraps around)
	task.Next()
	if task.Status != Todo {
		t.Errorf("after third Next(): Status = %d, want %d (Todo)", task.Status, Todo)
	}
}

// ── Table-driven version of the same test ───────────────────────
//
// This is cleaner when you have many transitions to verify.

func TestTaskNextTableDriven(t *testing.T) {
	tests := []struct {
		name  string
		start Status
		want  Status
	}{
		{"todo to in-progress", Todo, InProgress},
		{"in-progress to done", InProgress, Done},
		{"done wraps to todo", Done, Todo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := Task{Status: tt.start}
			task.Next()
			if task.Status != tt.want {
				t.Errorf("Next() from %d = %d, want %d", tt.start, task.Status, tt.want)
			}
		})
	}
}
