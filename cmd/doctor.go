package cmd

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check if puma-dev is installed and running; print config and runtime details",
	RunE: func(cmd *cobra.Command, args []string) error {
		installedPath, installed := findPumaDevBinary()
		if !installed {
			fmt.Fprintf(cmd.OutOrStdout(), "puma-dev is not installed or not found in PATH.\n\nInstall instructions: https://github.com/puma/puma-dev?tab=readme-ov-file#installation\n")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "puma-dev binary: %s\n", installedPath)
		if v := pumaDevVersion(installedPath); v != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "version: %s\n", v)
		}

		// Check if it's running by attempting TCP dials to common ports.
		running, port := isPumaDevRunning()
		if running {
			fmt.Fprintf(cmd.OutOrStdout(), "status: running (listening on 127.0.0.1:%d)\n", port)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "status: not reachable on typical ports (9280/80/3000).\n")
		}

		// Discover configuration files depending on OS and show their contents.
		paths := discoverConfigFiles()
		if len(paths) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "configuration files: not found in common locations")
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout(), "configuration files:")
		for _, p := range paths {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", p)
			if b, err := os.ReadFile(p); err == nil {
				trimmed := strings.TrimSpace(string(b))
				if len(trimmed) > 0 {
					fmt.Fprintln(cmd.OutOrStdout(), indent("--- begin file ---\n"+trimmed+"\n--- end file ---", 2))
				}
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  (unable to read: %v)\n", err)
			}
		}
		return nil
	},
}

func init() { rootCmd.AddCommand(doctorCmd) }

func findPumaDevBinary() (string, bool) {
	path, err := exec.LookPath("puma-dev")
	return path, err == nil
}

func pumaDevVersion(bin string) string {
	c := exec.Command(bin, "-V")
	var out bytes.Buffer
	c.Stdout = &out
	// best-effort; ignore errors
	_ = c.Run()
	return strings.TrimSpace(out.String())
}

// isPumaDevRunning tries ports commonly used by puma-dev.
func isPumaDevRunning() (bool, int) {
	ports := []int{9280, 80, 9292}
	for _, p := range ports {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", p), 250*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true, p
		}
	}
	return false, 0
}

// discoverConfigFiles returns likely config/service definitions for puma-dev
func discoverConfigFiles() []string {
	var paths []string
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths,
			filepath.Join(home, "Library", "LaunchAgents", "io.puma.dev.plist"),
			"/Library/LaunchAgents/io.puma.dev.plist",
		)
		// puma-dev also uses ~/.puma-dev for state/mappings; include it if exists
		if fi, err := os.Stat(filepath.Join(home, ".puma-dev")); err == nil && fi.IsDir() {
			paths = append(paths, filepath.Join(home, ".puma-dev", "config.yml"))
			paths = append(paths, filepath.Join(home, ".puma-dev", "log", "puma-dev.log"))
		}
	case "linux":
		paths = append(paths,
			filepath.Join(home, ".config", "systemd", "user", "puma-dev.service"),
			"/etc/systemd/system/puma-dev.service",
		)
		// Include common XDG locations
		paths = append(paths, filepath.Join(home, ".puma-dev", "config.yml"))
	default:
		// Fallback: try common paths
		paths = append(paths, filepath.Join(home, ".puma-dev", "config.yml"))
	}
	// filter to existing files
	var existing []string
	for _, p := range paths {
		if p == "" {
			continue
		}
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			existing = append(existing, p)
		}
	}
	return existing
}

// indent helper for pretty printing embedded file contents
func indent(s string, spaces int) string {
	pad := strings.Repeat(" ", spaces)
	var out []string
	for _, line := range strings.Split(s, "\n") {
		out = append(out, pad+line)
	}
	return strings.Join(out, "\n")
}

// cmdContext returns a context if available from root command in future; for now, a simple background context
func cmdContext() (ctx interface{ Done() <-chan struct{} }) { return dummyCtx{} }

type dummyCtx struct{}

func (d dummyCtx) Done() <-chan struct{} { return nil }
