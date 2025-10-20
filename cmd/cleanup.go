package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rolling-space/pumadevctl/internal"
	"github.com/spf13/cobra"
)

var cleanupYes bool
var cleanupDry bool

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove unreachable mappings (non-symlink)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := internal.ResolveDir(dirFlag)
		if err != nil {
			return err
		}
		entries, err := internal.LoadEntries(dir)
		if err != nil {
			return err
		}
		results := internal.ValidateEntries(entries, 300)
		toDelete := []internal.Entry{}
		for _, r := range results {
			if !r.IsSymlink && !r.Reachable {
				toDelete = append(toDelete, r.Entry)
			}
		}
		if jsonFlag {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(toDelete)
		}
		f := internal.NewFormatter(cmd.OutOrStdout())
		if len(toDelete) == 0 {
			if !quietFlag {
				f.Info("nothing to delete")
			}
			return nil
		}
		f.Header("Unreachable entries")
		for _, e := range toDelete {
			f.Bullet(fmt.Sprintf("%s → %s", e.Domain, e.Mapping))
		}
		if cleanupDry {
			f.Warn("--dry-run set; no deletions performed.")
			return nil
		}
		if !cleanupYes && !forceFlag {
			fmt.Fprint(cmd.OutOrStdout(), "Delete these? [y/N]: ")
			rdr := bufio.NewReader(os.Stdin)
			line, _ := rdr.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(line)) != "y" {
				f.Warn("aborted")
				return nil
			}
		}
		// delete
		for _, e := range toDelete {
			if err := internal.DeleteEntry(dir, e.Domain); err != nil {
				internal.NewFormatter(cmd.OutOrStdout()).Error("failed to delete %s: %v", e.Domain, err)
			} else if !quietFlag {
				f.Success("deleted: %s", e.Domain)
			}
		}
		return nil
	},
}

func init() {
	cleanupCmd.Flags().BoolVar(&cleanupYes, "yes", false, "assume yes; do not prompt")
	cleanupCmd.Flags().BoolVar(&cleanupDry, "dry-run", false, "show what would be deleted without doing it")
	rootCmd.AddCommand(cleanupCmd)
}
