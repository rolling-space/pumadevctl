package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/pumadevctl/internal"
)

var updateLinkTarget string

var updateCmd = &cobra.Command{
	Use:   "update <domain> <mapping>",
	Short: "Update an existing entry (file content) or use --link to repoint a symlink",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := internal.ResolveDir(dirFlag)
		if err != nil {
			return err
		}
		domain := args[0]
		if updateLinkTarget != "" {
			if err := internal.UpdateSymlink(dir, domain, updateLinkTarget); err != nil {
				return err
			}
			if !quietFlag && !jsonFlag {
				internal.NewFormatter(cmd.OutOrStdout()).Success("updated symlink: %s → %s", domain, updateLinkTarget)
			}
			if jsonFlag {
				out := map[string]string{"domain": domain, "link_target": updateLinkTarget, "type": "symlink"}
				b, _ := json.MarshalIndent(out, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(b))
			}
			return nil
		}
		if len(args) < 2 {
			return fmt.Errorf("mapping required unless --link is set")
		}
		mapping := args[1]
		if _, err := internal.ParseMapping(mapping); err != nil {
			return err
		}
		if err := internal.UpdateEntry(dir, domain, mapping); err != nil {
			return err
		}
		if !quietFlag && !jsonFlag {
			internal.NewFormatter(cmd.OutOrStdout()).Success("updated: %s → %s", domain, mapping)
		}
		if jsonFlag {
			out := map[string]string{"domain": domain, "mapping": mapping, "type": "file"}
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateLinkTarget, "link", "", "repoint an existing symlink to this path")
	rootCmd.AddCommand(updateCmd)
}
