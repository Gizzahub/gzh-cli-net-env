// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package netenv

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestProfileManager(t *testing.T) *ProfileManager {
	t.Helper()
	return NewProfileManager(t.TempDir())
}

func makeValidProfile(name string) *NetworkProfile {
	return &NetworkProfile{
		Name:        name,
		Description: "test profile",
		Priority:    50,
		Conditions: []NetworkCondition{
			{Type: "wifi_ssid", Value: "TestWiFi", Operator: "equals"},
		},
		Components: NetworkComponents{
			DNS: &DNSConfig{Servers: []string{"8.8.8.8"}},
		},
	}
}

func TestNewProfileManager(t *testing.T) {
	pm := NewProfileManager("/tmp/test")
	if pm == nil {
		t.Fatal("expected non-nil ProfileManager")
	}
	if len(pm.ListProfiles()) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(pm.ListProfiles()))
	}
}

func TestSaveProfile_NewAndLoad(t *testing.T) {
	pm := setupTestProfileManager(t)
	profile := makeValidProfile("test-net")

	err := pm.SaveProfile(profile)
	if err != nil {
		t.Fatalf("SaveProfile error: %v", err)
	}

	got, err := pm.GetProfile("test-net")
	if err != nil {
		t.Fatalf("GetProfile error: %v", err)
	}
	if got.Name != "test-net" {
		t.Errorf("name = %q, want %q", got.Name, "test-net")
	}
	if got.Description != "test profile" {
		t.Errorf("description = %q", got.Description)
	}
	if got.Priority != 50 {
		t.Errorf("priority = %d", got.Priority)
	}
	if len(got.Conditions) != 1 {
		t.Errorf("conditions len = %d", len(got.Conditions))
	}
	if got.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if got.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestSaveProfile_UpdateExisting(t *testing.T) {
	pm := setupTestProfileManager(t)
	profile := makeValidProfile("update-test")
	originalTime := time.Now().Add(-1 * time.Hour)
	profile.CreatedAt = originalTime

	_ = pm.SaveProfile(profile)
	time.Sleep(10 * time.Millisecond)

	profile.Description = "updated description"
	_ = pm.SaveProfile(profile)

	got, _ := pm.GetProfile("update-test")
	if got.Description != "updated description" {
		t.Errorf("description = %q, want 'updated description'", got.Description)
	}
	if !got.CreatedAt.Equal(originalTime) {
		t.Error("CreatedAt should not change on update")
	}
	if !got.UpdatedAt.After(got.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

func TestSaveProfile_InvalidName(t *testing.T) {
	pm := setupTestProfileManager(t)

	tests := []struct {
		name    string
		profile *NetworkProfile
	}{
		{name: "empty name", profile: &NetworkProfile{Name: ""}},
		{name: "spaces in name", profile: &NetworkProfile{Name: "has spaces"}},
		{name: "too long", profile: &NetworkProfile{Name: string(make([]byte, 65))}},
		{name: "special chars", profile: &NetworkProfile{Name: "test@home"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.SaveProfile(tt.profile)
			if err == nil {
				t.Error("expected error for invalid name")
			}
		})
	}
}

func TestSaveProfile_InvalidConditions(t *testing.T) {
	pm := setupTestProfileManager(t)

	tests := []struct {
		name    string
		profile *NetworkProfile
	}{
		{name: "missing condition type", profile: &NetworkProfile{Name: "test", Conditions: []NetworkCondition{{Value: "x"}}}},
		{name: "missing condition value", profile: &NetworkProfile{Name: "test", Conditions: []NetworkCondition{{Type: "wifi_ssid"}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.SaveProfile(tt.profile)
			if err == nil {
				t.Error("expected error for invalid condition")
			}
		})
	}
}

func TestLoadProfiles_EmptyDir(t *testing.T) {
	pm := setupTestProfileManager(t)
	err := pm.LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles error: %v", err)
	}
	if len(pm.ListProfiles()) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(pm.ListProfiles()))
	}
}

func TestLoadProfiles_MultipleProfiles(t *testing.T) {
	pm := setupTestProfileManager(t)

	_ = pm.SaveProfile(makeValidProfile("alpha"))
	_ = pm.SaveProfile(makeValidProfile("beta"))
	_ = pm.SaveProfile(makeValidProfile("gamma"))

	pm2 := NewProfileManager(pm.configDir)
	err := pm2.LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles error: %v", err)
	}

	profiles := pm2.ListProfiles()
	if len(profiles) != 3 {
		t.Errorf("expected 3 profiles, got %d", len(profiles))
	}
}

func TestLoadProfiles_MalformedYAML(t *testing.T) {
	pm := setupTestProfileManager(t)

	_ = pm.SaveProfile(makeValidProfile("good"))

	profilesDir := filepath.Join(pm.configDir, "net-env", "profiles")
	_ = os.WriteFile(filepath.Join(profilesDir, "bad.yaml"), []byte("invalid: yaml: ["), 0o600)

	err := pm.LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles should not fail on one bad file: %v", err)
	}

	if _, err := pm.GetProfile("good"); err != nil {
		t.Errorf("good profile should still be loaded: %v", err)
	}
}

