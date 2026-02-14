// renderer.go — Sprite and tile rendering utilities.
//
// Provides helpers for drawing sprites, tile layers, and animated frames
// onto an Ebitengine screen image. The renderer works in pixel coordinates
// and supports camera-relative drawing.
package engine

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Sprite represents a drawable image with position and optional source rect.
type Sprite struct {
	Image    *ebiten.Image
	X, Y     float64
	SrcRect  *image.Rectangle // nil = draw entire image
	ScaleX   float64
	ScaleY   float64
	FlipH    bool
}

// NewSprite creates a Sprite with default scale (1x).
func NewSprite(img *ebiten.Image, x, y float64) Sprite {
	return Sprite{
		Image:  img,
		X:      x,
		Y:      y,
		ScaleX: 1,
		ScaleY: 1,
	}
}

// DrawSprite draws a sprite onto the target image, offset by camera position.
func DrawSprite(target *ebiten.Image, s Sprite, cameraX, cameraY float64) {
	if s.Image == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}

	sx := s.ScaleX
	if s.FlipH {
		sx = -sx
	}
	op.GeoM.Scale(sx, s.ScaleY)
	op.GeoM.Translate(s.X-cameraX, s.Y-cameraY)

	if s.SrcRect != nil {
		target.DrawImage(s.Image.SubImage(*s.SrcRect).(*ebiten.Image), op)
	} else {
		target.DrawImage(s.Image, op)
	}
}

// TileMap holds a grid of tile indices and the tileset image.
// TODO: implement full tile rendering with multiple layers.
type TileMap struct {
	Tiles    [][]int
	TileSize int
	Tileset  *ebiten.Image
}

// DrawTileMap renders visible tiles onto the screen relative to the camera.
func DrawTileMap(target *ebiten.Image, tm TileMap, cameraX, cameraY float64, screenW, screenH int) {
	if tm.Tileset == nil || len(tm.Tiles) == 0 {
		return
	}
	// TODO: calculate visible tile range and draw only those tiles
}
