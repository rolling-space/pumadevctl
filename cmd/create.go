package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/rolling-space/pumadevctl/internal"
	"github.com/spf13/cobra"
)

var createLinkTarget string
var createAuto bool

var createCmd = &cobra.Command{
	Use:   "create <domain> [mapping]",
	Short: "Create a new entry (mapping file or symlink)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := internal.ResolveDir(dirFlag)
		if err != nil {
			return err
		}
		domain := args[0]
		// If --link is set, create symlink and ignore mapping args
		if createLinkTarget != "" {
			if err := internal.CreateSymlink(dir, domain, createLinkTarget, forceFlag); err != nil {
				return err
			}
			if !quietFlag && !jsonFlag {
				internal.NewFormatter(cmd.OutOrStdout()).Success("created symlink: %s → %s", domain, createLinkTarget)
			}
			if jsonFlag {
				out := map[string]string{"domain": domain, "link_target": createLinkTarget, "type": "symlink"}
				b, _ := json.MarshalIndent(out, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(b))
			}
			return nil
		}
		mapping := ""
		if len(args) == 2 {
			mapping = args[1]
			if _, err := internal.ParseMapping(mapping); err != nil {
				return err
			}
		} else {
			// auto port if not provided or --auto: allocate first available block within configured range
			entries, err := internal.LoadEntries(dir)
			if err != nil {
				return err
			}
			p, err := internal.FindNextAvailablePortBlock(entries, portMinFlag, portMaxFlag, portBlockSize)
			if err != nil {
				return err
			}
			mapping = strconv.Itoa(p)
		}
		if err := internal.WriteEntry(dir, domain, mapping, forceFlag); err != nil {
			return err
		}
		if !quietFlag && !jsonFlag {
			internal.NewFormatter(cmd.OutOrStdout()).Success("created: %s → %s", domain, mapping)
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
	createCmd.Flags().StringVar(&createLinkTarget, "link", "", "create a symlink entry pointing to this path instead of a port mapping")
	createCmd.Flags().BoolVar(&createAuto, "auto", true, "auto-pick a free port when mapping is omitted")
	rootCmd.AddCommand(createCmd)
}
