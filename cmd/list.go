package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sebsoto/gojira/pkg/jira"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "lists pending releases",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		issues, err := jira.Search(fmt.Sprintf("project = %s AND issuetype = Epic AND labels in (OperatorProductization) AND statusCategory != \"Done\"", project))
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		for _, issue := range issues {
			links, err := jira.GetRemoteLinks(issue.Key)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			fmt.Printf("%s: %s | %v \n", issue.Key, issue.Fields.EpicName, links)
		}
	},
}

func init() {
	releaseCmd.AddCommand(listCmd)
}
