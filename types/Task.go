package types

type Task struct {
	Id        int
	TaskTitle string
	Status    Status
}

func (t Task) Title() string { return t.TaskTitle }
func (t Task) Description() string {
	statuses := []string{"Todo", "In Progress", "Done"}
	return statuses[t.Status]
}
func (t Task) FilterValue() string { return t.TaskTitle }

func (t *Task) Next() {
	if t.Status == 2 {
		t.Status = 0
	} else {
		t.Status++
	}
}
