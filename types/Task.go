package types

type Task struct {
	Id        int
	TaskTitle string // rename the field
	Status    string
}

func (t Task) Title() string       { return t.TaskTitle }
func (t Task) Description() string { return t.Status }
func (t Task) FilterValue() string { return t.TaskTitle }
