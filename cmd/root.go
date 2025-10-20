package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	dirFlag       string
	forceFlag     bool
	jsonFlag      bool
	quietFlag     bool
	portMinFlag   int
	portMaxFlag   int
	portBlockSize int
	version       = "0.3.0"
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

	// Port allocation controls
	rootCmd.PersistentFlags().IntVar(&portMinFlag, "port-min", 36000, "minimum port for auto allocation (inclusive)")
	rootCmd.PersistentFlags().IntVar(&portMaxFlag, "port-max", 37000, "maximum port for auto allocation (inclusive)")
	rootCmd.PersistentFlags().IntVar(&portBlockSize, "port-block-size", 10, "number of consecutive ports reserved per domain")
}
