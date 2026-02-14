// wap.go — WAP/WWD level file parser.
//
// Captain Claw levels are stored in WAP (World Wide Data) format with
// the .WWD extension. Each file describes the tile layout, object
// placements, tile properties, and level metadata for a single level.
//
// WAP format overview:
//   - Header: magic, version, level name, tileset references, dimensions
//   - Tile descriptions: per-tile properties (solid, platform, etc.)
//   - Planes: multiple layers (background, action, foreground)
//   - Objects: enemy spawns, items, triggers, checkpoints
package assets

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// WAP file magic bytes
var wapMagic = [4]byte{'W', 'W', 'D', '\x00'}

// WapHeader contains the parsed header of a WAP level file.
type WapHeader struct {
	Version    uint32
	LevelName  string
	TileWidth  uint32
	TileHeight uint32
	PlaneCount uint32
}

// WapPlane represents a single tile plane (layer) in the level.
type WapPlane struct {
	Name          string
	Width, Height uint32
	TileData      []uint32 // flat array of tile indices, row-major
}

// WapObject represents a placed object in the level.
type WapObject struct {
	Name   string
	X, Y   int32
	Width  int32
	Height int32
	Type   string
	// TODO: add logic/param fields
}

// WapLevel holds all parsed data from a WAP/WWD file.
type WapLevel struct {
	Header  WapHeader
	Planes  []WapPlane
	Objects []WapObject
}

// LoadWap parses a WAP/WWD level file from disk.
func LoadWap(path string) (*WapLevel, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("wap: cannot open %s: %w", path, err)
	}
	defer f.Close()

	// Read and validate magic
	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		return nil, fmt.Errorf("wap: failed to read header: %w", err)
	}
	if magic != wapMagic {
		return nil, fmt.Errorf("wap: invalid magic bytes in %s", path)
	}

	var header WapHeader
	binary.Read(f, binary.LittleEndian, &header.Version)

	// TODO: parse full header, planes, tile data, and objects
	// This is a placeholder that returns an empty level structure

	return &WapLevel{
		Header: header,
	}, nil
}
