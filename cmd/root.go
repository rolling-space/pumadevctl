package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	dirFlag   string
	forceFlag bool
	jsonFlag  bool
	quietFlag bool
	version   = "0.2.0"
)

var rootCmd = &cobra.Command{
	Use:     "pumadevctl",
	Short:   "Manage puma-dev mappings (~/.puma-dev) with CRUD, list, validate, cleanup",
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	home, _ := os.UserHomeDir()
	defaultDir := filepath.Join(home, ".puma-dev")

	rootCmd.PersistentFlags().StringVarP(&dirFlag, "dir", "d", defaultDir, "directory for puma-dev entries")
	rootCmd.PersistentFlags().BoolVarP(&forceFlag, "force", "f", false, "force operation without interactive confirmations")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output JSON when supported")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress non-essential output")
}
