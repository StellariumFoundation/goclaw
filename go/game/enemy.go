// enemy.go — Enemy types, AI, and behaviors.
//
// Captain Claw features various enemy types per level (soldiers, officers,
// bears, tigers, etc.). Each enemy has patrol/chase/attack AI states.
package game

// EnemyType identifies the kind of enemy.
type EnemyType int

const (
	EnemySoldier EnemyType = iota
	EnemyOfficer
	EnemyRat
	EnemyBear
	EnemyTiger
	EnemyPirate
	// TODO: add all enemy types from each level
)

// EnemyState represents the AI state of an enemy.
type EnemyState int

const (
	EnemyPatrol EnemyState = iota
	EnemyChase
	EnemyAttack
	EnemyHurt
	EnemyDead
)

// Enemy represents a single enemy instance in the game world.
type Enemy struct {
	X, Y       float64
	VelX, VelY float64
	Width      float64
	Height     float64
	Type       EnemyType
	State      EnemyState
	Health     int
	FacingLeft bool

	// Patrol bounds
	PatrolMinX float64
	PatrolMaxX float64

	AnimFrame int
	AnimTimer float64
}

// NewEnemy creates an enemy of the given type at the specified position.
func NewEnemy(t EnemyType, x, y float64) *Enemy {
	return &Enemy{
		X:      x,
		Y:      y,
		Width:  32,
		Height: 48,
		Type:   t,
		State:  EnemyPatrol,
		Health: 10,
	}
}

// Update advances the enemy AI and physics by one tick.
// TODO: implement patrol, chase, and attack behaviors.
func (e *Enemy) Update(dt float64, playerX, playerY float64) {
	// placeholder — will implement AI state machine
}
