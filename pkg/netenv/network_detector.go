// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	platformWindows = "windows"
	platformDarwin  = "darwin"
	platformLinux   = "linux"
)

// NetworkDetector handles automatic network environment detection.
type NetworkDetector struct {
	profiles []NetworkProfile
}

// NewNetworkDetector creates a new network detector.
func NewNetworkDetector(profiles []NetworkProfile) *NetworkDetector {
	return &NetworkDetector{
		profiles: profiles,
	}
}

// DetectEnvironment automatically detects the current network environment.
func (nd *NetworkDetector) DetectEnvironment(ctx context.Context) (*NetworkProfile, error) {
	networkInfo, err := nd.getCurrentNetworkInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %w", err)
	}

	bestProfile := nd.findBestMatchingProfile(networkInfo)
	return bestProfile, nil
}

func (nd *NetworkDetector) getCurrentNetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	info := &NetworkInfo{
		Timestamp: time.Now(),
	}

	if ssid, err := nd.getWiFiSSID(ctx); err == nil {
		info.WiFiSSID = ssid
	}

	if ips, err := nd.getLocalIPs(); err == nil {
		info.LocalIPs = ips
	}

	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	if gateway, err := nd.getDefaultGateway(ctx); err == nil {
		info.DefaultGateway = gateway
	}

	if dns, err := nd.getDNSServers(); err == nil {
		info.DNSServers = dns
	}

	return info, nil
}

func (nd *NetworkDetector) getWiFiSSID(ctx context.Context) (string, error) {
	switch runtime.GOOS {
	case platformDarwin:
		return nd.getWiFiSSIDMacOS(ctx)
	case platformLinux:
		return nd.getWiFiSSIDLinux(ctx)
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (nd *NetworkDetector) getWiFiSSIDMacOS(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport", "-I")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, " SSID:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("SSID not found")
}

func (nd *NetworkDetector) getWiFiSSIDLinux(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "iwgetid", "-r")
	if output, err := cmd.Output(); err == nil {
		ssid := strings.TrimSpace(string(output))
		if ssid != "" {
			return ssid, nil
		}
	}

	cmd = exec.CommandContext(ctx, "nmcli", "-t", "-f", "active,ssid", "dev", "wifi")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "yes:") {
			return strings.TrimPrefix(line, "yes:"), nil
		}
	}

	return "", fmt.Errorf("SSID not found")
}

func (nd *NetworkDetector) getLocalIPs() ([]string, error) {
	var ips []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips, nil
}

func (nd *NetworkDetector) getDefaultGateway(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "route", "-n", "get", "default")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.CommandContext(ctx, "ip", "route", "show", "default")
		output, err = cmd.Output()
		if err != nil {
			return "", err
		}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "gateway") || strings.Contains(line, "via") {
			fields := strings.Fields(line)
			for i, field := range fields {
				if (field == "gateway" || field == "via") && i+1 < len(fields) {
					return fields[i+1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("gateway not found")
}

func (nd *NetworkDetector) getDNSServers() ([]string, error) {
	if runtime.GOOS == platformWindows {
		return []string{}, nil
	}

	content, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}

	var servers []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				servers = append(servers, fields[1])
			}
		}
	}

	return servers, nil
}

func (nd *NetworkDetector) findBestMatchingProfile(networkInfo *NetworkInfo) *NetworkProfile {
	var bestProfile *NetworkProfile
	bestScore := 0

	for i := range nd.profiles {
		score := nd.scoreProfile(&nd.profiles[i], networkInfo)
		if score > bestScore {
			bestScore = score
			bestProfile = &nd.profiles[i]
		}
	}

	return bestProfile
}

func (nd *NetworkDetector) scoreProfile(profile *NetworkProfile, networkInfo *NetworkInfo) int {
	score := 0

	for _, condition := range profile.Conditions {
		switch condition.Type {
		case "wifi_ssid":
			if nd.matchCondition(condition, networkInfo.WiFiSSID) {
				score += 100
			}
		case "ip_range":
			for _, ip := range networkInfo.LocalIPs {
				if nd.matchIPRange(condition.Value, ip) {
					score += 50
				}
			}
		case "hostname":
			if nd.matchCondition(condition, networkInfo.Hostname) {
				score += 30
			}
		case "gateway":
			if nd.matchCondition(condition, networkInfo.DefaultGateway) {
				score += 70
			}
		}
	}

	score += profile.Priority
	return score
}

func (nd *NetworkDetector) matchCondition(condition NetworkCondition, value string) bool {
	switch condition.Operator {
	case "contains":
		return strings.Contains(value, condition.Value)
	case "matches", "equals", "":
		return value == condition.Value
	default:
		return false
	}
}

func (nd *NetworkDetector) matchIPRange(cidr, ip string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return cidr == ip
	}

	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}

	return network.Contains(ipAddr)
}
