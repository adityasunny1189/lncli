package project

import "time"

// ProjectJSON is the schema for .project.json committed to the user's repo.
// The website reads this file from GitHub to track progress.
type ProjectJSON struct {
	ProjectID       string            `json:"project_id"`
	ProjectVersion  string            `json:"project_version"`
	CLIVersion      string            `json:"cli_version"`
	CurrentTask     string            `json:"current_task"`
	CompletedTasks  []string          `json:"completed_tasks"`
	Timestamps      map[string]string `json:"timestamps"`
}

// TaskResult holds the outcome of running tests for one task.
type TaskResult struct {
	TaskID  string
	Passed  bool
	Output  string
	Elapsed time.Duration
}
