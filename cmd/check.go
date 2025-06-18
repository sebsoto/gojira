package cmd

import (
	"fmt"
	"os"

	"github.com/sebsoto/gojira/pkg/konflux"
	"github.com/spf13/cobra"
)

var namespace string

var (
	// checkCmd represents the new command
	checkCmd = &cobra.Command{
		Use:   "check",
		Short: "deprecated: temporary command to check what will be in a release",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			projects := []string{project, "OCPBUGS"}
			release, err := konflux.NewRelease(namespace, releaseplan, version, projects, "")
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			release.PrintContents()
			fmt.Printf("\n%s\n", release.ReleaseYAML())
		},
	}
)

func init() {
	releaseCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVar(&releaseplan, "releaseplan", "", "Konflux releaseplan")
	checkCmd.MarkFlagRequired("releaseplan")
	checkCmd.Flags().StringVar(&version, "version", "", "Semver of the release")
	checkCmd.MarkFlagRequired("version")
	checkCmd.Flags().StringVar(&namespace, "namespace", "", "Konflux namespace")
	checkCmd.MarkFlagRequired("namespace")
}
