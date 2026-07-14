// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gizzahub/gzh-cli-net-env/pkg/netenv"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	_ = w.Close()
	buf := &bytes.Buffer{}
	_, _ = io.Copy(buf, r)
	_ = r.Close()
	return buf.String()
}

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	if cmd.Use != "net-env" {
		t.Errorf("expected Use 'net-env', got %q", cmd.Use)
	}

	subcommands := cmd.Commands()
	if len(subcommands) != 3 {
		t.Fatalf("expected 3 subcommands, got %d", len(subcommands))
	}

	expected := map[string]bool{"status": false, "watch": false, "profile": false}
	for _, sub := range subcommands {
		if _, ok := expected[sub.Name()]; ok {
			expected[sub.Name()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("subcommand %q not found", name)
		}
	}
}

func TestNewRootCmd_HelpOutput(t *testing.T) {
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.Args = nil
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	for _, expected := range []string{"network environment", "status", "watch", "profile"} {
		if !strings.Contains(output, expected) {
			t.Errorf("help output missing %q", expected)
		}
	}
}

func TestStatusCmd_TableOutput(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"status"})

	output := captureStdout(t, func() {
		_ = cmd.Execute()
	})

	if !strings.Contains(output, "Network Environment Status") {
		t.Errorf("expected status output to contain header, got: %s", output)
	}
}

func TestStatusCmd_JSONOutput(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"status", "--format", "json"})

	output := captureStdout(t, func() {
		_ = cmd.Execute()
	})

	if !strings.Contains(output, "{") {
		t.Errorf("expected JSON output to contain '{', got: %s", output)
	}
}

func TestStatusCmd_WithHealth(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"status", "--health"})

	output := captureStdout(t, func() {
		_ = cmd.Execute()
	})

	if !strings.Contains(output, "Network Health") {
		t.Errorf("expected health output, got: %s", output)
	}
}

func TestCheckProxyStatus_Disabled(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")

	status := checkProxyStatus()
	if status.Active {
		t.Error("expected proxy inactive when no env vars set")
	}
	if status.Status != "disabled" {
		t.Errorf("expected 'disabled', got %q", status.Status)
	}
}

func TestCheckProxyStatus_Enabled(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://proxy.example.com:8080")
	t.Setenv("HTTPS_PROXY", "http://proxy.example.com:8080")

	status := checkProxyStatus()
	if !status.Active {
		t.Error("expected proxy active when env vars set")
	}
	if status.Status != "enabled" {
		t.Errorf("expected 'enabled', got %q", status.Status)
	}
	if status.Details["http"] != "http://proxy.example.com:8080" {
		t.Errorf("expected http detail, got %v", status.Details["http"])
	}
}

func TestCheckWiFiStatus(t *testing.T) {
	status := checkWiFiStatus()
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if !status.Active {
		t.Error("expected WiFi active")
	}
}

func TestCheckVPNStatus(t *testing.T) {
	status := checkVPNStatus()
	if status == nil {
		t.Fatal("expected non-nil status")
	}
}

func TestCheckDNSStatus(t *testing.T) {
	status := checkDNSStatus()
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if !status.Active {
		t.Error("expected DNS active")
	}
}

func TestCalculateHealth(t *testing.T) {
	tests := []struct {
		name          string
		components    netenv.ComponentStatuses
		wantMinScore  int
		wantMaxScore  int
		wantStatusStr string
	}{
		{
			name: "all active",
			components: netenv.ComponentStatuses{
				WiFi:  &netenv.ComponentStatus{Active: true},
				VPN:   &netenv.ComponentStatus{Active: true},
				DNS:   &netenv.ComponentStatus{Active: true},
				Proxy: &netenv.ComponentStatus{Active: true},
			},
			wantMinScore:  100,
			wantMaxScore:  100,
			wantStatusStr: "excellent",
		},
		{
			name: "half active",
			components: netenv.ComponentStatuses{
				WiFi:  &netenv.ComponentStatus{Active: true},
				VPN:   &netenv.ComponentStatus{Active: false},
				DNS:   &netenv.ComponentStatus{Active: true},
				Proxy: &netenv.ComponentStatus{Active: false},
			},
			wantMinScore:  50,
			wantMaxScore:  50,
			wantStatusStr: "fair",
		},
		{
			name: "none active",
			components: netenv.ComponentStatuses{
				WiFi:  &netenv.ComponentStatus{Active: false},
				VPN:   &netenv.ComponentStatus{Active: false},
				DNS:   &netenv.ComponentStatus{Active: false},
				Proxy: &netenv.ComponentStatus{Active: false},
			},
			wantMinScore:  0,
			wantMaxScore:  0,
			wantStatusStr: "poor",
		},
		{
			name:          "all nil",
			components:    netenv.ComponentStatuses{},
			wantMinScore:  0,
			wantMaxScore:  0,
			wantStatusStr: "poor",
		},
		{
			name: "with errors",
			components: netenv.ComponentStatuses{
				WiFi: &netenv.ComponentStatus{Active: true, Error: "weak signal"},
				DNS:  &netenv.ComponentStatus{Active: true},
			},
			wantMinScore:  100,
			wantMaxScore:  100,
			wantStatusStr: "excellent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			health := calculateHealth(tt.components)
			if health.Score < tt.wantMinScore || health.Score > tt.wantMaxScore {
				t.Errorf("score = %d, want %d-%d", health.Score, tt.wantMinScore, tt.wantMaxScore)
			}
			if tt.wantStatusStr != "" && health.Status != tt.wantStatusStr {
				t.Errorf("status = %q, want %q", health.Status, tt.wantStatusStr)
			}
		})
	}
}

