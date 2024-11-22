package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var version = "1.0.0" // Default to "dev" for local builds.

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of git-auth",
	Long:  `All software has versions. This is the version of our git-auth.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("git-auth version: %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
