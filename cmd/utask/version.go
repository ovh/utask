package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ovh/utask"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the µTask version and current commit",
	Long: "Display information about current µTask version and\n" +
		"picked commit.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("µTask overview")
		fmt.Printf("Version: %s\nCommit: %s\n", utask.Version, utask.Commit)
	},
}
