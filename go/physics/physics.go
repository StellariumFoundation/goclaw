// Package physics provides the physics subsystem for GoClaw.
//
// physics.go — Gravity, collision detection, and movement resolution.
//
// Captain Claw uses tile-based collision with AABB (axis-aligned bounding
// box) checks. This package handles gravity application, ground detection,
// wall collision, and one-way platform logic.
package physics

// Gravity constant in pixels per second squared.
const Gravity = 800.0

// MaxFallSpeed caps the terminal velocity to prevent tunneling.
const MaxFallSpeed = 600.0

// AABB represents an axis-aligned bounding box for collision detection.
type AABB struct {
	X, Y          float64 // top-left corner
	Width, Height float64
}

// Overlaps checks whether two AABBs overlap.
func (a AABB) Overlaps(b AABB) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}

// Center returns the center point of the AABB.
func (a AABB) Center() (float64, float64) {
	return a.X + a.Width/2, a.Y + a.Height/2
}

// ApplyGravity adds gravitational acceleration to a vertical velocity.
// Returns the new velocity, clamped to MaxFallSpeed.
func ApplyGravity(velY, dt float64) float64 {
	velY += Gravity * dt
	if velY > MaxFallSpeed {
		velY = MaxFallSpeed
	}
	return velY
}

// ResolveCollisionX checks horizontal movement against solid tiles and
// returns the corrected X position.
// TODO: implement full tile-based horizontal collision resolution.
func ResolveCollisionX(x, y, width, height, velX, dt float64, isSolid func(tx, ty int) bool, tileW, tileH int) float64 {
	newX := x + velX*dt
	// placeholder — will check tiles along the movement path
	return newX
}

// ResolveCollisionY checks vertical movement against solid tiles and
// returns the corrected Y position and whether the entity is grounded.
// TODO: implement full tile-based vertical collision resolution.
func ResolveCollisionY(x, y, width, height, velY, dt float64, isSolid func(tx, ty int) bool, tileW, tileH int) (float64, bool) {
	newY := y + velY*dt
	grounded := false
	// placeholder — will check tiles below/above and resolve penetration
	return newY, grounded
}
