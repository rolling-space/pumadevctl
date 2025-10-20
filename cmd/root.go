package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/pumadevctl/internal"
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

	// Load config from XDG and use as defaults unless flags were provided.
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := internal.LoadAppConfig()
		if err != nil {
			return err
		}
		// If the flag was not set by the user, apply config values.
		if f := cmd.Flags().Lookup("dir"); f != nil && !f.Changed {
			dirFlag = cfg.Dir
		}
		if f := cmd.Flags().Lookup("port-min"); f != nil && !f.Changed && cfg.PortMin != 0 {
			portMinFlag = cfg.PortMin
		}
		if f := cmd.Flags().Lookup("port-max"); f != nil && !f.Changed && cfg.PortMax != 0 {
			portMaxFlag = cfg.PortMax
		}
		if f := cmd.Flags().Lookup("port-block-size"); f != nil && !f.Changed && cfg.PortBlockSize != 0 {
			portBlockSize = cfg.PortBlockSize
		}
		_ = runtime.GOOS // keep import used in case future OS-specific defaults are needed
		_ = time.Second  // keep import used for potential timeouts in future flags
		return nil
	}
}
