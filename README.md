# 🦞 GoClaw — AI Digital Worker

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](LICENSE)

**GoClaw** is a Golang rewrite of [OpenClaw](Openclaw/), the personal AI assistant. It is designed as an **AI digital worker** — an autonomous agent that can operate across messaging channels, execute tasks, and integrate with your existing workflows.

## Overview

GoClaw reimagines the OpenClaw personal AI assistant in Go, bringing the benefits of Go's performance, concurrency model, and single-binary deployment to the OpenClaw ecosystem.

### Key Goals

- **AI Digital Worker**: An autonomous agent capable of performing tasks, answering questions, and managing workflows across multiple channels.
- **Go-native**: Built from the ground up in Go for performance, reliability, and ease of deployment.
- **Channel Support**: WhatsApp, Telegram, Slack, Discord, Signal, and more — just like the original OpenClaw.
- **Single Binary**: Deploy anywhere with a single compiled binary.

## Project Structure

```
.
├── go.mod          # Go module definition
├── main.go         # Entry point for the AI digital worker
├── LICENSE         # MIT License
├── README.md       # This file
└── Openclaw/       # Original OpenClaw source (TypeScript/Node.js)
```

The original OpenClaw source code is preserved in the [`Openclaw/`](Openclaw/) directory for reference. The Go rewrite lives at the repository root.

## Getting Started

```bash
# Clone the repository
git clone https://github.com/StellariumFoundation/goclaw.git
cd goclaw

# Build
go build -o goclaw .

# Run
./goclaw
```

## License

This project is licensed under the [MIT License](LICENSE).

Copyright (c) 2025 StellariumFoundation
