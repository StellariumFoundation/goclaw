// input.go — Keyboard and mouse input handling.
//
// Wraps Ebitengine's input API to provide a clean interface for the game logic
// layer. Tracks key states (pressed, just pressed, just released) and mouse
// position/buttons per frame.
package engine

import "github.com/hajimehoshi/ebiten/v2"

// InputState captures the current frame's input snapshot.
type InputState struct {
	// Keyboard
	Left, Right, Up, Down bool
	Jump                  bool
	Attack                bool
	Pause                 bool

	// Mouse
	MouseX, MouseY int
	MousePressed   bool
}

// PollInput reads the current input devices and returns an InputState.
func PollInput() InputState {
	mx, my := ebiten.CursorPosition()
	return InputState{
		Left:         ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA),
		Right:        ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD),
		Up:           ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW),
		Down:         ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS),
		Jump:         ebiten.IsKeyPressed(ebiten.KeySpace),
		Attack:       ebiten.IsKeyPressed(ebiten.KeyZ) || ebiten.IsKeyPressed(ebiten.KeyControl),
		Pause:        ebiten.IsKeyPressed(ebiten.KeyEscape),
		MouseX:       mx,
		MouseY:       my,
		MousePressed: ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
	}
}
