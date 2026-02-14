// menu.go — Main menu and pause menu.
//
// Provides the title screen (main menu) with options like New Game,
// Load Game, Options, and Quit, as well as the in-game pause menu.
package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// MenuState represents which menu is currently active.
type MenuState int

const (
	MenuNone  MenuState = iota // no menu shown (gameplay)
	MenuMain                   // title screen
	MenuPause                  // in-game pause
)

// MenuItem represents a selectable menu entry.
type MenuItem struct {
	Label  string
	Action func()
}

// Menu manages menu state and rendering.
type Menu struct {
	State    MenuState
	Items    []MenuItem
	Selected int
}

// NewMainMenu creates the title screen menu.
func NewMainMenu(onNewGame, onQuit func()) *Menu {
	return &Menu{
		State: MenuMain,
		Items: []MenuItem{
			{Label: "New Game", Action: onNewGame},
			{Label: "Quit", Action: onQuit},
		},
	}
}

// NewPauseMenu creates the in-game pause menu.
func NewPauseMenu(onResume, onQuit func()) *Menu {
	return &Menu{
		State: MenuPause,
		Items: []MenuItem{
			{Label: "Resume", Action: onResume},
			{Label: "Quit to Menu", Action: onQuit},
		},
	}
}

// Update handles menu input (up/down navigation, enter to select).
func (m *Menu) Update() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		// TODO: debounce key presses
		if m.Selected > 0 {
			m.Selected--
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		if m.Selected < len(m.Items)-1 {
			m.Selected++
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		if m.Selected >= 0 && m.Selected < len(m.Items) {
			m.Items[m.Selected].Action()
		}
	}
}

// Draw renders the menu onto the screen.
func (m *Menu) Draw(screen *ebiten.Image, screenW, screenH int) {
	startY := screenH/2 - len(m.Items)*20/2
	for i, item := range m.Items {
		prefix := "  "
		if i == m.Selected {
			prefix = "> "
		}
		ebitenutil.DebugPrintAt(screen, prefix+item.Label, screenW/2-60, startY+i*20)
	}
	// TODO: replace debug text with proper menu graphics
}
