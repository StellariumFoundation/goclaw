// Package engine provides the core game engine: game loop, window management,
// input handling, and rendering.
//
// window.go — Window creation and management using Ebitengine.
package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// WindowConfig holds configuration for the game window.
type WindowConfig struct {
	Width  int
	Height int
	Title  string
	VSync  bool
}

// DefaultWindowConfig returns sensible defaults for the game window.
func DefaultWindowConfig() WindowConfig {
	return WindowConfig{
		Width:  800,
		Height: 600,
		Title:  "GoClaw - Captain Claw Reimplementation",
		VSync:  true,
	}
}

// ApplyWindowConfig applies the given configuration to the Ebitengine window.
func ApplyWindowConfig(cfg WindowConfig) {
	ebiten.SetWindowSize(cfg.Width, cfg.Height)
	ebiten.SetWindowTitle(cfg.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if cfg.VSync {
		ebiten.SetVsyncEnabled(true)
	}
}