func TestPrintStatusTable(t *testing.T) {
	status := &netenv.NetworkStatus{
		Profile: &netenv.NetworkProfile{
			Name:        "test-profile",
			Description: "Test Description",
		},
		Components: netenv.ComponentStatuses{
			WiFi:  &netenv.ComponentStatus{Active: true, Status: "connected"},
			VPN:   &netenv.ComponentStatus{Active: false, Status: "disconnected"},
			DNS:   &netenv.ComponentStatus{Active: true, Status: "configured"},
			Proxy: &netenv.ComponentStatus{Active: false, Status: "disabled"},
		},
		Health: netenv.HealthStatus{
			Status: "good",
			Score:  75,
		},
	}

	err := printStatusTable(status, false)
	if err != nil {
		t.Fatalf("printStatusTable() error = %v", err)
	}
}

func TestPrintStatusTable_NilProfile(t *testing.T) {
	status := &netenv.NetworkStatus{
		Health: netenv.HealthStatus{Status: "unknown"},
	}

	err := printStatusTable(status, false)
	if err != nil {
		t.Fatalf("printStatusTable() error = %v", err)
	}
}

func TestPrintStatusTable_NilComponents(t *testing.T) {
	status := &netenv.NetworkStatus{
		Components: netenv.ComponentStatuses{},
		Health:     netenv.HealthStatus{Status: "poor", Score: 0},
	}

	err := printStatusTable(status, true)
	if err != nil {
		t.Fatalf("printStatusTable() error = %v", err)
	}
}

func TestPrintComponent(t *testing.T) {
	printComponent("WiFi", &netenv.ComponentStatus{Active: true, Status: "connected"})
	printComponent("VPN", &netenv.ComponentStatus{Active: false, Status: "disconnected"})
	printComponent("DNS", &netenv.ComponentStatus{Active: true, Status: "ok", Error: "slow"})
	printComponent("Proxy", nil)
}

func TestProfileListCmd_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("GZH_CONFIG_DIR", tmpDir)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"profile", "list"})

	output := captureStdout(t, func() {
		_ = cmd.Execute()
	})

	if !strings.Contains(output, "No profiles") {
		t.Errorf("expected 'No profiles' message, got: %s", output)
	}
}

func TestProfileInitCmd(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("GZH_CONFIG_DIR", tmpDir)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"profile", "init"})

	output := captureStdout(t, func() {
		_ = cmd.Execute()
	})

	if !strings.Contains(output, "Default profiles created") {
		t.Errorf("expected 'Default profiles created', got: %s", output)
	}

	profilesDir := filepath.Join(tmpDir, "net-env", "profiles")
	files, err := os.ReadDir(profilesDir)
	if err != nil {
		t.Fatalf("failed to read profiles dir: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 profile files, got %d", len(files))
	}
}

func TestProfileListCmd_WithProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("GZH_CONFIG_DIR", tmpDir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"profile", "init"})
	captureStdout(t, func() {
		_ = initCmd.Execute()
	})

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"profile", "list"})

	output := captureStdout(t, func() {
		_ = cmd.Execute()
	})

	for _, name := range []string{"home", "office", "cafe"} {
		if !strings.Contains(output, name) {
			t.Errorf("expected profile %q in output, got: %s", name, output)
		}
	}
}

func TestProfileShowCmd(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("GZH_CONFIG_DIR", tmpDir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"profile", "init"})
	captureStdout(t, func() {
		_ = initCmd.Execute()
	})

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"profile", "show", "home"})

	output := captureStdout(t, func() {
		_ = cmd.Execute()
	})

	if !strings.Contains(output, "Name:") {
		t.Errorf("expected 'Name:' in output, got: %s", output)
	}
	if !strings.Contains(output, "home") {
		t.Errorf("expected 'home' in output, got: %s", output)
	}
}

func TestProfileShowCmd_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("GZH_CONFIG_DIR", tmpDir)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"profile", "show", "nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent profile")
	}
}

func TestProfileDeleteCmd(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("GZH_CONFIG_DIR", tmpDir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"profile", "init"})
	captureStdout(t, func() {
		_ = initCmd.Execute()
	})

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"profile", "delete", "home"})

	output := captureStdout(t, func() {
		_ = cmd.Execute()
	})

	if !strings.Contains(output, "deleted") {
		t.Errorf("expected 'deleted' message, got: %s", output)
	}

	profilePath := filepath.Join(tmpDir, "net-env", "profiles", "home.yaml")
	if _, err := os.Stat(profilePath); !os.IsNotExist(err) {
		t.Errorf("expected profile file to be deleted")
	}
}

func TestProfileDeleteCmd_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("GZH_CONFIG_DIR", tmpDir)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"profile", "delete", "nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for deleting non-existent profile")
	}
}

func TestCollectStatus(t *testing.T) {
	ctx := t.Context()
	profiles := []netenv.NetworkProfile{}
	detector := netenv.NewNetworkDetector(profiles)

	status := collectStatus(ctx, detector)

	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if status.Components.WiFi == nil {
		t.Error("expected WiFi component")
	}
	if status.Components.DNS == nil {
		t.Error("expected DNS component")
	}
	if status.Components.Proxy == nil {
		t.Error("expected Proxy component")
	}
}
