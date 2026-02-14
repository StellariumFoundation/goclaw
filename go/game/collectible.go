// collectible.go — Treasures, health pickups, lives, and powerups.
//
// Captain Claw levels are filled with collectible items: gold/silver/bronze
// treasures for score, health potions, extra lives, ammo pickups, and
// special powerups (catnip invincibility, ghost invisibility).
package game

// CollectibleType identifies the kind of pickup.
type CollectibleType int

const (
	CollectTreasureGold CollectibleType = iota
	CollectTreasureSilver
	CollectTreasureBronze
	CollectTreasureCrown
	CollectTreasureRing
	CollectTreasureChalice
	CollectTreasureCross
	CollectTreasureScepter
	CollectTreasureGecko
	CollectHealth
	CollectExtraLife
	CollectPistolAmmo
	CollectMagicAmmo
	CollectDynamiteAmmo
	CollectPowerupCatnip    // invincibility
	CollectPowerupGhost     // invisibility
	CollectPowerupFireSword // fire sword
	CollectMapPiece         // end-of-level collectible
)

// Collectible represents a pickup item in the game world.
type Collectible struct {
	X, Y      float64
	Width     float64
	Height    float64
	Type      CollectibleType
	Value     int  // score value or ammo amount
	Collected bool // already picked up

	AnimFrame int
	AnimTimer float64
}

// NewCollectible creates a collectible at the given position.
func NewCollectible(t CollectibleType, x, y float64, value int) *Collectible {
	return &Collectible{
		X:      x,
		Y:      y,
		Width:  24,
		Height: 24,
		Type:   t,
		Value:  value,
	}
}

// Apply gives the collectible's effect to the player.
// TODO: implement per-type effects (score, health, ammo, powerups).
func (c *Collectible) Apply(p *Player) {
	if c.Collected {
		return
	}
	c.Collected = true

	switch c.Type {
	case CollectTreasureGold, CollectTreasureSilver, CollectTreasureBronze,
		CollectTreasureCrown, CollectTreasureRing, CollectTreasureChalice,
		CollectTreasureCross, CollectTreasureScepter, CollectTreasureGecko:
		p.Score += c.Value
	case CollectHealth:
		p.Health += c.Value
		if p.Health > 100 {
			p.Health = 100
		}
	case CollectExtraLife:
		p.Lives++
	case CollectPistolAmmo:
		p.Pistol += c.Value
	case CollectMagicAmmo:
		p.Magic += c.Value
	case CollectDynamiteAmmo:
		p.Dynamite += c.Value
	}
}
