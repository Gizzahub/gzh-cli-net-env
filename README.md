# gzh-cli-net-env

> Network environment detection and profile management for the gzh-cli ecosystem.

`net-env` inspects the active network — WiFi SSID, IP, gateway, and DNS — and
manages named network profiles stored as YAML. It ships as a library consumed by
the main `gzh-cli` wrapper and as a standalone command.

**Module**: `github.com/gizzahub/gzh-cli-net-env` · **Root cmd**: `net-env` · **Go**: 1.25.7

> **Status**: Local-only repository — no public GitHub remote yet. It is registered
> in the devbox `.gz-project.yaml` (path-based) but not in `.gz-git.yaml`.

## Features

- **Network detection** — WiFi SSID, IP address, default gateway, DNS servers
- **Profile management** — create, list, show, and delete YAML network profiles
- **Change monitoring** — watch for network state changes
- **Cross-platform** — macOS and Linux detection backends

## Install

```bash
make build     # build the binary
make check     # fmt + lint + test
```

## Commands

| Command                         | Purpose                     |
|---------------------------------|-----------------------------|
| `net-env status`                | Show current network state  |
| `net-env watch`                 | Monitor network changes     |
| `net-env profile list`          | List saved profiles         |
| `net-env profile show <name>`   | Show one profile            |
| `net-env profile init`          | Create/initialize a profile |
| `net-env profile delete <name>` | Delete a profile            |

## Configuration

Profiles are stored as YAML under:

```
~/.config/gzh-manager/net-env/profiles/*.yaml
```

Override the config directory with the `GZH_CONFIG_DIR` environment variable.

## Platform Support

| Feature             | macOS               | Linux             |
|---------------------|---------------------|-------------------|
| WiFi SSID detection | airport             | iwgetid / nmcli   |
| IP detection        | net.InterfaceAddrs  | net.InterfaceAddrs|
| Gateway detection   | route               | ip route          |
| DNS servers         | /etc/resolv.conf    | /etc/resolv.conf  |

## Library Usage

```go
import netenv "github.com/gizzahub/gzh-cli-net-env/cmd/netenv"

cmd := netenv.NewRootCmd()  // mounted by the gzh-cli wrapper
```

## Architecture

```
cmd/netenv/        CLI entry point (NewRootCmd) + status/watch/profile
pkg/netenv/
  types.go         NetworkProfile, ComponentStatus, ...
  utils.go         config directory utilities
  network_detector.go   WiFi SSID / IP / gateway detection
  profile_manager.go    YAML-based profile CRUD
```

## Development

```bash
make check     # fmt + lint + test (pre-commit)
make test      # run tests
make fmt       # format
make lint      # lint
make tidy      # go mod tidy
```

> **Note**: This module depends on `gzh-cli-core` for shared utilities (config
> directory resolution). Import `github.com/gizzahub/gzh-cli-core/config`.
