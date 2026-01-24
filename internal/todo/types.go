package todo

// Session represents a work session
type Session struct {
	ID              string
	Type            string // quick, code, research
	Goal            string
	SuccessCriteria string
	Boundaries      string // JSON array
	Deliverables    string // JSON array
	Status          string // active, completed
	Metadata        string // JSON
	CreatedAt       string
	UpdatedAt       string
}

// Task represents a todo item
type Task struct {
	ID             int
	SessionID      string
	Type           string // task, research, coordination, analysis, deliverable
	Priority       string // p0, p1, p2, p3, p4
	PriorityOrder  int
	Status         string // pending, in_progress, completed, blocked
	Task           string
	ActiveForm     string
	Files          string // JSON array
	Refs           string // JSON array
	WaitingOn      string
	Output         string
	Audience       string
	Instructions   string // JSON
	Notes          string
	BlockingReason string
	DependantIDs   string // JSON array
	Effort         string // xs, s, m
	Metadata       string // JSON
	CreatedAt      string
	UpdatedAt      string
}

// NextOutput represents the structured output for 'todo next'
type NextOutput struct {
	Task           *Task
	Session        *Session
	TotalTasks     int
	CompletedTasks int
	UpcomingTasks  []*Task
	Instructions   []string
	MustNotDo      []string
	Files          []string
	Prerequisites  []string
	Boundaries     []string
}