func TestLoadProfiles_NonYAMLIgnored(t *testing.T) {
	pm := setupTestProfileManager(t)

	_ = pm.SaveProfile(makeValidProfile("real"))

	profilesDir := filepath.Join(pm.configDir, "net-env", "profiles")
	_ = os.WriteFile(filepath.Join(profilesDir, "readme.txt"), []byte("not yaml"), 0o600)
	_ = os.WriteFile(filepath.Join(profilesDir, "data.json"), []byte(`{"name":"json"}`), 0o600)

	err := pm.LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles error: %v", err)
	}

	if len(pm.ListProfiles()) != 1 {
		t.Errorf("expected 1 profile, got %d", len(pm.ListProfiles()))
	}
}

func TestGetProfile_Found(t *testing.T) {
	pm := setupTestProfileManager(t)
	_ = pm.SaveProfile(makeValidProfile("found"))

	got, err := pm.GetProfile("found")
	if err != nil {
		t.Fatalf("GetProfile error: %v", err)
	}
	if got.Name != "found" {
		t.Errorf("name = %q", got.Name)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	pm := setupTestProfileManager(t)
	_, err := pm.GetProfile("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent profile")
	}
}

func TestListProfiles_SortedByPriority(t *testing.T) {
	pm := setupTestProfileManager(t)

	low := makeValidProfile("low")
	low.Priority = 10
	_ = pm.SaveProfile(low)

	high := makeValidProfile("high")
	high.Priority = 100
	_ = pm.SaveProfile(high)

	mid := makeValidProfile("mid")
	mid.Priority = 50
	_ = pm.SaveProfile(mid)

	profiles := pm.ListProfiles()
	if profiles[0].Name != "high" {
		t.Errorf("expected 'high' first, got %q", profiles[0].Name)
	}
	if profiles[1].Name != "mid" {
		t.Errorf("expected 'mid' second, got %q", profiles[1].Name)
	}
	if profiles[2].Name != "low" {
		t.Errorf("expected 'low' third, got %q", profiles[2].Name)
	}
}

func TestListProfiles_SortedByName(t *testing.T) {
	pm := setupTestProfileManager(t)

	for _, name := range []string{"charlie", "alpha", "bravo"} {
		p := makeValidProfile(name)
		p.Priority = 50
		_ = pm.SaveProfile(p)
	}

	profiles := pm.ListProfiles()
	if profiles[0].Name != "alpha" {
		t.Errorf("expected 'alpha' first, got %q", profiles[0].Name)
	}
	if profiles[1].Name != "bravo" {
		t.Errorf("expected 'bravo' second, got %q", profiles[1].Name)
	}
	if profiles[2].Name != "charlie" {
		t.Errorf("expected 'charlie' third, got %q", profiles[2].Name)
	}
}

func TestListProfiles_Empty(t *testing.T) {
	pm := setupTestProfileManager(t)
	profiles := pm.ListProfiles()
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}
}

func TestDeleteProfile_Existing(t *testing.T) {
	pm := setupTestProfileManager(t)
	_ = pm.SaveProfile(makeValidProfile("delete-me"))

	err := pm.DeleteProfile("delete-me")
	if err != nil {
		t.Fatalf("DeleteProfile error: %v", err)
	}

	_, err = pm.GetProfile("delete-me")
	if err == nil {
		t.Error("expected error after delete")
	}

	profilePath := filepath.Join(pm.configDir, "net-env", "profiles", "delete-me.yaml")
	if _, err := os.Stat(profilePath); !os.IsNotExist(err) {
		t.Error("expected file to be deleted")
	}
}

func TestDeleteProfile_NotFound(t *testing.T) {
	pm := setupTestProfileManager(t)
	err := pm.DeleteProfile("nonexistent")
	if err == nil {
		t.Error("expected error for deleting non-existent profile")
	}
}

func TestExportProfile(t *testing.T) {
	pm := setupTestProfileManager(t)
	_ = pm.SaveProfile(makeValidProfile("export-test"))

	exportPath := filepath.Join(t.TempDir(), "exported.yaml")
	err := pm.ExportProfile("export-test", exportPath)
	if err != nil {
		t.Fatalf("ExportProfile error: %v", err)
	}

	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("export file should exist")
	}
}

func TestExportProfile_NotFound(t *testing.T) {
	pm := setupTestProfileManager(t)
	err := pm.ExportProfile("nonexistent", "/tmp/out.yaml")
	if err == nil {
		t.Error("expected error for exporting non-existent profile")
	}
}

func TestImportProfile(t *testing.T) {
	pm := setupTestProfileManager(t)
	_ = pm.SaveProfile(makeValidProfile("source"))

	exportPath := filepath.Join(t.TempDir(), "import-source.yaml")
	_ = pm.ExportProfile("source", exportPath)

	pm2 := setupTestProfileManager(t)
	err := pm2.ImportProfile(exportPath)
	if err != nil {
		t.Fatalf("ImportProfile error: %v", err)
	}

	got, err := pm2.GetProfile("source")
	if err != nil {
		t.Fatalf("GetProfile after import error: %v", err)
	}
	if got.Name != "source" {
		t.Errorf("name = %q", got.Name)
	}
}

