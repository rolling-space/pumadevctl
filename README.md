# pumadevctl

A practical CLI to manage your `~/.puma-dev` domain mappings. Because editing random files in that folder at 2am is a cry for help.

## Features

- **List** with grouping of duplicate mappings (same port/host:port across multiple domains)
- **CRUD**: create, read, update, delete
- Create **symlinks** with `--link` (for puma-dev app symlink style)
- **Auto-port allocation** when you omit the mapping (`create myapp`): picks the first available port block within the configurable range (default 36000-37000, reserving 10 ports per domain)
- **Validate**: TCP dial each mapping and report reachable vs unreachable
- **Cleanup**: delete unreachable mappings, with `--dry-run` and `--yes`
- Fancy output with color; `--json` for machine-friendly output
- `--dir` to target a different directory than `~/.puma-dev`

## Install

```bash
go install github.com/rolling-space/pumadevctl@latest
```

## Versioning and releases

This CLI embeds version metadata at build time. By default, local builds display `dev`.

To produce a release binary with version, commit, and build date:

```bash
VERSION=v0.4.0
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Build to ./exe as per project guidelines
mkdir -p ./exe
GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) \
  go build -ldflags "-s -w \
    -X github.com/rolling-space/pumadevctl/internal.Version=$VERSION \
    -X github.com/rolling-space/pumadevctl/internal.Commit=$COMMIT \
    -X github.com/rolling-space/pumadevctl/internal.Date=$DATE" \
  -o ./exe/pumadevctl .
```

Check the result:

```bash
./exe/pumadevctl --version         # short summary
./exe/pumadevctl version           # detailed info
```

## Usage

```bash
pumadevctl list
pumadevctl create myapp 36777
pumadevctl create myapi                 # auto-allocates a port >= 30000
pumadevctl create myapp --link ~/dev/myapp   # symlink entry
pumadevctl read myapp
pumadevctl update myapp 36888
pumadevctl update myapp --link ~/dev/other   # repoint symlink
pumadevctl delete myapp
pumadevctl validate --timeout 500
pumadevctl cleanup --dry-run
```

## Notes

- Mapping accepts `PORT` or `HOST:PORT` (supports `[::1]:3000` style IPv6)
- Validation only dials non-symlink entries; symlinks are listed as-is
- Auto-port allocation checks used ports in mappings and also tries listening to confirm availability
- Deletion prompts unless `--force` or `cleanup --yes`

MIT licensed. You break it, you get to keep both pieces.
