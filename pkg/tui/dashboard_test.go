// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/gizzahub/gzh-cli-net-env/pkg/netenv"
)

func TestStatusIcon(t *testing.T) {
	tests := []struct {
		name     string
		active   bool
		hasError bool
		want     string
	}{
		{name: "active", active: true, hasError: false, want: iconConnected},
		{name: "inactive", active: false, hasError: false, want: iconDisconnected},
		{name: "error active", active: true, hasError: true, want: iconWarning},
		{name: "error inactive", active: false, hasError: true, want: iconWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := statusIcon(tt.active, tt.hasError)
			if got != tt.want {
				t.Errorf("statusIcon(%v, %v) = %q, want %q", tt.active, tt.hasError, got, tt.want)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		name   string
		active bool
		status string
		want   string
	}{
		{name: "active connected", active: true, status: "connected", want: colorGreen + iconConnected + " connected" + colorReset},
		{name: "inactive disconnected", active: false, status: "disconnected", want: colorGray + iconDisconnected + " disconnected" + colorReset},
		{name: "active with custom status", active: true, status: "configured", want: colorGreen + iconConnected + " configured" + colorReset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatStatus(tt.active, tt.status)
			if got != tt.want {
				t.Errorf("FormatStatus(%v, %q) = %q, want %q", tt.active, tt.status, got, tt.want)
			}
		})
	}
}

func TestFormatHealth(t *testing.T) {
	tests := []struct {
		name   string
		status string
		score  int
		color  string
	}{
		{name: "excellent", status: "excellent", score: 90, color: colorGreen},
		{name: "good", status: "good", score: 75, color: colorCyan},
		{name: "fair", status: "fair", score: 50, color: colorYellow},
		{name: "poor", status: "poor", score: 20, color: colorRed},
		{name: "unknown", status: "unknown", score: 0, color: colorGray},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatHealth(tt.status, tt.score)
			if !strings.HasPrefix(got, tt.color) {
				t.Errorf("FormatHealth(%q, %d) should start with color %q, got %q", tt.status, tt.score, tt.color, got)
			}
			if !strings.Contains(got, tt.status) {
				t.Errorf("should contain status %q", tt.status)
			}
			if !strings.Contains(got, colorReset) {
				t.Error("should contain reset code")
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{name: "short enough", input: "hello", maxLen: 10, want: "hello"},
		{name: "exact length", input: "hello", maxLen: 5, want: "hello"},
		{name: "truncated with ellipsis", input: "hello world", maxLen: 8, want: "hello..."},
		{name: "very short maxLen", input: "hello", maxLen: 2, want: "he"},
		{name: "zero maxLen", input: "hello", maxLen: 0, want: ""},
		{name: "unicode", input: "안녕하세요", maxLen: 4, want: "안..."},
		{name: "empty input", input: "", maxLen: 5, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{name: "needs padding", input: "hi", width: 5, want: "hi   "},
		{name: "exact width", input: "hello", width: 5, want: "hello"},
		{name: "too long truncated", input: "hello world", width: 5, want: "hello"},
		{name: "empty", input: "", width: 3, want: "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PadRight(tt.input, tt.width)
			if len(got) != tt.width {
				t.Errorf("PadRight(%q, %d) len = %d, want %d", tt.input, tt.width, len(got), tt.width)
			}
		})
	}
}

func TestNewDashboard(t *testing.T) {
	d := NewDashboard(true)
	if d == nil {
		t.Fatal("expected non-nil Dashboard")
	}
	if !d.verbose {
		t.Error("expected verbose=true")
	}
	if d.width != 80 {
		t.Errorf("expected width=80, got %d", d.width)
	}
}

func TestNewDashboardWidth(t *testing.T) {
	d := NewDashboardWidth(false, 120)
	if d.verbose {
		t.Error("expected verbose=false")
	}
	if d.width != 120 {
		t.Errorf("expected width=120, got %d", d.width)
	}
}

func TestRender_NilStatus(t *testing.T) {
	d := NewDashboard(false)
	output := d.Render(nil)
	if !strings.Contains(output, "Network Environment Dashboard") {
		t.Errorf("expected header in nil render, got: %s", output)
	}
}

func TestRender_WithProfile(t *testing.T) {
	d := NewDashboard(false)
	status := &netenv.NetworkStatus{
		Profile: &netenv.NetworkProfile{Name: "home", Description: "Home network"},
		Components: netenv.ComponentStatuses{
			WiFi: &netenv.ComponentStatus{Active: true, Status: "connected"},
			DNS:  &netenv.ComponentStatus{Active: true, Status: "configured"},
		},
		Health: netenv.HealthStatus{Status: "good", Score: 75},
	}

	output := d.Render(status)
	if !strings.Contains(output, "home") {
		t.Error("expected profile name in output")
	}
	if !strings.Contains(output, "connected") {
		t.Error("expected component status in output")
	}
}

func TestRender_VerboseWithIssues(t *testing.T) {
	d := NewDashboard(true)
	status := &netenv.NetworkStatus{
		Health: netenv.HealthStatus{
			Status: "fair",
			Score:  50,
			Issues: []string{"WiFi signal weak", "DNS slow"},
		},
	}

	output := d.Render(status)
	if !strings.Contains(output, "Issues:") {
		t.Error("expected Issues section in verbose mode")
	}
	if !strings.Contains(output, "WiFi signal weak") {
		t.Error("expected issue text in output")
	}
}

func TestRenderHeader_NilProfile(t *testing.T) {
	d := NewDashboard(false)
	output := d.RenderHeader(nil)
	if !strings.Contains(output, "No active profile") {
		t.Errorf("expected 'No active profile', got: %s", output)
	}
}

func TestRenderHeader_WithProfile(t *testing.T) {
	d := NewDashboard(false)
	profile := &netenv.NetworkProfile{Name: "office", Description: "Corporate network"}
	output := d.RenderHeader(profile)
	if !strings.Contains(output, "office") {
		t.Error("expected profile name")
	}
	if !strings.Contains(output, "Corporate network") {
		t.Error("expected description")
	}
}

func TestRenderHeader_LongDescription(t *testing.T) {
	d := NewDashboard(false)
	longDesc := strings.Repeat("A", 200)
	profile := &netenv.NetworkProfile{Name: "test", Description: longDesc}
	output := d.RenderHeader(profile)
	if strings.Contains(output, strings.Repeat("A", 200)) {
		t.Error("description should be truncated")
	}
}

func TestRenderComponents_AllNil(t *testing.T) {
	d := NewDashboard(false)
	output := d.RenderComponents(netenv.ComponentStatuses{})
	if !strings.Contains(output, iconUnknown) {
		t.Error("expected unknown icon for nil components")
	}
}

func TestRenderComponents_ActiveAndInactive(t *testing.T) {
	d := NewDashboard(false)
	components := netenv.ComponentStatuses{
		WiFi: &netenv.ComponentStatus{Active: true, Status: "connected"},
		VPN:  &netenv.ComponentStatus{Active: false, Status: "disconnected"},
	}
	output := d.RenderComponents(components)
	if !strings.Contains(output, iconConnected) {
		t.Error("expected connected icon")
	}
	if !strings.Contains(output, iconDisconnected) {
		t.Error("expected disconnected icon")
	}
}

func TestRenderComponents_WithError(t *testing.T) {
	d := NewDashboard(false)
	components := netenv.ComponentStatuses{
		DNS: &netenv.ComponentStatus{Active: true, Status: "error", Error: "timeout"},
	}
	output := d.RenderComponents(components)
	if !strings.Contains(output, iconWarning) {
		t.Error("expected warning icon for error")
	}
	if !strings.Contains(output, "timeout") {
		t.Error("expected error message in output")
	}
}

func TestRenderHealth(t *testing.T) {
	d := NewDashboard(false)
	health := netenv.HealthStatus{Status: "excellent", Score: 95}
	output := d.RenderHealth(health)
	if !strings.Contains(output, "excellent") {
		t.Error("expected status text")
	}
	if !strings.Contains(output, "95") {
		t.Error("expected score in output")
	}
}

func TestRenderIssues_Empty(t *testing.T) {
	d := NewDashboard(false)
	output := d.RenderIssues(nil)
	if output != "" {
		t.Errorf("expected empty string for no issues, got %q", output)
	}
}

func TestRenderIssues_WithItems(t *testing.T) {
	d := NewDashboard(false)
	issues := []string{"DNS timeout", "VPN disconnected"}
	output := d.RenderIssues(issues)
	if !strings.Contains(output, "DNS timeout") {
		t.Error("expected first issue")
	}
	if !strings.Contains(output, "VPN disconnected") {
		t.Error("expected second issue")
	}
}

func TestRenderIssues_LongIssue(t *testing.T) {
	d := NewDashboardWidth(false, 40)
	longIssue := strings.Repeat("X", 200)
	output := d.RenderIssues([]string{longIssue})
	if strings.Contains(output, strings.Repeat("X", 200)) {
		t.Error("issue should be truncated")
	}
}

func TestRenderNetworkInfo_Nil(t *testing.T) {
	d := NewDashboard(false)
	output := d.RenderNetworkInfo(nil)
	if output != "" {
		t.Errorf("expected empty for nil info, got %q", output)
	}
}

func TestRenderNetworkInfo_Complete(t *testing.T) {
	d := NewDashboard(true)
	info := &netenv.NetworkInfo{
		WiFiSSID:       "HomeWiFi",
		LocalIPs:       []string{"192.168.1.100", "10.0.0.5"},
		Hostname:       "my-mac",
		DefaultGateway: "192.168.1.1",
		DNSServers:     []string{"8.8.8.8", "1.1.1.1"},
		Timestamp:      time.Now(),
	}

	output := d.RenderNetworkInfo(info)
	for _, expected := range []string{"HomeWiFi", "my-mac", "192.168.1.100", "192.168.1.1", "8.8.8.8"} {
		if !strings.Contains(output, expected) {
			t.Errorf("expected %q in output", expected)
		}
	}
}

func TestRenderNetworkInfo_Partial(t *testing.T) {
	d := NewDashboard(true)
	info := &netenv.NetworkInfo{
		WiFiSSID: "OnlyWiFi",
	}

	output := d.RenderNetworkInfo(info)
	if !strings.Contains(output, "OnlyWiFi") {
		t.Error("expected WiFi SSID")
	}
	if strings.Contains(output, "Hostname") {
		t.Error("should not contain Hostname when empty")
	}
}
