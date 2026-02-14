// Package ui provides menus and HUD rendering for GoClaw.
//
// hud.go — In-game heads-up display showing score, health, lives,
// and ammo counts (pistol, magic, dynamite).
package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// HUD renders the in-game status overlay.
type HUD struct {
	// No state needed yet; reads from game state each frame.
}

// NewHUD creates a new HUD instance.
func NewHUD() *HUD {
	return &HUD{}
}

// Draw renders the HUD onto the screen.
func (h *HUD) Draw(screen *ebiten.Image, score, health, lives, pistol, magic, dynamite int) {
	// Top-left: score and lives
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Score: %d    Lives: %d", score, lives),
		10, 10,
	)

	// Below that: health bar
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Health: %d/100", health),
		10, 28,
	)

	// Ammo counts
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Pistol: %d  Magic: %d  Dynamite: %d", pistol, magic, dynamite),
		10, 46,
	)

	// TODO: replace debug text with proper sprite-based HUD graphics
}
