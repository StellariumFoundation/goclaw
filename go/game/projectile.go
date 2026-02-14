// projectile.go — Player and enemy projectiles.
//
// Handles pistol bullets, magic claw projectiles, dynamite, and enemy
// projectiles (swords, cannonballs, etc.). Each projectile has velocity,
// lifetime, damage, and collision detection.
package game

// ProjectileType identifies the kind of projectile.
type ProjectileType int

const (
	ProjPistol   ProjectileType = iota // player pistol bullet
	ProjMagic                          // player magic claw
	ProjDynamite                       // player dynamite (arc trajectory)
	ProjFireSword                      // player fire sword projectile
	ProjEnemySword                     // enemy thrown sword
	ProjCannonball                     // cannon trap
)

// Projectile represents a moving projectile in the game world.
type Projectile struct {
	X, Y       float64
	VelX, VelY float64
	Width      float64
	Height     float64
	Type       ProjectileType
	Damage     int
	Lifetime   float64 // seconds remaining before despawn
	FromPlayer bool    // true if fired by the player
	Active     bool
}

// NewProjectile creates a projectile at the given position with velocity.
func NewProjectile(t ProjectileType, x, y, vx, vy float64, fromPlayer bool) *Projectile {
	dmg := 10
	switch t {
	case ProjPistol:
		dmg = 10
	case ProjMagic:
		dmg = 25
	case ProjDynamite:
		dmg = 50
	case ProjFireSword:
		dmg = 30
	case ProjEnemySword, ProjCannonball:
		dmg = 15
	}

	return &Projectile{
		X:          x,
		Y:          y,
		VelX:       vx,
		VelY:       vy,
		Width:      8,
		Height:     8,
		Type:       t,
		Damage:     dmg,
		Lifetime:   3.0,
		FromPlayer: fromPlayer,
		Active:     true,
	}
}

// Update moves the projectile and decrements its lifetime.
func (p *Projectile) Update(dt float64) {
	if !p.Active {
		return
	}
	p.X += p.VelX * dt
	p.Y += p.VelY * dt

	// Dynamite has gravity
	if p.Type == ProjDynamite {
		p.VelY += 400 * dt
	}

	p.Lifetime -= dt
	if p.Lifetime <= 0 {
		p.Active = false
	}
}
