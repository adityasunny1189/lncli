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
	"github.com/spf13/cobra"
)

var projectFlag string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a project workspace locally",
	Long: `lncli init --project <id>

Fetches the starter code and tests for the given project from the LearnHub
GitHub repository and creates a local workspace directory ready to work in.

Example:
  lncli init --project basic-programming`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectFlag == "" {
			return fmt.Errorf("--project is required (e.g. lncli init --project basic-programming)")
		}
		return runInit(projectFlag)
	},
}

func init() {
	initCmd.Flags().StringVar(&projectFlag, "project", "", "Project ID to initialize (required)")
}

func runInit(projectID string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	destDir := filepath.Join(cwd, projectID)

	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("directory %q already exists — delete it first or pick a different location", destDir)
	}

	fmt.Printf("\n  Initializing project: %s\n", projectID)
	fmt.Printf("  Fetching starter files from github.com/%s/%s...\n\n",
		github.RepoOwner, github.RepoName)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Download starter files from the public GitHub repo
	if err := github.FetchStarter(projectID, destDir); err != nil {
		_ = os.RemoveAll(destDir)
		return fmt.Errorf("could not fetch starter files: %w\n\nMake sure you are connected to the internet and the project ID is correct.", err)
	}

	// Write .project.json
	pj := project.ProjectJSON{
		ProjectID:      projectID,
		ProjectVersion: "1.0.0",
		CLIVersion:     Version,
		CurrentTask:    "",
		CompletedTasks: []string{},
		Timestamps:     map[string]string{"initialized": time.Now().UTC().Format(time.RFC3339)},
	}
	pjData, err := json.MarshalIndent(pj, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destDir, ".project.json"), pjData, 0644); err != nil {
		return err
	}

	// Write .lncli.json
	cfg := &config.WorkspaceConfig{ProjectID: projectID}
	if err := config.Save(destDir, cfg); err != nil {
		return err
	}

	// Write .gitignore
	_ = os.WriteFile(filepath.Join(destDir, ".gitignore"), []byte("# lncli workspace\n.DS_Store\n"), 0644)

	fmt.Printf("  ✓  Created %s/\n", projectID)
	fmt.Printf("  ✓  .project.json initialized\n")
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Printf("    cd %s\n", projectID)
	fmt.Println("    git init && git remote add origin https://github.com/<you>/" + projectID)
	fmt.Println("    # edit task files in your editor")
	fmt.Println("    lncli run")
	fmt.Println()

	return nil
}
