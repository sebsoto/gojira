package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"

	"github.com/sebsoto/gojira/pkg/release"
)

var (
	version      string
	date         string
	majorRelease bool
	// newCmd represents the new command
	newCmd = &cobra.Command{
		Use:   "new",
		Short: "Adds a new release to JIRA",
		Long:  `Creates a release epic and other related JIRA issues required for a tracking a release`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("new called")
			parsedDate, err := time.Parse(time.DateOnly, date)
			if err != nil {
				fmt.Fprintf(os.Stderr, "given date has the wrong format")
				os.Exit(1)
			}
			version = strings.TrimPrefix(version, "v")
			if !semver.IsValid("v" + version) {
				fmt.Fprintf(os.Stderr, "version is not a valid semver")
				os.Exit(1)
			}
			if err = release.CreateIssues(project, version, majorRelease, parsedDate); err != nil {
				fmt.Fprintf(os.Stderr, "%s", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	releaseCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(&version, "version", "", "Semver of the release")
	newCmd.MarkFlagRequired("version")
	newCmd.Flags().StringVar(&date, "date", "", "Planned date of the release")
	newCmd.MarkFlagRequired("date")
	newCmd.Flags().BoolVar(&majorRelease, "major", false, "Indicate this is a major release")
	newCmd.MarkFlagRequired("major")
}
