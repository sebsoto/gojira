package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/sebsoto/gojira/pkg/konflux"
	"github.com/sebsoto/gojira/pkg/release"
)

var issue string
var tailCommit string

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "updates pending releases",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		projects := []string{project, "OCPBUGS"}
		kRelease, err := konflux.NewRelease(namespace, releaseplan, version, projects, tailCommit)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		err = release.UpdateRelease(issue, project, version, true, time.Now(), kRelease.Release)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	releaseCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVar(&issue, "issue", "", "ticket to update")
	updateCmd.MarkFlagRequired("issue")
	updateCmd.Flags().StringVar(&releaseplan, "releaseplan", "", "Konflux releaseplan")
	updateCmd.MarkFlagRequired("releaseplan")
	updateCmd.Flags().StringVar(&version, "version", "", "Semver of the release")
	updateCmd.MarkFlagRequired("version")
	updateCmd.Flags().StringVar(&tailCommit, "tail", "", "tail commit of the release")
	updateCmd.Flags().StringVar(&namespace, "namespace", "", "Konflux namespace")
	updateCmd.MarkFlagRequired("namespace")
}
