package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repo is the GitHub repo that hosts all LearnHub project content.
const (
	RepoOwner = "adityasunny1189"
	RepoName  = "learning-notes"
	Branch    = "master"
)

// contentsEntry is one item from the GitHub Contents API response.
type contentsEntry struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "dir"
	DownloadURL string `json:"download_url"`
}

// FetchStarter downloads the starter files for projectID from the public repo
// and writes them into destDir, preserving the subdirectory structure.
func FetchStarter(projectID, destDir string) error {
	starterPath := fmt.Sprintf("projects/%s/starter", projectID)
	return fetchDir(starterPath, destDir)
}

// fetchDir recursively fetches all files under apiPath into localDir.
func fetchDir(apiPath, localDir string) error {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
		RepoOwner, RepoName, apiPath, Branch,
	)

	entries, err := listContents(url)
	if err != nil {
		return fmt.Errorf("fetching %s: %w", apiPath, err)
	}

	for _, e := range entries {
		dest := filepath.Join(localDir, e.Name)
		switch e.Type {
		case "dir":
			if err := os.MkdirAll(dest, 0755); err != nil {
				return err
			}
			if err := fetchDir(e.Path, dest); err != nil {
				return err
			}
		case "file":
			if err := downloadFile(e.DownloadURL, dest); err != nil {
				return fmt.Errorf("downloading %s: %w", e.Path, err)
			}
		}
	}
	return nil
}

func listContents(url string) ([]contentsEntry, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("path not found on GitHub (404)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}
	var entries []contentsEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func downloadFile(rawURL, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	resp, err := http.Get(rawURL) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, rawURL)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0644)
}

// ValidateUser checks if a GitHub username exists via the public API.
func ValidateUser(username string) (bool, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/users/%s", username)) //nolint:gosec
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

// PushProgress stages .project.json, commits, and pushes.
func PushProgress(repoDir, taskID, projectID string) error {
	run := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, string(out))
		}
		return nil
	}

	if err := run("add", ".project.json"); err != nil {
		return err
	}

	msg := fmt.Sprintf("lncli: complete task %s [%s]", taskID, projectID)
	if err := run("commit", "-m", msg); err != nil {
		if strings.Contains(err.Error(), "nothing to commit") {
			return nil
		}
		return err
	}

	return run("push")
}
