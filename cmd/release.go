package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// releaseCmd represents the release command
	releaseCmd = &cobra.Command{
		Use:   "release",
		Short: "manage releases in JIRA",
		Long:  `Manage releases in JIRA`,
	}
)

func init() {
	rootCmd.AddCommand(releaseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// releaseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// releaseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
