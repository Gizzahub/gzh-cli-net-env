// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-net-env/pkg/netenv"
)

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage network environment profiles",
		Long: `Manage network environment profiles including listing, viewing, creating, and deleting.

Network profiles define comprehensive network configurations that can be
automatically applied when switching network environments.

Examples:
  # List all profiles
  net-env profile list

  # Show profile details
  net-env profile show office

  # Create default example profiles
  net-env profile init`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newProfileListCmd())
	cmd.AddCommand(newProfileShowCmd())
	cmd.AddCommand(newProfileInitCmd())
	cmd.AddCommand(newProfileDeleteCmd())

	return cmd
}

func newProfileListCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all network profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			pm := newProfileManager()
			if err := pm.LoadProfiles(); err != nil {
				return err
			}

			profiles := pm.ListProfiles()
			if len(profiles) == 0 {
				fmt.Println("No profiles configured.")
				fmt.Println("Run 'net-env profile init' to create default profiles.")
				return nil
			}

			fmt.Printf("%-20s %-8s %-50s\n", "NAME", "PRIORITY", "DESCRIPTION")
			fmt.Println("─────────────────────────────────────────────────────────────────────────────")
			for _, p := range profiles {
				fmt.Printf("%-20s %-8d %s\n", p.Name, p.Priority, p.Description)
				if verbose && len(p.Conditions) > 0 {
					for _, c := range p.Conditions {
						fmt.Printf("    Condition: %s %s %s\n", c.Type, c.Operator, c.Value)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed profile information")
	return cmd
}

func newProfileShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show details of a specific profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm := newProfileManager()
			if err := pm.LoadProfiles(); err != nil {
				return err
			}

			profile, err := pm.GetProfile(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("Name:        %s\n", profile.Name)
			fmt.Printf("Description: %s\n", profile.Description)
			fmt.Printf("Priority:    %d\n", profile.Priority)
			fmt.Printf("Auto:        %v\n", profile.Auto)

			if len(profile.Conditions) > 0 {
				fmt.Println("\nConditions:")
				for _, c := range profile.Conditions {
					op := c.Operator
					if op == "" {
						op = "equals"
					}
					fmt.Printf("  - %s %s %q\n", c.Type, op, c.Value)
				}
			}

			if profile.Components.WiFi != nil {
				fmt.Printf("\nWiFi: SSID=%s\n", profile.Components.WiFi.SSID)
			}

			if profile.Components.VPN != nil {
				fmt.Printf("VPN:  %s (%s)\n", profile.Components.VPN.Name, profile.Components.VPN.Type)
			}

			if profile.Components.DNS != nil {
				fmt.Printf("DNS:  %v\n", profile.Components.DNS.Servers)
			}

			if profile.Components.Proxy != nil && (profile.Components.Proxy.HTTP != "" || profile.Components.Proxy.HTTPS != "") {
				fmt.Printf("Proxy: HTTP=%s HTTPS=%s\n", profile.Components.Proxy.HTTP, profile.Components.Proxy.HTTPS)
			}

			return nil
		},
	}
}

func newProfileInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create default example profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			pm := newProfileManager()
			if err := pm.LoadProfiles(); err != nil {
				return err
			}

			if err := pm.CreateDefaultProfiles(); err != nil {
				return fmt.Errorf("failed to create default profiles: %w", err)
			}

			fmt.Println("Default profiles created: home, office, cafe")
			return nil
		},
	}
}

func newProfileDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm := newProfileManager()
			if err := pm.LoadProfiles(); err != nil {
				return err
			}

			if err := pm.DeleteProfile(args[0]); err != nil {
				return err
			}

			fmt.Printf("Profile '%s' deleted.\n", args[0])
			return nil
		},
	}
}

func newProfileManager() *netenv.ProfileManager {
	return netenv.NewProfileManager(netenv.GetConfigDirectory())
}
