package templates

// Template represents a task template that can be applied to create multiple tasks
type Template struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Tasks       []TaskSpec `yaml:"tasks"`
}

// TaskSpec defines a task in a template
type TaskSpec struct {
	Title        string              `yaml:"title"`
	Priority     string              `yaml:"priority,omitempty"`
	Effort       string              `yaml:"effort,omitempty"`
	Type         string              `yaml:"type,omitempty"`
	Files        []string            `yaml:"files,omitempty"`
	DependsOn    []string            `yaml:"depends_on,omitempty"` // Logical IDs within template
	Instructions map[string][]string `yaml:"instructions,omitempty"`
}

// TemplateInfo is metadata about a template (for listing)
type TemplateInfo struct {
	Name        string
	Description string
	TaskCount   int
	Source      string // "embedded" or "user"
}
