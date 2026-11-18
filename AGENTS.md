# gzh-cli-net-env

## Development workflow — user decision 2026-11-18

- Use the CE task runtime for worktree lifecycle and integration:
  `ce task run-doctor`, `run-start <card-id> --type <type>`, work and push in
  the task worktree, `run-status`, then `run-finish` from the primary checkout.
  `.ce/task-runtime.yaml` is the tracked declaration; worktree placement comes
  from the machine's CE defaults. This replaces the 2026-11-06 plain-Git rule.
- This repository keeps no task board. Its cards live only in the
  gzh-cli-devbox board; name the run after that card's ID. `tasks/` is ignored
  here; do not create cards or receipts in this repository.
- Run the repository's applicable source checks and keep failures visible.
  Documentation and workflow-only changes require diff checks and relevant
  Markdown/configuration or workflow checks; do not require unrelated Go test
  suites for prose edits.
- GitLab CI is not required to build, test, or contribute. GitHub Actions and
  server-side branch rules are separate and must still be respected where present.
- Preserve in-file locked regions. Do not change machine-global installed tools
  or unrelated repositories as part of this repository's workflow.

LLM-optimized guidance for gzh-cli-net-env.

**Module**: `github.com/gizzahub/gzh-cli-net-env` | **Binary root cmd**: `net-env` |
**Consumer Go**: 1.26+ | **Development/CI toolchain**: 1.27.4

Network Environment Management Library for the gzh-cli ecosystem: detects the
active network (WiFi SSID / IP / gateway / DNS) and manages named profiles.

> **Status**: Remote `github.com/Gizzahub/gzh-cli-net-env` exists. Registered in the
> devbox `.gz-project.yaml` (path-based); may still be absent from `.gz-git.yaml`
> until workspace sync registration is updated.

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

**DON'T**: Hard-code the config path — honor `GZH_CONFIG_DIR` · Duplicate utilities
that belong in `gzh-cli-core`.

> **Core dependency**: This module depends on `gzh-cli-core` for config directory
> resolution (`config.GetConfigDirectory`, `config.EnsureConfigDirectory`). Import
> `github.com/gizzahub/gzh-cli-core/config` for shared utilities.

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

`github.com/gizzahub/gzh-cli-net-env` — Consumer Go 1.26+; development and CI
use Go 1.27.4.
