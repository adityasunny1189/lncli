package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/adityasunny1189/lncli/internal/project"
)

// RunTask executes `go test ./...` inside taskDir and returns the result.
func RunTask(taskDir string) project.TaskResult {
	taskID := filepath.Base(taskDir)
	start := time.Now()

	// Download any missing dependencies silently before running tests.
	// This is a no-op when all deps are already cached (e.g. stdlib-only tasks).
	dl := exec.Command("go", "mod", "download")
	dl.Dir = taskDir
	_ = dl.Run()

	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Dir = taskDir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	elapsed := time.Since(start)

	return project.TaskResult{
		TaskID:  taskID,
		Passed:  err == nil,
		Output:  out.String(),
		Elapsed: elapsed,
	}
}

// FindCurrentTask returns the task directory that should be run next.
// It looks for the first task directory whose tests have not yet passed
// (i.e. not listed in completedTasks).
func FindCurrentTask(projectDir string, completedTasks []string) (string, error) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return "", err
	}

	completed := make(map[string]bool, len(completedTasks))
	for _, t := range completedTasks {
		completed[t] = true
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		// Task directories follow the pattern NN-slug (e.g. task-01-even-odd)
		if !strings.HasPrefix(name, "task-") {
			continue
		}
		if !completed[name] {
			return filepath.Join(projectDir, name), nil
		}
	}
	return "", nil
}

// PrintResult writes a human-friendly summary to stdout.
func PrintResult(r project.TaskResult) {
	if r.Passed {
		fmt.Printf("\n  ✓  %s — PASSED (%s)\n\n", r.TaskID, r.Elapsed.Round(time.Millisecond))
	} else {
		fmt.Printf("\n  ✗  %s — FAILED (%s)\n\n", r.TaskID, r.Elapsed.Round(time.Millisecond))
		fmt.Println(r.Output)
	}
}
