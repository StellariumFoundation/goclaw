// Package assets handles loading and caching of game assets.
//
// loader.go — Functions to load sprites, sounds, and level data from
// the original Captain Claw game assets (REZ archives and WAP level files).
package assets

import (
	"fmt"
	"os"
)

// AssetManager caches loaded assets to avoid redundant disk reads.
type AssetManager struct {
	basePath string
	// TODO: add caches for images, sounds, level data
}

// NewAssetManager creates an asset manager rooted at the given base path.
// The base path should point to the directory containing the original
// Captain Claw game files (CLAW.REZ, etc.).
func NewAssetManager(basePath string) *AssetManager {
	return &AssetManager{basePath: basePath}
}

// LoadFile reads a raw file from the asset base path.
func (am *AssetManager) LoadFile(relPath string) ([]byte, error) {
	fullPath := am.basePath + "/" + relPath
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("assets: failed to load %s: %w", relPath, err)
	}
	return data, nil
}

// TODO: add methods to load sprites (from PID/PCX), sounds (WAV), and
// level data (WAP/WWD) using the rez and wap parsers.
