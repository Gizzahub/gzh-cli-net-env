// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

// Package tui provides terminal UI helpers for network environment status display.
package tui

import (
	"fmt"
	"strings"

	"github.com/gizzahub/gzh-cli-net-env/pkg/netenv"
)

// Dashboard renders network environment status for the terminal.
type Dashboard struct {
	verbose bool
	width   int
}

// NewDashboard creates a Dashboard with default width.
func NewDashboard(verbose bool) *Dashboard {
	return &Dashboard{verbose: verbose, width: 80}
}

// NewDashboardWidth creates a Dashboard with a custom width.
func NewDashboardWidth(verbose bool, width int) *Dashboard {
	return &Dashboard{verbose: verbose, width: width}
}

// Render returns a full status dashboard for the given network status.
func (d *Dashboard) Render(status *netenv.NetworkStatus) string {
	if status == nil {
		return d.RenderHeader(nil)
	}

	var b strings.Builder
	b.WriteString(d.RenderHeader(status.Profile))
	b.WriteString("\n")
	b.WriteString(d.RenderComponents(status.Components))
	b.WriteString("\n")
	b.WriteString(d.RenderHealth(status.Health))

	if d.verbose && len(status.Health.Issues) > 0 {
		b.WriteString("\n")
		b.WriteString(d.RenderIssues(status.Health.Issues))
	}

	return b.String()
}

// RenderHeader returns the dashboard header for the active profile.
func (d *Dashboard) RenderHeader(profile *netenv.NetworkProfile) string {
	var b strings.Builder
	b.WriteString(colorCyan)
	b.WriteString(strings.Repeat("=", d.width))
	b.WriteString(colorReset)
	b.WriteString("\n")
	b.WriteString(colorCyan)
	b.WriteString("  Network Environment Dashboard")
	b.WriteString(colorReset)
	b.WriteString("\n")
	b.WriteString(colorCyan)
	b.WriteString(strings.Repeat("=", d.width))
	b.WriteString(colorReset)
	b.WriteString("\n\n")

	if profile != nil {
		fmt.Fprintf(&b, "  Active Profile: %s%s%s", colorGreen, profile.Name, colorReset)
		if profile.Description != "" {
			desc := TruncateString(profile.Description, d.width-30)
			fmt.Fprintf(&b, " (%s)", desc)
		}
		b.WriteString("\n")
	} else {
		fmt.Fprintf(&b, "  %sNo active profile (manual configuration)%s\n", colorGray, colorReset)
	}

	return b.String()
}

// RenderComponents returns a formatted list of network component statuses.
func (d *Dashboard) RenderComponents(components netenv.ComponentStatuses) string {
	var b strings.Builder
	b.WriteString("  Components:\n")

	entries := []struct {
		name string
		s    *netenv.ComponentStatus
	}{
		{"WiFi", components.WiFi},
		{"VPN", components.VPN},
		{"DNS", components.DNS},
		{"Proxy", components.Proxy},
		{"Docker", components.Docker},
		{"Kubernetes", components.Kubernetes},
	}

	for _, e := range entries {
		b.WriteString(d.renderComponentLine(e.name, e.s))
	}

	return b.String()
}

func (d *Dashboard) renderComponentLine(name string, s *netenv.ComponentStatus) string {
	label := PadRight("  "+name+":", 16)

	if s == nil {
		return fmt.Sprintf("%s %s%s%s\n", label, colorGray, iconUnknown+" unknown", colorReset)
	}

	statusStr := s.Status
	if s.Error != "" {
		return fmt.Sprintf("%s %s%s (%s)%s\n", label, colorRed, iconWarning, s.Error, colorReset)
	}

	if s.Active {
		return fmt.Sprintf("%s %s%s%s\n", label, colorGreen, iconConnected+" "+statusStr, colorReset)
	}

	return fmt.Sprintf("%s %s%s%s\n", label, colorGray, iconDisconnected+" "+statusStr, colorReset)
}

// RenderHealth returns a formatted network health summary.
func (d *Dashboard) RenderHealth(health netenv.HealthStatus) string {
	var b strings.Builder
	b.WriteString("\n  ")
	b.WriteString(FormatHealth(health.Status, health.Score))
	b.WriteString("\n")

	if d.verbose && len(health.Issues) > 0 {
		b.WriteString("\n")
		b.WriteString(d.RenderIssues(health.Issues))
	}

	return b.String()
}

// RenderIssues returns a formatted list of health issues.
func (d *Dashboard) RenderIssues(issues []string) string {
	if len(issues) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(colorYellow)
	b.WriteString("  Issues:\n")
	b.WriteString(colorReset)

	for _, issue := range issues {
		truncated := TruncateString(issue, d.width-6)
		fmt.Fprintf(&b, "    %s%s%s\n", colorYellow, truncated, colorReset)
	}

	return b.String()
}

// RenderNetworkInfo returns formatted network detection details.
func (d *Dashboard) RenderNetworkInfo(info *netenv.NetworkInfo) string {
	if info == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString(colorGray)
	b.WriteString("  Network Details:\n")
	b.WriteString(colorReset)

	if info.WiFiSSID != "" {
		fmt.Fprintf(&b, "    WiFi SSID:     %s\n", info.WiFiSSID)
	}
	if info.Hostname != "" {
		fmt.Fprintf(&b, "    Hostname:      %s\n", info.Hostname)
	}
	if len(info.LocalIPs) > 0 {
		fmt.Fprintf(&b, "    Local IPs:     %s\n", strings.Join(info.LocalIPs, ", "))
	}
	if info.DefaultGateway != "" {
		fmt.Fprintf(&b, "    Gateway:       %s\n", info.DefaultGateway)
	}
	if len(info.DNSServers) > 0 {
		fmt.Fprintf(&b, "    DNS Servers:   %s\n", strings.Join(info.DNSServers, ", "))
	}
	fmt.Fprintf(&b, "    Detected:      %s\n", info.Timestamp.Format("2006-01-02 15:04:05"))

	return b.String()
}
