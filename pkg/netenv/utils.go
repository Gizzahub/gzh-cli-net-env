// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package netenv

import (
	"path/filepath"

	"github.com/gizzahub/gzh-cli-core/config"
)

// GetProfilesPath returns the path to the network profiles configuration file.
func GetProfilesPath() string {
	return filepath.Join(config.GetConfigDirectory(), "network-profiles.yaml")
}
