Project development guidelines for pumadevctl

Scope
- Audience: advanced Go developers contributing to a small CLI that manages Puma-dev style domain→port/symlink mappings.
- This document captures project-specific details only: build/config nuances, testing strategy, and development tips derived from current code.

Build and configuration
- Go version: 1.21 (see go.mod). Use a recent Go toolchain matching or newer than 1.21.
- Modules: standard Go modules; no vendoring required.
- Binary entrypoint: main.go wires up Cobra commands defined under cmd/.
- Third-party deps: github.com/spf13/cobra for CLI, github.com/fatih/color for TTY-styled output.
- Building
  - Package-level build (recommended during development): build only the root module or individual packages to avoid unrelated breakages.
    - go build ./internal  # builds the library layer
    - go build .           # builds the root command (main.go) when cmd packages compile
  - Full recursive build can fail if some CLI source is mid-change; prefer targeted builds per package while iterating.
  - Always output the built binary under the `./exe` folder.
- Running the CLI
  - go run . [command] [flags]
  - The commands assume a directory of mappings (files or symlinks). A mapping file contains either a port (e.g., 3000) or host:port (e.g., 127.0.0.1:3000). Symlinks represent a different class of mapping (LinkTarget is read via os.Readlink).
  - The directory path must be provided via the global --dir flag exposed by the root command (see internal/config.ResolveDir). ResolveDir enforces:
    - Non-empty path
    - Absolute path exists and is a directory

Testing
- Philosophy
  - Prefer unit tests in internal/ against pure functions. Avoid tests that depend on the local network, filesystem outside t.TempDir, or terminal state.
  - CLI-level E2E can be added later using exec.Command and t.TempDir; however, keep the default test suite hermetic.
- Running tests (per-package)
  - To run tests for the internal package only (recommended while the CLI is evolving):
    - go test ./internal
  - Running all packages recursively can fail if CLI code does not compile at some point:
    - go test ./...  # only when cmd/ builds successfully
- Adding tests
  - File naming: *_test.go in the same package directory as the code under test.
  - Use t.TempDir() for ephemeral files and symlinks representing puma-dev entries.
  - For functions that may touch the network (e.g., ValidateEntries which dials TCP):
    - Prefer testing parsing/transform layers (ParseMapping, GroupByMapping) without opening sockets.
    - If you must exercise reachability, stand up a temporary listener on 127.0.0.1 with net.Listen in the test and close it at the end; or extract seams to avoid real dials.
  - For colorized output (internal/format.go), assert on structured data (GroupByMapping) rather than terminal color control sequences.
- Demonstration: simple, working unit test (what we ran successfully)
  - Target: internal.GroupByMapping (pure function; no I/O, deterministic ordering within groups, stable output).
  - Steps we executed locally:
    1) Create a test file internal/format_test.go with:

       package internal

       import (
           "reflect"
           "testing"
       )

       func TestGroupByMapping_Basic(t *testing.T) {
           entries := []Entry{
               {Domain: "a.test", Mapping: "127.0.0.1:3000"},
               {Domain: "b.test", Mapping: "127.0.0.1:3000"},
               {Domain: "c.test", Mapping: "127.0.0.1:4000"},
               {Domain: "link.test", IsSymlink: true},
           }
           groups := GroupByMapping(entries)

           got := map[string][]string{}
           for _, g := range groups {
               got[g.Mapping] = g.Domains
           }

           want := map[string][]string{
               "127.0.0.1:3000": {"a.test", "b.test"},
               "127.0.0.1:4000": {"c.test"},
               "(symlink)":      {"link.test"},
           }
           if !reflect.DeepEqual(got, want) {
               t.Fatalf("unexpected groups: %#v", got)
           }
       }

    2) Run only the internal package tests:
       go test ./internal

    3) Remove the test file after verification (to keep the repo clean if the test was illustrative-only).

  - Notes:
    - GroupByMapping sorts the domain list within each bucket, and sorts group buckets by their Mapping key. When validating, prefer key-based assertions (mapping->domains) instead of positional indexes to avoid assumptions about bucket order.

Additional development information
- Directory/mapping model (internal/entries.go)
  - LoadEntries reads non-directory entries in the target folder. Symlinks are classified with IsSymlink=true and LinkTarget set via os.Readlink; regular files carry Mapping=trimmed file content.
  - Sorting: LoadEntries returns entries sorted by Domain.
  - WriteEntry/CreateSymlink/UpdateEntry/UpdateSymlink/DeleteEntry provide primitive operations; prefer these helpers in CLI code to keep behavior consistent (e.g., overwrite semantics, chmod 0644 for files).
- Mapping parsing and reachability (internal/portutil.go, internal/validation.go)
  - ParseMapping accepts a single port (assumes 127.0.0.1) or host:port, including IPv6 in [::1]:3000 form; validates the numeric port range.
  - ValidateEntries skips reachability checks for symlinks; for non-symlinks it attempts a TCP connection with a configurable timeout (milliseconds) via IsPortReachable.
  - Tests should avoid real network flakiness; if needed, start a local listener inside tests.
- Output/formatting (internal/format.go)
  - PrintListFancy writes styled output to stdout; PrintListJSON writes JSON to color.Output. For programmatic uses, prefer GroupByMapping and PrintListJSON.
- CLI structure (cmd/*)
  - Commands are built on Cobra; root wiring is in cmd/root.go and main.go. Each subcommand resolves the mappings directory via internal.ResolveDir before performing operations.
  - JSON mode: where supported, a --json flag should render machine-readable output. Quiet mode is typically honored to suppress "nothing to do" chatter.
- Error-handling and UX conventions
  - Prefer RunE in Cobra commands and return errors rather than exiting; let Cobra print the error with a non-zero exit code.
  - For potentially-destructive ops (e.g., cleanup), commands support --yes/--dry-run and/or interactive confirmation.

Known caveats while developing
- A full module build/test (go build ./... or go test ./...) may fail if certain cmd/* sources are mid-refactor or contain formatting issues. When iterating, target the package(s) you changed (e.g., ./internal) to keep feedback loops fast. Ensure the cmd package compiles before publishing CLI changes.
- Colorized output depends on a TTY; JSON printing writes to color.Output which handles Windows terminals via go-colorable; avoid hardcoding ANSI expectations in tests.

Release hygiene
- Keep go.mod and go.sum tidy: `go mod tidy` after dependency updates.
- Ensure CLI compiles on linux/amd64 and darwin/amd64 at minimum: `GOOS=linux GOARCH=amd64 go build .` etc.
- Document any new global flags in README.md and ensure JSON output remains stable for scripting consumers.
