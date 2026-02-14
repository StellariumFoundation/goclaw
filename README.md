# GoClaw — AI Digital Worker

**GoClaw** is a Golang rewrite of [OpenClaw](Openclaw/), reimagined as an **AI Digital Worker**. While OpenClaw is a personal AI assistant built with Node.js/TypeScript, GoClaw brings the same powerful capabilities to Go — delivering a fast, compiled, single-binary AI worker that can be deployed anywhere.

The original OpenClaw Python/TypeScript source code is preserved in the [`Openclaw/`](Openclaw/) directory for reference.

## Overview

GoClaw is designed to be a high-performance AI digital worker that:

- **Connects to the channels you already use** — WhatsApp, Telegram, Slack, Discord, Google Chat, Signal, iMessage, Microsoft Teams, and more
- **Runs locally or on any server** — single binary, no runtime dependencies
- **Processes tasks autonomously** — leveraging LLMs (Anthropic Claude, OpenAI GPT) to understand and execute complex workflows
- **Manages multi-channel communication** — unified inbox across all messaging platforms
- **Executes tools and automations** — browser control, cron jobs, webhooks, and custom skills

## Features

- **Single Binary Deployment** — compile once, run anywhere with Go's cross-compilation
- **Multi-Channel Inbox** — WhatsApp, Telegram, Slack, Discord, Google Chat, Signal, BlueBubbles (iMessage), Microsoft Teams, Matrix, Zalo, WebChat
- **AI Agent Runtime** — built-in agent loop with tool streaming and block streaming
- **Gateway Control Plane** — WebSocket-based control plane for sessions, channels, tools, and events
- **Multi-Agent Routing** — route inbound channels/accounts to isolated agents with per-agent sessions
- **Voice Support** — voice wake and talk mode integration
- **Browser Automation** — headless browser control via CDP
- **Skills Platform** — extensible skill system for custom capabilities
- **Session Management** — persistent sessions with compaction, pruning, and context management
- **Security First** — DM pairing, sandboxing, and per-session isolation

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git

### Clone and Build

```bash
git clone https://github.com/StellariumFoundation/goclaw.git
cd goclaw
make build
```

### Run

```bash
make run
# or directly:
./goclaw
```

## Building

The project uses a standard Go build system with a Makefile:

```bash
# Build the binary
make build

# Run directly (without building)
make run

# Format code
make fmt

# Run vet checks
make vet

# Clean build artifacts
make clean
```

## Project Structure

```
.
├── main.go          # Application entry point
├── go.mod           # Go module definition
├── Makefile         # Build targets
├── LICENSE          # MIT License
├── README.md        # This file
└── Openclaw/        # Original OpenClaw source code (preserved)
    ├── README.md    # Original OpenClaw documentation
    ├── src/         # TypeScript source
    ├── apps/        # Companion apps (macOS, iOS, Android)
    ├── docs/        # Documentation
    ├── extensions/  # Channel extensions
    └── ...
```

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

Copyright (c) 2026 StellariumFoundation
