// Package game contains the core game logic for GoClaw.
//
// player.go — Player character (Captain Claw) state, movement, physics,
// and animation management.
package game

// PlayerState enumerates the possible states of the player character.
type PlayerState int

const (
	PlayerIdle PlayerState = iota
	PlayerRunning
	PlayerJumping
	PlayerFalling
	PlayerAttacking
	PlayerClimbing
	PlayerDying
)

// Player represents Captain Claw — the player-controlled character.
type Player struct {
	X, Y       float64 // world position
	VelX, VelY float64 // velocity
	Width      float64 // collision box width
	Height     float64 // collision box height
	State      PlayerState
	FacingLeft bool

	// Stats
	Health    int
	Lives     int
	Score     int
	Pistol    int // pistol ammo
	Magic     int // magic ammo
	Dynamite  int // dynamite ammo

	// Animation
	AnimFrame int
	AnimTimer float64
}

// NewPlayer creates a player with default starting values.
func NewPlayer(x, y float64) *Player {
	return &Player{
		X:      x,
		Y:      y,
		Width:  32,
		Height: 64,
		State:  PlayerIdle,
		Health: 100,
		Lives:  6,
		Pistol: 25,
		Magic:  5,
	}
}

// Update advances the player state by one tick.
// TODO: implement full movement physics, animation cycling, and state transitions.
func (p *Player) Update(dt float64, left, right, jump, attack bool) {
	const speed = 200.0
	p.VelX = 0
	if left {
		p.VelX = -speed
		p.FacingLeft = true
	}
	if right {
		p.VelX = speed
		p.FacingLeft = false
	}

	// Apply simple horizontal movement (physics package will handle gravity)
	p.X += p.VelX * dt
	p.Y += p.VelY * dt
}
