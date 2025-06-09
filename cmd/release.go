package cmd

import (
	"github.com/spf13/cobra"
)

var (
	project     string
	releaseplan string
	// releaseCmd represents the release command
	releaseCmd = &cobra.Command{
		Use:   "release",
		Short: "manage releases in JIRA",
		Long:  `Manage releases in JIRA`,
	}
)

func init() {
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.PersistentFlags().StringVar(&project, "project", "", "JIRA project")
	releaseCmd.MarkPersistentFlagRequired("project")
}
