package cmd

import (
	"fmt"
	"github.com/sebsoto/gojira/pkg/konflux"
	"github.com/spf13/cobra"
	"os"
)

var (
	v           string
	releaseplan string
	// checkCmd represents the new command
	checkCmd = &cobra.Command{
		Use:   "check",
		Short: "deprecated: temporary command to check what will be in a release",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			projects := []string{project, "OCPBUGS"}
			err := konflux.NewRelease(releaseplan, v, projects)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		},
	}
)

func init() {
	releaseCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVar(&v, "version", "", "Semver of the release, example v10.18.1")
	checkCmd.MarkFlagRequired("version")
	checkCmd.Flags().StringVar(&releaseplan, "releaseplan", "", "Konflux releaseplan")
	checkCmd.MarkFlagRequired("repoOwner")
}
