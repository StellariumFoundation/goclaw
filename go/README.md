# GoClaw

A Go rewrite of [OpenClaw](https://github.com/openclaw/openclaw) — the open-source reimplementation of **Captain Claw**, the classic 1997 2D side-scrolling platformer by Monolith Productions.

## Status

🚧 **Early development / scaffolding phase** — the project structure is in place but gameplay is not yet functional.

## Project Structure

```
go/
├── main.go              # Entry point — initializes window and game loop
├── engine/              # Core game engine
│   ├── window.go        # Window creation and management (Ebitengine)
│   ├── input.go         # Keyboard/mouse input handling
│   ├── renderer.go      # Sprite and tile rendering
│   └── audio.go         # Sound/music playback
├── game/                # Game logic
│   ├── player.go        # Captain Claw — movement, physics, animations
│   ├── enemy.go         # Enemy types, AI, behaviors
│   ├── level.go         # Level loading, tile maps, collision geometry
│   ├── camera.go        # Camera/viewport following the player
│   ├── collectible.go   # Treasures, health, lives, powerups
│   └── projectile.go    # Player and enemy projectiles
├── assets/              # Asset loading and management
│   ├── loader.go        # Asset loading functions
│   ├── rez.go           # REZ archive parser (Captain Claw asset format)
│   └── wap.go           # WAP/WWD level file parser
├── physics/             # Physics subsystem
│   └── physics.go       # Gravity, collision detection, movement resolution
└── ui/                  # User interface
    ├── hud.go           # In-game HUD (score, health, lives, ammo)
    └── menu.go          # Main menu, pause menu
```

## Dependencies

- **[Ebitengine](https://ebitengine.org/)** (`github.com/hajimehoshi/ebiten/v2`) — 2D game library for Go providing window management, rendering, input handling, and audio playback.

## Building and Running

### Prerequisites

- Go 1.22 or later
- On Linux: `libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config`
- On macOS: Xcode command line tools
- On Windows: no additional dependencies

### Build

```bash
cd go
go build -o goclaw .
```

### Run

```bash
cd go
go run .
```

This will open a window titled "GoClaw - Captain Claw Reimplementation" with a placeholder screen.

## Goals

- Faithful reimplementation of Captain Claw gameplay in Go
- Parse and use original game assets (REZ archives, WAP level files)
- All 14 levels with enemies, collectibles, and boss fights
- Sound effects and music playback
- Cross-platform support (Linux, macOS, Windows)

## License

See the repository root [LICENSE](../LICENSE) file.
