package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print lncli version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("lncli v%s\n", Version)
	},
}
