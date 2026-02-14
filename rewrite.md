# OpenClaw → GoClaw: Complete Go 1.26.0 Rewrite Blueprint

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Original Architecture Overview](#2-original-architecture-overview)
3. [Module/Component Map](#3-modulecomponent-map)
4. [Data Flow and APIs](#4-data-flow-and-apis)
5. [Dependency Analysis](#5-dependency-analysis)
6. [Go 1.26.0 Rewrite Plan](#6-go-1260-rewrite-plan)
7. [Proposed Go Package Structure](#7-proposed-go-package-structure)
8. [Implementation Priority](#8-implementation-priority)
9. [Technical Considerations](#9-technical-considerations)
10. [Migration Strategy](#10-migration-strategy)

---

## 1. Executive Summary

### What is OpenClaw?

OpenClaw is a **multi-channel personal AI assistant gateway** built primarily in TypeScript/Node.js (≥22). It acts as a unified control plane that connects to messaging platforms (WhatsApp, Telegram, Slack, Discord, Google Chat, Signal, iMessage, Microsoft Teams, Matrix, WebChat, and more) and routes conversations through AI model providers (Anthropic Claude, OpenAI GPT/Codex, Google Gemini, local Ollama, etc.). The system includes:

- A **WebSocket-based Gateway server** (the control plane)
- A **CLI** (`openclaw`) for management, onboarding, and direct agent interaction
- **Channel adapters** for 15+ messaging platforms
- An **AI agent runtime** (Pi agent, embedded runner) with tool execution
- **Companion apps** for macOS (menu bar), iOS, and Android (Swift/Kotlin)
- A **plugin/extension system** for third-party integrations
- **Browser automation** via Playwright/CDP
- **Voice/TTS** capabilities (ElevenLabs, Edge TTS)
- **Canvas** (A2UI) for agent-driven visual workspaces
- **Session management**, cron jobs, webhooks, and a full configuration system

### Goal of the Go Rewrite

The GoClaw rewrite aims to port the entire OpenClaw server-side platform from TypeScript/Node.js to **Go 1.26.0**, targeting:

- **Single binary distribution** — no Node.js runtime dependency, no `npm install`
- **Superior concurrency** — goroutines for WebSocket handling, channel adapters, and agent runs
- **Lower memory footprint** — critical for self-hosted/VPS deployments
- **Faster startup** — sub-second cold start vs. Node.js JIT warmup
- **Type safety at compile time** — Go's static typing eliminates runtime type errors
- **Simplified deployment** — cross-compile for Linux/macOS/Windows/ARM
- **Production reliability** — Go's mature stdlib for HTTP, WebSocket, TLS, and process management

The Go module is `github.com/StellariumFoundation/goclaw`.

---

## 2. Original Architecture Overview

### Languages & Frameworks

| Layer | Language | Framework/Runtime |
|-------|----------|-------------------|
| Gateway server | TypeScript | Node.js ≥22, Express 5, ws (WebSocket) |
| CLI | TypeScript | Commander.js, @clack/prompts |
| AI Agent runtime | TypeScript | @mariozechner/pi-agent-core, pi-ai, pi-coding-agent |
| Configuration | TypeScript | Zod schemas, JSON5, YAML |
| Build system | TypeScript | tsdown (Rolldown-based), Vitest, oxlint, oxfmt |
| Control UI | TypeScript | Lit (Web Components), Rolldown |
| macOS app | Swift | SwiftUI, Bonjour/mDNS |
| iOS app | Swift | SwiftUI, OpenClawKit |
| Android app | Kotlin | Gradle, Android SDK |
| Docs i18n tool | Go | Custom translation pipeline |
| Package manager | — | pnpm 10.23.0 (monorepo workspace) |

### Repository Structure

```
Openclaw/
├── src/                    # Core TypeScript source (~200+ modules)
│   ├── entry.ts            # CLI entrypoint (process respawn, profile)
│   ├── cli/                # CLI commands (Commander.js program)
│   ├── gateway/            # WebSocket gateway server (control plane)
│   ├── agents/             # AI agent runtime, tools, sandbox
│   ├── auto-reply/         # Inbound message → agent reply pipeline
│   ├── channels/           # Channel abstraction layer & plugin system
│   ├── config/             # Configuration schema, loading, validation (Zod)
│   ├── browser/            # Browser automation (Playwright, CDP)
│   ├── telegram/           # Telegram channel adapter (grammY)
│   ├── discord/            # Discord channel adapter (discord.js/Carbon)
│   ├── slack/              # Slack channel adapter (Bolt)
│   ├── signal/             # Signal channel adapter (signal-cli)
│   ├── imessage/           # iMessage channel adapter (legacy)
│   ├── whatsapp/           # WhatsApp channel adapter (Baileys)
│   ├── infra/              # Infrastructure (env, retry, heartbeat, bonjour, etc.)
│   ├── hooks/              # Lifecycle hooks (bundled + workspace)
│   ├── plugins/            # Plugin system (discovery, loading, runtime)
│   ├── sessions/           # Session management
│   ├── routing/            # Multi-agent routing
│   ├── process/            # Process management, command queue
│   ├── media/              # Media pipeline (images, audio, video)
│   ├── media-understanding/# Media analysis
│   ├── tts/                # Text-to-speech
│   ├── logging/            # Structured logging, redaction
│   ├── security/           # Security utilities
│   ├── wizard/             # Onboarding wizard
│   ├── tui/                # Terminal UI
│   ├── cron/               # Cron job system
│   ├── daemon/             # Daemon management (launchd/systemd)
│   ├── canvas-host/        # Canvas A2UI host
│   ├── node-host/          # Node host for device nodes
│   ├── pairing/            # Device pairing
│   ├── providers/          # Auth provider adapters (Copilot, Gemini, Qwen)
│   ├── plugin-sdk/         # Plugin SDK exports
│   ├── types/              # Shared type definitions
│   ├── utils/              # General utilities
│   ├── web/                # Web channel
│   ├── acp/                # Agent Client Protocol bridge
│   ├── commands/           # CLI command implementations
│   ├── link-understanding/ # URL/link analysis
│   ├── markdown/           # Markdown processing
│   └── terminal/           # Terminal utilities
├── extensions/             # Plugin extensions (36+ extensions)
│   ├── bluebubbles/        # iMessage via BlueBubbles
│   ├── matrix/             # Matrix protocol
│   ├── msteams/            # Microsoft Teams
│   ├── googlechat/         # Google Chat
│   ├── line/               # LINE messaging
│   ├── feishu/             # Feishu/Lark
│   ├── irc/                # IRC
│   ├── nostr/              # Nostr protocol
│   ├── twitch/             # Twitch chat
│   ├── voice-call/         # Voice calling
│   ├── memory-core/        # Memory system core
│   ├── memory-lancedb/     # LanceDB memory backend
│   ├── diagnostics-otel/   # OpenTelemetry diagnostics
│   ├── copilot-proxy/      # GitHub Copilot proxy
│   ├── llm-task/           # LLM task tool
│   ├── talk-voice/         # Voice conversation
│   ├── open-prose/         # Creative writing skills
│   └── ...                 # 18+ more extensions
├── apps/                   # Native companion apps
│   ├── macos/              # macOS menu bar app (Swift)
│   ├── ios/                # iOS node app (Swift)
│   ├── android/            # Android node app (Kotlin)
│   └── shared/OpenClawKit/ # Shared Swift package
├── docs/                   # Documentation (Mintlify)
├── packages/               # Sub-packages (clawdbot, moltbot wrappers)
├── scripts/                # Build/deploy/test scripts
├── skills/                 # Bundled skills
├── assets/                 # Static assets, Chrome extension
├── ui/                     # Control UI (Lit web components)
└── Swabble/                # Additional tooling
```

### Key Configuration Files

| File | Purpose |
|------|---------|
| `package.json` | Root package, dependencies, scripts |
| `pnpm-workspace.yaml` | Monorepo workspace (root, ui, packages/*, extensions/*) |
| `tsdown.config.ts` | Build config (6 entry points) |
| `vitest.config.ts` | Test config (unit, e2e, live, gateway, extensions) |
| `Dockerfile` | Production Docker image (Node 22 bookworm) |
| `docker-compose.yml` | Gateway + CLI services |
| `fly.toml` | Fly.io deployment |
| `.env.example` | Environment variable template |

---

## 3. Module/Component Map

### 3.1 Gateway Server (`src/gateway/`)

The heart of OpenClaw. A WebSocket + HTTP server that manages all state.

| File/Dir | Purpose |
|----------|---------|
| `server.impl.ts` | Main server initialization, wiring all subsystems |
| `server-http.ts` | HTTP listener (Express 5), REST endpoints |
| `server-ws-runtime.ts` | WebSocket connection handling |
| `server-methods.ts` | Core WS method handlers |
| `server-methods/` | Individual method handlers (agent, chat, config, sessions, etc.) |
| `server-channels.ts` | Channel manager lifecycle |
| `server-chat.ts` | Chat event handler, agent event routing |
| `server-cron.ts` | Cron service integration |
| `server-discovery.ts` | Bonjour/mDNS discovery |
| `server-plugins.ts` | Plugin loading and lifecycle |
| `server-tailscale.ts` | Tailscale Serve/Funnel integration |
| `server-broadcast.ts` | Event broadcasting to connected clients |
| `server-lanes.ts` | Concurrency lane management |
| `server-maintenance.ts` | Maintenance timers |
| `server-mobile-nodes.ts` | Mobile node management |
| `server-model-catalog.ts` | Model catalog loading |
| `server-node-events.ts` | Node event handling |
| `server-node-subscriptions.ts` | Node subscription management |
| `server-reload-handlers.ts` | Hot-reload handlers |
| `server-runtime-config.ts` | Runtime configuration resolution |
| `server-runtime-state.ts` | Runtime state management |
| `server-session-key.ts` | Session key resolution |
| `server-startup.ts` | Startup sidecars |
| `server-wizard-sessions.ts` | Wizard session tracking |
| `protocol/` | Protocol schema definitions (TypeBox) |
| `protocol/schema/` | Individual schema modules (agent, channels, config, etc.) |
| `server/` | Server internals (health, TLS, WS connection, HTTP listen) |
| `auth.ts` | Authentication (token, password, Tailscale identity) |
| `auth-rate-limit.ts` | Auth rate limiting |
| `client.ts` | Gateway client (for CLI → Gateway communication) |
| `config-reload.ts` | Config file watcher and reloader |
| `node-registry.ts` | Device node registry |
| `openai-http.ts` | OpenAI-compatible HTTP API |
| `openresponses-http.ts` | Open Responses HTTP API |

**Key interactions:** The gateway server is the central hub. All channels, CLI clients, mobile nodes, and the Control UI connect via WebSocket. The server manages sessions, routes messages to the agent runtime, and broadcasts events.

### 3.2 CLI (`src/cli/`)

Full-featured CLI built with Commander.js.

| File/Dir | Purpose |
|----------|---------|
| `program.ts` | Main Commander program definition |
| `program/` | Command registration (agent, configure, maintenance, etc.) |
| `run-main.ts` | CLI entry runner |
| `gateway-cli.ts` | `openclaw gateway` command |
| `daemon-cli.ts` | `openclaw daemon` command (install/start/stop) |
| `channels-cli.ts` | `openclaw channels` command |
| `nodes-cli.ts` | `openclaw nodes` command |
| `browser-cli.ts` | `openclaw browser` command |
| `config-cli.ts` | `openclaw config` command |
| `skills-cli.ts` | `openclaw skills` command |
| `plugins-cli.ts` | `openclaw plugins` command |
| `pairing-cli.ts` | `openclaw pairing` command |
| `update-cli.ts` | `openclaw update` command |
| `profile.ts` | CLI profile management |
| `completion-cli.ts` | Shell completion |

### 3.3 Agent Runtime (`src/agents/`)

The AI agent execution engine.

| File/Dir | Purpose |
|----------|---------|
| `pi-embedded-runner.ts` | Main embedded Pi agent runner |
| `pi-embedded-runner/` | Runner internals (abort, compact, history, model, etc.) |
| `pi-embedded-subscribe.ts` | Agent event subscription and streaming |
| `pi-embedded-helpers/` | Helper utilities (errors, images, thinking, turns) |
| `pi-tools.ts` | Tool definition and registration |
| `pi-tool-definition-adapter.ts` | Tool schema adaptation |
| `tools/` | Individual tool implementations |
| `bash-tools.ts` | Bash/shell execution tool |
| `system-prompt.ts` | System prompt construction |
| `auth-profiles.ts` | Auth profile management and rotation |
| `models-config.ts` | Model configuration and provider setup |
| `model-catalog.ts` | Model catalog |
| `sandbox/` | Docker sandbox for untrusted execution |
| `skills/` | Skills system |
| `cli-runner.ts` | CLI-based agent runner |
| `session-slug.ts` | Session slug generation |
| `identity-file.ts` | Agent identity file management |
| `lanes.ts` | Agent lane management |

**Agent Tools** (`src/agents/tools/`):
- `browser-tool.ts` — Browser automation
- `canvas-tool.ts` — Canvas A2UI control
- `cron-tool.ts` — Cron job management
- `discord-actions.ts` — Discord-specific actions
- `slack-actions.ts` — Slack-specific actions
- `telegram-actions.ts` — Telegram-specific actions
- `whatsapp-actions.ts` — WhatsApp-specific actions
- `gateway-tool.ts` — Gateway control
- `image-tool.ts` — Image generation
- `memory-tool.ts` — Memory/knowledge base
- `message-tool.ts` — Cross-channel messaging
- `nodes-tool.ts` — Device node invocation
- `sessions-*.ts` — Multi-session tools (list, history, send, spawn)
- `tts-tool.ts` — Text-to-speech
- `web-fetch.ts` — Web fetching/scraping
- `web-search.ts` — Web search

### 3.4 Auto-Reply Pipeline (`src/auto-reply/`)

Handles inbound messages → directive parsing → agent execution → reply delivery.

| File/Dir | Purpose |
|----------|---------|
| `reply.ts` | Main reply orchestration |
| `reply/` | Reply subsystem (agent-runner, commands, directives, streaming, etc.) |
| `dispatch.ts` | Message dispatch |
| `envelope.ts` | Message envelope construction |
| `chunk.ts` | Message chunking for platform limits |
| `command-detection.ts` | Chat command detection (`/status`, `/reset`, etc.) |
| `commands-registry.ts` | Command registry |
| `group-activation.ts` | Group activation logic |
| `heartbeat.ts` | Heartbeat/typing indicators |
| `status.ts` | Status command handler |
| `thinking.ts` | Thinking level management |

### 3.5 Channel System (`src/channels/`)

Abstraction layer for messaging platforms.

| File/Dir | Purpose |
|----------|---------|
| `registry.ts` | Channel registry |
| `plugins/` | Channel plugin system |
| `plugins/index.ts` | Plugin discovery and loading |
| `plugins/types.ts` | Channel plugin type definitions |
| `plugins/catalog.ts` | Plugin catalog |
| `plugins/onboarding/` | Per-channel onboarding flows |
| `plugins/outbound/` | Per-channel outbound message sending |
| `plugins/normalize/` | Per-channel message normalization |
| `plugins/actions/` | Per-channel action handlers |
| `plugins/status-issues/` | Per-channel status diagnostics |
| `allowlist-match.ts` | Sender allowlist matching |
| `mention-gating.ts` | Mention-based activation gating |
| `command-gating.ts` | Command access gating |
| `typing.ts` | Typing indicator management |
| `targets.ts` | Target resolution |
| `session.ts` | Channel session management |
| `web/` | WebChat channel |

### 3.6 Configuration (`src/config/`)

Comprehensive configuration system with Zod validation.

| File/Dir | Purpose |
|----------|---------|
| `config.ts` | Config loading, migration, writing |
| `schema.ts` | TypeBox schema definitions |
| `zod-schema.ts` | Zod validation schemas |
| `types.ts` | Config type definitions (30+ type files) |
| `defaults.ts` | Default values |
| `validation.ts` | Config validation |
| `sessions/` | Session config and storage |
| `legacy.ts` | Legacy config migration |
| `includes.ts` | Config file includes |
| `env-vars.ts` | Environment variable mapping |
| `env-substitution.ts` | Env var substitution in config |
| `io.ts` | Config file I/O |
| `paths.ts` | Config path resolution |

### 3.7 Infrastructure (`src/infra/`)

Core infrastructure utilities.

| File/Dir | Purpose |
|----------|---------|
| `env.ts` | Environment variable handling |
| `retry.ts` | Retry logic with backoff |
| `heartbeat-runner.ts` | Heartbeat/cron runner |
| `bonjour.ts` | Bonjour/mDNS service |
| `device-identity.ts` | Device identity management |
| `device-pairing.ts` | Device pairing |
| `update-check.ts` | Update checking |
| `update-runner.ts` | Update execution |
| `tailscale.ts` | Tailscale integration |
| `ssh-tunnel.ts` | SSH tunnel management |
| `provider-usage.ts` | API usage tracking |
| `session-cost-usage.ts` | Session cost tracking |
| `system-events.ts` | System event bus |
| `system-presence.ts` | System presence detection |
| `gateway-lock.ts` | Gateway lock (single instance) |
| `file-lock.ts` | File locking |
| `archive.ts` | Archive (tar) utilities |
| `fetch.ts` | HTTP fetch utilities |
| `net/` | Network utilities (SSRF protection) |
| `tls/` | TLS utilities (fingerprinting) |
| `format-time/` | Time formatting |

### 3.8 Browser Automation (`src/browser/`)

Full browser control via Playwright and CDP.

| File/Dir | Purpose |
|----------|---------|
| `server.ts` | Browser control HTTP server |
| `pw-session.ts` | Playwright session management |
| `pw-ai.ts` | AI-assisted browser actions |
| `pw-tools-core.ts` | Core browser tool operations |
| `cdp.ts` | Chrome DevTools Protocol |
| `chrome.ts` | Chrome/Chromium management |
| `profiles.ts` | Browser profile management |
| `client.ts` | Browser client |
| `config.ts` | Browser configuration |
| `extension-relay.ts` | Chrome extension relay |
| `screenshot.ts` | Screenshot capture |

### 3.9 Media Pipeline (`src/media/`)

Media handling for images, audio, and video.

| File/Dir | Purpose |
|----------|---------|
| `store.ts` | Media store (temp file lifecycle) |
| `fetch.ts` | Media fetching |
| `host.ts` | Media hosting |
| `server.ts` | Media HTTP server |
| `image-ops.ts` | Image operations (sharp) |
| `audio.ts` | Audio processing |
| `mime.ts` | MIME type detection |
| `parse.ts` | Media parsing |

### 3.10 Plugin System (`src/plugins/`)

Extensible plugin architecture.

| File/Dir | Purpose |
|----------|---------|
| `discovery.ts` | Plugin discovery |
| `loader.ts` | Plugin loading (jiti) |
| `registry.ts` | Plugin registry |
| `runtime.ts` | Plugin runtime |
| `services.ts` | Plugin services |
| `hooks.ts` | Plugin hook system |
| `install.ts` | Plugin installation |
| `manifest.ts` | Plugin manifest parsing |
| `tools.ts` | Plugin tool registration |
| `slots.ts` | Plugin slot system |

### 3.11 Extensions (`extensions/`)

36+ extension plugins, each with:
- `openclaw.plugin.json` — Plugin manifest
- `package.json` — Dependencies
- `index.ts` — Entry point
- `src/` — Implementation

**Channel extensions:** bluebubbles, discord, feishu, googlechat, imessage, irc, line, matrix, mattermost, msteams, nextcloud-talk, nostr, signal, slack, telegram, tlon, twitch, voice-call, whatsapp, zalo, zalouser

**Utility extensions:** copilot-proxy, device-pair, diagnostics-otel, google-antigravity-auth, google-gemini-cli-auth, llm-task, lobster, memory-core, memory-lancedb, minimax-portal-auth, open-prose, phone-control, qwen-portal-auth, talk-voice, thread-ownership

### 3.12 Native Apps (`apps/`)

| App | Language | Purpose |
|-----|----------|---------|
| `apps/macos/` | Swift | macOS menu bar app, Voice Wake, Canvas |
| `apps/ios/` | Swift | iOS node (camera, screen, Canvas, Talk) |
| `apps/android/` | Kotlin | Android node (camera, screen, Canvas, Talk) |
| `apps/shared/OpenClawKit/` | Swift | Shared Swift package (protocol, chat UI, gateway models) |

---

## 4. Data Flow and APIs

### 4.1 Gateway WebSocket Protocol

The Gateway exposes a JSON-RPC-like WebSocket protocol on port 18789 (default).

**Connection flow:**
```
Client → ws://127.0.0.1:18789 → Auth (token/password) → Session
```

**Key WS methods (server-methods/):**
- `agent.*` — Agent operations (run, abort, status)
- `chat.*` — Chat operations (send, inject, transcript)
- `config.*` — Configuration CRUD
- `sessions.*` — Session management (list, patch, reset, send)
- `channels.*` — Channel status and control
- `nodes.*` — Device node operations (list, describe, invoke)
- `models.*` — Model catalog and selection
- `health.*` — Health checks
- `cron.*` — Cron job management
- `skills.*` — Skills management
- `browser.*` — Browser control
- `logs.*` — Log retrieval
- `devices.*` — Device management
- `usage.*` — Usage statistics
- `wizard.*` — Onboarding wizard
- `web.*` — Web/WebChat operations
- `talk.*` — Voice/Talk mode
- `tts.*` — Text-to-speech
- `voicewake.*` — Voice wake detection
- `update.*` — Update management
- `system.*` — System operations
- `connect.*` — Connection management

**Events (server → client broadcasts):**
- Agent events (text, tool calls, errors, completion)
- Channel events (message received, status change)
- Node events (connected, disconnected, capability change)
- Session events (created, updated, reset)
- Health events
- Presence events

### 4.2 HTTP API

The Gateway also exposes HTTP endpoints:
- `GET /health` — Health check
- `GET /api/v1/*` — REST API
- OpenAI-compatible API (`/v1/chat/completions`, etc.)
- Open Responses API
- Media server endpoints
- Browser control endpoints
- Control UI static files
- WebChat endpoints
- Plugin HTTP routes
- Webhook endpoints

### 4.3 Message Flow (Inbound)

```
1. Channel adapter receives message (e.g., Telegram bot update)
2. Channel plugin normalizes message → InboundMessage
3. Allowlist/mention gating check
4. Session key resolution (routing rules)
5. Command detection (/status, /reset, etc.)
6. If not a command → auto-reply pipeline:
   a. Directive parsing (inline /think, /model, etc.)
   b. Media staging (images, audio → temp files)
   c. Agent runner invocation (Pi embedded runner)
   d. Tool execution loop (bash, browser, web, etc.)
   e. Response streaming → block reply coalescing
   f. Chunking for platform limits
   g. Outbound delivery via channel adapter
7. Gateway broadcasts events to connected clients
```

### 4.4 Message Flow (Outbound / Agent-initiated)

```
1. Agent tool call (e.g., message_tool, discord_actions)
2. Target resolution (channel + recipient)
3. Message formatting for target platform
4. Channel adapter sends message
5. Delivery confirmation
```

### 4.5 Configuration Flow

```
~/.openclaw/openclaw.json (JSON5)
    ↓ loadConfig()
    ↓ Zod validation
    ↓ Legacy migration
    ↓ Env var substitution
    ↓ Runtime overrides
    → Typed OpenClawConfig object
```

---

## 5. Dependency Analysis

### 5.1 Core Runtime Dependencies → Go Equivalents

| Node.js Dependency | Purpose | Go Equivalent |
|--------------------|---------|---------------|
| `express` (v5) | HTTP server | `net/http` (stdlib) or `chi`/`echo` |
| `ws` | WebSocket server | `nhooyr.io/websocket` or `gorilla/websocket` |
| `commander` | CLI framework | `cobra` + `pflag` |
| `@clack/prompts` | Interactive CLI prompts | `charmbracelet/huh` or `survey` |
| `zod` (v4) | Schema validation | Go structs + `go-playground/validator` |
| `@sinclair/typebox` | JSON Schema / TypeBox | `encoding/json` + custom schemas |
| `dotenv` | Env file loading | `joho/godotenv` |
| `chalk` | Terminal colors | `fatih/color` or `charmbracelet/lipgloss` |
| `chokidar` | File watching | `fsnotify/fsnotify` |
| `yaml` | YAML parsing | `gopkg.in/yaml.v3` |
| `json5` | JSON5 parsing | `yosuke-furukawa/json5` or custom parser |
| `tslog` | Structured logging | `log/slog` (stdlib) or `zerolog` |
| `undici` | HTTP client | `net/http` (stdlib) |
| `proper-lockfile` | File locking | `os.OpenFile` with `syscall.Flock` |
| `croner` | Cron scheduling | `robfig/cron/v3` |
| `ajv` | JSON Schema validation | `santhosh-tekuri/jsonschema` |
| `sharp` | Image processing | `disintegration/imaging` or `h2non/bimg` |
| `tar` | Archive handling | `archive/tar` (stdlib) |
| `jszip` | ZIP handling | `archive/zip` (stdlib) |
| `markdown-it` | Markdown rendering | `yuin/goldmark` |
| `qrcode-terminal` | QR code display | `skip2/go-qrcode` |
| `osc-progress` | Progress bars | `schollz/progressbar` |
| `signal-utils` | Signal handling | `os/signal` (stdlib) |

### 5.2 Channel/Messaging Dependencies → Go Equivalents

| Node.js Dependency | Purpose | Go Equivalent |
|--------------------|---------|---------------|
| `grammy` | Telegram Bot API | `go-telegram-bot-api/telegram-bot-api` or `gotd/td` |
| `@buape/carbon` / `discord-api-types` | Discord Bot | `bwmarrin/discordgo` |
| `@slack/bolt` + `@slack/web-api` | Slack Bot | `slack-go/slack` |
| `@whiskeysockets/baileys` | WhatsApp (Web) | `tulir/whatsmeow` |
| `@line/bot-sdk` | LINE messaging | Custom HTTP client |
| `@larksuiteoapi/node-sdk` | Feishu/Lark | `larksuite/oapi-sdk-go` |
| `@matrix-org/matrix-sdk-crypto-nodejs` | Matrix E2EE | `mautrix/go` |

### 5.3 AI/ML Dependencies → Go Equivalents

| Node.js Dependency | Purpose | Go Equivalent |
|--------------------|---------|---------------|
| `@mariozechner/pi-*` | Pi agent runtime | Custom Go agent runtime |
| `@agentclientprotocol/sdk` | ACP bridge | Custom Go ACP implementation |
| `@aws-sdk/client-bedrock` | AWS Bedrock | `aws/aws-sdk-go-v2` |
| `node-llama-cpp` | Local LLM | `go-skynet/go-llama.cpp` |
| `ollama` | Ollama client | `ollama/ollama` (Go native) |
| `pdfjs-dist` | PDF parsing | `ledongthuc/pdf` or `unidoc/unipdf` |
| `@mozilla/readability` | Web readability | `nicholasgasior/goinern` or custom |
| `linkedom` | DOM parsing | `PuerkitoBio/goquery` |

### 5.4 Browser/Automation Dependencies → Go Equivalents

| Node.js Dependency | Purpose | Go Equivalent |
|--------------------|---------|---------------|
| `playwright-core` | Browser automation | `chromedp/chromedp` (CDP) or `go-rod/rod` |
| `file-type` | File type detection | `h2non/filetype` |

### 5.5 Voice/TTS Dependencies → Go Equivalents

| Node.js Dependency | Purpose | Go Equivalent |
|--------------------|---------|---------------|
| `node-edge-tts` | Edge TTS | Custom HTTP client to Edge TTS API |
| ElevenLabs API | Voice synthesis | Custom HTTP client |

### 5.6 Dev/Build Dependencies (Not needed in Go)

These are TypeScript-specific and have no Go equivalent needed:
- `tsdown`, `tsx`, `typescript`, `vitest`, `oxlint`, `oxfmt`, `rolldown`
- `@types/*` packages
- `lit`, `@lit-labs/signals`, `@lit/context` (Control UI — separate concern)

---

## 6. Go 1.26.0 Rewrite Plan

### 6.1 Gateway Server

**Original:** `src/gateway/` (Express 5 + ws WebSocket)

**Go implementation:**
- Use `net/http` for HTTP server with `nhooyr.io/websocket` for WebSocket
- JSON-RPC message handling with typed Go structs
- Connection management with goroutine-per-connection model
- Auth middleware (token, password, rate limiting)
- Health endpoint, OpenAI-compatible API, webhook endpoints

**Key mappings:**
| TypeScript | Go |
|------------|-----|
| `server.impl.ts` → `startGatewayServer()` | `internal/gateway/server.go` → `NewServer()` |
| `server-methods.ts` → handler map | `internal/gateway/methods/` → method registry |
| `server-ws-runtime.ts` → WS handling | `internal/gateway/ws/` → connection handler |
| `protocol/schema/` → TypeBox schemas | `internal/gateway/protocol/` → Go structs |
| `server-broadcast.ts` → event fan-out | `internal/gateway/broadcast.go` → channel-based fan-out |
| `auth.ts` → auth middleware | `internal/gateway/auth/` → middleware |
| `config-reload.ts` → file watcher | `internal/gateway/reload.go` → fsnotify watcher |

### 6.2 CLI

**Original:** `src/cli/` (Commander.js)

**Go implementation:**
- Use `cobra` for CLI framework
- Use `charmbracelet/huh` or `charmbracelet/bubbletea` for interactive prompts
- Gateway client via WebSocket

**Key mappings:**
| TypeScript | Go |
|------------|-----|
| `program.ts` → Commander program | `cmd/goclaw/main.go` → cobra root |
| `gateway-cli.ts` | `cmd/goclaw/gateway.go` |
| `daemon-cli.ts` | `cmd/goclaw/daemon.go` |
| `channels-cli.ts` | `cmd/goclaw/channels.go` |
| `nodes-cli.ts` | `cmd/goclaw/nodes.go` |
| `config-cli.ts` | `cmd/goclaw/config.go` |
| `skills-cli.ts` | `cmd/goclaw/skills.go` |

### 6.3 Agent Runtime

**Original:** `src/agents/` (Pi agent runtime)

**Go implementation:**
- Custom agent runtime with goroutine-based tool execution
- Streaming response handling via channels
- Tool registry with typed tool definitions
- Auth profile rotation
- Session management with compaction

**Key mappings:**
| TypeScript | Go |
|------------|-----|
| `pi-embedded-runner.ts` | `internal/agent/runner.go` |
| `pi-embedded-subscribe.ts` | `internal/agent/subscribe.go` |
| `pi-tools.ts` | `internal/agent/tools/registry.go` |
| `tools/*.ts` | `internal/agent/tools/*.go` |
| `bash-tools.ts` | `internal/agent/tools/bash.go` |
| `system-prompt.ts` | `internal/agent/prompt.go` |
| `auth-profiles.ts` | `internal/agent/auth.go` |
| `sandbox/` | `internal/agent/sandbox/` |

### 6.4 Channel Adapters

**Original:** `src/telegram/`, `src/discord/`, `src/slack/`, etc.

**Go implementation:**
- Each channel as a separate package implementing a common `Channel` interface
- Plugin-based channel loading

| TypeScript | Go |
|------------|-----|
| `src/telegram/` | `internal/channels/telegram/` |
| `src/discord/` | `internal/channels/discord/` |
| `src/slack/` | `internal/channels/slack/` |
| `src/whatsapp/` | `internal/channels/whatsapp/` |
| `src/signal/` | `internal/channels/signal/` |
| `src/imessage/` | `internal/channels/imessage/` |
| `src/channels/web/` | `internal/channels/webchat/` |

### 6.5 Configuration

**Original:** `src/config/` (Zod schemas, JSON5)

**Go implementation:**
- Go structs with JSON tags
- `go-playground/validator` for validation
- JSON5 parser for config file reading
- Environment variable overlay
- Config file watching via fsnotify

| TypeScript | Go |
|------------|-----|
| `config.ts` | `internal/config/loader.go` |
| `schema.ts` + `zod-schema.ts` | `internal/config/types.go` |
| `types.*.ts` (30+ files) | `internal/config/types/` |
| `defaults.ts` | `internal/config/defaults.go` |
| `validation.ts` | `internal/config/validate.go` |
| `sessions/` | `internal/config/sessions/` |
| `legacy.ts` | `internal/config/migrate.go` |

### 6.6 Auto-Reply Pipeline

**Original:** `src/auto-reply/`

**Go implementation:**
- Pipeline pattern with Go channels for streaming
- Command detection and routing
- Directive parsing
- Agent runner integration

| TypeScript | Go |
|------------|-----|
| `reply.ts` | `internal/reply/pipeline.go` |
| `reply/agent-runner.ts` | `internal/reply/agent.go` |
| `reply/commands.ts` | `internal/reply/commands.go` |
| `reply/directive-handling.ts` | `internal/reply/directives.go` |
| `chunk.ts` | `internal/reply/chunk.go` |
| `dispatch.ts` | `internal/reply/dispatch.go` |

### 6.7 Infrastructure

**Original:** `src/infra/`

**Go implementation:**
- Leverage Go stdlib heavily
- Custom retry with exponential backoff
- Goroutine-based heartbeat runner
- mDNS via `hashicorp/mdns`

| TypeScript | Go |
|------------|-----|
| `env.ts` | `internal/infra/env.go` |
| `retry.ts` | `internal/infra/retry.go` |
| `heartbeat-runner.ts` | `internal/infra/heartbeat.go` |
| `bonjour.ts` | `internal/infra/bonjour.go` |
| `device-identity.ts` | `internal/infra/identity.go` |
| `gateway-lock.ts` | `internal/infra/lock.go` |
| `tailscale.ts` | `internal/infra/tailscale.go` |
| `provider-usage.ts` | `internal/infra/usage.go` |
| `system-events.ts` | `internal/infra/events.go` |

### 6.8 Plugin System

**Original:** `src/plugins/`

**Go implementation:**
- Plugin interface with Go plugin system or hashicorp/go-plugin
- Plugin discovery from filesystem
- HTTP route registration
- Hook system

| TypeScript | Go |
|------------|-----|
| `discovery.ts` | `internal/plugins/discovery.go` |
| `loader.ts` | `internal/plugins/loader.go` |
| `registry.ts` | `internal/plugins/registry.go` |
| `runtime.ts` | `internal/plugins/runtime.go` |
| `hooks.ts` | `internal/plugins/hooks.go` |
| `manifest.ts` | `internal/plugins/manifest.go` |

### 6.9 Media Pipeline

**Original:** `src/media/`

**Go implementation:**
- Image processing via `disintegration/imaging`
- Temp file management
- Media hosting via HTTP

| TypeScript | Go |
|------------|-----|
| `store.ts` | `internal/media/store.go` |
| `fetch.ts` | `internal/media/fetch.go` |
| `host.ts` | `internal/media/host.go` |
| `server.ts` | `internal/media/server.go` |
| `image-ops.ts` | `internal/media/image.go` |
| `audio.ts` | `internal/media/audio.go` |

### 6.10 Browser Automation

**Original:** `src/browser/`

**Go implementation:**
- Use `chromedp/chromedp` for CDP-based browser control
- Or `go-rod/rod` for higher-level API

| TypeScript | Go |
|------------|-----|
| `server.ts` | `internal/browser/server.go` |
| `pw-session.ts` | `internal/browser/session.go` |
| `cdp.ts` | `internal/browser/cdp.go` |
| `chrome.ts` | `internal/browser/chrome.go` |
| `screenshot.ts` | `internal/browser/screenshot.go` |

---

## 7. Proposed Go Package Structure

```
github.com/StellariumFoundation/goclaw/
├── cmd/
│   └── goclaw/
│       ├── main.go              # Entry point
│       ├── root.go              # Root cobra command
│       ├── gateway.go           # gateway command
│       ├── agent.go             # agent command
│       ├── daemon.go            # daemon command
│       ├── channels.go          # channels command
│       ├── config.go            # config command
│       ├── nodes.go             # nodes command
│       ├── browser.go           # browser command
│       ├── skills.go            # skills command
│       ├── plugins.go           # plugins command
│       ├── pairing.go           # pairing command
│       ├── update.go            # update command
│       ├── onboard.go           # onboard command
│       ├── doctor.go            # doctor command
│       ├── message.go           # message command
│       ├── acp.go               # acp command
│       └── tui.go               # tui command
├── internal/
│   ├── gateway/
│   │   ├── server.go            # Main gateway server
│   │   ├── options.go           # Server options
│   │   ├── broadcast.go         # Event broadcasting
│   │   ├── lanes.go             # Concurrency lanes
│   │   ├── maintenance.go       # Maintenance timers
│   │   ├── reload.go            # Config hot-reload
│   │   ├── startup.go           # Startup sequence
│   │   ├── auth/
│   │   │   ├── auth.go          # Authentication
│   │   │   ├── ratelimit.go     # Rate limiting
│   │   │   └── tailscale.go     # Tailscale identity
│   │   ├── ws/
│   │   │   ├── connection.go    # WebSocket connection
│   │   │   ├── handler.go       # Message handler
│   │   │   └── types.go         # WS types
│   │   ├── http/
│   │   │   ├── server.go        # HTTP server
│   │   │   ├── health.go        # Health endpoint
│   │   │   ├── openai.go        # OpenAI-compatible API
│   │   │   ├── media.go         # Media endpoints
│   │   │   └── webhooks.go      # Webhook endpoints
│   │   ├── methods/
│   │   │   ├── registry.go      # Method registry
│   │   │   ├── agent.go         # Agent methods
│   │   │   ├── chat.go          # Chat methods
│   │   │   ├── config.go        # Config methods
│   │   │   ├── sessions.go      # Session methods
│   │   │   ├── channels.go      # Channel methods
│   │   │   ├── nodes.go         # Node methods
│   │   │   ├── models.go        # Model methods
│   │   │   ├── health.go        # Health methods
│   │   │   ├── cron.go          # Cron methods
│   │   │   ├── skills.go        # Skills methods
│   │   │   ├── browser.go       # Browser methods
│   │   │   ├── logs.go          # Log methods
│   │   │   ├── devices.go       # Device methods
│   │   │   ├── usage.go         # Usage methods
│   │   │   ├── wizard.go        # Wizard methods
│   │   │   ├── talk.go          # Talk methods
│   │   │   ├── tts.go           # TTS methods
│   │   │   └── update.go        # Update methods
│   │   └── protocol/
│   │       ├── types.go         # Protocol types
│   │       ├── agent.go         # Agent protocol
│   │       ├── channels.go      # Channel protocol
│   │       ├── config.go        # Config protocol
│   │       ├── sessions.go      # Session protocol
│   │       ├── nodes.go         # Node protocol
│   │       └── frames.go        # Frame types
│   ├── agent/
│   │   ├── runner.go            # Agent runner
│   │   ├── subscribe.go         # Event subscription
│   │   ├── prompt.go            # System prompt builder
│   │   ├── auth.go              # Auth profile management
│   │   ├── models.go            # Model configuration
│   │   ├── catalog.go           # Model catalog
│   │   ├── identity.go          # Agent identity
│   │   ├── lanes.go             # Agent lanes
│   │   ├── compact.go           # Session compaction
│   │   ├── history.go           # Session history
│   │   ├── tools/
│   │   │   ├── registry.go      # Tool registry
│   │   │   ├── bash.go          # Bash tool
│   │   │   ├── browser.go       # Browser tool
│   │   │   ├── canvas.go        # Canvas tool
│   │   │   ├── cron.go          # Cron tool
│   │   │   ├── discord.go       # Discord actions
│   │   │   ├── slack.go         # Slack actions
│   │   │   ├── telegram.go      # Telegram actions
│   │   │   ├── whatsapp.go      # WhatsApp actions
│   │   │   ├── gateway.go       # Gateway tool
│   │   │   ├── image.go         # Image tool
│   │   │   ├── memory.go        # Memory tool
│   │   │   ├── message.go       # Message tool
│   │   │   ├── nodes.go         # Nodes tool
│   │   │   ├── sessions.go      # Session tools
│   │   │   ├── tts.go           # TTS tool
│   │   │   ├── webfetch.go      # Web fetch tool
│   │   │   └── websearch.go     # Web search tool
│   │   └── sandbox/
│   │       ├── docker.go        # Docker sandbox
│   │       ├── config.go        # Sandbox config
│   │       ├── registry.go      # Sandbox registry
│   │       └── workspace.go     # Sandbox workspace
│   ├── channels/
│   │   ├── interface.go         # Channel interface
│   │   ├── registry.go          # Channel registry
│   │   ├── allowlist.go         # Allowlist matching
│   │   ├── mention.go           # Mention gating
│   │   ├── typing.go            # Typing indicators
│   │   ├── targets.go           # Target resolution
│   │   ├── session.go           # Channel session
│   │   ├── telegram/
│   │   │   ├── bot.go           # Telegram bot
│   │   │   ├── send.go          # Send messages
│   │   │   ├── monitor.go       # Update monitor
│   │   │   ├── format.go        # Message formatting
│   │   │   ├── download.go      # Media download
│   │   │   ├── webhook.go       # Webhook handling
│   │   │   └── accounts.go      # Account management
│   │   ├── discord/
│   │   │   ├── bot.go           # Discord bot
│   │   │   ├── send.go          # Send messages
│   │   │   ├── monitor.go       # Event monitor
│   │   │   ├── chunk.go         # Message chunking
│   │   │   └── accounts.go      # Account management
│   │   ├── slack/
│   │   │   ├── bot.go           # Slack bot
│   │   │   ├── send.go          # Send messages
│   │   │   ├── monitor.go       # Event monitor
│   │   │   └── accounts.go      # Account management
│   │   ├── whatsapp/
│   │   │   ├── client.go        # WhatsApp client
│   │   │   ├── send.go          # Send messages
│   │   │   ├── monitor.go       # Message monitor
│   │   │   └── accounts.go      # Account management
│   │   ├── signal/
│   │   │   ├── client.go        # Signal client
│   │   │   └── send.go          # Send messages
│   │   ├── webchat/
│   │   │   └── handler.go       # WebChat handler
│   │   └── imessage/
│   │       └── client.go        # iMessage client
│   ├── config/
│   │   ├── loader.go            # Config loading
│   │   ├── types.go             # Config types (main struct)
│   │   ├── defaults.go          # Default values
│   │   ├── validate.go          # Validation
│   │   ├── migrate.go           # Legacy migration
│   │   ├── paths.go             # Path resolution
│   │   ├── env.go               # Env var handling
│   │   ├── io.go                # File I/O
│   │   ├── types/
│   │   │   ├── agent.go         # Agent config types
│   │   │   ├── channels.go      # Channel config types
│   │   │   ├── gateway.go       # Gateway config types
│   │   │   ├── models.go        # Model config types
│   │   │   ├── tools.go         # Tool config types
│   │   │   ├── sandbox.go       # Sandbox config types
│   │   │   ├── skills.go        # Skills config types
│   │   │   └── ...              # Other config types
│   │   └── sessions/
│   │       ├── store.go         # Session store
│   │       ├── transcript.go    # Transcript management
│   │       ├── metadata.go      # Session metadata
│   │       └── types.go         # Session types
│   ├── reply/
│   │   ├── pipeline.go          # Reply pipeline
│   │   ├── agent.go             # Agent runner integration
│   │   ├── commands.go          # Chat commands
│   │   ├── directives.go        # Directive handling
│   │   ├── chunk.go             # Message chunking
│   │   ├── dispatch.go          # Message dispatch
│   │   ├── streaming.go         # Response streaming
│   │   └── formatting.go        # Reply formatting
│   ├── infra/
│   │   ├── env.go               # Environment utilities
│   │   ├── retry.go             # Retry with backoff
│   │   ├── heartbeat.go         # Heartbeat runner
│   │   ├── bonjour.go           # mDNS/Bonjour
│   │   ├── identity.go          # Device identity
│   │   ├── lock.go              # File/gateway locking
│   │   ├── tailscale.go         # Tailscale integration
│   │   ├── usage.go             # Usage tracking
│   │   ├── events.go            # System event bus
│   │   ├── presence.go          # System presence
│   │   ├── update.go            # Update checking
│   │   ├── archive.go           # Archive utilities
│   │   ├── fetch.go             # HTTP fetch
│   │   ├── ssrf.go              # SSRF protection
│   │   └── tls.go               # TLS utilities
│   ├── browser/
│   │   ├── server.go            # Browser control server
│   │   ├── session.go           # Browser session
│   │   ├── cdp.go               # CDP client
│   │   ├── chrome.go            # Chrome management
│   │   ├── screenshot.go        # Screenshot capture
│   │   └── profiles.go          # Browser profiles
│   ├── media/
│   │   ├── store.go             # Media store
│   │   ├── fetch.go             # Media fetching
│   │   ├── host.go              # Media hosting
│   │   ├── server.go            # Media HTTP server
│   │   ├── image.go             # Image operations
│   │   └── audio.go             # Audio processing
│   ├── plugins/
│   │   ├── discovery.go         # Plugin discovery
│   │   ├── loader.go            # Plugin loading
│   │   ├── registry.go          # Plugin registry
│   │   ├── runtime.go           # Plugin runtime
│   │   ├── hooks.go             # Hook system
│   │   └── manifest.go          # Manifest parsing
│   ├── hooks/
│   │   ├── runner.go            # Hook runner
│   │   ├── types.go             # Hook types
│   │   ├── gmail.go             # Gmail hook
│   │   └── bundled/             # Bundled hooks
│   ├── logging/
│   │   ├── logger.go            # Structured logger
│   │   ├── redact.go            # Log redaction
│   │   ├── subsystem.go         # Subsystem logger
│   │   └── config.go            # Log config
│   ├── tts/
│   │   ├── tts.go               # TTS engine
│   │   └── edge.go              # Edge TTS
│   ├── wizard/
│   │   ├── onboarding.go        # Onboarding wizard
│   │   ├── session.go           # Wizard session
│   │   └── prompts.go           # Wizard prompts
│   ├── cron/
│   │   └── service.go           # Cron service
│   ├── daemon/
│   │   ├── install.go           # Daemon installation
│   │   ├── lifecycle.go         # Start/stop/status
│   │   └── systemd.go           # systemd integration
│   ├── routing/
│   │   ├── resolve.go           # Route resolution
│   │   ├── bindings.go          # Channel bindings
│   │   └── session_key.go       # Session key resolution
│   ├── security/
│   │   └── ssrf.go              # SSRF protection
│   └── acp/
│       └── bridge.go            # ACP bridge
├── pkg/
│   ├── protocol/
│   │   └── types.go             # Public protocol types
│   └── sdk/
│       └── plugin.go            # Plugin SDK
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── README.md
└── rewrite.md
```

---

## 8. Implementation Priority

### Phase 1: Core Foundation (Weeks 1–3)

1. **Configuration system** (`internal/config/`)
   - Config types, loader, defaults, validation, paths
   - JSON5 parsing, env var overlay
   - Session store basics

2. **Logging** (`internal/logging/`)
   - Structured logger (slog-based)
   - Redaction, subsystem loggers

3. **Infrastructure** (`internal/infra/`)
   - Env utilities, retry, file locking
   - Device identity, paths

4. **CLI skeleton** (`cmd/goclaw/`)
   - Cobra root command, version, help
   - Config command (get/set/list)

### Phase 2: Gateway Server (Weeks 4–7)

5. **Gateway WebSocket server** (`internal/gateway/`)
   - HTTP server with WebSocket upgrade
   - Connection management, auth
   - Method registry and dispatch
   - Event broadcasting

6. **Gateway protocol** (`internal/gateway/protocol/`)
   - All protocol types as Go structs
   - JSON serialization/deserialization

7. **Gateway methods** (`internal/gateway/methods/`)
   - Health, config, sessions, channels, models
   - Connect, system, logs

8. **Gateway CLI** (`cmd/goclaw/gateway.go`)
   - `goclaw gateway` command with all flags

### Phase 3: Agent Runtime (Weeks 8–11)

9. **Agent runner** (`internal/agent/`)
   - Core agent loop with streaming
   - System prompt construction
   - Auth profile management
   - Model selection and failover

10. **Agent tools** (`internal/agent/tools/`)
    - Bash tool (process execution)
    - Web fetch/search tools
    - Session tools (list, history, send, spawn)
    - Message tool

11. **Auto-reply pipeline** (`internal/reply/`)
    - Inbound message processing
    - Command detection and handling
    - Directive parsing
    - Agent runner integration
    - Response chunking and delivery

### Phase 4: Channel Adapters (Weeks 12–16)

12. **Channel interface** (`internal/channels/interface.go`)
    - Common channel interface
    - Registry, allowlist, mention gating

13. **Telegram adapter** (`internal/channels/telegram/`)
    - Bot creation, update handling
    - Message sending, media download
    - Webhook support

14. **Discord adapter** (`internal/channels/discord/`)
    - Bot connection, event handling
    - Message sending, chunking

15. **Slack adapter** (`internal/channels/slack/`)
    - Bolt-equivalent event handling
    - Message sending

16. **WhatsApp adapter** (`internal/channels/whatsapp/`)
    - WhatsApp Web connection (whatsmeow)
    - Message sending/receiving

17. **WebChat** (`internal/channels/webchat/`)
    - WebSocket-based chat

### Phase 5: Advanced Features (Weeks 17–22)

18. **Browser automation** (`internal/browser/`)
    - Chrome management, CDP control
    - Screenshot, navigation, interaction

19. **Media pipeline** (`internal/media/`)
    - Image processing, audio handling
    - Media store, hosting

20. **Plugin system** (`internal/plugins/`)
    - Plugin discovery, loading
    - Hook system, HTTP routes

21. **Cron system** (`internal/cron/`)
    - Cron job scheduling and execution

22. **TTS** (`internal/tts/`)
    - Edge TTS, ElevenLabs integration

23. **Onboarding wizard** (`internal/wizard/`)
    - Interactive setup flow

### Phase 6: Operations & Polish (Weeks 23–26)

24. **Daemon management** (`internal/daemon/`)
    - systemd/launchd service installation
    - Start/stop/status

25. **Bonjour/mDNS** (`internal/infra/bonjour.go`)
    - Service discovery

26. **Tailscale integration** (`internal/infra/tailscale.go`)
    - Serve/Funnel automation

27. **Update system** (`internal/infra/update.go`)
    - Self-update mechanism

28. **Doctor command** — Diagnostic checks

29. **Signal, iMessage, Matrix** adapters

30. **Extension system** — Port key extensions

---

## 9. Technical Considerations

### 9.1 Concurrency Model

**TypeScript (original):** Single-threaded event loop with async/await. Concurrency via Promise.all, setTimeout, and event emitters.

**Go (target):**
- **Goroutine per WebSocket connection** — each client gets its own goroutine for reading/writing
- **Channel-based event broadcasting** — use Go channels for fan-out to connected clients
- **Context-based cancellation** — `context.Context` for request lifecycle, agent run cancellation
- **sync.Mutex / sync.RWMutex** — for shared state (session store, config, node registry)
- **Worker pools** — for agent runs, tool execution (bounded concurrency via semaphores)
- **errgroup** — for coordinated goroutine lifecycle in server startup/shutdown

```go
// Example: Gateway server with graceful shutdown
func (s *Server) Run(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)
    
    g.Go(func() error { return s.httpServer.ListenAndServe() })
    g.Go(func() error { return s.runHeartbeat(ctx) })
    g.Go(func() error { return s.runConfigWatcher(ctx) })
    g.Go(func() error { return s.runMaintenanceTimers(ctx) })
    
    <-ctx.Done()
    return g.Wait()
}
```

### 9.2 Error Handling

**TypeScript (original):** try/catch with custom error classes, unhandled rejection handlers.

**Go (target):**
- **Explicit error returns** — `(result, error)` pattern throughout
- **Custom error types** — `GatewayError`, `ChannelError`, `AgentError` with error codes
- **Error wrapping** — `fmt.Errorf("failed to start channel: %w", err)`
- **Sentinel errors** — `var ErrSessionNotFound = errors.New("session not found")`
- **Panic recovery** — middleware for HTTP/WS handlers

```go
type GatewayError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Cause   error  `json:"-"`
}

func (e *GatewayError) Error() string { return e.Message }
func (e *GatewayError) Unwrap() error { return e.Cause }
```

### 9.3 Testing Strategy

**TypeScript (original):** Vitest with colocated `*.test.ts` files, e2e tests, live tests.

**Go (target):**
- **Unit tests** — `*_test.go` colocated with source (Go convention)
- **Table-driven tests** — standard Go pattern for parameterized tests
- **Integration tests** — build tag `//go:build integration`
- **E2E tests** — build tag `//go:build e2e`
- **Test helpers** — `internal/testutil/` package
- **Mocking** — interfaces + mock implementations (no framework needed)
- **Coverage** — `go test -cover ./...`
- **Race detection** — `go test -race ./...`

```go
func TestSessionKeyResolution(t *testing.T) {
    tests := []struct {
        name     string
        channel  string
        sender   string
        expected string
    }{
        {"telegram DM", "telegram", "user123", "telegram:dm:user123"},
        {"discord group", "discord", "guild:chan", "discord:group:guild:chan"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ResolveSessionKey(tt.channel, tt.sender)
            if got != tt.expected {
                t.Errorf("got %q, want %q", got, tt.expected)
            }
        })
    }
}
```

### 9.4 Build & Distribution

**TypeScript (original):** tsdown → dist/, npm publish, Docker image.

**Go (target):**
- **Single binary** — `go build -o goclaw ./cmd/goclaw`
- **Cross-compilation** — `GOOS=linux GOARCH=amd64 go build`
- **Version embedding** — `go build -ldflags "-X main.version=..."` 
- **Docker** — Multi-stage build (build stage + scratch/distroless)
- **Release** — GoReleaser for multi-platform binaries + checksums
- **Install** — `go install github.com/StellariumFoundation/goclaw/cmd/goclaw@latest`

```dockerfile
# Multi-stage Dockerfile
FROM golang:1.26 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /goclaw ./cmd/goclaw

FROM gcr.io/distroless/static-debian12
COPY --from=builder /goclaw /goclaw
ENTRYPOINT ["/goclaw"]
CMD ["gateway"]
```

### 9.5 Configuration Compatibility

The Go rewrite should maintain **full backward compatibility** with existing `~/.openclaw/openclaw.json` config files. This means:
- Parse JSON5 format
- Support all existing config keys
- Maintain the same default values
- Support environment variable overrides with the same names
- Support config file includes

### 9.6 Protocol Compatibility

The WebSocket protocol must remain **wire-compatible** so existing clients (macOS app, iOS app, Control UI) can connect to the Go gateway without changes. This means:
- Same JSON-RPC message format
- Same method names and parameters
- Same event names and payloads
- Same auth flow

### 9.7 Embedded Assets

The Go binary should embed static assets:
- Control UI (HTML/JS/CSS) via `embed.FS`
- Default skill files
- Template files (AGENTS.md, SOUL.md, TOOLS.md)
- Chrome extension assets

```go
//go:embed assets/control-ui/*
var controlUIFS embed.FS

//go:embed assets/templates/*
var templatesFS embed.FS
```

### 9.8 Database / State Storage

**TypeScript (original):** JSON files on disk (`~/.openclaw/`), SQLite (sqlite-vec for memory).

**Go (target):**
- JSON files for config and session state (maintain compatibility)
- `modernc.org/sqlite` (pure Go SQLite) or `mattn/go-sqlite3` for memory/vector storage
- Consider `bbolt` for high-performance key-value storage

### 9.9 WebSocket Implementation Details

The gateway WebSocket server is the most critical component. Key requirements:
- Support thousands of concurrent connections
- JSON message framing with method dispatch
- Binary message support for media
- Ping/pong keepalive
- Graceful connection draining on shutdown
- Per-connection auth state
- Subscription management (events, node subscriptions)

### 9.10 Agent Runtime Architecture

The agent runtime is the most complex subsystem. Key considerations:
- **Streaming responses** — Use Go channels for token-by-token streaming from LLM APIs
- **Tool execution** — Goroutine-based with context cancellation
- **Session management** — Thread-safe session state with compaction
- **Multi-provider support** — Interface-based provider abstraction
- **Auth rotation** — Automatic failover between auth profiles
- **Sandbox execution** — Docker container management for untrusted code

---

## 10. Migration Strategy

### 10.1 Phased Rollout

1. **Phase A: Parallel Development** — Build GoClaw alongside OpenClaw. Both run simultaneously.
2. **Phase B: Feature Parity** — GoClaw reaches feature parity for core gateway + top channels.
3. **Phase C: Beta Testing** — Users can opt into GoClaw via `goclaw gateway` instead of `openclaw gateway`.
4. **Phase D: Migration** — Provide migration tool that validates config compatibility.
5. **Phase E: Default** — GoClaw becomes the default, OpenClaw enters maintenance mode.

### 10.2 Compatibility Testing

- **Config compatibility tests** — Parse all known config patterns from OpenClaw test suite
- **Protocol compatibility tests** — Record/replay WebSocket sessions from OpenClaw
- **Channel integration tests** — Verify each channel adapter against real APIs
- **Agent behavior tests** — Compare agent responses between TypeScript and Go implementations

### 10.3 Data Migration

User data lives in `~/.openclaw/`:
- `openclaw.json` — Config file (read directly, no migration needed)
- `credentials/` — Channel credentials (read directly)
- `sessions/` — Session transcripts (JSON, read directly)
- `workspace/` — Agent workspace (filesystem, no migration)
- `skills/` — Installed skills (filesystem, no migration)
- `plugins/` — Installed plugins (may need re-installation for Go plugin format)

### 10.4 Extension Migration

Extensions are currently TypeScript/Node.js packages. Options:
1. **Go plugins** — Rewrite extensions in Go (preferred for core channels)
2. **Process-based plugins** — Run TypeScript extensions as child processes with JSON-RPC
3. **WASM plugins** — Compile extensions to WASM (future consideration)
4. **HTTP plugins** — Extensions expose HTTP endpoints (already partially supported)

Recommended approach: Rewrite core channel extensions in Go, support process-based plugins for community extensions.

### 10.5 Native App Compatibility

The macOS/iOS/Android apps communicate via WebSocket. As long as the Go gateway maintains protocol compatibility, no changes are needed to the native apps. The `OpenClawKit` Swift package defines the protocol types — these must match exactly.

### 10.6 Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Protocol incompatibility | Record/replay testing against TypeScript gateway |
| Missing channel features | Prioritize channels by user count (Telegram > Discord > Slack > WhatsApp) |
| Agent behavior differences | Extensive comparison testing with same prompts/tools |
| Performance regression | Benchmark critical paths (WS throughput, agent latency) |
| Plugin ecosystem breakage | Support process-based plugins as bridge |
| Config migration failures | Comprehensive config parsing test suite |

### 10.7 Success Criteria

- [ ] All gateway WS methods implemented and protocol-compatible
- [ ] Top 4 channels (Telegram, Discord, Slack, WhatsApp) fully functional
- [ ] Agent runtime with all core tools working
- [ ] CLI feature parity for common commands
- [ ] Config file backward compatibility
- [ ] Single binary < 50MB
- [ ] Startup time < 500ms
- [ ] Memory usage < 100MB idle
- [ ] All existing macOS/iOS apps connect without changes
- [ ] Docker image < 30MB (distroless)

---

*This document serves as the complete blueprint for rewriting OpenClaw from TypeScript/Node.js to Go 1.26.0 as GoClaw. It should be updated as implementation progresses and new insights emerge.*
