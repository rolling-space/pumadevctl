package cmd

import (
	"encoding/json"

	"github.com/rolling-space/pumadevctl/internal"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <domain>",
	Short: "Read a single mapping or symlink",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := internal.ResolveDir(dirFlag)
		if err != nil {
			return err
		}
		e, err := internal.ReadEntry(dir, args[0])
		if err != nil {
			return err
		}
		if jsonFlag {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(e)
		}
		f := internal.NewFormatter(cmd.OutOrStdout())
		if e.IsSymlink {
			f.Info("%s → %s (symlink)", e.Domain, e.LinkTarget)
			return nil
		}
		f.Info("%s → %s", e.Domain, e.Mapping)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
