# OpenClaw → GoClaw: Architecture Analysis & Go Rewrite Blueprint

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Directory & File Structure](#2-directory--file-structure)
3. [Core Modules & Responsibilities](#3-core-modules--responsibilities)
4. [Data Flow & Control Flow](#4-data-flow--control-flow)
5. [External Dependencies & Integrations](#5-external-dependencies--integrations)
6. [Configuration & Environment Setup](#6-configuration--environment-setup)
7. [Entry Points & CLI Commands](#7-entry-points--cli-commands)
8. [Key Classes, Functions & Interfaces](#8-key-classes-functions--interfaces)
9. [State Management & Persistence](#9-state-management--persistence)
10. [Error Handling Patterns](#10-error-handling-patterns)
11. [Tests & CI/CD](#11-tests--cicd)
12. [Go Rewrite Plan](#12-go-rewrite-plan)
13. [Migration Strategy](#13-migration-strategy)
14. [Challenges & Design Decisions](#14-challenges--design-decisions)

---

## 1. Project Overview

**OpenClaw** is a multi-channel AI gateway and personal AI assistant platform built in TypeScript (ESM, Node.js ≥22). It acts as a unified control plane that:

- Connects to **multiple messaging channels** (WhatsApp, Telegram, Slack, Discord, Signal, iMessage, Google Chat, Microsoft Teams, Matrix, IRC, LINE, Nostr, Twitch, Zalo, BlueBubbles, Nextcloud Talk, Mattermost, Tlon)
- Routes messages to **AI model providers** (OpenAI, Anthropic Claude, Google Gemini, Bedrock, Ollama, OpenRouter, GitHub Copilot, HuggingFace, vLLM, LiteLLM, MiniMax, Qwen, and more)
- Provides an **agent loop** with tool use, session management, context compaction, and multi-agent orchestration
- Supports **browser automation** via Playwright/CDP for web interaction tools
- Offers a **plugin/extension system** for adding channels, tools, and capabilities
- Includes a **cron scheduler** for periodic agent tasks
- Provides **TUI** (terminal UI), **Web UI**, and native **macOS/iOS/Android** apps
- Implements the **Agent Client Protocol (ACP)** for standardized agent communication
- Features **skills** (markdown-based tool definitions), **hooks** (event-driven automation), and **memory** (vector search via sqlite-vec/LanceDB)

The system is designed as a single-user, self-hosted assistant that runs as a background daemon (launchd on macOS, systemd on Linux).

---

## 2. Directory & File Structure

```
Openclaw/
├── openclaw.mjs              # CLI entry point (Node.js shim → dist/entry.js)
├── package.json              # Dependencies, scripts, bin config
├── tsdown.config.ts          # Build configuration (tsdown bundler)
├── tsconfig.json             # TypeScript config
├── vitest.config.ts          # Test config (+ e2e, unit, live, gateway, extensions variants)
├── Dockerfile                # Production Docker image
├── docker-compose.yml        # Docker Compose for gateway + CLI
├── fly.toml                  # Fly.io deployment config
├── render.yaml               # Render deployment config
├── .env.example              # Environment variable reference
├── pnpm-workspace.yaml       # pnpm workspace (monorepo)
│
├── src/                      # Main TypeScript source
│   ├── entry.ts              # Build entry point
│   ├── index.ts              # Library exports
│   ├── runtime.ts            # Runtime environment abstraction
│   ├── extensionAPI.ts       # Extension API exports
│   │
│   ├── cli/                  # CLI framework (Commander.js)
│   │   ├── program/          # Command registration & routing
│   │   ├── gateway-cli/      # Gateway subcommands
│   │   ├── daemon-cli/       # Daemon management subcommands
│   │   ├── nodes-cli/        # Node management subcommands
│   │   ├── cron-cli/         # Cron subcommands
│   │   ├── browser-cli*.ts   # Browser automation CLI
│   │   └── *.ts              # Individual CLI modules
│   │
│   ├── commands/             # Command implementations
│   │   ├── agent.ts          # Agent command (core AI interaction)
│   │   ├── onboard.ts        # Onboarding wizard
│   │   ├── configure.ts      # Configuration wizard
│   │   ├── doctor.ts         # Diagnostic/repair tool
│   │   ├── health.ts         # Health check
│   │   ├── status.ts         # Status reporting
│   │   ├── message.ts        # Message send/receive
│   │   ├── models.ts         # Model management
│   │   ├── channels.ts       # Channel management
│   │   ├── sessions.ts       # Session management
│   │   ├── sandbox.ts        # Sandbox management
│   │   └── agent/            # Agent command helpers
│   │
│   ├── gateway/              # Gateway server (WebSocket + HTTP)
│   │   ├── server-startup.ts # Server initialization & sidecars
│   │   ├── server-http.ts    # HTTP routes (Express)
│   │   ├── server-chat.ts    # Chat/conversation handling
│   │   ├── server-channels.ts# Channel lifecycle
│   │   ├── server-cron.ts    # Cron integration
│   │   ├── server-plugins.ts # Plugin lifecycle
│   │   ├── server-discovery.ts# mDNS/Bonjour discovery
│   │   ├── server-broadcast.ts# Event broadcasting
│   │   ├── server-browser.ts # Browser control server
│   │   ├── openai-http.ts    # OpenAI-compatible HTTP API
│   │   ├── openresponses-http.ts # Open Responses API
│   │   ├── boot.ts           # Boot sequence (BOOT.md execution)
│   │   ├── auth.ts           # Gateway authentication
│   │   ├── client.ts         # Gateway WebSocket client
│   │   ├── protocol/         # WebSocket protocol schema (TypeBox)
│   │   ├── server/           # Server internals (TLS, WS, health)
│   │   └── server-methods/   # RPC method handlers
│   │
│   ├── agents/               # AI agent system
│   │   ├── pi-embedded-runner.ts  # Pi agent runner (embedded mode)
│   │   ├── pi-embedded-subscribe.ts # Stream subscription
│   │   ├── system-prompt.ts  # System prompt construction
│   │   ├── compaction.ts     # Context window compaction
│   │   ├── model-catalog.ts  # Model registry & capabilities
│   │   ├── model-selection.ts# Model resolution & failover
│   │   ├── model-fallback.ts # Failover chain
│   │   ├── auth-profiles.ts  # Auth profile rotation
│   │   ├── cli-runner.ts     # CLI agent runner
│   │   ├── skills.ts         # Skills loading & prompt injection
│   │   ├── identity.ts       # Agent identity management
│   │   ├── bash-tools.*.ts   # Shell/bash tool implementations
│   │   ├── sandbox/          # Sandbox (Docker) execution
│   │   ├── tools/            # Agent tool definitions
│   │   └── schema/           # Agent data schemas
│   │
│   ├── config/               # Configuration system
│   │   ├── config.ts         # Config loading/writing (JSON5)
│   │   ├── schema.ts         # Config schema (TypeBox)
│   │   ├── zod-schema.ts     # Config validation (Zod)
│   │   ├── paths.ts          # Config file paths
│   │   ├── types.ts          # Config type definitions
│   │   ├── legacy*.ts        # Legacy config migration
│   │   ├── sessions/         # Session store & metadata
│   │   └── env-*.ts          # Environment variable handling
│   │
│   ├── channels/             # Channel abstraction layer
│   │   ├── registry.ts       # Channel registry
│   │   ├── plugins/          # Channel plugin interfaces
│   │   ├── allowlists/       # Sender allowlist logic
│   │   └── web/              # Web channel (WhatsApp Web)
│   │
│   ├── plugins/              # Plugin system
│   │   ├── loader.ts         # Plugin discovery & loading
│   │   ├── registry.ts       # Plugin registry
│   │   ├── services.ts       # Plugin service lifecycle
│   │   ├── hooks.ts          # Plugin hook wiring
│   │   ├── tools.ts          # Plugin tool registration
│   │   ├── install.ts        # Plugin installation
│   │   └── runtime/          # Plugin runtime environment
│   │
│   ├── cron/                 # Cron scheduler
│   │   ├── service.ts        # Cron service (timer management)
│   │   ├── schedule.ts       # Schedule parsing (croner)
│   │   ├── store.ts          # Cron job persistence
│   │   ├── delivery.ts       # Cron result delivery
│   │   ├── isolated-agent/   # Isolated agent execution
│   │   └── service/          # Service internals
│   │
│   ├── hooks/                # Event hooks system
│   │   ├── internal-hooks.ts # Internal hook registry
│   │   ├── loader.ts         # Hook loading
│   │   ├── gmail*.ts         # Gmail integration hooks
│   │   └── bundled/          # Built-in hooks
│   │
│   ├── browser/              # Browser automation
│   │   ├── pw-session.ts     # Playwright session management
│   │   ├── pw-tools-core.ts  # Browser tool implementations
│   │   ├── cdp.ts            # Chrome DevTools Protocol
│   │   ├── server.ts         # Browser control HTTP server
│   │   ├── chrome.ts         # Chrome process management
│   │   └── profiles*.ts      # Browser profile management
│   │
│   ├── infra/                # Infrastructure & utilities
│   │   ├── fetch.ts          # HTTP client (undici)
│   │   ├── retry.ts          # Retry logic
│   │   ├── gateway-lock.ts   # File-based gateway lock
│   │   ├── heartbeat-runner.ts # Heartbeat system
│   │   ├── bonjour*.ts       # mDNS/Bonjour discovery
│   │   ├── tailscale.ts      # Tailscale integration
│   │   ├── outbound/         # Outbound message delivery
│   │   ├── net/              # Network security (SSRF protection)
│   │   ├── tls/              # TLS utilities
│   │   └── *.ts              # Various infrastructure utilities
│   │
│   ├── logging/              # Logging system (tslog)
│   │   ├── logger.ts         # Logger factory
│   │   ├── subsystem.ts      # Subsystem-scoped loggers
│   │   ├── redact.ts         # Secret redaction
│   │   └── console.ts        # Console output formatting
│   │
│   ├── media/                # Media processing
│   │   ├── store.ts          # Media file storage
│   │   ├── fetch.ts          # Media download
│   │   ├── image-ops.ts      # Image processing (sharp)
│   │   └── audio.ts          # Audio processing
│   │
│   ├── tts/                  # Text-to-speech
│   │   └── tts.ts            # TTS engine (Edge TTS, ElevenLabs)
│   │
│   ├── tui/                  # Terminal UI
│   │   ├── tui.ts            # TUI main loop
│   │   ├── components/       # TUI components
│   │   └── theme/            # TUI theming
│   │
│   ├── acp/                  # Agent Client Protocol
│   │   ├── server.ts         # ACP server
│   │   ├── client.ts         # ACP client
│   │   └── session.ts        # ACP session mapping
│   │
│   ├── sessions/             # Session management
│   ├── routing/              # Message routing
│   ├── security/             # Security auditing
│   ├── markdown/             # Markdown processing
│   ├── terminal/             # Terminal utilities
│   ├── daemon/               # Daemon management (launchd/systemd)
│   ├── pairing/              # Device pairing
│   ├── wizard/               # Setup wizards
│   ├── auto-reply/           # Auto-reply logic
│   ├── providers/            # Provider-specific auth
│   ├── plugin-sdk/           # Plugin SDK exports
│   │
│   ├── telegram/             # Telegram channel (grammy)
│   ├── discord/              # Discord channel (@buape/carbon)
│   ├── slack/                # Slack channel (@slack/bolt)
│   ├── signal/               # Signal channel
│   ├── imessage/             # iMessage channel
│   ├── whatsapp/             # WhatsApp channel (Baileys)
│   ├── line/                 # LINE channel
│   └── web/                  # WhatsApp Web channel
│
├── extensions/               # Extension plugins (workspace packages)
│   ├── bluebubbles/          # BlueBubbles (iMessage bridge)
│   ├── discord/              # Discord extension
│   ├── telegram/             # Telegram extension
│   ├── slack/                # Slack extension
│   ├── signal/               # Signal extension
│   ├── whatsapp/             # WhatsApp extension
│   ├── matrix/               # Matrix extension
│   ├── msteams/              # Microsoft Teams extension
│   ├── googlechat/           # Google Chat extension
│   ├── irc/                  # IRC extension
│   ├── mattermost/           # Mattermost extension
│   ├── nostr/                # Nostr extension
│   ├── twitch/               # Twitch extension
│   ├── zalo/                 # Zalo extension
│   ├── zalouser/             # Zalo Personal extension
│   ├── line/                 # LINE extension
│   ├── feishu/               # Feishu/Lark extension
│   ├── tlon/                 # Tlon extension
│   ├── nextcloud-talk/       # Nextcloud Talk extension
│   ├── memory-core/          # Memory (vector search) core
│   ├── memory-lancedb/       # Memory with LanceDB backend
│   ├── voice-call/           # Voice call extension
│   ├── talk-voice/           # Talk voice extension
│   ├── open-prose/           # Prose generation extension
│   ├── llm-task/             # LLM task extension
│   ├── lobster/              # Lobster UI extension
│   ├── copilot-proxy/        # GitHub Copilot proxy
│   ├── device-pair/          # Device pairing extension
│   ├── diagnostics-otel/     # OpenTelemetry diagnostics
│   ├── thread-ownership/     # Thread ownership extension
│   ├── phone-control/        # Phone control extension
│   └── *-auth/               # Various auth provider extensions
│
├── skills/                   # Skill definitions (markdown + tools)
│   ├── github/               # GitHub integration skill
│   ├── slack/                # Slack skill
│   ├── discord/              # Discord skill
│   ├── coding-agent/         # Coding agent skill
│   ├── weather/              # Weather skill
│   ├── spotify-player/       # Spotify skill
│   ├── obsidian/             # Obsidian notes skill
│   ├── notion/               # Notion skill
│   ├── trello/               # Trello skill
│   └── ...                   # 50+ skill definitions
│
├── packages/                 # Internal packages
│   ├── clawdbot/             # ClawdBot package
│   └── moltbot/              # MoltBot package
│
├── apps/                     # Native applications
│   ├── macos/                # macOS app (Swift/SwiftUI)
│   ├── ios/                  # iOS app (Swift/SwiftUI)
│   ├── android/              # Android app (Kotlin)
│   └── shared/               # Shared native code (OpenClawKit)
│
├── ui/                       # Web UI (Lit, Vite)
│   ├── src/                  # UI source
│   └── package.json          # UI dependencies
│
├── docs/                     # Documentation (Mintlify)
├── scripts/                  # Build/dev scripts
├── test/                     # Test fixtures & helpers
├── .github/                  # CI/CD workflows
└── .agents/                  # Agent configuration
```

---

## 3. Core Modules & Responsibilities

### 3.1 CLI (`src/cli/`)

The CLI is built on **Commander.js** and serves as the primary user interface. Key components:

- **`program/build-program.ts`**: Constructs the Commander program, registers all commands
- **`program/command-registry.ts`**: Central command registration (agent, gateway, onboard, configure, status, health, message, models, channels, sessions, cron, browser, daemon, nodes, plugins, skills, hooks, etc.)
- **`deps.ts`**: Dependency injection container (`CliDeps`) for testability
- **`gateway-cli/`**: Gateway start/stop/discover/call subcommands
- **`daemon-cli/`**: Daemon install/uninstall/status/restart
- **`run-main.ts`**: Top-level CLI runner with error handling

### 3.2 Gateway Server (`src/gateway/`)

The gateway is the core runtime — an **Express HTTP server** + **WebSocket server** that:

- Listens on a configurable port (default 18789)
- Authenticates clients via token or password
- Manages channel connections (start/stop/monitor)
- Routes inbound messages to the agent loop
- Broadcasts events to connected clients (native apps, TUI, web UI)
- Exposes an OpenAI-compatible HTTP API
- Runs cron jobs, hooks, and plugin services
- Provides mDNS/Bonjour discovery for local network

Key files:
- **`server-http.ts`**: Express route setup
- **`server-chat.ts`**: Chat message handling
- **`server-methods/`**: RPC method handlers (agent, chat, config, channels, sessions, etc.)
- **`protocol/`**: WebSocket protocol schema using TypeBox
- **`auth.ts`**: Token/password authentication with rate limiting
- **`client.ts`**: WebSocket client for connecting to remote gateways

### 3.3 Agent System (`src/agents/`)

The agent system is the AI brain. It:

- Runs an **embedded Pi agent** (`pi-embedded-runner.ts`) — a loop that sends messages to LLM providers and processes tool calls
- Manages **system prompts** (`system-prompt.ts`) with dynamic context injection
- Handles **context compaction** (`compaction.ts`) when conversations exceed model limits
- Supports **model failover** (`model-fallback.ts`) across providers
- Manages **auth profiles** (`auth-profiles.ts`) for OAuth/API key rotation
- Provides **tool definitions** (`tools/`) including bash execution, file operations, browser control, messaging, and more
- Supports **multi-agent** orchestration with subagent spawning
- Loads **skills** (`skills.ts`) from markdown files that define tool schemas and instructions
- Runs in **sandbox** mode (`sandbox/`) via Docker for isolated execution

### 3.4 Configuration (`src/config/`)

Configuration is stored as **JSON5** at `~/.openclaw/openclaw.json`:

- **`schema.ts`**: Full config schema using TypeBox
- **`zod-schema.ts`**: Zod validation schemas
- **`types.ts`**: TypeScript type definitions for all config sections
- **`io.ts`**: Config file read/write with caching
- **`legacy*.ts`**: Migration from older config formats
- **`sessions/`**: Session store (JSON file-based)
- **`env-vars.ts`**: Environment variable mapping
- **`defaults.ts`**: Default configuration values

Config sections include: gateway, channels (telegram, discord, slack, signal, whatsapp, etc.), models, agents, hooks, cron, browser, sandbox, tools, memory, TTS, and more.

### 3.5 Channel System (`src/channels/` + channel-specific dirs)

Channels are abstracted through a **plugin interface** (`channels/plugins/types.ts`):

- **`ChannelPlugin`**: Main plugin interface with adapters for auth, messaging, outbound, directory, status, setup, pairing, etc.
- **`registry.ts`**: Channel registry for discovering and managing channels
- Built-in channels: WhatsApp Web (`src/web/`), Telegram (`src/telegram/`), Discord (`src/discord/`), Slack (`src/slack/`), Signal (`src/signal/`), iMessage (`src/imessage/`), LINE (`src/line/`)
- Extension channels: loaded via the plugin system from `extensions/`

### 3.6 Plugin System (`src/plugins/`)

Plugins extend OpenClaw with new channels, tools, hooks, and providers:

- **`loader.ts`**: Discovers plugins from `extensions/` and installed npm packages
- **`registry.ts`**: Plugin registry with lifecycle management
- **`services.ts`**: Plugin service start/stop
- **`hooks.ts`**: Plugin hook wiring (before/after tool call, message, session, compaction, gateway events)
- **`tools.ts`**: Plugin tool registration
- **`install.ts`**: Plugin installation via npm
- **Plugin SDK** (`src/plugin-sdk/`): Exported types and utilities for plugin authors

### 3.7 Cron Scheduler (`src/cron/`)

The cron system runs periodic agent tasks:

- **`service.ts`**: Main cron service with timer management
- **`schedule.ts`**: Cron expression parsing (via `croner` library)
- **`store.ts`**: Persistent cron job storage
- **`delivery.ts`**: Result delivery to channels
- **`isolated-agent/`**: Runs agent in isolated context for cron jobs

### 3.8 Browser Automation (`src/browser/`)

Browser automation for web interaction tools:

- **`pw-session.ts`**: Playwright session management
- **`pw-tools-core.ts`**: Tool implementations (click, type, screenshot, evaluate, navigate, etc.)
- **`cdp.ts`**: Chrome DevTools Protocol integration
- **`chrome.ts`**: Chrome process discovery and management
- **`server.ts`**: HTTP control server for browser operations
- **`profiles*.ts`**: Browser profile management

### 3.9 Infrastructure (`src/infra/`)

Shared infrastructure utilities:

- **`fetch.ts`**: HTTP client with retry, timeout, and SSRF protection
- **`retry.ts`**: Configurable retry with exponential backoff
- **`gateway-lock.ts`**: File-based lock to prevent multiple gateway instances
- **`heartbeat-runner.ts`**: Periodic heartbeat system for channel health
- **`bonjour*.ts`**: mDNS/Bonjour service discovery
- **`tailscale.ts`**: Tailscale VPN integration
- **`outbound/`**: Message delivery pipeline (queue, envelope, format, target resolution)
- **`net/`**: Network security (SSRF pinning, fetch guards)
- **`provider-usage.*`**: Usage tracking per provider

### 3.10 Logging (`src/logging/`)

Structured logging via **tslog**:

- **`subsystem.ts`**: Subsystem-scoped loggers (e.g., `gateway/boot`, `cron/service`)
- **`redact.ts`**: Automatic secret redaction in logs
- **`console.ts`**: Console output formatting with timestamps

### 3.11 Media Processing (`src/media/`)

Media handling for images, audio, and documents:

- **`store.ts`**: Media file storage and retrieval
- **`fetch.ts`**: Media download with content-type detection
- **`image-ops.ts`**: Image processing via `sharp`
- **`audio.ts`**: Audio format handling

### 3.12 TUI (`src/tui/`)

Terminal User Interface for interactive agent chat:

- **`tui.ts`**: Main TUI loop using Pi TUI library
- **`components/`**: UI components
- **`theme/`**: Color theming

### 3.13 ACP (`src/acp/`)

Agent Client Protocol implementation for standardized agent communication:

- **`server.ts`**: ACP server endpoint
- **`client.ts`**: ACP client
- **`session.ts`**: Session mapping between ACP and internal sessions

---

## 4. Data Flow & Control Flow

### 4.1 Inbound Message Flow

```
Channel (WhatsApp/Telegram/Slack/...) 
  → Channel Plugin (monitor/webhook)
    → Channel Registry
      → Auto-Reply Logic (allowlist check, mention gating, group policy)
        → Session Resolution (session key from channel + sender)
          → Agent Loop (Pi embedded runner)
            → LLM Provider (OpenAI/Anthropic/Gemini/...)
              → Tool Execution (if tool calls)
                → Response Assembly
                  → Outbound Delivery (back to channel)
```

### 4.2 CLI Agent Flow

```
CLI (openclaw agent --message "...")
  → Agent Command (src/commands/agent.ts)
    → Session Resolution
      → Model Selection (with failover)
        → Pi Embedded Runner
          → LLM API Call
            → Stream Processing (subscribe)
              → Tool Execution Loop
                → Final Response
                  → Optional Channel Delivery (--to flag)
```

### 4.3 Gateway WebSocket Flow

```
Client (macOS app / TUI / Web UI)
  → WebSocket Connection
    → Authentication (token/password)
      → Method Dispatch (server-methods/)
        → Handler Execution
          → Event Broadcasting (to all connected clients)
```

### 4.4 Cron Execution Flow

```
Cron Service (timer fires)
  → Job Resolution (from store)
    → Isolated Agent Spawn
      → Agent Loop Execution
        → Result Delivery (to configured channel)
          → Run Log Update
```

---

## 5. External Dependencies & Integrations

### 5.1 AI/LLM Providers

| Provider | Integration | Library |
|----------|------------|---------|
| OpenAI | API + OAuth (Codex) | `undici` (HTTP) |
| Anthropic | API + OAuth (Claude Pro/Max) | `undici` (HTTP) |
| Google Gemini | API + OAuth (Antigravity) | `undici` (HTTP) |
| AWS Bedrock | API | `@aws-sdk/client-bedrock` |
| Ollama | Local API | `ollama` (dev) |
| OpenRouter | API | `undici` (HTTP) |
| GitHub Copilot | OAuth | Custom auth flow |
| HuggingFace | API | `undici` (HTTP) |
| vLLM | API | `undici` (HTTP) |
| LiteLLM | API | `undici` (HTTP) |
| MiniMax | API + OAuth | `undici` (HTTP) |
| Qwen | API + OAuth | `undici` (HTTP) |

### 5.2 Messaging Channels

| Channel | Library/Protocol |
|---------|-----------------|
| WhatsApp | `@whiskeysockets/baileys` (Web API) |
| Telegram | `grammy` (Bot API) |
| Discord | `@buape/carbon` + `discord-api-types` |
| Slack | `@slack/bolt` + `@slack/web-api` |
| Signal | Signal CLI (subprocess) |
| iMessage | AppleScript/macOS APIs |
| LINE | `@line/bot-sdk` |
| Feishu/Lark | `@larksuiteoapi/node-sdk` |
| Matrix | `@matrix-org/matrix-sdk-crypto-nodejs` |
| Others | Custom HTTP/WebSocket implementations |

### 5.3 Core Dependencies

| Dependency | Purpose |
|-----------|---------|
| `commander` | CLI framework |
| `express` | HTTP server |
| `ws` | WebSocket server/client |
| `playwright-core` | Browser automation |
| `sharp` | Image processing |
| `sqlite-vec` | Vector search (memory) |
| `croner` | Cron expression parsing |
| `@sinclair/typebox` | Runtime type schemas |
| `zod` | Config validation |
| `yaml` / `json5` | Config file parsing |
| `dotenv` | Environment variable loading |
| `tslog` | Structured logging |
| `chalk` | Terminal colors |
| `@clack/prompts` | Interactive CLI prompts |
| `proper-lockfile` | File locking |
| `chokidar` | File watching |
| `linkedom` | HTML parsing |
| `markdown-it` | Markdown rendering |
| `pdfjs-dist` | PDF text extraction |
| `node-edge-tts` | Text-to-speech |
| `@lydell/node-pty` | PTY for shell tools |
| `undici` | HTTP client |
| `tar` | Archive handling |
| `jszip` | ZIP handling |
| `@homebridge/ciao` | mDNS/Bonjour |
| `qrcode-terminal` | QR code display |
| `@mariozechner/pi-*` | Pi agent framework |
| `@agentclientprotocol/sdk` | ACP protocol |

---

## 6. Configuration & Environment Setup

### 6.1 Config File

Primary config: `~/.openclaw/openclaw.json` (JSON5 format)

```json5
{
  gateway: {
    port: 18789,
    auth: { token: "..." },
    mode: "local",
  },
  models: {
    default: { provider: "anthropic", model: "claude-sonnet-4-20250514" },
    providers: {
      anthropic: { apiKey: "sk-ant-..." },
      openai: { apiKey: "sk-..." },
    },
  },
  channels: {
    telegram: { botToken: "..." },
    discord: { botToken: "..." },
    whatsapp: { enabled: true },
  },
  agents: {
    default: { workspace: "~/.openclaw/workspace" },
  },
  hooks: { /* ... */ },
  cron: { jobs: [] },
  browser: { enabled: true },
  sandbox: { docker: { image: "openclaw-sandbox" } },
}
```

### 6.2 Environment Variables

Key env vars (from `.env.example`):

- `OPENCLAW_GATEWAY_TOKEN` / `OPENCLAW_GATEWAY_PASSWORD` — Gateway auth
- `OPENCLAW_STATE_DIR` — State directory (default `~/.openclaw`)
- `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, `GEMINI_API_KEY` — Provider keys
- `TELEGRAM_BOT_TOKEN`, `DISCORD_BOT_TOKEN`, `SLACK_BOT_TOKEN` — Channel tokens
- `BRAVE_API_KEY`, `PERPLEXITY_API_KEY` — Tool API keys
- `ELEVENLABS_API_KEY` — TTS API key

### 6.3 State Directory

```
~/.openclaw/
├── openclaw.json          # Main config
├── .env                   # Environment overrides
├── credentials/           # OAuth credentials
├── agents/                # Per-agent state
│   └── <agentId>/
│       ├── sessions/      # Session transcripts (JSONL)
│       └── workspace/     # Agent workspace
├── sessions.json          # Session store
├── cron.json              # Cron job store
├── plugins/               # Installed plugins
├── media/                 # Media file cache
├── memory/                # Vector memory store
└── logs/                  # Log files
```

---

## 7. Entry Points & CLI Commands

### 7.1 Entry Point Chain

```
openclaw.mjs (shim)
  → dist/entry.js (built)
    → src/entry.ts
      → src/cli/run-main.ts
        → src/cli/program/build-program.ts
          → Commander program with all commands registered
```

### 7.2 Primary CLI Commands

| Command | Description |
|---------|-------------|
| `openclaw onboard` | Interactive setup wizard |
| `openclaw gateway` | Start/manage gateway server |
| `openclaw agent` | Run agent with a message |
| `openclaw message send` | Send a message to a channel |
| `openclaw status` | Show system status |
| `openclaw health` | Health check |
| `openclaw doctor` | Diagnostic & repair |
| `openclaw configure` | Configuration wizard |
| `openclaw models` | Model management |
| `openclaw channels` | Channel management |
| `openclaw sessions` | Session management |
| `openclaw cron` | Cron job management |
| `openclaw browser` | Browser automation |
| `openclaw daemon` | Daemon install/manage |
| `openclaw plugins` | Plugin management |
| `openclaw skills` | Skills management |
| `openclaw hooks` | Hooks management |
| `openclaw tui` | Terminal UI |
| `openclaw nodes` | Node management |
| `openclaw update` | Self-update |
| `openclaw reset` | Reset state |

---

## 8. Key Classes, Functions & Interfaces

### 8.1 Core Types

```typescript
// Config type (src/config/types.ts)
interface OpenClawConfig {
  gateway: GatewayConfig;
  models: ModelsConfig;
  channels: ChannelsConfig;
  agents: AgentsConfig;
  hooks: HooksConfig;
  cron: CronConfig;
  browser: BrowserConfig;
  sandbox: SandboxConfig;
  tools: ToolsConfig;
  memory: MemoryConfig;
  tts: TtsConfig;
  // ... 20+ config sections
}

// Channel Plugin (src/channels/plugins/types.plugin.ts)
interface ChannelPlugin {
  meta: ChannelMeta;
  auth?: ChannelAuthAdapter;
  messaging?: ChannelMessagingAdapter;
  outbound?: ChannelOutboundAdapter;
  directory?: ChannelDirectoryAdapter;
  status?: ChannelStatusAdapter;
  setup?: ChannelSetupAdapter;
  pairing?: ChannelPairingAdapter;
  gateway?: ChannelGatewayAdapter;
  // ...
}

// CLI Dependencies (src/cli/deps.ts)
interface CliDeps {
  fetch: typeof fetch;
  // ... injectable dependencies
}

// Agent Command Options (src/commands/agent/types.ts)
interface AgentCommandOpts {
  message?: string;
  to?: string;
  sessionId?: string;
  sessionKey?: string;
  agentId?: string;
  thinking?: string;
  verbose?: string;
  mode?: string;
  // ...
}
```

### 8.2 Key Functions

- **`agentCommand()`** (`src/commands/agent.ts`): Main agent execution entry point
- **`runEmbeddedPiAgent()`** (`src/agents/pi-embedded.ts`): Runs the AI agent loop
- **`loadConfig()`** (`src/config/io.ts`): Loads and caches configuration
- **`startGatewaySidecars()`** (`src/gateway/server-startup.ts`): Initializes gateway services
- **`buildProgram()`** (`src/cli/program/build-program.ts`): Constructs CLI program
- **`resolveConfiguredModelRef()`** (`src/agents/model-selection.ts`): Resolves model from config
- **`runWithModelFallback()`** (`src/agents/model-fallback.ts`): Executes with failover
- **`buildWorkspaceSkillSnapshot()`** (`src/agents/skills.ts`): Loads skills for agent

### 8.3 Protocol Schema

The gateway WebSocket protocol is defined using TypeBox schemas in `src/gateway/protocol/schema/`:

- **Frames**: Request/response/event frame types
- **Agent**: Agent run, status, events
- **Chat**: Chat messages, streaming
- **Config**: Config read/write
- **Sessions**: Session CRUD
- **Channels**: Channel status, management
- **Cron**: Cron job management
- **Nodes**: Node registration, events
- **Devices**: Device pairing

---

## 9. State Management & Persistence

### 9.1 File-Based State

OpenClaw uses **file-based persistence** (no database):

- **Config**: `~/.openclaw/openclaw.json` (JSON5, with file locking via `proper-lockfile`)
- **Sessions**: `~/.openclaw/sessions.json` (session metadata) + `~/.openclaw/agents/<id>/sessions/*.jsonl` (transcripts)
- **Cron**: `~/.openclaw/cron.json` (job definitions and run logs)
- **Media**: `~/.openclaw/media/` (cached media files)
- **Memory**: `~/.openclaw/memory/` (sqlite-vec or LanceDB vector store)
- **Credentials**: `~/.openclaw/credentials/` (OAuth tokens)
- **Gateway Lock**: `~/.openclaw/gateway.lock` (prevents multiple instances)

### 9.2 In-Memory State

- **Channel connections**: Maintained in gateway runtime state
- **WebSocket clients**: Tracked in server runtime
- **Cron timers**: Managed by cron service
- **Plugin instances**: Held in plugin registry
- **Agent sessions**: Active sessions in memory during execution

### 9.3 Caching

- **Config cache**: In-memory cache with file-watch invalidation
- **Model catalog**: Cached model capabilities
- **Media cache**: Downloaded media files
- **Session cache**: Recent session metadata

---

## 10. Error Handling Patterns

### 10.1 TypeScript Patterns

```typescript
// Result types (common pattern)
type BootRunResult =
  | { status: "skipped"; reason: "missing" | "empty" }
  | { status: "ran" }
  | { status: "failed"; reason: string };

// Try-catch with typed errors
try {
  const result = await loadBootFile(workspaceDir);
} catch (err) {
  const anyErr = err as { code?: string };
  if (anyErr.code === "ENOENT") {
    return { status: "missing" };
  }
  throw err;
}

// Retry with backoff (src/infra/retry.ts)
await retry(async () => {
  return await fetchWithTimeout(url, options);
}, { maxRetries: 3, backoff: "exponential" });
```

### 10.2 Error Categories

- **Configuration errors**: Validated at load time, reported via `doctor` command
- **Auth errors**: Handled with profile rotation and failover
- **Network errors**: Retried with exponential backoff
- **Channel errors**: Logged and reported via status; channels auto-reconnect
- **Agent errors**: Context overflow triggers compaction; model errors trigger failover
- **Plugin errors**: Isolated; plugin failures don't crash the gateway

---

## 11. Tests & CI/CD

### 11.1 Test Framework

- **Vitest** with V8 coverage (70% threshold for lines/branches/functions/statements)
- Test configs: `vitest.config.ts` (default), `vitest.unit.config.ts`, `vitest.e2e.config.ts`, `vitest.live.config.ts`, `vitest.gateway.config.ts`, `vitest.extensions.config.ts`
- Colocated tests: `*.test.ts` next to source files
- E2E tests: `*.e2e.test.ts`
- Live tests: `*.live.test.ts` (require real API keys)

### 11.2 CI/CD

GitHub Actions workflows (`.github/workflows/`):

- **`ci.yml`**: Main CI (lint, typecheck, build, test)
- **`docker-release.yml`**: Docker image build & push
- **`install-smoke.yml`**: Installation smoke tests
- **`formal-conformance.yml`**: Protocol conformance tests
- **`stale.yml`**: Stale issue/PR management
- **`labeler.yml`**: Auto-labeling
- **`workflow-sanity.yml`**: Workflow validation

### 11.3 Docker Testing

- `pnpm test:docker:all` — Full Docker test suite
- `pnpm test:docker:live-models` — Live model tests in Docker
- `pnpm test:docker:onboard` — Onboarding E2E in Docker
- `pnpm test:docker:gateway-network` — Gateway network tests

---

## 12. Go Rewrite Plan

### 12.1 Go Project Structure

```
goclaw/
├── go.mod                    # Module: github.com/StellariumFoundation/goclaw
├── go.sum
├── main.go                   # Entry point
├── Makefile                  # Build targets
├── Dockerfile
│
├── cmd/                      # CLI entry points
│   └── goclaw/
│       └── main.go           # CLI main
│
├── internal/                 # Internal packages (not exported)
│   ├── cli/                  # CLI framework
│   │   ├── program.go        # Command registration
│   │   ├── deps.go           # Dependency injection
│   │   ├── gateway.go        # Gateway subcommands
│   │   ├── daemon.go         # Daemon subcommands
│   │   ├── agent.go          # Agent subcommands
│   │   ├── channels.go       # Channel subcommands
│   │   ├── models.go         # Model subcommands
│   │   ├── cron.go           # Cron subcommands
│   │   ├── browser.go        # Browser subcommands
│   │   ├── plugins.go        # Plugin subcommands
│   │   └── onboard.go        # Onboarding wizard
│   │
│   ├── gateway/              # Gateway server
│   │   ├── server.go         # HTTP + WebSocket server
│   │   ├── auth.go           # Authentication
│   │   ├── chat.go           # Chat handling
│   │   ├── channels.go       # Channel lifecycle
│   │   ├── broadcast.go      # Event broadcasting
│   │   ├── discovery.go      # mDNS discovery
│   │   ├── openai_http.go    # OpenAI-compatible API
│   │   ├── boot.go           # Boot sequence
│   │   ├── methods/          # RPC method handlers
│   │   └── protocol/         # WebSocket protocol
│   │       ├── schema.go     # Protocol schema definitions
│   │       └── frames.go     # Frame types
│   │
│   ├── agent/                # AI agent system
│   │   ├── runner.go         # Agent loop runner
│   │   ├── prompt.go         # System prompt builder
│   │   ├── compaction.go     # Context compaction
│   │   ├── model.go          # Model catalog & selection
│   │   ├── failover.go       # Model failover
│   │   ├── auth.go           # Auth profile management
│   │   ├── skills.go         # Skills loading
│   │   ├── identity.go       # Agent identity
│   │   ├── sandbox.go        # Sandbox execution
│   │   ├── tools/            # Tool implementations
│   │   │   ├── bash.go       # Shell execution
│   │   │   ├── browser.go    # Browser tools
│   │   │   ├── file.go       # File operations
│   │   │   ├── message.go    # Messaging tools
│   │   │   └── registry.go   # Tool registry
│   │   └── stream/           # LLM stream processing
│   │       ├── subscriber.go
│   │       └── assembler.go
│   │
│   ├── config/               # Configuration
│   │   ├── config.go         # Config loading/writing
│   │   ├── schema.go         # Config schema
│   │   ├── types.go          # Config types
│   │   ├── defaults.go       # Default values
│   │   ├── paths.go          # Config paths
│   │   ├── env.go            # Environment variables
│   │   ├── validate.go       # Validation
│   │   ├── migrate.go        # Legacy migration
│   │   └── sessions/         # Session store
│   │       ├── store.go
│   │       ├── metadata.go
│   │       └── transcript.go
│   │
│   ├── channel/              # Channel abstraction
│   │   ├── plugin.go         # Channel plugin interface
│   │   ├── registry.go       # Channel registry
│   │   ├── allowlist.go      # Allowlist logic
│   │   ├── routing.go        # Message routing
│   │   └── adapters/         # Built-in channel adapters
│   │       ├── telegram.go
│   │       ├── discord.go
│   │       ├── slack.go
│   │       ├── signal.go
│   │       ├── whatsapp.go
│   │       └── line.go
│   │
│   ├── plugin/               # Plugin system
│   │   ├── loader.go         # Plugin discovery
│   │   ├── registry.go       # Plugin registry
│   │   ├── hooks.go          # Hook wiring
│   │   ├── tools.go          # Tool registration
│   │   └── install.go        # Plugin installation
│   │
│   ├── cron/                 # Cron scheduler
│   │   ├── service.go        # Cron service
│   │   ├── schedule.go       # Schedule parsing
│   │   ├── store.go          # Job persistence
│   │   └── delivery.go       # Result delivery
│   │
│   ├── browser/              # Browser automation
│   │   ├── session.go        # Browser session
│   │   ├── cdp.go            # CDP client
│   │   ├── chrome.go         # Chrome management
│   │   ├── tools.go          # Browser tools
│   │   └── server.go         # Control server
│   │
│   ├── hooks/                # Event hooks
│   │   ├── registry.go       # Hook registry
│   │   ├── loader.go         # Hook loading
│   │   └── gmail.go          # Gmail integration
│   │
│   ├── media/                # Media processing
│   │   ├── store.go          # Media storage
│   │   ├── fetch.go          # Media download
│   │   ├── image.go          # Image processing
│   │   └── audio.go          # Audio processing
│   │
│   ├── tts/                  # Text-to-speech
│   │   └── engine.go         # TTS engine
│   │
│   ├── tui/                  # Terminal UI
│   │   ├── app.go            # TUI application
│   │   └── components/       # UI components
│   │
│   ├── acp/                  # Agent Client Protocol
│   │   ├── server.go
│   │   ├── client.go
│   │   └── session.go
│   │
│   ├── infra/                # Infrastructure
│   │   ├── fetch.go          # HTTP client
│   │   ├── retry.go          # Retry logic
│   │   ├── lock.go           # File locking
│   │   ├── heartbeat.go      # Heartbeat system
│   │   ├── bonjour.go        # mDNS discovery
│   │   ├── tailscale.go      # Tailscale integration
│   │   ├── outbound/         # Message delivery
│   │   └── net/              # Network security
│   │
│   ├── logging/              # Logging
│   │   ├── logger.go         # Logger
│   │   ├── redact.go         # Secret redaction
│   │   └── subsystem.go      # Subsystem loggers
│   │
│   ├── daemon/               # Daemon management
│   │   ├── launchd.go        # macOS launchd
│   │   ├── systemd.go        # Linux systemd
│   │   └── service.go        # Service abstraction
│   │
│   └── security/             # Security
│       ├── audit.go          # Security auditing
│       └── ssrf.go           # SSRF protection
│
├── pkg/                      # Public packages (for plugins/extensions)
│   ├── pluginsdk/            # Plugin SDK
│   │   ├── types.go          # Plugin types
│   │   ├── channel.go        # Channel plugin interface
│   │   └── tools.go          # Tool definitions
│   │
│   └── protocol/             # Gateway protocol
│       ├── schema.go         # Protocol schema
│       └── frames.go         # Frame types
│
└── extensions/               # Extension plugins (separate Go modules)
    ├── telegram/
    ├── discord/
    ├── slack/
    └── ...
```

### 12.2 TypeScript → Go Pattern Mapping

#### Interfaces & Types

```typescript
// TypeScript
interface ChannelPlugin {
  meta: ChannelMeta;
  auth?: ChannelAuthAdapter;
  messaging?: ChannelMessagingAdapter;
}
```

```go
// Go — use interfaces for behavior, structs for data
type ChannelPlugin interface {
    Meta() ChannelMeta
    Auth() (ChannelAuthAdapter, bool)       // bool = "is implemented"
    Messaging() (ChannelMessagingAdapter, bool)
}

type ChannelMeta struct {
    ID          string
    DisplayName string
    // ...
}
```

#### Optional Fields

```typescript
// TypeScript
interface Config {
  gateway?: GatewayConfig;
  models?: ModelsConfig;
}
```

```go
// Go — use pointers for optional fields
type Config struct {
    Gateway *GatewayConfig `json:"gateway,omitempty"`
    Models  *ModelsConfig  `json:"models,omitempty"`
}
```

#### Discriminated Unions / Result Types

```typescript
// TypeScript
type BootRunResult =
  | { status: "skipped"; reason: "missing" | "empty" }
  | { status: "ran" }
  | { status: "failed"; reason: string };
```

```go
// Go — use error returns + typed errors, or a result struct
type BootRunResult struct {
    Status string // "skipped", "ran", "failed"
    Reason string // populated for "skipped" and "failed"
}

// Or use idiomatic Go error handling:
var ErrBootSkipped = errors.New("boot skipped")

func RunBootOnce(cfg *Config) error {
    // return ErrBootSkipped, nil, or wrapped error
}
```

#### Async/Await → Goroutines

```typescript
// TypeScript
async function startGatewaySidecars(params: StartParams) {
  const browserControl = await startBrowserControlServer();
  const gmailResult = await startGmailWatcher(cfg);
  // ...
}
```

```go
// Go — use goroutines + errgroup for concurrent startup
func (g *Gateway) StartSidecars(ctx context.Context) error {
    eg, ctx := errgroup.WithContext(ctx)

    eg.Go(func() error {
        return g.startBrowserControl(ctx)
    })

    eg.Go(func() error {
        return g.startGmailWatcher(ctx)
    })

    return eg.Wait()
}
```

#### Event Emitters → Channels

```typescript
// TypeScript
emitAgentEvent({ type: "tool-start", tool: name });
```

```go
// Go — use channels for event streaming
type AgentEvent struct {
    Type string
    Tool string
    // ...
}

type AgentRunner struct {
    events chan AgentEvent
}

func (r *AgentRunner) Events() <-chan AgentEvent {
    return r.events
}
```

#### Dependency Injection

```typescript
// TypeScript
interface CliDeps {
  fetch: typeof fetch;
}
function createDefaultDeps(): CliDeps { ... }
```

```go
// Go — use interfaces + constructor injection
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

type Deps struct {
    HTTP   HTTPClient
    Logger *slog.Logger
    Config *config.Config
}

func NewDefaultDeps() *Deps {
    return &Deps{
        HTTP:   http.DefaultClient,
        Logger: slog.Default(),
    }
}
```

#### Commander.js → Cobra

```typescript
// TypeScript (Commander.js)
const program = new Command();
program
  .command("gateway")
  .option("--port <port>", "Port", "18789")
  .action(async (opts) => { ... });
```

```go
// Go (Cobra)
var gatewayCmd = &cobra.Command{
    Use:   "gateway",
    Short: "Start the gateway server",
    RunE: func(cmd *cobra.Command, args []string) error {
        port, _ := cmd.Flags().GetInt("port")
        return runGateway(cmd.Context(), port)
    },
}

func init() {
    gatewayCmd.Flags().IntP("port", "p", 18789, "Port to listen on")
    rootCmd.AddCommand(gatewayCmd)
}
```

#### Express → Standard Library / Chi

```typescript
// TypeScript (Express)
app.get("/health", (req, res) => {
  res.json({ status: "ok" });
});
```

```go
// Go (net/http + chi router)
r := chi.NewRouter()
r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
})
```

#### WebSocket (ws) → gorilla/websocket or nhooyr.io/websocket

```typescript
// TypeScript (ws)
const wss = new WebSocketServer({ server });
wss.on("connection", (ws) => {
  ws.on("message", (data) => { ... });
});
```

```go
// Go (nhooyr.io/websocket)
func handleWS(w http.ResponseWriter, r *http.Request) {
    conn, err := websocket.Accept(w, r, nil)
    if err != nil { return }
    defer conn.Close(websocket.StatusNormalClosure, "")

    for {
        _, msg, err := conn.Read(r.Context())
        if err != nil { break }
        // handle msg
    }
}
```

#### TypeBox/Zod → Go Struct Tags + Validation

```typescript
// TypeScript (TypeBox)
const GatewayConfig = Type.Object({
  port: Type.Number({ default: 18789 }),
  auth: Type.Optional(Type.Object({
    token: Type.String(),
  })),
});
```

```go
// Go — struct tags + validator
type GatewayConfig struct {
    Port int            `json:"port" validate:"min=1,max=65535" default:"18789"`
    Auth *GatewayAuth   `json:"auth,omitempty"`
}

type GatewayAuth struct {
    Token string `json:"token" validate:"required"`
}
```

### 12.3 Go Library Replacements

| TypeScript Library | Go Replacement | Notes |
|-------------------|---------------|-------|
| `commander` | `github.com/spf13/cobra` | CLI framework |
| `express` | `net/http` + `github.com/go-chi/chi/v5` | HTTP router |
| `ws` | `nhooyr.io/websocket` | WebSocket |
| `@sinclair/typebox` | Go struct tags + `github.com/go-playground/validator` | Schema validation |
| `zod` | `github.com/go-playground/validator` | Validation |
| `tslog` | `log/slog` (stdlib) | Structured logging |
| `chalk` | `github.com/fatih/color` or `github.com/charmbracelet/lipgloss` | Terminal colors |
| `@clack/prompts` | `github.com/charmbracelet/huh` | Interactive prompts |
| `dotenv` | `github.com/joho/godotenv` | Env loading |
| `yaml` | `gopkg.in/yaml.v3` | YAML parsing |
| `json5` | `github.com/yosuke-furukawa/json5` or custom | JSON5 parsing |
| `undici` | `net/http` (stdlib) | HTTP client |
| `playwright-core` | `github.com/playwright-community/playwright-go` | Browser automation |
| `sharp` | `github.com/disintegration/imaging` or CGo bindings | Image processing |
| `sqlite-vec` | `github.com/mattn/go-sqlite3` + vector extension | Vector search |
| `croner` | `github.com/robfig/cron/v3` | Cron scheduling |
| `proper-lockfile` | `github.com/gofrs/flock` | File locking |
| `chokidar` | `github.com/fsnotify/fsnotify` | File watching |
| `markdown-it` | `github.com/yuin/goldmark` | Markdown |
| `linkedom` | `golang.org/x/net/html` | HTML parsing |
| `pdfjs-dist` | `github.com/ledongthuc/pdf` or `github.com/unidoc/unipdf` | PDF extraction |
| `node-edge-tts` | Custom HTTP client to Edge TTS API | TTS |
| `@lydell/node-pty` | `github.com/creack/pty` | PTY |
| `tar` | `archive/tar` (stdlib) | Archive |
| `jszip` | `archive/zip` (stdlib) | ZIP |
| `@homebridge/ciao` | `github.com/grandcat/zeroconf` | mDNS |
| `qrcode-terminal` | `github.com/mdp/qrterminal` | QR codes |
| `grammy` (Telegram) | `github.com/go-telegram-bot-api/telegram-bot-api/v5` | Telegram |
| `@buape/carbon` (Discord) | `github.com/bwmarrin/discordgo` | Discord |
| `@slack/bolt` | `github.com/slack-go/slack` | Slack |
| `@whiskeysockets/baileys` | `github.com/nickstenning/go-whatsapp` or `go.mau.fi/whatsmeow` | WhatsApp |
| `@line/bot-sdk` | `github.com/line/line-bot-sdk-go` | LINE |
| `@larksuiteoapi/node-sdk` | `github.com/larksuite/oapi-sdk-go` | Feishu |
| `@aws-sdk/client-bedrock` | `github.com/aws/aws-sdk-go-v2/service/bedrockruntime` | Bedrock |
| `signal-utils` | Custom Signal CLI wrapper | Signal |
| `osc-progress` | `github.com/charmbracelet/bubbles/progress` | Progress bars |
| Pi agent framework | Custom implementation | Agent loop |
| ACP SDK | Custom implementation | ACP protocol |

### 12.4 Go Module Dependencies

```go
// go.mod
module github.com/StellariumFoundation/goclaw

go 1.26.0

require (
    // CLI
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.0

    // HTTP & WebSocket
    github.com/go-chi/chi/v5 v5.0.12
    nhooyr.io/websocket v1.8.10

    // Logging
    // (use stdlib log/slog)

    // Terminal UI
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.0
    github.com/charmbracelet/huh v0.3.0
    github.com/fatih/color v1.16.0

    // Config & Validation
    github.com/go-playground/validator/v10 v10.17.0
    github.com/joho/godotenv v1.5.1
    gopkg.in/yaml.v3 v3.0.1

    // File system
    github.com/fsnotify/fsnotify v1.7.0
    github.com/gofrs/flock v0.8.1

    // Browser automation
    github.com/playwright-community/playwright-go v0.4.0

    // Image processing
    github.com/disintegration/imaging v1.6.2

    // Database & Vector search
    github.com/mattn/go-sqlite3 v1.14.22

    // Cron
    github.com/robfig/cron/v3 v3.0.1

    // mDNS
    github.com/grandcat/zeroconf v1.0.0

    // QR codes
    github.com/mdp/qrterminal v1.0.1

    // Messaging channels
    github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
    github.com/bwmarrin/discordgo v0.27.1
    github.com/slack-go/slack v0.12.3
    go.mau.fi/whatsmeow v0.0.0
    github.com/line/line-bot-sdk-go v7.21.0

    // AWS
    github.com/aws/aws-sdk-go-v2 v1.24.0
    github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.7.0

    // PTY
    github.com/creack/pty v1.1.21

    // Markdown
    github.com/yuin/goldmark v1.7.0

    // HTML parsing
    golang.org/x/net v0.20.0

    // Concurrency
    golang.org/x/sync v0.6.0

    // Archive
    // (use stdlib archive/tar, archive/zip)

    // Crypto
    // (use stdlib crypto/*)

    // Testing
    github.com/stretchr/testify v1.8.4
)
```

---

## 13. Migration Strategy

### 13.1 Phase 1: Foundation (Weeks 1-3)

**Goal**: Core infrastructure that everything else depends on.

1. **Config system** (`internal/config/`)
   - JSON5 config loading/writing
   - Type definitions for all config sections
   - Environment variable loading
   - Config validation
   - Config paths and state directory management

2. **Logging** (`internal/logging/`)
   - Structured logging with `slog`
   - Subsystem-scoped loggers
   - Secret redaction

3. **Infrastructure** (`internal/infra/`)
   - HTTP client with retry and timeout
   - File locking
   - SSRF protection
   - Path utilities

4. **CLI skeleton** (`internal/cli/`)
   - Cobra command structure
   - Dependency injection
   - Basic commands: `version`, `config`, `status`

### 13.2 Phase 2: Gateway Core (Weeks 4-6)

**Goal**: Working gateway server with WebSocket protocol.

1. **Gateway server** (`internal/gateway/`)
   - HTTP server with Chi router
   - WebSocket server
   - Authentication (token/password)
   - Health endpoint
   - Protocol schema (Go structs)

2. **Gateway client** (`internal/gateway/`)
   - WebSocket client for connecting to gateway
   - Protocol frame encoding/decoding

3. **Session management** (`internal/config/sessions/`)
   - Session store (file-based)
   - Session metadata
   - Transcript reading/writing (JSONL)

4. **CLI gateway commands**
   - `goclaw gateway run`
   - `goclaw gateway status`
   - `goclaw health`

### 13.3 Phase 3: Agent System (Weeks 7-10)

**Goal**: Working AI agent with LLM integration.

1. **LLM client** (`internal/agent/`)
   - OpenAI API client (streaming)
   - Anthropic API client (streaming)
   - Model catalog and selection
   - Auth profile management
   - Model failover

2. **Agent runner** (`internal/agent/`)
   - Agent loop (message → LLM → tool calls → response)
   - System prompt construction
   - Context compaction
   - Stream processing

3. **Tool system** (`internal/agent/tools/`)
   - Tool registry
   - Bash/shell execution
   - File operations
   - Basic messaging tools

4. **Skills** (`internal/agent/`)
   - Markdown skill loading
   - Skill prompt injection

5. **CLI agent commands**
   - `goclaw agent --message "..."`
   - `goclaw message send`

### 13.4 Phase 4: Channels (Weeks 11-14)

**Goal**: Multi-channel messaging support.

1. **Channel abstraction** (`internal/channel/`)
   - Channel plugin interface
   - Channel registry
   - Message routing
   - Allowlist logic

2. **Channel adapters** (one at a time)
   - Telegram (highest priority — most common)
   - WhatsApp (via whatsmeow)
   - Discord
   - Slack
   - Signal
   - LINE

3. **Auto-reply system**
   - Inbound message → agent → outbound response
   - Group message handling
   - Mention gating

4. **Outbound delivery**
   - Message queue
   - Channel-specific formatting
   - Delivery confirmation

### 13.5 Phase 5: Advanced Features (Weeks 15-18)

**Goal**: Feature parity with TypeScript version.

1. **Browser automation** (`internal/browser/`)
   - Playwright-go integration
   - CDP client
   - Browser tools

2. **Cron scheduler** (`internal/cron/`)
   - Cron service with robfig/cron
   - Job persistence
   - Isolated agent execution

3. **Plugin system** (`internal/plugin/`)
   - Plugin discovery and loading
   - Hook wiring
   - Tool registration

4. **Hooks** (`internal/hooks/`)
   - Event hook system
   - Gmail integration

5. **Media processing** (`internal/media/`)
   - Image processing
   - Audio handling
   - Media storage

6. **TTS** (`internal/tts/`)
   - Edge TTS integration
   - ElevenLabs integration

### 13.6 Phase 6: UI & Platform (Weeks 19-22)

**Goal**: Complete user experience.

1. **TUI** (`internal/tui/`)
   - Bubbletea-based terminal UI
   - Chat interface
   - Status display

2. **Web UI**
   - Serve static web UI assets
   - WebSocket integration

3. **Daemon management** (`internal/daemon/`)
   - launchd (macOS)
   - systemd (Linux)
   - Service lifecycle

4. **Onboarding wizard**
   - Interactive setup
   - Channel configuration
   - Model selection

5. **Doctor/diagnostics**
   - Health checks
   - Config validation
   - Repair tools

6. **ACP** (`internal/acp/`)
   - Agent Client Protocol server
   - Session mapping

### 13.7 Dependency Graph

```
Phase 1: config → logging → infra → cli-skeleton
Phase 2: gateway-server → gateway-client → sessions → cli-gateway
Phase 3: llm-client → agent-runner → tools → skills → cli-agent
Phase 4: channel-abstraction → channel-adapters → auto-reply → outbound
Phase 5: browser → cron → plugins → hooks → media → tts
Phase 6: tui → web-ui → daemon → onboard → doctor → acp
```

---

## 14. Challenges & Design Decisions

### 14.1 Pi Agent Framework

The TypeScript version uses `@mariozechner/pi-*` packages for the agent loop. This is a custom framework that handles:
- LLM API communication with streaming
- Tool call parsing and execution
- Session management
- Context window management

**Go approach**: Implement a custom agent loop from scratch. The core loop is:
1. Build messages array (system prompt + history + user message)
2. Call LLM API with streaming
3. Parse streaming response for text and tool calls
4. Execute tool calls
5. Append results to history
6. Repeat until no more tool calls

This is well-suited to Go's goroutine model — use a goroutine for the stream reader and channels for tool execution coordination.

### 14.2 Plugin System

TypeScript plugins are loaded dynamically via `jiti` (runtime TypeScript execution). Go doesn't have equivalent dynamic loading.

**Go approach**: 
- Use Go plugin system (`plugin` package) for `.so` plugins (Linux/macOS only)
- Or use **HashiCorp go-plugin** (gRPC-based) for cross-platform plugin support
- Or compile extensions statically and use build tags for optional features
- Recommended: **go-plugin** for channel extensions, static compilation for core features

### 14.3 JSON5 Config

Go doesn't have native JSON5 support. Options:
- Use a JSON5 parser library
- Migrate config format to YAML or TOML (breaking change)
- Support both JSON5 (read) and JSON (write) during migration

**Recommendation**: Support JSON5 reading for backward compatibility, write as standard JSON. Use `github.com/yosuke-furukawa/json5` or similar.

### 14.4 TypeBox Schema → Go Types

TypeBox provides runtime type checking and schema generation. In Go:
- Define types as Go structs with JSON tags
- Use `go-playground/validator` for validation
- Generate JSON Schema from Go types if needed (for protocol compatibility)

### 14.5 Streaming LLM Responses

TypeScript uses async iterators and event emitters for streaming. Go approach:
- Use `io.Reader` for HTTP streaming (SSE)
- Parse SSE events in a goroutine
- Send parsed events through a channel
- Consumer goroutine processes events and executes tools

```go
type StreamEvent struct {
    Type    string
    Content string
    ToolCall *ToolCall
}

func StreamLLM(ctx context.Context, req *LLMRequest) (<-chan StreamEvent, error) {
    events := make(chan StreamEvent, 100)
    go func() {
        defer close(events)
        // HTTP request with streaming
        // Parse SSE events
        // Send to channel
    }()
    return events, nil
}
```

### 14.6 WebSocket Protocol

The TypeScript version uses a custom WebSocket protocol with TypeBox schemas. In Go:
- Define protocol messages as Go structs
- Use JSON encoding/decoding
- Implement a message router that dispatches to handlers

```go
type Frame struct {
    ID     string          `json:"id"`
    Method string          `json:"method"`
    Params json.RawMessage `json:"params,omitempty"`
}

type Handler func(ctx context.Context, params json.RawMessage) (any, error)

type Router struct {
    handlers map[string]Handler
}
```

### 14.7 Concurrency Model

TypeScript is single-threaded with async/await. Go is natively concurrent:

- **Gateway**: Each WebSocket connection gets its own goroutine
- **Channels**: Each channel monitor runs in its own goroutine
- **Agent**: Tool execution can be parallelized where safe
- **Cron**: Timer goroutines for scheduled jobs
- Use `context.Context` for cancellation propagation
- Use `sync.Mutex` / `sync.RWMutex` for shared state
- Use `errgroup` for coordinated goroutine lifecycle

### 14.8 Error Handling

TypeScript uses try/catch with typed error checking. Go uses explicit error returns:

```go
// Wrap errors with context
if err := loadConfig(); err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Sentinel errors for specific cases
var (
    ErrConfigNotFound = errors.New("config file not found")
    ErrAuthFailed     = errors.New("authentication failed")
)

// Check specific errors
if errors.Is(err, ErrConfigNotFound) {
    // handle missing config
}
```

### 14.9 Testing Strategy

- Use `testing` package (stdlib) + `testify` for assertions
- Table-driven tests for comprehensive coverage
- Use `httptest` for HTTP handler testing
- Use `testcontainers-go` for Docker-based integration tests
- Mock interfaces for unit testing (no need for complex DI frameworks)

### 14.10 Build & Distribution

- Single binary distribution (Go's strength)
- Cross-compilation for Linux, macOS, Windows (arm64 + amd64)
- Docker image based on `scratch` or `alpine` (much smaller than Node.js)
- Use `goreleaser` for release automation
- Embed static assets (web UI) using `embed` package

### 14.11 Performance Considerations

Go advantages over TypeScript/Node.js:
- **Lower memory footprint**: No V8 heap overhead
- **Faster startup**: No JIT compilation
- **Better concurrency**: Native goroutines vs event loop
- **Single binary**: No `node_modules` dependency tree
- **Smaller Docker images**: ~20MB vs ~500MB+

### 14.12 Backward Compatibility

- Read existing `~/.openclaw/` state directory
- Parse existing `openclaw.json` config (JSON5)
- Read existing session transcripts (JSONL)
- Support existing WebSocket protocol for native app compatibility
- Provide migration tool for any format changes

---

## Appendix: File Count Summary

| Directory | Approx. Files | Description |
|-----------|--------------|-------------|
| `src/agents/` | ~200 | Agent system (largest module) |
| `src/gateway/` | ~150 | Gateway server |
| `src/cli/` | ~120 | CLI framework |
| `src/commands/` | ~130 | Command implementations |
| `src/config/` | ~100 | Configuration |
| `src/infra/` | ~100 | Infrastructure |
| `src/browser/` | ~60 | Browser automation |
| `src/channels/` | ~30 | Channel abstraction |
| `src/plugins/` | ~40 | Plugin system |
| `src/cron/` | ~40 | Cron scheduler |
| `src/web/` | ~40 | WhatsApp Web |
| `src/hooks/` | ~25 | Event hooks |
| `src/logging/` | ~18 | Logging |
| `src/media/` | ~20 | Media processing |
| `src/line/` | ~30 | LINE channel |
| `src/slack/` | ~15 | Slack channel |
| `src/telegram/` | ~10 | Telegram channel |
| `src/tui/` | ~25 | Terminal UI |
| `extensions/` | ~36 dirs | Extension plugins |
| `skills/` | ~50 dirs | Skill definitions |
| **Total** | **~1200+** | Source files (excluding tests) |

This is a large codebase. The Go rewrite should prioritize the core path (config → gateway → agent → channels) and add features incrementally. The phased approach in Section 13 ensures each phase produces a working, testable system.
