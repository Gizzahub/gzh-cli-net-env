// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// ProfileManager handles network profile management operations.
type ProfileManager struct {
	configDir string
	profiles  map[string]*NetworkProfile
}

// NewProfileManager creates a new profile manager.
func NewProfileManager(configDir string) *ProfileManager {
	return &ProfileManager{
		configDir: configDir,
		profiles:  make(map[string]*NetworkProfile),
	}
}

// LoadProfiles loads all network profiles from the configuration directory.
func (pm *ProfileManager) LoadProfiles() error {
	profilesDir := filepath.Join(pm.configDir, "net-env", "profiles")

	if err := os.MkdirAll(profilesDir, 0o750); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return fmt.Errorf("failed to read profiles directory: %w", err)
	}

	pm.profiles = make(map[string]*NetworkProfile)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if !isYAMLFile(fileName) {
			continue
		}

		profilePath := filepath.Join(profilesDir, fileName)
		if err := pm.loadProfile(profilePath); err != nil {
			fmt.Printf("Warning: failed to load profile %s: %v\n", fileName, err)
			continue
		}
	}

	return nil
}

func (pm *ProfileManager) loadProfile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read profile file: %w", err)
	}

	var profile NetworkProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return fmt.Errorf("failed to parse profile YAML: %w", err)
	}

	if err := pm.validateProfile(&profile); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	pm.profiles[profile.Name] = &profile
	return nil
}

// SaveProfile saves a network profile to disk.
func (pm *ProfileManager) SaveProfile(profile *NetworkProfile) error {
	if err := pm.validateProfile(profile); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	now := time.Now()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = now
	}
	profile.UpdatedAt = now

	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	profilesDir := filepath.Join(pm.configDir, "net-env", "profiles")
	if err := os.MkdirAll(profilesDir, 0o750); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	fileName := fmt.Sprintf("%s.yaml", profile.Name)
	filePath := filepath.Join(profilesDir, fileName)

	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	pm.profiles[profile.Name] = profile
	return nil
}

// GetProfile returns a profile by name.
func (pm *ProfileManager) GetProfile(name string) (*NetworkProfile, error) {
	profile, exists := pm.profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}
	return profile, nil
}

// ListProfiles returns all available profiles sorted by priority then name.
func (pm *ProfileManager) ListProfiles() []*NetworkProfile {
	profiles := make([]*NetworkProfile, 0, len(pm.profiles))
	for _, profile := range pm.profiles {
		profiles = append(profiles, profile)
	}

	sort.Slice(profiles, func(i, j int) bool {
		if profiles[i].Priority != profiles[j].Priority {
			return profiles[i].Priority > profiles[j].Priority
		}
		return profiles[i].Name < profiles[j].Name
	})

	return profiles
}

// DeleteProfile removes a profile.
func (pm *ProfileManager) DeleteProfile(name string) error {
	if _, exists := pm.profiles[name]; !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	fileName := fmt.Sprintf("%s.yaml", name)
	filePath := filepath.Join(pm.configDir, "net-env", "profiles", fileName)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete profile file: %w", err)
	}

	delete(pm.profiles, name)
	return nil
}

// ExportProfile exports a profile to a file.
func (pm *ProfileManager) ExportProfile(name, outputPath string) error {
	profile, err := pm.GetProfile(name)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// ImportProfile imports a profile from a file.
func (pm *ProfileManager) ImportProfile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	var profile NetworkProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return fmt.Errorf("failed to parse profile YAML: %w", err)
	}

	return pm.SaveProfile(&profile)
}

// CreateDefaultProfiles creates default example profiles if they don't exist.
func (pm *ProfileManager) CreateDefaultProfiles() error {
	profiles := []*NetworkProfile{
		{
			Name:        "home",
			Description: "Home network profile",
			Priority:    50,
			Conditions: []NetworkCondition{
				{Type: "wifi_ssid", Value: "Home-WiFi", Operator: "equals"},
				{Type: "ip_range", Value: "192.168.1.0/24", Operator: "equals"},
			},
			Components: NetworkComponents{
				DNS:   &DNSConfig{Servers: []string{"1.1.1.1", "1.0.0.1"}},
				Proxy: &ProxyConfig{},
			},
		},
		{
			Name:        "office",
			Description: "Corporate office network profile",
			Priority:    100,
			Conditions: []NetworkCondition{
				{Type: "wifi_ssid", Value: "Corporate-WiFi", Operator: "equals"},
				{Type: "ip_range", Value: "10.0.0.0/8", Operator: "equals"},
			},
			Components: NetworkComponents{
				VPN: &VPNConfig{Name: "corp-vpn", Type: "openvpn", AutoConnect: true, Priority: 100},
				DNS: &DNSConfig{Servers: []string{"10.0.0.1", "10.0.0.2"}, Override: true},
				Proxy: &ProxyConfig{
					HTTP:  "proxy.corp.com:8080",
					HTTPS: "proxy.corp.com:8080",
					Auth:  &ProxyAuth{Username: "${PROXY_USER}", Password: "${PROXY_PASS}"},
				},
			},
		},
		{
			Name:        "cafe",
			Description: "Public WiFi / Cafe network profile",
			Priority:    25,
			Conditions: []NetworkCondition{
				{Type: "wifi_ssid", Value: "Starbucks", Operator: "contains"},
			},
			Components: NetworkComponents{
				VPN: &VPNConfig{Name: "personal-vpn", Type: "wireguard", AutoConnect: true, Priority: 100},
				DNS: &DNSConfig{Servers: []string{"1.1.1.1", "8.8.8.8"}, Override: true},
			},
		},
	}

	for _, profile := range profiles {
		if _, err := pm.GetProfile(profile.Name); err != nil {
			if err := pm.SaveProfile(profile); err != nil {
				return fmt.Errorf("failed to create default profile %s: %w", profile.Name, err)
			}
		}
	}

	return nil
}

// GetAutoProfile returns the best auto-matching profile for current environment.
func (pm *ProfileManager) GetAutoProfile() (*NetworkProfile, error) {
	profiles := pm.ListProfiles()
	autoProfiles := make([]NetworkProfile, 0)

	for _, profile := range profiles {
		if profile.Auto {
			autoProfiles = append(autoProfiles, *profile)
		}
	}

	if len(autoProfiles) == 0 {
		return nil, fmt.Errorf("no auto profiles configured")
	}

	detector := NewNetworkDetector(autoProfiles)
	return detector.DetectEnvironment(context.TODO())
}

func (pm *ProfileManager) validateProfile(profile *NetworkProfile) error {
	if profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if !isValidProfileName(profile.Name) {
		return fmt.Errorf("profile name contains invalid characters")
	}

	for i, condition := range profile.Conditions {
		if condition.Type == "" {
			return fmt.Errorf("condition %d: type is required", i)
		}
		if condition.Value == "" {
			return fmt.Errorf("condition %d: value is required", i)
		}
	}

	return nil
}

func isYAMLFile(fileName string) bool {
	ext := filepath.Ext(fileName)
	return ext == ".yaml" || ext == ".yml"
}

func isValidProfileName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}

	for _, r := range name {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') &&
			(r < '0' || r > '9') && r != '-' && r != '_' {
			return false
		}
	}

	return true
}