func TestImportProfile_Malformed(t *testing.T) {
	pm := setupTestProfileManager(t)

	badPath := filepath.Join(t.TempDir(), "bad.yaml")
	_ = os.WriteFile(badPath, []byte("invalid: yaml: ["), 0o600)

	err := pm.ImportProfile(badPath)
	if err == nil {
		t.Error("expected error for malformed import file")
	}
}

func TestCreateDefaultProfiles(t *testing.T) {
	pm := setupTestProfileManager(t)
	err := pm.CreateDefaultProfiles()
	if err != nil {
		t.Fatalf("CreateDefaultProfiles error: %v", err)
	}

	for _, name := range []string{"home", "office", "cafe"} {
		if _, err := pm.GetProfile(name); err != nil {
			t.Errorf("expected profile %q: %v", name, err)
		}
	}
}

func TestCreateDefaultProfiles_Idempotent(t *testing.T) {
	pm := setupTestProfileManager(t)

	_ = pm.CreateDefaultProfiles()
	_ = pm.CreateDefaultProfiles()

	profiles := pm.ListProfiles()
	if len(profiles) != 3 {
		t.Errorf("expected 3 profiles after double init, got %d", len(profiles))
	}
}

func TestIsValidProfileName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "simple lowercase", input: "home", want: true},
		{name: "with hyphen", input: "home-wifi", want: true},
		{name: "with underscore", input: "home_wifi", want: true},
		{name: "with numbers", input: "net123", want: true},
		{name: "mixed case", input: "HomeWiFi", want: true},
		{name: "empty", input: "", want: false},
		{name: "spaces", input: "has space", want: false},
		{name: "dots", input: "home.wifi", want: false},
		{name: "slashes", input: "home/wifi", want: false},
		{name: "special chars", input: "home@wifi", want: false},
		{name: "at max length 64", input: string(make([]byte, 64)), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == string(make([]byte, 64)) {
				valid := make([]byte, 64)
				for i := range valid {
					valid[i] = 'a'
				}
				got := isValidProfileName(string(valid))
				if got != true {
					t.Errorf("isValidProfileName(64 chars) = %v, want true", got)
				}
				return
			}
			got := isValidProfileName(tt.input)
			if got != tt.want {
				t.Errorf("isValidProfileName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidProfileName_TooLong(t *testing.T) {
	long := make([]byte, 65)
	for i := range long {
		long[i] = 'a'
	}
	if isValidProfileName(string(long)) {
		t.Error("expected false for 65-char name")
	}
}

func TestIsYAMLFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{name: ".yaml", filename: "profile.yaml", want: true},
		{name: ".yml", filename: "profile.yml", want: true},
		{name: ".txt", filename: "readme.txt", want: false},
		{name: ".json", filename: "data.json", want: false},
		{name: "no extension", filename: "profile", want: false},
		{name: "empty", filename: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isYAMLFile(tt.filename)
			if got != tt.want {
				t.Errorf("isYAMLFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestValidateProfile_Valid(t *testing.T) {
	pm := setupTestProfileManager(t)
	err := pm.validateProfile(makeValidProfile("valid"))
	if err != nil {
		t.Errorf("expected no error: %v", err)
	}
}

func TestValidateProfile_EmptyName(t *testing.T) {
	pm := setupTestProfileManager(t)
	err := pm.validateProfile(&NetworkProfile{Name: ""})
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestValidateProfile_InvalidNameChars(t *testing.T) {
	pm := setupTestProfileManager(t)
	tests := []string{"has space", "has.dot", "has/slash", "has@at"}
	for _, name := range tests {
		err := pm.validateProfile(&NetworkProfile{Name: name})
		if err == nil {
			t.Errorf("expected error for name %q", name)
		}
	}
}

func TestValidateProfile_MissingConditionType(t *testing.T) {
	pm := setupTestProfileManager(t)
	profile := &NetworkProfile{
		Name:       "test",
		Conditions: []NetworkCondition{{Value: "some-value"}},
	}
	err := pm.validateProfile(profile)
	if err == nil {
		t.Error("expected error for missing condition type")
	}
}

func TestValidateProfile_MissingConditionValue(t *testing.T) {
	pm := setupTestProfileManager(t)
	profile := &NetworkProfile{
		Name:       "test",
		Conditions: []NetworkCondition{{Type: "wifi_ssid"}},
	}
	err := pm.validateProfile(profile)
	if err == nil {
		t.Error("expected error for missing condition value")
	}
}

func TestGetAutoProfile_NoAutoProfiles(t *testing.T) {
	pm := setupTestProfileManager(t)
	profile := makeValidProfile("manual")
	profile.Auto = false
	_ = pm.SaveProfile(profile)

	_, err := pm.GetAutoProfile()
	if err == nil {
		t.Error("expected error when no auto profiles configured")
	}
}
