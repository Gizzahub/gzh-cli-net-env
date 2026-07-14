# Common Tasks - gzh-cli-net-env

## Architecture: Single-Package Design

All core logic lives in `pkg/netenv/` — a flat package with pure functions and platform dispatch:

```text
cmd/netenv/          # CLI commands (NewRootCmd export)
├── root.go          # NewRootCmd() — used by gzh-cli wrapper
├── status.go        # status + collectStatus + health calculation
├── watch.go         # real-time monitor
└── profile.go       # profile list/show/init/delete

pkg/netenv/          # Core library
├── types.go         # NetworkProfile, ComponentStatus, HealthStatus, ...
├── network_detector.go  # WiFi/IP/gateway/DNS detection + profile matching
├── profile_manager.go   # YAML profile CRUD + validation
└── utils.go         # config path helpers (delegates to gzh-cli-core)

pkg/tui/             # Terminal UI dashboard
├── styles.go        # ANSI colors, FormatStatus, FormatHealth
└── dashboard.go     # Dashboard renderer (Render, RenderHeader, RenderComponents)
```

## Pure Function Extraction Pattern

Platform-dependent code (exec.Command) is split into thin wrappers + testable pure parsers:

```go
// Thin wrapper (platform-specific, hard to test)
func (nd *NetworkDetector) getWiFiSSIDMacOS(ctx context.Context) (string, error) {
    output, err := exec.CommandContext(ctx, "airport", "-I").Output()
    ...
    return parseSSIDFromAirportOutput(string(output))
}

// Pure parser (testable, no I/O)
func parseSSIDFromAirportOutput(output string) (string, error) { ... }
```

## Profile Matching

Profiles are scored against current network state:

| Condition Type | Points on Match |
|----------------|-----------------|
| `wifi_ssid` | 100 |
| `gateway` | 70 |
| `ip_range` | 50 |
| `hostname` | 30 |
| + Priority bonus | profile.Priority |

Highest score wins. Score 0 = no match (nil profile returned).

## CLI Commands

| Command | Purpose |
|---------|---------|
| `net-env status` | Show current network state + component statuses |
| `net-env status --json` | JSON output |
| `net-env status --health` | Include health score |
| `net-env watch` | Real-time monitor (refresh every 5s) |
| `net-env watch --interval 10s` | Custom refresh interval |
| `net-env profile list` | List saved profiles (sorted by priority) |
| `net-env profile show <name>` | Show profile details |
| `net-env profile init` | Create default home/office/cafe profiles |
| `net-env profile delete <name>` | Delete a profile |

## Platform Support

| Feature | macOS | Linux |
|---------|-------|-------|
| WiFi SSID | `airport -I` | `iwgetid` / `nmcli` |
| Local IPs | `net.InterfaceAddrs` | `net.InterfaceAddrs` |
| Gateway | `route -n get default` | `ip route show default` |
| DNS | `/etc/resolv.conf` | `/etc/resolv.conf` |

## Configuration

Profiles stored as YAML under:

```
~/.config/gzh-manager/net-env/profiles/*.yaml
```

Override with `GZH_CONFIG_DIR` environment variable.

## Test Coverage

| Package | Coverage |
|---------|----------|
| `pkg/netenv` | 84.5% |
| `pkg/tui` | 99.2% |
| `cmd/netenv` | 81.0% |
