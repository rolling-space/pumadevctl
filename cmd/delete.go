package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/pumadevctl/internal"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <domain>",
	Short: "Delete an entry (file or symlink)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := internal.ResolveDir(dirFlag)
		if err != nil {
			return err
		}
		domain := args[0]
		if !forceFlag {
			fmt.Fprintf(cmd.OutOrStdout(), "Delete %s? [y/N]: ", domain)
			rdr := bufio.NewReader(os.Stdin)
			line, _ := rdr.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(line)) != "y" {
				internal.NewFormatter(cmd.OutOrStdout()).Warn("aborted")
				return nil
			}
		}
		if err := internal.DeleteEntry(dir, domain); err != nil {
			return err
		}
		if !quietFlag && !jsonFlag {
			internal.NewFormatter(cmd.OutOrStdout()).Success("deleted: %s", domain)
		}
		if jsonFlag {
			out := map[string]string{"domain": domain, "status": "deleted"}
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
		}
		return nil
	},
}

func init() { rootCmd.AddCommand(deleteCmd) }
