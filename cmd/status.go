package cmd

import (
	"fmt"
	"os"

	"github.com/sebsoto/gojira/pkg/konflux"
	"github.com/spf13/cobra"
)

var namespace string

var (
	// statusCmd represents the new command
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Current status of a potential release",
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
	releaseCmd.AddCommand(statusCmd)
	statusCmd.Flags().StringVar(&releaseplan, "releaseplan", "", "Konflux releaseplan")
	statusCmd.MarkFlagRequired("releaseplan")
	statusCmd.Flags().StringVar(&version, "version", "", "Semver of the release")
	statusCmd.MarkFlagRequired("version")
	statusCmd.Flags().StringVar(&namespace, "namespace", "", "Konflux namespace")
	statusCmd.MarkFlagRequired("namespace")
}
