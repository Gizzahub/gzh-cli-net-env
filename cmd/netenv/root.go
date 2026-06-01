// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for network environment management.
// Designed for direct use or wrapping by a parent CLI.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "net-env",
		Short: "Manage network environment configurations",
		Long: `Manage network environment transitions on-demand.

This command helps you manage network configurations when
switching between different network environments. It provides:
- Network status checking (WiFi, VPN, DNS, Proxy)
- Real-time network monitoring
- Network profile management and automatic switching
- Cross-platform support (macOS and Linux)

Examples:
  # Show current network status
  net-env status

  # Monitor network changes in real-time
  net-env watch

  # List configured network profiles
  net-env profile list`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newWatchCmd())
	cmd.AddCommand(newProfileCmd())

	return cmd
}
