package cmd

import (
	"fmt"
	"github.com/sebsoto/gojira/pkg/konflux"
	"github.com/sebsoto/gojira/pkg/release"
	"os"
	"time"

	"github.com/spf13/cobra"
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
		kRelease, err := konflux.NewRelease(releaseplan, version, projects, tailCommit)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		err = release.UpdateRelease(issue, project, version, true, time.Now(), kRelease)
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
	updateCmd.MarkFlagRequired("tail")
}
