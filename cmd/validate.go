package cmd

import (
    "encoding/json"
    "fmt"

    "github.com/spf13/cobra"
    "github.com/yourusername/pumadevctl/internal"
)

var timeoutMs int

var validateCmd = &cobra.Command{
    Use:   "validate",
    Short: "Validate reachability of mappings (TCP dial)",
    RunE: func(cmd *cobra.Command, args []string) error {
        dir, err := internal.ResolveDir(dirFlag)
        if err != nil {
            return err
        }
        entries, err := internal.LoadEntries(dir)
        if err != nil {
            return err
        }
        results := internal.ValidateEntries(entries, timeoutMs)
        if jsonFlag {
            enc := json.NewEncoder(cmd.OutOrStdout())
            enc.SetIndent("", "  ")
            return enc.Encode(results)
        }
        // pretty print
        ok := 0
        bad := 0
        for _, r := range results {
            if r.IsSymlink {
                fmt.Fprintf(cmd.OutOrStdout(), "%s (symlink) → %s
", r.Domain, r.LinkTarget)
                continue
            }
            if r.Reachable {
                fmt.Fprintf(cmd.OutOrStdout(), "✔ %s → %s
", r.Domain, r.Mapping)
                ok++
            } else {
                fmt.Fprintf(cmd.OutOrStdout(), "✖ %s → %s  (%s)
", r.Domain, r.Mapping, r.Reason)
                bad++
            }
        }
        if !quietFlag {
            fmt.Fprintf(cmd.OutOrStdout(), "
%d reachable, %d unreachable
", ok, bad)
        }
        return nil
    },
}

func init() {
    validateCmd.Flags().IntVar(&timeoutMs, "timeout", 500, "TCP dial timeout in milliseconds")
    rootCmd.AddCommand(validateCmd)
}
