package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adityasunny1189/lncli/internal/config"
	"github.com/adityasunny1189/lncli/internal/github"
	"github.com/adityasunny1189/lncli/internal/project"
	"github.com/adityasunny1189/lncli/internal/runner"
	"github.com/spf13/cobra"
)

var skipPushFlag bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tests for the current task and push progress to GitHub",
	Long: `lncli run

Must be run from inside an initialized project directory (one that contains
.lncli.json). It:

  1. Detects the current task from .project.json
  2. Runs go test ./... inside that task's directory
  3. On success: updates .project.json and pushes to GitHub
  4. On failure: shows test output so you can fix the issues

Example:
  cd basic-programming
  lncli run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRun()
	},
}

func init() {
	runCmd.Flags().BoolVar(&skipPushFlag, "no-push", false, "Run tests but skip the git push")
}

func runRun() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Load workspace config
	cfg, err := config.Load(cwd)
	if err != nil {
		return fmt.Errorf("could not read .lncli.json: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("no .lncli.json found — run `lncli init --project <id>` first")
	}

	// Load .project.json
	pjPath := filepath.Join(cwd, ".project.json")
	pjData, err := os.ReadFile(pjPath)
	if err != nil {
		return fmt.Errorf("could not read .project.json: %w", err)
	}
	var pj project.ProjectJSON
	if err := json.Unmarshal(pjData, &pj); err != nil {
		return fmt.Errorf("could not parse .project.json: %w", err)
	}

	// Find next task to run
	taskDir, err := runner.FindCurrentTask(cwd, pj.CompletedTasks)
	if err != nil {
		return err
	}
	if taskDir == "" {
		fmt.Println("\n  All tasks complete! Visit the website to see your results.\n")
		return nil
	}

	taskID := filepath.Base(taskDir)
	fmt.Printf("\n  Running tests for: %s\n", taskID)
	fmt.Println("  ─────────────────────────────────────")

	result := runner.RunTask(taskDir)
	runner.PrintResult(result)

	if !result.Passed {
		fmt.Println("  Fix the failing tests and run `lncli run` again.\n")
		return nil
	}

	// Update .project.json
	pj.CompletedTasks = append(pj.CompletedTasks, taskID)
	pj.Timestamps[taskID] = time.Now().UTC().Format(time.RFC3339)

	// Find next task
	nextDir, _ := runner.FindCurrentTask(cwd, pj.CompletedTasks)
	if nextDir != "" {
		pj.CurrentTask = filepath.Base(nextDir)
	} else {
		pj.CurrentTask = ""
	}

	updatedData, err := json.MarshalIndent(pj, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(pjPath, updatedData, 0644); err != nil {
		return fmt.Errorf("could not write .project.json: %w", err)
	}

	fmt.Printf("  .project.json updated — task %q marked complete\n", taskID)

	if skipPushFlag {
		fmt.Println("  (skipping git push — --no-push flag set)\n")
		return nil
	}

	// Push to GitHub
	fmt.Println("  Pushing to GitHub...")
	if err := github.PushProgress(cwd, taskID, cfg.ProjectID); err != nil {
		fmt.Printf("\n  Warning: could not push to GitHub: %v\n", err)
		fmt.Println("  Your local progress is saved. Run `git push` manually when ready.\n")
		return nil
	}

	fmt.Println("  ✓  Progress pushed to GitHub")

	if pj.CurrentTask != "" {
		fmt.Printf("\n  Next task: %s\n", pj.CurrentTask)
		fmt.Println("  Go back to the website and click \"Check Progress\" to unlock it.")
	} else {
		fmt.Println("\n  Project complete! Go to the website to see your badge.")
	}
	fmt.Println()

	return nil
}
