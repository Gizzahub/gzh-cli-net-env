# CLAUDE.md - gzh-cli-net-env

LLM-optimized guidance for gzh-cli-net-env.

**Module**: `github.com/gizzahub/gzh-cli-net-env` | **Binary root cmd**: `net-env` | **Go**: 1.25.7

Network Environment Management Library for the gzh-cli ecosystem: detects the
active network (WiFi SSID / IP / gateway / DNS) and manages named profiles.

> **Status**: Local-only repo — no GitHub remote yet. Registered in the devbox
> `.gz-project.yaml` (path-based) but absent from `.gz-git.yaml` (needs a url).
> Do not assume `git push`/`gz-git workspace sync` targets exist for this repo.

## Quick Start

```bash
make build     # Build
make test      # Run tests
make check     # fmt + lint + test
make tidy      # go mod tidy
```

## Commands

| Command                 | Purpose                                    |
|-------------------------|--------------------------------------------|
| `net-env status`        | Show current network state                 |
| `net-env watch`         | Monitor network changes                    |
| `net-env profile list`  | List saved profiles                        |
| `net-env profile show <name>` | Show one profile                     |
| `net-env profile init`  | Create/initialize a profile                |
| `net-env profile delete <name>` | Delete a profile                   |

## Absolute Rules

**DO**: Run `make check` before every commit · Keep platform branches (macOS/Linux)
symmetric in `network_detector.go` · Store profiles as YAML under the config dir.

**DON'T**: Assume a remote exists (local-only) · Hard-code the config path — honor
`GZH_CONFIG_DIR` · Duplicate utilities that belong in `gzh-cli-core`.

> **Core dependency**: This module depends on `gzh-cli-core` for config directory
> resolution (`config.GetConfigDirectory`, `config.EnsureConfigDirectory`). Import
> `github.com/gizzahub/gzh-cli-core/config` for shared utilities.

## Architecture

```
gzh-cli-net-env/
├── cmd/netenv/          # CLI entry point (NewRootCmd)
│   ├── root.go          # NewRootCmd() — used by gzh-cli wrapper
│   ├── status.go        # status subcommand
│   ├── watch.go         # watch subcommand
│   └── profile.go       # profile subcommand
└── pkg/netenv/          # Core library packages
    ├── types.go         # NetworkProfile, ComponentStatus, etc.
    ├── utils.go         # Profile path helpers (delegates to gzh-cli-core/config)
    ├── network_detector.go  # WiFi SSID / IP / gateway detection
    └── profile_manager.go   # YAML-based profile CRUD
```

## Key API

```go
import netenv "github.com/gizzahub/gzh-cli-net-env/cmd/netenv"

cmd := netenv.NewRootCmd()  // used by gzh-cli wrapper
```

## Configuration

Profiles stored in `~/.config/gzh-manager/net-env/profiles/*.yaml`

Override with `GZH_CONFIG_DIR` environment variable.

## Platform Support

| Feature | macOS | Linux |
|---------|-------|-------|
| WiFi SSID detection | airport | iwgetid/nmcli |
| IP detection | net.InterfaceAddrs | net.InterfaceAddrs |
| Gateway detection | route | ip route |
| DNS servers | /etc/resolv.conf | /etc/resolv.conf |

## Module

`github.com/gizzahub/gzh-cli-net-env` — Go 1.25.7+
