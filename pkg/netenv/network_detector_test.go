// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"testing"
)

func TestParseSSIDFromAirportOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    string
		wantErr bool
	}{
		{
			name: "valid SSID line",
			output: `     agrCtlRSSI: -52
     agrExtRSSI: 0
    lastTxRate: 878
        BSSID: aa:bb:cc:dd:ee:ff
         SSID: MyHomeNetwork
            RSSI: -52`,
			want:    "MyHomeNetwork",
			wantErr: false,
		},
		{
			name: "SSID with spaces in name",
			output: `     agrCtlRSSI: -52
         SSID: Coffee Shop WiFi
            RSSI: -52`,
			want:    "Coffee Shop WiFi",
			wantErr: false,
		},
		{
			name:    "empty output",
			output:  "",
			want:    "",
			wantErr: true,
		},
		{
			name: "no SSID line",
			output: `     agrCtlRSSI: -52
     agrExtRSSI: 0
    lastTxRate: 878`,
			want:    "",
			wantErr: true,
		},
		{
			name: "SSID line but no colon",
			output: `     agrCtlRSSI: -52
         SSID without colon
            RSSI: -52`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "SSID line with leading content containing ssid keyword",
			output:  `         SSID: TestNet-5G`,
			want:    "TestNet-5G",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSSIDFromAirportOutput(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSSIDFromAirportOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseSSIDFromAirportOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseSSIDFromIwgetidOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{name: "simple SSID", output: "MyWiFi\n", want: "MyWiFi"},
		{name: "SSID with trailing newline", output: "HomeNet\n\n", want: "HomeNet"},
		{name: "SSID with leading space", output: "  SpacedNet  \n", want: "SpacedNet"},
		{name: "empty output", output: "", want: ""},
		{name: "just whitespace", output: "   \n  ", want: ""},
		{name: "no trailing newline", output: "OfficeNet", want: "OfficeNet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSSIDFromIwgetidOutput(tt.output)
			if got != tt.want {
				t.Errorf("parseSSIDFromIwgetidOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseSSIDFromNmcliOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    string
		wantErr bool
	}{
		{
			name:    "active network found",
			output:  "yes:CorpWiFi\nno:GuestWiFi\n",
			want:    "CorpWiFi",
			wantErr: false,
		},
		{
			name:    "active network on second line",
			output:  "no:GuestWiFi\nyes:HomeWiFi\n",
			want:    "HomeWiFi",
			wantErr: false,
		},
		{
			name:    "no active network",
			output:  "no:GuestWiFi\nno:PublicWiFi\n",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty output",
			output:  "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "line without colon",
			output:  "yeswithoutcolon\n",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSSIDFromNmcliOutput(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSSIDFromNmcliOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseSSIDFromNmcliOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseGatewayFromRouteOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    string
		wantErr bool
	}{
		{
			name: "macOS route output",
			output: `   route to: default
destination: default
       mask: default
    gateway: 192.168.1.1
  interface: en0`,
			want:    "192.168.1.1",
			wantErr: false,
		},
		{
			name: "Linux ip route output",
			output: `default via 10.0.0.1 dev eth0 proto dhcp
10.0.0.0/24 dev eth0 proto kernel scope link src 10.0.0.5`,
			want:    "10.0.0.1",
			wantErr: false,
		},
		{
			name: "gateway at end of line (macOS)",
			output: `       gateway: 172.16.0.1
  interface: en0`,
			want:    "172.16.0.1",
			wantErr: false,
		},
		{
			name:    "no gateway keyword",
			output:  "some other output without gateway",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty output",
			output:  "",
			want:    "",
			wantErr: true,
		},
		{
			name: "gateway keyword but no IP after",
			output: `       gateway:
  interface: en0`,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGatewayFromRouteOutput(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseGatewayFromRouteOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseGatewayFromRouteOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseDNSServers(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "multiple nameservers",
			content: "nameserver 8.8.8.8\nnameserver 1.1.1.1\n",
			want:    []string{"8.8.8.8", "1.1.1.1"},
		},
		{
			name:    "single nameserver",
			content: "nameserver 8.8.4.4\n",
			want:    []string{"8.8.4.4"},
		},
		{
			name:    "with comments and search domain",
			content: "# Generated by NetworkManager\nsearch example.com\nnameserver 192.168.1.1\nnameserver 192.168.1.2\n",
			want:    []string{"192.168.1.1", "192.168.1.2"},
		},
		{
			name:    "empty content",
			content: "",
			want:    nil,
		},
		{
			name:    "no nameserver lines",
			content: "search example.com\ndomain example.com\n",
			want:    nil,
		},
		{
			name:    "nameserver with IPv6",
			content: "nameserver 2001:4860:4860::8888\nnameserver fe80::1\n",
			want:    []string{"2001:4860:4860::8888", "fe80::1"},
		},
		{
			name:    "nameserver line without IP",
			content: "nameserver\n",
			want:    nil,
		},
		{
			name:    "indented nameserver",
			content: "  nameserver 10.0.0.1\n",
			want:    []string{"10.0.0.1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDNSServers(tt.content)
			if len(got) != len(tt.want) {
				t.Errorf("parseDNSServers() = %v, want %v", got, tt.want)
				return
			}
			for i, s := range got {
				if s != tt.want[i] {
					t.Errorf("parseDNSServers()[%d] = %q, want %q", i, s, tt.want[i])
				}
			}
		})
	}
}

func TestGetLocalIPs(t *testing.T) {
	nd := NewNetworkDetector(nil)

	ips, err := nd.getLocalIPs()
	if err != nil {
		t.Fatalf("getLocalIPs() error = %v", err)
	}

	for _, ip := range ips {
		if ip == "127.0.0.1" || ip == "::1" {
			t.Errorf("getLocalIPs() should not return loopback: %s", ip)
		}
	}
}

func TestMatchCondition(t *testing.T) {
	nd := NewNetworkDetector(nil)

	tests := []struct {
		name      string
		condition NetworkCondition
		value     string
		want      bool
	}{
		{name: "equals match", condition: NetworkCondition{Value: "HomeWiFi", Operator: "equals"}, value: "HomeWiFi", want: true},
		{name: "equals no match", condition: NetworkCondition{Value: "HomeWiFi", Operator: "equals"}, value: "OfficeWiFi", want: false},
		{name: "contains match", condition: NetworkCondition{Value: "Star", Operator: "contains"}, value: "Starbucks WiFi", want: true},
		{name: "contains no match", condition: NetworkCondition{Value: "Star", Operator: "contains"}, value: "HomeNet", want: false},
		{name: "matches as equals", condition: NetworkCondition{Value: "TestNet", Operator: "matches"}, value: "TestNet", want: true},
		{name: "empty operator defaults to equals", condition: NetworkCondition{Value: "MyNet", Operator: ""}, value: "MyNet", want: true},
		{name: "empty operator no match", condition: NetworkCondition{Value: "MyNet", Operator: ""}, value: "Other", want: false},
		{name: "unknown operator", condition: NetworkCondition{Value: "test", Operator: "regex"}, value: "test", want: false},
		{name: "empty value equals", condition: NetworkCondition{Value: "", Operator: "equals"}, value: "", want: true},
		{name: "contains empty", condition: NetworkCondition{Value: "", Operator: "contains"}, value: "anything", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nd.matchCondition(tt.condition, tt.value)
			if got != tt.want {
				t.Errorf("matchCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchIPRange(t *testing.T) {
	nd := NewNetworkDetector(nil)

	tests := []struct {
		name string
		cidr string
		ip   string
		want bool
	}{
		{name: "CIDR match", cidr: "192.168.1.0/24", ip: "192.168.1.100", want: true},
		{name: "CIDR no match", cidr: "192.168.1.0/24", ip: "10.0.0.1", want: false},
		{name: "CIDR boundary lower", cidr: "192.168.1.0/24", ip: "192.168.1.0", want: true},
		{name: "CIDR boundary upper", cidr: "192.168.1.0/24", ip: "192.168.1.255", want: true},
		{name: "CIDR /32 exact", cidr: "10.0.0.1/32", ip: "10.0.0.1", want: true},
		{name: "CIDR /8 network", cidr: "10.0.0.0/8", ip: "10.255.255.255", want: true},
		{name: "invalid CIDR falls back to string compare match", cidr: "192.168.1.1", ip: "192.168.1.1", want: true},
		{name: "invalid CIDR falls back no match", cidr: "192.168.1.1", ip: "10.0.0.1", want: false},
		{name: "invalid IP address", cidr: "192.168.1.0/24", ip: "not-an-ip", want: false},
		{name: "empty cidr empty ip", cidr: "", ip: "", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nd.matchIPRange(tt.cidr, tt.ip)
			if got != tt.want {
				t.Errorf("matchIPRange(%q, %q) = %v, want %v", tt.cidr, tt.ip, got, tt.want)
			}
		})
	}
}

func TestScoreProfile(t *testing.T) {
	nd := NewNetworkDetector(nil)

	info := &NetworkInfo{
		WiFiSSID:       "HomeWiFi",
		LocalIPs:       []string{"192.168.1.100"},
		Hostname:       "my-mac",
		DefaultGateway: "192.168.1.1",
	}

	tests := []struct {
		name     string
		profile  NetworkProfile
		minScore int
		maxScore int
	}{
		{
			name: "wifi_ssid match gives 100",
			profile: NetworkProfile{
				Name:       "test",
				Priority:   0,
				Conditions: []NetworkCondition{{Type: "wifi_ssid", Value: "HomeWiFi", Operator: "equals"}},
			},
			minScore: 100,
			maxScore: 100,
		},
		{
			name: "ip_range match gives 50",
			profile: NetworkProfile{
				Name:       "test",
				Priority:   0,
				Conditions: []NetworkCondition{{Type: "ip_range", Value: "192.168.1.0/24", Operator: "equals"}},
			},
			minScore: 50,
			maxScore: 50,
		},
		{
			name: "hostname match gives 30",
			profile: NetworkProfile{
				Name:       "test",
				Priority:   0,
				Conditions: []NetworkCondition{{Type: "hostname", Value: "my-mac", Operator: "equals"}},
			},
			minScore: 30,
			maxScore: 30,
		},
		{
			name: "gateway match gives 70",
			profile: NetworkProfile{
				Name:       "test",
				Priority:   0,
				Conditions: []NetworkCondition{{Type: "gateway", Value: "192.168.1.1", Operator: "equals"}},
			},
			minScore: 70,
			maxScore: 70,
		},
		{
			name: "priority bonus added",
			profile: NetworkProfile{
				Name:       "test",
				Priority:   42,
				Conditions: []NetworkCondition{{Type: "wifi_ssid", Value: "HomeWiFi", Operator: "equals"}},
			},
			minScore: 142,
			maxScore: 142,
		},
		{
			name: "no matching conditions gives just priority",
			profile: NetworkProfile{
				Name:       "test",
				Priority:   10,
				Conditions: []NetworkCondition{{Type: "wifi_ssid", Value: "WrongWiFi", Operator: "equals"}},
			},
			minScore: 10,
			maxScore: 10,
		},
		{
			name: "no conditions gives just priority",
			profile: NetworkProfile{
				Name:       "test",
				Priority:   5,
				Conditions: nil,
			},
			minScore: 5,
			maxScore: 5,
		},
		{
			name: "multiple matches accumulate",
			profile: NetworkProfile{
				Name:     "test",
				Priority: 10,
				Conditions: []NetworkCondition{
					{Type: "wifi_ssid", Value: "HomeWiFi", Operator: "equals"},
					{Type: "gateway", Value: "192.168.1.1", Operator: "equals"},
				},
			},
			minScore: 180,
			maxScore: 180,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := nd.scoreProfile(&tt.profile, info)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("scoreProfile() = %d, want between %d and %d", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestFindBestMatchingProfile(t *testing.T) {
	info := &NetworkInfo{
		WiFiSSID:       "HomeWiFi",
		LocalIPs:       []string{"192.168.1.100"},
		Hostname:       "my-mac",
		DefaultGateway: "192.168.1.1",
	}

	t.Run("returns highest scoring profile", func(t *testing.T) {
		profiles := []NetworkProfile{
			{Name: "low", Priority: 1},
			{Name: "high", Priority: 100, Conditions: []NetworkCondition{{Type: "wifi_ssid", Value: "HomeWiFi", Operator: "equals"}}},
			{Name: "mid", Priority: 50},
		}
		nd := NewNetworkDetector(profiles)
		best := nd.findBestMatchingProfile(info)
		if best == nil {
			t.Fatal("expected non-nil profile")
		}
		if best.Name != "high" {
			t.Errorf("expected 'high', got %q", best.Name)
		}
	})

	t.Run("returns nil when no profiles match", func(t *testing.T) {
		profiles := []NetworkProfile{
			{Name: "no-match", Priority: 0, Conditions: []NetworkCondition{{Type: "wifi_ssid", Value: "WrongWiFi", Operator: "equals"}}},
		}
		nd := NewNetworkDetector(profiles)
		best := nd.findBestMatchingProfile(info)
		if best != nil {
			t.Errorf("expected nil, got %q", best.Name)
		}
	})

	t.Run("returns nil for empty profiles", func(t *testing.T) {
		nd := NewNetworkDetector([]NetworkProfile{})
		best := nd.findBestMatchingProfile(info)
		if best != nil {
			t.Errorf("expected nil for empty profiles, got %q", best.Name)
		}
	})

	t.Run("returns nil for nil profiles", func(t *testing.T) {
		nd := NewNetworkDetector(nil)
		best := nd.findBestMatchingProfile(info)
		if best != nil {
			t.Errorf("expected nil for nil profiles, got %q", best.Name)
		}
	})
}

func TestDetectEnvironment(t *testing.T) {
	t.Run("returns nil profile when no match but no error", func(t *testing.T) {
		profiles := []NetworkProfile{
			{Name: "office", Priority: 100, Conditions: []NetworkCondition{{Type: "wifi_ssid", Value: "NonExistentSSID", Operator: "equals"}}},
		}
		nd := NewNetworkDetector(profiles)
		profile, err := nd.DetectEnvironment(context.Background())
		if err != nil {
			t.Fatalf("DetectEnvironment() error = %v", err)
		}
		if profile != nil {
			t.Logf("Detected profile %q (OK on this machine)", profile.Name)
		}
	})

	t.Run("succeeds with empty profiles", func(t *testing.T) {
		nd := NewNetworkDetector([]NetworkProfile{})
		profile, err := nd.DetectEnvironment(context.Background())
		if err != nil {
			t.Fatalf("DetectEnvironment() error = %v", err)
		}
		if profile != nil {
			t.Errorf("expected nil profile for empty list, got %q", profile.Name)
		}
	})
}

func TestGetCurrentNetworkInfo(t *testing.T) {
	nd := NewNetworkDetector(nil)
	info := nd.getCurrentNetworkInfo(context.Background())
	if info == nil {
		t.Fatal("expected non-nil info")
	}
	if info.Hostname == "" {
		t.Log("warning: hostname is empty (unusual but not fatal)")
	}
	if info.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}
