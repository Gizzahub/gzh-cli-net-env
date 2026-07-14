// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"path/filepath"
	"testing"

	"github.com/gizzahub/gzh-cli-core/config"
)

func TestGetProfilesPath_Override(t *testing.T) {
	customDir := "/custom/netenv/path"
	t.Setenv("GZH_CONFIG_DIR", customDir)

	path := GetProfilesPath()
	expected := filepath.Join(customDir, "network-profiles.yaml")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestGetProfilesPath_Default(t *testing.T) {
	t.Setenv("GZH_CONFIG_DIR", "")

	path := GetProfilesPath()
	coreDir := config.GetConfigDirectory()
	expected := filepath.Join(coreDir, "network-profiles.yaml")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestCoreConfigDirectoryIntegration(t *testing.T) {
	customDir := "/tmp/netenv-integration-test"
	t.Setenv("GZH_CONFIG_DIR", customDir)

	dir := config.GetConfigDirectory()
	if dir != customDir {
		t.Errorf("core config dir: expected %s, got %s", customDir, dir)
	}
}
