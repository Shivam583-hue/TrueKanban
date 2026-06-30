package db

import (
	"testing"

	"github.com/Shivam583-hue/TrueKanban/types"
)

// ── Testing with setup/teardown ─────────────────────────────────
//
// The db package uses a package-level `db` variable. We need to
// call Init() before tests and Close() after. There are two ways:
//
//   1. TestMain — runs before/after ALL tests in this package
//   2. Helper functions — called at the start of each test
//
// We'll use TestMain here because every test needs the DB.

// TestMain is a special function. Go runs it instead of running
// tests directly. You control when tests run via m.Run().
func TestMain(m *testing.M) {
	// Setup: initialize DB (creates ./task.db in the test directory)
	Init()

	// Run all tests. m.Run() returns an exit code.
	m.Run()

	// Teardown: close the database connection.
	Close()

	// Tip: In production code, you'd also delete the test DB file.
	// We skip that here so you can inspect it after tests.
}

// ── Helper function ─────────────────────────────────────────────
//
// t.Helper() marks this as a helper — when it fails, Go reports
// the line in the CALLING test, not inside this function.
// This makes error messages much more useful.

func cleanTable(t *testing.T) {
	t.Helper()
	_, err := db.Exec("DELETE FROM tasks")
	if err != nil {
		t.Fatalf("failed to clean table: %v", err)
	}
}

// ── Basic CRUD tests ────────────────────────────────────────────

func TestInsertAndFetch(t *testing.T) {
	cleanTable(t) // start with a clean slate

	// Insert a task
	Insert("Write tests", types.Todo)

	// Fetch it back
	items := Fetch(types.Todo)

	if len(items) != 1 {
		t.Fatalf("Fetch(Todo) returned %d items, want 1", len(items))
		// t.Fatalf stops here. Use it when continuing makes no sense
		// (we'd panic on items[0] if the slice is empty).
	}

	// Type-assert the list.Item back to a Task
	task, ok := items[0].(types.Task)
	if !ok {
		t.Fatal("Fetch returned an item that isn't a types.Task")
	}

	if task.TaskTitle != "Write tests" {
		t.Errorf("task.TaskTitle = %q, want %q", task.TaskTitle, "Write tests")
	}
	if task.Status != types.Todo {
		t.Errorf("task.Status = %d, want %d", task.Status, types.Todo)
	}
}

func TestFetchReturnsOnlyMatchingStatus(t *testing.T) {
	cleanTable(t)

	Insert("Task A", types.Todo)
	Insert("Task B", types.InProgress)
	Insert("Task C", types.Done)

	todoItems := Fetch(types.Todo)
	if len(todoItems) != 1 {
		t.Errorf("Fetch(Todo) returned %d items, want 1", len(todoItems))
	}

	inProgressItems := Fetch(types.InProgress)
	if len(inProgressItems) != 1 {
		t.Errorf("Fetch(InProgress) returned %d items, want 1", len(inProgressItems))
	}

	doneItems := Fetch(types.Done)
	if len(doneItems) != 1 {
		t.Errorf("Fetch(Done) returned %d items, want 1", len(doneItems))
	}
}

func TestUpdate(t *testing.T) {
	cleanTable(t)

	Insert("Move me", types.Todo)

	// Get the inserted task's ID
	items := Fetch(types.Todo)
	if len(items) == 0 {
		t.Fatal("no items found after insert")
	}
	task := items[0].(types.Task)

	// Update status from Todo -> InProgress
	Update(task.Id, types.InProgress)

	// Should no longer appear in Todo
	todoItems := Fetch(types.Todo)
	if len(todoItems) != 0 {
		t.Errorf("Fetch(Todo) after update returned %d items, want 0", len(todoItems))
	}

	// Should now appear in InProgress
	inProgressItems := Fetch(types.InProgress)
	if len(inProgressItems) != 1 {
		t.Errorf("Fetch(InProgress) after update returned %d items, want 1", len(inProgressItems))
	}
}

func TestDelete(t *testing.T) {
	cleanTable(t)

	Insert("Delete me", types.Todo)

	items := Fetch(types.Todo)
	if len(items) == 0 {
		t.Fatal("no items found after insert")
	}
	task := items[0].(types.Task)

	// Delete the task
	Delete(task.Id)

	// Should be gone
	remaining := Fetch(types.Todo)
	if len(remaining) != 0 {
		t.Errorf("Fetch(Todo) after delete returned %d items, want 0", len(remaining))
	}
}

func TestDeleteNonExistentTaskDoesNotPanic(t *testing.T) {
	cleanTable(t)

	// Deleting an ID that doesn't exist should not panic or error.
	// This test passes if it simply doesn't crash.
	Delete(99999)
}

func TestFetchEmptyTable(t *testing.T) {
	cleanTable(t)

	items := Fetch(types.Todo)

	// Go returns nil for an empty slice from append, not []list.Item{}.
	// Both are fine — we just check the length.
	if len(items) != 0 {
		t.Errorf("Fetch on empty table returned %d items, want 0", len(items))
	}
}
