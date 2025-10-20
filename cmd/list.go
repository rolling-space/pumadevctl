package cmd

import (
	"github.com/rolling-space/pumadevctl/internal"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List mappings and group duplicates",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := internal.ResolveDir(dirFlag)
		if err != nil {
			return err
		}
		entries, err := internal.LoadEntries(dir)
		if err != nil {
			return err
		}
		if jsonFlag {
			return internal.PrintListJSON(entries)
		}
		internal.PrintListFancy(entries)
		// Also print simple JSON when quiet requested? quiet suppresses extras, so nothing more.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
