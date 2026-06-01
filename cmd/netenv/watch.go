// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-net-env/pkg/netenv"
)

func newWatchCmd() *cobra.Command {
	var interval time.Duration

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Continuously monitor network changes",
		Long: `Monitor network status in real-time, refreshing at a configured interval.

Examples:
  # Watch with default 5-second interval
  net-env watch

  # Watch with custom interval
  net-env watch --interval 10s`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWatch(cmd, interval)
		},
	}

	cmd.Flags().DurationVarP(&interval, "interval", "i", 5*time.Second, "Refresh interval")

	return cmd
}

func runWatch(cmd *cobra.Command, interval time.Duration) error {
	configDir := netenv.GetConfigDirectory()

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

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ctx := cmd.Context()

	for {
		fmt.Print("\033[2J\033[H")
		fmt.Printf("Network Status (Updated: %s) — Press Ctrl+C to exit\n\n", time.Now().Format("15:04:05"))

		status := collectStatus(ctx, detector)
		if err := printStatusTable(status, false); err != nil {
			fmt.Printf("Display error: %v\n", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
