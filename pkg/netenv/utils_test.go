// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gizzahub/gzh-cli-core/config"
)

// TestGetProfilesPath_Override verifies that GZH_CONFIG_DIR is respected
// through the core config.GetConfigDirectory() call chain.
func TestGetProfilesPath_Override(t *testing.T) {
	customDir := "/custom/netenv/path"
	os.Setenv("GZH_CONFIG_DIR", customDir)
	defer os.Unsetenv("GZH_CONFIG_DIR")

	path := GetProfilesPath()
	expected := filepath.Join(customDir, "network-profiles.yaml")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

// TestGetProfilesPath_Default verifies the fallback to ~/.config/gzh-manager
// when GZH_CONFIG_DIR is not set.
func TestGetProfilesPath_Default(t *testing.T) {
	os.Unsetenv("GZH_CONFIG_DIR")

	path := GetProfilesPath()
	coreDir := config.GetConfigDirectory()
	expected := filepath.Join(coreDir, "network-profiles.yaml")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

// TestCoreConfigDirectoryIntegration is a smoke test verifying that net-env
// correctly delegates to core's GetConfigDirectory.
func TestCoreConfigDirectoryIntegration(t *testing.T) {
	customDir := "/tmp/netenv-integration-test"
	os.Setenv("GZH_CONFIG_DIR", customDir)
	defer os.Unsetenv("GZH_CONFIG_DIR")

	dir := config.GetConfigDirectory()
	if dir != customDir {
		t.Errorf("core config dir: expected %s, got %s", customDir, dir)
	}
}
