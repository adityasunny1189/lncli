package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

Fetches starter code from GitHub, creates the local workspace, creates the
GitHub repo, and wires up git — so you can go straight to writing code.

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

	// ── Step 1: GitHub username ──────────────────────────────────────────
	username := detectGitHubUsername()
	if username == "" {
		fmt.Print("  Enter your GitHub username: ")
		reader := bufio.NewReader(os.Stdin)
		username, _ = reader.ReadString('\n')
		username = strings.TrimSpace(username)
	} else {
		fmt.Printf("  GitHub user: %s\n", username)
	}
	if username == "" {
		return fmt.Errorf("GitHub username is required")
	}

	fmt.Printf("\n  Initializing project: %s\n", projectID)
	fmt.Printf("  Fetching starter files from github.com/%s/%s...\n\n",
		github.RepoOwner, github.RepoName)

	// ── Step 2: Create local directory + fetch starter files ─────────────
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	if err := github.FetchStarter(projectID, destDir); err != nil {
		_ = os.RemoveAll(destDir)
		return fmt.Errorf("could not fetch starter files: %w\n\nMake sure you are connected to the internet and the project ID is correct.", err)
	}
	fmt.Printf("  ✓  Starter files downloaded\n")

	// ── Step 3: Write .project.json ──────────────────────────────────────
	pj := project.ProjectJSON{
		ProjectID:      projectID,
		ProjectVersion: "1.0.0",
		CLIVersion:     Version,
		CurrentTask:    "",
		CompletedTasks: []string{},
		Timestamps:     map[string]string{"initialized": time.Now().UTC().Format(time.RFC3339)},
	}
	pjData, _ := json.MarshalIndent(pj, "", "  ")
	if err := os.WriteFile(filepath.Join(destDir, ".project.json"), pjData, 0644); err != nil {
		return err
	}

	// ── Step 4: Write .lncli.json ─────────────────────────────────────────
	cfg := &config.WorkspaceConfig{ProjectID: projectID, GitHubUsername: username}
	if err := config.Save(destDir, cfg); err != nil {
		return err
	}

	// ── Step 5: Write .gitignore ──────────────────────────────────────────
	_ = os.WriteFile(filepath.Join(destDir, ".gitignore"), []byte(".DS_Store\n"), 0644)

	// ── Step 6: git init ──────────────────────────────────────────────────
	if err := runGit(destDir, "init", "-b", "main"); err != nil {
		// older git versions don't support -b
		if err2 := runGit(destDir, "init"); err2 != nil {
			return fmt.Errorf("git init failed: %w", err2)
		}
	}
	fmt.Printf("  ✓  git init\n")

	// ── Step 7: Create GitHub repo ────────────────────────────────────────
	repoURL := fmt.Sprintf("https://github.com/%s/%s", username, projectID)
	if ghAvailable() {
		fmt.Printf("  Creating github.com/%s/%s...\n", username, projectID)
		out, err := exec.Command("gh", "repo", "create", projectID,
			"--public",
			"--description", fmt.Sprintf("LearnHub project: %s", projectID),
		).CombinedOutput()
		if err != nil {
			// repo might already exist
			if !strings.Contains(string(out), "already exists") {
				fmt.Printf("  Warning: could not create repo automatically: %s\n", strings.TrimSpace(string(out)))
				fmt.Printf("  Create it manually at https://github.com/new (name: %s, public)\n", projectID)
			}
		} else {
			fmt.Printf("  ✓  Created github.com/%s/%s\n", username, projectID)
		}
	} else {
		fmt.Printf("\n  ⚠  GitHub CLI not found. Create the repo manually:\n")
		fmt.Printf("     https://github.com/new  →  name: %s  →  Public\n\n", projectID)
		fmt.Println("  Press Enter once the repo is created...")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}

	// ── Step 8: Set remote ────────────────────────────────────────────────
	_ = runGit(destDir, "remote", "remove", "origin") // remove if it exists
	if err := runGit(destDir, "remote", "add", "origin", repoURL); err != nil {
		return fmt.Errorf("could not set git remote: %w", err)
	}
	fmt.Printf("  ✓  Remote → %s\n", repoURL)

	// ── Step 9: Initial commit + push ─────────────────────────────────────
	_ = runGit(destDir, "config", "user.email", username+"@users.noreply.github.com")
	if err := runGit(destDir, "add", "."); err != nil {
		return err
	}
	if err := runGit(destDir, "commit", "-m", "lncli: initial project setup"); err != nil {
		return err
	}
	if err := runGit(destDir, "push", "-u", "origin", "main"); err != nil {
		fmt.Printf("  Warning: initial push failed — run `git push -u origin main` manually\n")
	} else {
		fmt.Printf("  ✓  Pushed to GitHub\n")
	}

	// ── Done ──────────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Printf("  ✓  %s is ready\n\n", projectID)
	fmt.Printf("  cd %s\n", projectID)
	fmt.Println("  # open any task file in your editor and write your solution")
	fmt.Println("  lncli run")
	fmt.Println()

	return nil
}

// detectGitHubUsername tries to get the username from the gh CLI.
func detectGitHubUsername() string {
	out, err := exec.Command("gh", "api", "user", "--jq", ".login").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ghAvailable returns true if the gh CLI is installed and authenticated.
func ghAvailable() bool {
	err := exec.Command("gh", "auth", "status").Run()
	return err == nil
}

// runGit runs a git command in dir and returns any error.
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return nil
}
