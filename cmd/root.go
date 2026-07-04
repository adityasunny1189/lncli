package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lncli",
	Short: "LearnHub CLI — manage and verify hands-on projects",
	Long: `lncli is the official CLI for LearnHub projects.

It initializes project workspaces from templates, runs test suites
against your solutions, and commits progress to GitHub so the website
can track your advancement.

Get started:
  lncli init --project basic-programming
  lncli run`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)
}
