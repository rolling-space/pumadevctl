package cmd

import (
	"fmt"

	"github.com/rolling-space/pumadevctl/internal"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show detailed version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(internal.VersionLong())
		},
	}
	rootCmd.AddCommand(cmd)
}
