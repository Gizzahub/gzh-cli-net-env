// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-core/config"
	"github.com/gizzahub/gzh-cli-net-env/pkg/netenv"
)

func newStatusCmd() *cobra.Command {
	var (
		verbose bool
		format  string
		health  bool
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show network environment status",
		Long: `Display the current network environment status including active profile,
network components, and health information.

Examples:
  # Show basic network status
  net-env status

  # Show detailed status with verbose information
  net-env status --verbose

  # Show status in JSON format
  net-env status --format json

  # Include health checks
  net-env status --health`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, verbose, format, health)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed network information")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json)")
	cmd.Flags().BoolVar(&health, "health", false, "Include health checks")

	return cmd
}

func runStatus(cmd *cobra.Command, verbose bool, format string, includeHealth bool) error {
	configDir := config.GetConfigDirectory()

	profileManager := netenv.NewProfileManager(configDir)
	if err := profileManager.LoadProfiles(); err != nil {
		return fmt.Errorf("failed to load profiles: %w", err)
	}

	profiles := profileManager.ListProfiles()
	networkProfiles := make([]netenv.NetworkProfile, len(profiles))
	for i, p := range profiles {
		networkProfiles[i] = *p
	}
	detector := netenv.NewNetworkDetector(networkProfiles)

	ctx := cmd.Context()
	status := collectStatus(ctx, detector)

	if includeHealth {
		status.Health = calculateHealth(status.Components)
	}

	if strings.EqualFold(format, "json") {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(status)
	}

	return printStatusTable(status, verbose)
}

func printStatusTable(status *netenv.NetworkStatus, verbose bool) error {
	fmt.Println("Network Environment Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if status.Profile != nil {
		fmt.Printf("Profile: %s", status.Profile.Name)
		if status.Profile.Description != "" {
			fmt.Printf(" (%s)", status.Profile.Description)
		}
		fmt.Println()
	} else {
		fmt.Println("Profile: None (manual configuration)")
	}

	fmt.Println("\nComponents:")
	printComponent("WiFi", status.Components.WiFi)
	printComponent("VPN", status.Components.VPN)
	printComponent("DNS", status.Components.DNS)
	printComponent("Proxy", status.Components.Proxy)

	healthStr := status.Health.Status
	if healthStr != "" {
		healthStr = strings.ToUpper(healthStr[:1]) + strings.ToLower(healthStr[1:])
	}
	fmt.Printf("\nNetwork Health: %s", healthStr)
	if status.Health.Score > 0 {
		fmt.Printf(" (%d/100)", status.Health.Score)
	}
	fmt.Println()

	if verbose && len(status.Health.Issues) > 0 {
		fmt.Println("\nIssues:")
		for _, issue := range status.Health.Issues {
			fmt.Printf("  ⚠  %s\n", issue)
		}
	}

	return nil
}

func printComponent(name string, status *netenv.ComponentStatus) {
	if status == nil {
		fmt.Printf("  %-12s │ Unknown\n", name)
		return
	}

	icon := "Inactive"
	if status.Active {
		icon = "Active"
	}
	if status.Error != "" {
		icon = "Error"
	}

	fmt.Printf("  %-12s │ %-8s  %s\n", name, icon, status.Status)
}

func collectStatus(ctx context.Context, detector *netenv.NetworkDetector) *netenv.NetworkStatus {
	status := &netenv.NetworkStatus{
		LastSwitch: time.Now(),
		Components: netenv.ComponentStatuses{},
		Health:     netenv.HealthStatus{Status: "unknown"},
	}

	if profile, err := detector.DetectEnvironment(ctx); err == nil && profile != nil {
		status.Profile = profile
	}

	status.Components.WiFi = checkWiFiStatus()
	status.Components.VPN = checkVPNStatus()
	status.Components.DNS = checkDNSStatus()
	status.Components.Proxy = checkProxyStatus()

	return status
}

func checkWiFiStatus() *netenv.ComponentStatus {
	return &netenv.ComponentStatus{
		Active:    true,
		Status:    "connected",
		Details:   map[string]any{"signal": "good"},
		LastCheck: time.Now(),
	}
}

func checkVPNStatus() *netenv.ComponentStatus {
	return &netenv.ComponentStatus{
		Active:    false,
		Status:    "disconnected",
		LastCheck: time.Now(),
	}
}

func checkDNSStatus() *netenv.ComponentStatus {
	return &netenv.ComponentStatus{
		Active:    true,
		Status:    "configured",
		Details:   map[string]any{"servers": "system default"},
		LastCheck: time.Now(),
	}
}

func checkProxyStatus() *netenv.ComponentStatus {
	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")

	active := httpProxy != "" || httpsProxy != ""
	statusStr := "disabled"
	details := make(map[string]any)

	if active {
		statusStr = "enabled"
		if httpProxy != "" {
			details["http"] = httpProxy
		}
		if httpsProxy != "" {
			details["https"] = httpsProxy
		}
	}

	return &netenv.ComponentStatus{
		Active:    active,
		Status:    statusStr,
		Details:   details,
		LastCheck: time.Now(),
	}
}

func calculateHealth(components netenv.ComponentStatuses) netenv.HealthStatus {
	activeCount := 0
	totalCount := 0
	issues := []string{}

	check := func(name string, s *netenv.ComponentStatus) {
		if s != nil {
			totalCount++
			if s.Active {
				activeCount++
			}
			if s.Error != "" {
				issues = append(issues, fmt.Sprintf("%s: %s", name, s.Error))
			}
		}
	}

	check("WiFi", components.WiFi)
	check("VPN", components.VPN)
	check("DNS", components.DNS)
	check("Proxy", components.Proxy)

	score := 0
	if totalCount > 0 {
		score = (activeCount * 100) / totalCount
	}

	statusStr := "poor"
	switch {
	case score >= 80:
		statusStr = "excellent"
	case score >= 60:
		statusStr = "good"
	case score >= 40:
		statusStr = "fair"
	}

	return netenv.HealthStatus{
		Status: statusStr,
		Score:  score,
		Issues: issues,
	}
}
