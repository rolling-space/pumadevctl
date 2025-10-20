package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/pumadevctl/internal"
)

var doctorRaw bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check if puma-dev is installed and running; print config and runtime details",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := internal.NewFormatter(cmd.OutOrStdout())
		f.Header("puma-dev doctor")
		installedPath, installed := findPumaDevBinary()
		if !installed {
			f.Warn("puma-dev is not installed or not found in PATH.")
			f.Info("Install instructions: %s", "https://github.com/puma/puma-dev?tab=readme-ov-file#installation")
			return nil
		}

		f.KV("binary", installedPath)
		if v := pumaDevVersion(installedPath); v != "" {
			f.KV("version", v)
		}

		// Check if it's running by attempting TCP dials to common ports.
		running, port := isPumaDevRunning()
		if running {
			f.KV("status", fmt.Sprintf("running (127.0.0.1:%d)", port))
		} else {
			f.KV("status", "not reachable on typical ports (9280/80/3000)")
		}

		// Discover configuration files depending on OS and show a human-readable summary by default.
		paths := discoverConfigFiles()
		if len(paths) == 0 {
			f.KV("configuration", "not found in common locations")
			return nil
		}
		f.Subheader("configuration")
		for _, p := range paths {
			f.Bullet(p)
			b, err := os.ReadFile(p)
			if err != nil {
				f.Warn("  (unable to read: %v)", err)
				continue
			}
			ff := f.IndentBy(2)
			summary := summarizeConfigFile(p, string(b))
			if len(summary) == 0 {
				ff.Info("(no summary available)")
			} else {
				for _, kv := range summary {
					ff.KV(kv.Key, kv.Val)
				}
			}
			if doctorRaw {
				ff.Info("--- begin file ---")
				for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
					ff.Info("%s", line)
				}
				ff.Info("--- end file ---")
			}
		}
		return nil
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorRaw, "raw", false, "print raw configuration files in addition to the summary")
	rootCmd.AddCommand(doctorCmd)
}

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

type kvPair struct{ Key, Val string }

func summarizeConfigFile(path, content string) []kvPair {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".yml"), strings.HasSuffix(lower, ".yaml"):
		m := parseTopLevelYAML(content)
		if len(m) == 0 {
			return nil
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make([]kvPair, 0, len(keys))
		for _, k := range keys {
			out = append(out, kvPair{Key: k, Val: m[k]})
		}
		return out
	case strings.HasSuffix(lower, ".service"):
		return summarizeSystemdService(content)
	case strings.HasSuffix(lower, ".plist"):
		return summarizePlist(content)
	default:
		return nil
	}
}

func parseTopLevelYAML(s string) map[string]string {
	res := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		line := scanner.Text()
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		// Skip nested keys (indented)
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		if idx := strings.Index(v, " #"); idx >= 0 {
			v = strings.TrimSpace(v[:idx])
		}
		v = strings.Trim(v, "'\"")
		if k != "" && v != "" {
			res[k] = v
		}
	}
	return res
}

func summarizeSystemdService(s string) []kvPair {
	var out []kvPair
	scanner := bufio.NewScanner(strings.NewReader(s))
	inService := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			inService = line == "[Service]"
			continue
		}
		if !inService {
			continue
		}
		if strings.HasPrefix(line, "ExecStart=") {
			out = append(out, kvPair{Key: "exec", Val: strings.TrimPrefix(line, "ExecStart=")})
		}
		if strings.HasPrefix(line, "Environment=") {
			out = append(out, kvPair{Key: "environment", Val: strings.TrimPrefix(line, "Environment=")})
		}
		if strings.HasPrefix(line, "User=") {
			out = append(out, kvPair{Key: "user", Val: strings.TrimPrefix(line, "User=")})
		}
	}
	return out
}

func summarizePlist(s string) []kvPair {
	var out []kvPair
	// Label
	if lbl := extractXMLStringBetween(s, "Label"); lbl != "" {
		out = append(out, kvPair{Key: "label", Val: lbl})
	}
	// ProgramArguments (joined)
	args := extractXMLArrayStrings(s, "ProgramArguments")
	if len(args) > 0 {
		out = append(out, kvPair{Key: "args", Val: strings.Join(args, " ")})
	}
	// KeepAlive
	if hasXMLBool(s, "KeepAlive") {
		out = append(out, kvPair{Key: "keepalive", Val: "true"})
	}
	// RunAtLoad
	if hasXMLBool(s, "RunAtLoad") {
		out = append(out, kvPair{Key: "run_at_load", Val: "true"})
	}
	return out
}

func extractXMLStringBetween(s, key string) string {
	// naive search: <key>Label</key> <string>VALUE</string>
	k := fmt.Sprintf("<key>%s</key>", key)
	idx := strings.Index(s, k)
	if idx < 0 {
		return ""
	}
	s = s[idx+len(k):]
	op := strings.Index(s, "<string>")
	cl := strings.Index(s, "</string>")
	if op >= 0 && cl > op {
		return s[op+len("<string>") : cl]
	}
	return ""
}

func extractXMLArrayStrings(s, key string) []string {
	k := fmt.Sprintf("<key>%s</key>", key)
	idx := strings.Index(s, k)
	if idx < 0 {
		return nil
	}
	s = s[idx+len(k):]
	arrStart := strings.Index(s, "<array>")
	arrEnd := strings.Index(s, "</array>")
	if arrStart < 0 || arrEnd < 0 || arrEnd <= arrStart {
		return nil
	}
	arr := s[arrStart+len("<array>") : arrEnd]
	var vals []string
	for {
		op := strings.Index(arr, "<string>")
		if op < 0 {
			break
		}
		cl := strings.Index(arr[op+8:], "</string>")
		if cl < 0 {
			break
		}
		val := arr[op+8 : op+8+cl]
		vals = append(vals, val)
		arr = arr[op+8+cl+9:]
	}
	return vals
}

func hasXMLBool(s, key string) bool {
	k := fmt.Sprintf("<key>%s</key>", key)
	idx := strings.Index(s, k)
	if idx < 0 {
		return false
	}
	s = s[idx+len(k):]
	return strings.Contains(s, "<true/>")
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
