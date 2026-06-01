# CLAUDE.md - gzh-cli-net-env

Network Environment Management Library for gzh-cli ecosystem.

## Quick Start

```bash
make build     # Build
make test      # Run tests
make check     # fmt + lint + test
```

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
    ├── utils.go         # Config directory utilities
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

`github.com/gizzahub/gzh-cli-net-env` — Go 1.23+
