// camera.go — Camera/viewport that follows the player.
//
// The camera smoothly tracks the player character and clamps to level
// boundaries so the viewport never shows out-of-bounds areas.
package game

// Camera represents the viewport into the game world.
type Camera struct {
	X, Y          float64 // top-left corner in world coordinates
	Width, Height float64 // viewport size in pixels
	Smoothing     float64 // lerp factor (0..1); 1 = instant snap
}

// NewCamera creates a camera with the given viewport size.
func NewCamera(w, h float64) *Camera {
	return &Camera{
		Width:     w,
		Height:    h,
		Smoothing: 0.1,
	}
}

// Follow updates the camera position to track a target (typically the player).
// The target is centered in the viewport, clamped to level bounds.
func (c *Camera) Follow(targetX, targetY, levelW, levelH float64) {
	// Desired position: center target in viewport
	desiredX := targetX - c.Width/2
	desiredY := targetY - c.Height/2

	// Smooth interpolation
	c.X += (desiredX - c.X) * c.Smoothing
	c.Y += (desiredY - c.Y) * c.Smoothing

	// Clamp to level boundaries
	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	maxX := levelW - c.Width
	maxY := levelH - c.Height
	if maxX > 0 && c.X > maxX {
		c.X = maxX
	}
	if maxY > 0 && c.Y > maxY {
		c.Y = maxY
	}
}
