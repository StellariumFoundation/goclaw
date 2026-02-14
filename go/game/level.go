// level.go — Level loading, tile maps, and collision geometry.
//
// Captain Claw levels are stored in WAP/WWD format. This package will
// parse level data and construct the tile map, object placements, and
// collision rectangles for each level.
package game

// Tile represents a single tile in the level grid.
type Tile struct {
	ID       int  // tile index in the tileset
	Solid    bool // whether this tile blocks movement
	Platform bool // one-way platform (passable from below)
	Ladder   bool // climbable
	Deadly   bool // instant kill (spikes, lava)
}

// LevelObject represents a placed object in the level (enemy spawn, item, etc.).
type LevelObject struct {
	Type string
	X, Y float64
	// TODO: add properties map for object-specific data
}

// Level holds all data for a single game level.
type Level struct {
	Name       string
	Width      int // in tiles
	Height     int // in tiles
	TileWidth  int // pixel width of each tile
	TileHeight int // pixel height of each tile
	Tiles      [][]Tile
	Objects    []LevelObject

	// Player start position
	StartX, StartY float64
}

// NewLevel creates an empty level with the given dimensions.
func NewLevel(name string, w, h, tw, th int) *Level {
	tiles := make([][]Tile, h)
	for y := range tiles {
		tiles[y] = make([]Tile, w)
	}
	return &Level{
		Name:       name,
		Width:      w,
		Height:     h,
		TileWidth:  tw,
		TileHeight: th,
		Tiles:      tiles,
	}
}

// IsSolid checks whether the tile at grid position (tx, ty) is solid.
func (l *Level) IsSolid(tx, ty int) bool {
	if tx < 0 || ty < 0 || tx >= l.Width || ty >= l.Height {
		return true // out of bounds = solid
	}
	return l.Tiles[ty][tx].Solid
}

// WorldToTile converts world pixel coordinates to tile grid coordinates.
func (l *Level) WorldToTile(wx, wy float64) (int, int) {
	return int(wx) / l.TileWidth, int(wy) / l.TileHeight
}
