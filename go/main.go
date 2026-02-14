// GoClaw — A Go rewrite of OpenClaw / Captain Claw
// Entry point: initializes the game window and starts the game loop.
package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 800
	screenHeight = 600
	gameTitle    = "GoClaw - Captain Claw Reimplementation"
)

// Game implements ebiten.Game interface and holds the top-level game state.
type Game struct{}

// Update advances the game state by one tick (called every frame).
func (g *Game) Update() error {
	// Quit on Escape key
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}
	return nil
}

// Draw renders the current frame.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 20, G: 12, B: 28, A: 255}) // dark purple background

	msg := "GoClaw — Coming Soon"
	// Draw centered placeholder text using the debug utility
	ebitenutil.DebugPrintAt(screen, msg, screenWidth/2-len(msg)*3, screenHeight/2-8)

	// Additional info
	info := fmt.Sprintf("Captain Claw Reimplementation in Go\n\nPress ESC to quit\n\nResolution: %dx%d", screenWidth, screenHeight)
	ebitenutil.DebugPrintAt(screen, info, screenWidth/2-120, screenHeight/2+20)
}

// Layout returns the logical screen dimensions.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle(gameTitle)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
