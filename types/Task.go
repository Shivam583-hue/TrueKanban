package types

type Task struct {
	Id        int
	TaskTitle string // rename the field
	Status    int
}

// we'll use 0= toto, 1= inProgress, 2= done

func (t Task) Title() string       { return t.TaskTitle }
func (t Task) Description() int    { return t.Status }
func (t Task) FilterValue() string { return t.TaskTitle }

func (t *Task) Next() {
	if t.Status == 2 {
		t.Status = 0
	} else {
		t.Status++
	}
}
