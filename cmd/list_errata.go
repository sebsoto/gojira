package cmd

import (
	"fmt"
	"github.com/sebsoto/gojira/pkg/errata"
	"os"

	"github.com/spf13/cobra"
)

// listErrataCmd represents the listErrata command
var listErrataCmd = &cobra.Command{
	Use:   "list-errata",
	Short: "Lists pending errata",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		errataList, err := errata.List()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Printf("ID\tNAME\tURL\n")
		for _, e := range errataList {
			fmt.Printf("%d\t%s\t%s\n", e.ID, e.Synopsis, errata.URL(e.ID))
		}
	},
}

func init() {
	releaseCmd.AddCommand(listErrataCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listErrataCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listErrataCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
