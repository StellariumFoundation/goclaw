// rez.go — REZ archive parser.
//
// Captain Claw stores all game assets in a proprietary REZ archive format
// (typically CLAW.REZ). This file implements reading the archive directory
// and extracting individual files by path.
//
// REZ format overview:
//   - Header: magic bytes, version, directory offset
//   - Directory: tree of folders and file entries with offsets/sizes
//   - File data: raw bytes at the specified offsets
package assets

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// REZ archive magic bytes
var rezMagic = [4]byte{'R', 'E', 'Z', '\x00'}

// RezArchive represents an opened REZ archive.
type RezArchive struct {
	file    *os.File
	entries map[string]rezEntry
}

// rezEntry is a single file entry in the archive.
type rezEntry struct {
	Offset uint32
	Size   uint32
}

// OpenRez opens a REZ archive file and reads its directory.
func OpenRez(path string) (*RezArchive, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("rez: cannot open %s: %w", path, err)
	}

	// Read and validate magic bytes
	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		f.Close()
		return nil, fmt.Errorf("rez: failed to read header: %w", err)
	}
	if magic != rezMagic {
		f.Close()
		return nil, fmt.Errorf("rez: invalid magic bytes in %s", path)
	}

	// Read version and directory offset
	var version uint32
	var dirOffset uint32
	binary.Read(f, binary.LittleEndian, &version)
	binary.Read(f, binary.LittleEndian, &dirOffset)

	// TODO: parse the full directory tree from dirOffset
	// For now, return an empty archive structure
	return &RezArchive{
		file:    f,
		entries: make(map[string]rezEntry),
	}, nil
}

// Extract reads a file from the archive by its internal path.
func (r *RezArchive) Extract(name string) ([]byte, error) {
	entry, ok := r.entries[name]
	if !ok {
		return nil, fmt.Errorf("rez: file not found: %s", name)
	}

	buf := make([]byte, entry.Size)
	if _, err := r.file.ReadAt(buf, int64(entry.Offset)); err != nil {
		return nil, fmt.Errorf("rez: failed to read %s: %w", name, err)
	}
	return buf, nil
}

// Close releases the archive file handle.
func (r *RezArchive) Close() error {
	return r.file.Close()
}

// List returns all file paths in the archive.
func (r *RezArchive) List() []string {
	paths := make([]string, 0, len(r.entries))
	for p := range r.entries {
		paths = append(paths, p)
	}
	return paths
}
