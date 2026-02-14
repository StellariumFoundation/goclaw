# Openclaw → GoClaw: Go 1.26.0 Rewrite Architecture Document

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Original Architecture Overview](#original-architecture-overview)
3. [Module/Component Map](#modulecomponent-map)
4. [Data Flow & APIs](#data-flow--apis)
5. [Dependency Analysis](#dependency-analysis)
6. [Go Rewrite Plan](#go-rewrite-plan)
7. [Package Structure](#package-structure)
8. [Implementation Priority](#implementation-priority)
9. [Technical Considerations](#technical-considerations)
10. [Migration Strategy](#migration-strategy)

---

## Executive Summary

### What is Openclaw?

Openclaw is a **multi-channel AI gateway** with extensible messaging integrations. It acts as a unified bridge between AI model providers (OpenAI, Anthropic, Google Gemini, AWS Bedrock, Ollama, vLLM, etc.) and messaging platforms (Telegram, Discord, Slack, WhatsApp, Signal, MS Teams, Matrix, LINE, iMessage, IRC, and a web UI). It provides:

- A **WebSocket-based gateway server** that manages AI agent sessions, routing, and configuration
- A **CLI** for managing the gateway, agents, sessions, channels, cron jobs, browser automation, and more
- A **plugin/extension system** for adding new channels and capabilities
- An **Agent Client Protocol (ACP)** bridge for IDE integration
- **Multi-agent support** with sub-agent spawning, session isolation, and model failover
- **Cron scheduling**, **heartbeat/proactive messaging**, **media understanding**, **TTS**, **browser automation**, and **sandbox execution**
- Native **macOS**, **iOS**, and **Android** apps (Swift/Kotlin) with shared OpenClawKit framework
- A **TUI** (terminal UI) for interactive chat sessions

The current implementation is a **TypeScript/Node.js monorepo** (pnpm workspace) with ~218 dependencies, targeting Node.js ≥22.12.0.

### Goal of the Go Rewrite

Rewrite the Openclaw server-side core as **GoClaw** using **Go 1.26.0**, targeting:

- **Single static binary** deployment (no Node.js runtime dependency)
- **Superior concurrency** via goroutines for WebSocket handling, agent sessions, channel monitors, and cron
- **Lower memory footprint** and faster startup
- **Type safety** with Go's strong typing and compile-time checks
- **Simplified deployment** (Docker, systemd, launchd) without npm/pnpm toolchain
- **Maintainability** through Go's explicit error handling and standard library

---

## Original Architecture Overview

### Languages & Frameworks

| Layer | Language | Framework/Runtime |
|-------|----------|-------------------|
| Core server & CLI | TypeScript (ESM) | Node.js ≥22.12.0 |
| Build system | tsdown (Rolldown-based) | pnpm 10.23.0 workspace |
| Testing | TypeScript | Vitest 4.x (unit, e2e, live) |
| Linting/Formatting | TypeScript | oxlint, oxfmt |
| macOS/iOS apps | Swift | SwiftUI, OpenClawKit |
| Android app | Kotlin | Gradle |
| Web UI | TypeScript | Lit (Web Components) |
| Extensions | TypeScript | Plugin SDK (openclaw/plugin-sdk) |
| Skills | Markdown + scripts | SKILL.md definitions, Python/Bash scripts |

### Repository Structure

```
Openclaw/
├── src/                    # Core TypeScript source (~40+ subdirectories)
│   ├── entry.ts            # CLI entrypoint (process respawn, profile loading)
│   ├── cli/                # CLI command definitions (Commander.js)
│   ├── gateway/            # WebSocket gateway server (Express + ws)
│   ├── agents/             # AI agent runner, model selection, tools, sandbox
│   ├── auto-reply/         # Inbound message → agent reply pipeline
│   ├── channels/           # Channel abstraction layer & plugin adapters
│   ├── config/             # Configuration schema (Zod), loading, validation
│   ├── plugins/            # Plugin system (discovery, loading, hooks, runtime)
│   ├── hooks/              # Lifecycle hooks (bundled + user-defined)
│   ├── cron/               # Cron job scheduling and execution
│   ├── sessions/           # Session management, transcripts, metadata
│   ├── routing/            # Message routing, session key resolution
│   ├── security/           # Security audit, secret comparison, SSRF protection
│   ├── infra/              # Infrastructure utilities (env, dotenv, retry, etc.)
│   ├── logging/            # Structured logging, redaction, subsystems
│   ├── media/              # Media handling (fetch, store, MIME, image ops)
│   ├── media-understanding/# Vision, audio transcription, video analysis
│   ├── link-understanding/ # URL extraction and content fetching
│   ├── browser/            # Playwright-based browser automation
│   ├── canvas-host/        # Canvas/A2UI hosting server
│   ├── tts/                # Text-to-speech (ElevenLabs, Edge TTS, etc.)
│   ├── acp/                # Agent Client Protocol bridge (stdio NDJSON)
│   ├── daemon/             # Daemon management (systemd, launchd, schtasks)
│   ├── process/            # Process management, child process bridge
│   ├── tui/                # Terminal UI (interactive chat)
│   ├── wizard/             # Onboarding wizard
│   ├── providers/          # Provider-specific auth (GitHub Copilot, Qwen)
│   ├── terminal/           # Terminal formatting (ANSI, tables, themes)
│   ├── markdown/           # Markdown processing
│   ├── shared/             # Shared utilities (reasoning tags)
│   ├── utils/              # General utilities
│   ├── types/              # TypeScript type declarations
│   ├── test-helpers/       # Test utilities
│   ├── test-utils/         # Additional test utilities
│   │
│   │   # Channel-specific implementations
│   ├── discord/            # Discord bot (discord-api-types, @buape/carbon)
│   ├── slack/              # Slack bot (@slack/bolt, @slack/web-api)
│   ├── telegram/           # Telegram bot (grammy)
│   ├── whatsapp/           # WhatsApp (Baileys)
│   ├── signal/             # Signal messenger
│   ├── imessage/           # iMessage integration
│   ├── line/               # LINE messaging
│   ├── web/                # Web/WhatsApp Web interface
│   ├── pairing/            # Device pairing
│   ├── node-host/          # Node host for mobile devices
│   └── plugin-sdk/         # Plugin SDK exports
│
├── extensions/             # Extension plugins (separate packages)
│   ├── bluebubbles/        # BlueBubbles iMessage bridge
│   ├── matrix/             # Matrix protocol
│   ├── msteams/            # Microsoft Teams
│   ├── line/               # LINE extension
│   ├── copilot-proxy/      # GitHub Copilot proxy
│   ├── device-pair/        # Device pairing extension
│   ├── diagnostics-otel/   # OpenTelemetry diagnostics
│   ├── google-gemini-cli-auth/ # Google Gemini CLI auth
│   ├── llm-task/           # LLM task tool
│   ├── lobster/            # Lobster tool
│   ├── memory-core/        # Memory/RAG extension
│   ├── phone-control/      # Phone control
│   ├── qwen-portal-auth/   # Qwen portal auth
│   ├── signal/             # Signal extension
│   ├── thread-ownership/   # Thread ownership gating
│   └── whatsapp/           # WhatsApp extension
│
├── skills/                 # Skill definitions (50+ skills)
│   ├── github/             # GitHub CLI skill
│   ├── discord/            # Discord actions skill
│   ├── slack/              # Slack actions skill
│   ├── canvas/             # Canvas/A2UI skill
│   ├── coding-agent/       # Coding agent skill
│   └── ...                 # 45+ more skills
│
├── apps/                   # Native applications
│   ├── shared/OpenClawKit/ # Shared Swift framework
│   │   ├── Sources/OpenClawKit/       # Core protocol, commands, gateway
│   │   ├── Sources/OpenClawChatUI/    # Chat UI components
│   │   └── Sources/OpenClawProtocol/  # Protocol models
│   ├── macos/              # macOS app
│   ├── ios/                # iOS app
│   └── android/            # Android app
│
├── packages/               # Internal npm packages
│   ├── clawdbot/           # Legacy compatibility package
│   └── moltbot/            # Legacy compatibility package
│
├── docs/                   # Documentation (Mintlify)
├── scripts/                # Build/deploy scripts
├── assets/                 # Static assets
├── patches/                # pnpm patch overrides
└── git-hooks/              # Git hooks
```

### Key Configuration Files

| File | Purpose |
|------|---------|
| `package.json` | Dependencies, scripts, build config |
| `pnpm-workspace.yaml` | Monorepo workspace definition |
| `tsdown.config.ts` | Build configuration (6 entry points) |
| `vitest.config.ts` | Test configuration |
| `vitest.e2e.config.ts` | E2E test configuration |
| `.env.example` | Environment variable documentation |
| `fly.private.toml` | Fly.io deployment config |
| `Dockerfile.sandbox-browser` | Sandbox browser container |

---

## Module/Component Map

### 1. CLI (`src/cli/`)

**Purpose**: Command-line interface for all Openclaw operations.

**Key files**:
- `program.ts` — Main Commander.js program builder with command registration
- `run-main.ts` — CLI bootstrap and execution
- `gateway-cli.ts` — Gateway start/stop/discover commands
- `daemon-cli.ts` — Daemon install/uninstall/status
- `config-cli.ts` — Configuration get/set/apply/patch
- `models-cli.ts` — Model listing, auth, switching
- `channels-cli.ts` — Channel management
- `nodes-cli.ts` — Remote node management (camera, canvas, screen, location)
- `browser-cli.ts` — Browser automation commands
- `cron-cli.ts` — Cron job management
- `plugins-cli.ts` — Plugin install/uninstall/status
- `skills-cli.ts` — Skill management
- `security-cli.ts` — Security audit commands
- `tui-cli.ts` — TUI launcher
- `acp-cli.ts` — ACP bridge launcher

**Interactions**: Calls into gateway client, config, daemon, and all subsystems.

### 2. Gateway Server (`src/gateway/`)

**Purpose**: Central WebSocket server that manages all agent sessions, channels, and API endpoints.

**Key files**:
- `server.impl.ts` — Main server implementation (Express HTTP + WebSocket)
- `server-http.ts` — HTTP route registration
- `server-chat.ts` — Chat session management
- `server-channels.ts` — Channel lifecycle management
- `server-plugins.ts` — Plugin loading and management
- `server-cron.ts` — Cron service integration
- `server-discovery.ts` — mDNS/Bonjour service discovery
- `server-broadcast.ts` — WebSocket broadcast to connected clients
- `server-mobile-nodes.ts` — Mobile node management
- `server-model-catalog.ts` — Model catalog aggregation
- `auth.ts` — Token/password authentication
- `client.ts` — Gateway WebSocket client
- `openai-http.ts` — OpenAI-compatible HTTP API
- `openresponses-http.ts` — OpenAI Responses API compatibility
- `hooks.ts` — Gateway hook execution
- `tools-invoke-http.ts` — HTTP tool invocation endpoint
- `protocol/` — Protocol schema definitions (TypeBox)

**Interactions**: Central hub — connects to all channels, agents, plugins, cron, sessions.

### 3. Agents (`src/agents/`)

**Purpose**: AI agent execution engine — model selection, tool execution, session management.

**Key files**:
- `pi-embedded-runner/run.ts` — Core agent execution loop
- `pi-embedded-subscribe.ts` — Streaming subscription handler
- `pi-tools.ts` — Tool definition and execution
- `bash-tools.ts` — Bash/shell tool execution
- `models-config.ts` — Model configuration and resolution
- `model-catalog.test.ts` — Model catalog management
- `system-prompt.ts` — System prompt construction
- `sandbox/` — Docker sandbox management (container lifecycle, fs bridge)
- `auth-profiles/` — Auth profile rotation and failover
- `skills/` — Skill loading and prompt injection
- `identity-file.ts` — Agent identity (IDENTITY.md)
- `context-window-guard.ts` — Context window overflow protection
- `session-transcript-repair.ts` — Transcript corruption repair

**Interactions**: Called by auto-reply and gateway; calls model providers, tools, sandbox.

### 4. Auto-Reply Pipeline (`src/auto-reply/`)

**Purpose**: Inbound message processing → agent execution → outbound reply delivery.

**Key files**:
- `reply.ts` — Main reply orchestration
- `reply/agent-runner.ts` — Agent execution wrapper with typing, memory flush
- `reply/get-reply.ts` — Reply generation pipeline
- `reply/commands.ts` — Inline command processing (!status, !model, !reset, etc.)
- `reply/directive-handling.ts` — Model/thinking directive parsing
- `reply/queue.ts` — Reply queue management
- `reply/block-streaming.ts` — Streaming block reply coalescing
- `dispatch.ts` — Message dispatch to appropriate handler
- `envelope.ts` — Message envelope construction
- `heartbeat.ts` — Proactive heartbeat message handling
- `thinking.ts` — Thinking/reasoning tag processing

**Interactions**: Receives from channels; calls agents; sends via channel outbound adapters.

### 5. Channels (`src/channels/`)

**Purpose**: Abstraction layer for messaging platform integrations.

**Key files**:
- `registry.ts` — Channel registry (built-in + plugin channels)
- `plugins/types.ts` — Channel plugin interface definitions
- `plugins/index.ts` — Plugin channel loader
- `plugins/catalog.ts` — Channel catalog
- `plugins/onboarding/` — Per-channel onboarding flows
- `plugins/outbound/` — Per-channel outbound message adapters
- `plugins/normalize/` — Per-channel message normalization
- `plugins/actions/` — Per-channel action handlers
- `allowlists/` — Sender allowlist resolution
- `web/` — Web channel implementation

**Interactions**: Bridges between auto-reply pipeline and platform-specific implementations.

### 6. Channel Implementations

#### Discord (`src/discord/`)
- Bot using discord-api-types + @buape/carbon
- Gateway event monitoring, message sending, thread management
- Voice message support, PluralKit integration

#### Telegram (`src/telegram/`)
- Bot using grammy framework
- Webhook and polling modes, inline buttons, sticker cache
- Custom command menu, group migration handling

#### Slack (`src/slack/`)
- Bot using @slack/bolt + @slack/web-api
- HTTP mode support, thread context, channel migration

#### WhatsApp (`src/whatsapp/` + `src/web/`)
- Using @whiskeysockets/baileys
- QR code login, session persistence, media handling

#### Signal (`src/signal/`)
- signal-cli integration

#### iMessage (`src/imessage/`)
- macOS-only AppleScript/JXA integration

### 7. Configuration (`src/config/`)

**Purpose**: Configuration schema, loading, validation, and persistence.

**Key files**:
- `config.ts` — Main config type and loader
- `schema.ts` — TypeBox schema definition
- `zod-schema.ts` — Zod validation schemas
- `io.ts` — Config file I/O with env variable preservation
- `env-preserve.ts` — `${VAR}` reference preservation in config writes
- `env-substitution.ts` — Environment variable substitution
- `paths.ts` — Config/state directory resolution
- `sessions/` — Session store, transcript management
- `types.ts` — Comprehensive config type definitions (agents, channels, models, etc.)
- `legacy.ts` — Legacy config migration

**Interactions**: Used by every subsystem; loaded at startup, reloaded on config changes.

### 8. Plugin System (`src/plugins/`)

**Purpose**: Extensible plugin architecture for adding capabilities.

**Key files**:
- `discovery.ts` — Plugin discovery from extensions/ directory
- `loader.ts` — Plugin loading and initialization
- `registry.ts` — Plugin registry
- `manifest.ts` — Plugin manifest parsing (openclaw.plugin.json)
- `hooks.ts` — Plugin hook execution
- `runtime/` — Plugin runtime environment
- `tools.ts` — Plugin-provided tool registration
- `services.ts` — Plugin service management

**Interactions**: Loaded by gateway; provides channels, tools, hooks, and HTTP routes.

### 9. Extensions (`extensions/`)

Each extension is a separate npm package with:
- `openclaw.plugin.json` — Plugin manifest
- `index.ts` — Entry point
- `src/` — Implementation

**Notable extensions**:
- **matrix** — Full Matrix protocol support with E2EE
- **msteams** — Microsoft Teams with Graph API, file consent, polls
- **bluebubbles** — iMessage via BlueBubbles server
- **line** — LINE messaging platform
- **memory-core** — RAG/memory with sqlite-vec
- **diagnostics-otel** — OpenTelemetry integration
- **copilot-proxy** — GitHub Copilot auth proxy
- **thread-ownership** — Thread ownership gating hooks

### 10. Sessions (`src/sessions/` + `src/config/sessions/`)

**Purpose**: Session lifecycle, transcript storage, metadata management.

**Key files**:
- `config/sessions/store.ts` — Session store (filesystem-based)
- `config/sessions/transcript.ts` — Transcript read/write
- `config/sessions/metadata.ts` — Session metadata
- `config/sessions/paths.ts` — Session file path resolution
- `config/sessions/reset.ts` — Session reset and archival
- `sessions/send-policy.ts` — Outbound send policy per session

### 11. Cron (`src/cron/`)

**Purpose**: Scheduled job execution with isolated agent runs.

**Key files**:
- `service.ts` — Cron service (using croner library)
- `schedule.ts` — Schedule parsing and management
- `isolated-agent.ts` — Isolated agent execution for cron jobs
- `delivery.ts` — Cron result delivery to channels
- `store.ts` — Cron job persistence
- `session-reaper.ts` — Stale session cleanup

### 12. Infrastructure (`src/infra/`)

**Purpose**: Cross-cutting infrastructure utilities.

**Key files**:
- `env.ts` — Environment variable handling
- `dotenv.ts` — .env file loading
- `retry.ts` — Retry with backoff
- `fetch.ts` — HTTP fetch wrapper
- `bonjour.ts` — mDNS/Bonjour service discovery
- `device-identity.ts` — Device identity management
- `device-pairing.ts` — Device pairing protocol
- `heartbeat-runner.ts` — Heartbeat scheduling and execution
- `provider-usage.ts` — Provider usage tracking
- `session-cost-usage.ts` — Session cost calculation
- `update-check.ts` — Version update checking
- `gateway-lock.ts` — Gateway instance locking
- `ssh-tunnel.ts` — SSH tunnel management
- `tailscale.ts` — Tailscale integration
- `system-events.ts` — System event bus
- `net/ssrf.ts` — SSRF protection
- `tls/` — TLS certificate management

### 13. Security (`src/security/`)

**Purpose**: Security auditing, secret management, content sanitization.

**Key files**:
- `audit.ts` — Security audit engine
- `fix.ts` — Auto-fix security issues
- `secret-equal.ts` — Timing-safe secret comparison
- `external-content.ts` — External content sanitization
- `skill-scanner.ts` — Skill security scanning

### 14. Media & Understanding (`src/media/`, `src/media-understanding/`, `src/link-understanding/`)

**Purpose**: Media handling, vision/audio analysis, URL content extraction.

**Key files**:
- `media/store.ts` — Media file storage
- `media/fetch.ts` — Media download
- `media-understanding/runner.ts` — Vision/audio analysis pipeline
- `media-understanding/providers/` — Provider-specific implementations
- `link-understanding/detect.ts` — URL detection
- `link-understanding/runner.ts` — Content extraction

### 15. Browser Automation (`src/browser/`)

**Purpose**: Playwright-based browser automation for AI agents.

**Key files**:
- `server.ts` — Browser automation HTTP server
- `pw-session.ts` — Playwright session management
- `pw-tools-core.ts` — Browser tool implementations
- `cdp.ts` — Chrome DevTools Protocol integration
- `chrome.ts` — Chrome browser management
- `profiles-service.ts` — Browser profile management

### 16. TTS (`src/tts/`)

**Purpose**: Text-to-speech synthesis.

**Key files**:
- `tts.ts` — TTS orchestration
- `tts-core.ts` — Core TTS engine (ElevenLabs, Edge TTS, system)

### 17. ACP Bridge (`src/acp/`)

**Purpose**: Agent Client Protocol bridge for IDE integration.

**Key files**:
- `server.ts` — ACP stdio server
- `client.ts` — Gateway WebSocket client
- `session-mapper.ts` — ACP ↔ Gateway session mapping
- `event-mapper.ts` — Event translation

### 18. Daemon (`src/daemon/`)

**Purpose**: System daemon management.

**Key files**:
- `launchd.ts` — macOS launchd plist generation
- `systemd.ts` — Linux systemd unit generation
- `schtasks.ts` — Windows scheduled tasks
- `service.ts` — Daemon service lifecycle

### 19. TUI (`src/tui/`)

**Purpose**: Terminal-based interactive chat UI.

**Key files**:
- `tui.ts` — Main TUI application
- `gateway-chat.ts` — Gateway chat integration
- `tui-command-handlers.ts` — TUI command processing
- `tui-stream-assembler.ts` — Streaming response assembly
- `components/` — TUI UI components

### 20. Native Apps (`apps/shared/OpenClawKit/`)

**Purpose**: Shared Swift framework for macOS/iOS apps.

**Key modules**:
- `OpenClawKit` — Core protocol, gateway communication, device commands
- `OpenClawChatUI` — Chat UI components (SwiftUI)
- `OpenClawProtocol` — Protocol model definitions

---

## Data Flow & APIs

### Primary Data Flow

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐     ┌──────────────┐
│  Messaging   │────▶│   Channel    │────▶│  Auto-Reply  │────▶│    Agent     │
│  Platform    │     │   Adapter    │     │   Pipeline   │     │   Runner     │
│ (Telegram,   │     │ (normalize,  │     │ (commands,   │     │ (model call, │
│  Discord,    │     │  allowlist,  │     │  directives, │     │  tools,      │
│  Slack, etc) │     │  routing)    │     │  queue)      │     │  sandbox)    │
└─────────────┘     └──────────────┘     └─────────────┘     └──────────────┘
                                                                      │
                                                                      ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────┐     ┌──────────────┐
│  Messaging   │◀────│   Channel    │◀────│  Outbound    │◀────│   Model      │
│  Platform    │     │   Outbound   │     │  Dispatcher  │     │  Provider    │
│              │     │   Adapter    │     │  (threading, │     │ (OpenAI,     │
│              │     │              │     │   streaming)  │     │  Anthropic,  │
└─────────────┘     └──────────────┘     └─────────────┘     │  Gemini...)  │
                                                              └──────────────┘
```

### Gateway WebSocket Protocol

The gateway uses a **JSON-over-WebSocket** protocol with typed frames:

- **Client → Server**: `prompt`, `sessions.list`, `sessions.reset`, `config.get`, `config.set`, `agents.list`, `models.list`, `cron.*`, `nodes.*`, `tools.invoke`, etc.
- **Server → Client**: `reply.start`, `reply.chunk`, `reply.end`, `reply.error`, `snapshot`, `event`, etc.
- **Authentication**: Bearer token or password-based, with rate limiting

### HTTP API Endpoints

- `GET /health` — Health check
- `POST /chat` — Chat completion (OpenAI-compatible)
- `POST /v1/chat/completions` — OpenAI API compatibility
- `POST /v1/responses` — OpenAI Responses API compatibility
- `POST /tools/invoke` — Tool invocation
- `GET /media/:id` — Media file serving
- Plugin-registered HTTP routes

### ACP Protocol

- **Transport**: stdio with NDJSON
- **Operations**: `tasks/send`, `tasks/get`, `tasks/cancel`
- **Session mapping**: ACP session IDs ↔ Gateway session keys

### Configuration Hierarchy

```
Process env → ./.env → ~/.openclaw/.env → openclaw.json `env` block
                                            ↓
                                    openclaw.json (main config)
                                            ↓
                                    Agent-specific overrides
                                            ↓
                                    Session-level overrides
```

---

## Dependency Analysis

### Core Dependencies → Go Equivalents

| Node.js Dependency | Purpose | Go Equivalent |
|-------------------|---------|---------------|
| `express` 5.x | HTTP server | `net/http` (stdlib) + `chi` or `echo` |
| `ws` 8.x | WebSocket server/client | `nhooyr.io/websocket` or `gorilla/websocket` |
| `commander` 14.x | CLI framework | `cobra` + `viper` |
| `zod` 4.x | Schema validation | `go-playground/validator` + struct tags |
| `@sinclair/typebox` | JSON Schema types | `encoding/json` + custom schemas |
| `dotenv` 17.x | .env loading | `joho/godotenv` |
| `chalk` 5.x | Terminal colors | `fatih/color` or `charmbracelet/lipgloss` |
| `chokidar` 5.x | File watching | `fsnotify/fsnotify` |
| `croner` 10.x | Cron scheduling | `robfig/cron/v3` |
| `yaml` 2.x | YAML parsing | `gopkg.in/yaml.v3` |
| `json5` 2.x | JSON5 parsing | `yosuke-furukawa/json5` or custom parser |
| `ajv` 8.x | JSON Schema validation | `santhosh-tekuri/jsonschema` |
| `sharp` 0.34.x | Image processing | `disintegration/imaging` or `h2non/bimg` |
| `jszip` 3.x | ZIP handling | `archive/zip` (stdlib) |
| `tar` 7.x | TAR handling | `archive/tar` (stdlib) |
| `proper-lockfile` 4.x | File locking | `gofrs/flock` |
| `tslog` 4.x | Structured logging | `log/slog` (stdlib, Go 1.21+) |
| `undici` 7.x | HTTP client | `net/http` (stdlib) |
| `markdown-it` 14.x | Markdown rendering | `yuin/goldmark` |
| `pdfjs-dist` 5.x | PDF parsing | `unidoc/unipdf` or `ledongthuc/pdf` |
| `qrcode-terminal` | QR code display | `skip2/go-qrcode` |
| `sqlite-vec` | Vector DB | `mattn/go-sqlite3` + custom vec extension |
| `signal-utils` | Reactive signals | Go channels + sync primitives |

### Channel-Specific Dependencies → Go Equivalents

| Node.js Dependency | Purpose | Go Equivalent |
|-------------------|---------|---------------|
| `grammy` 1.x | Telegram bot | `go-telegram-bot-api/telegram-bot-api` or `gotd/td` |
| `discord-api-types` + `@buape/carbon` | Discord bot | `bwmarrin/discordgo` |
| `@slack/bolt` + `@slack/web-api` | Slack bot | `slack-go/slack` |
| `@whiskeysockets/baileys` | WhatsApp | `tulir/whatsmeow` |
| `@line/bot-sdk` | LINE bot | `line/line-bot-sdk-go` |
| `@grammyjs/runner` | Telegram runner | Built into Go bot library |
| `@larksuiteoapi/node-sdk` | Lark/Feishu | `larksuite/oapi-sdk-go` |
| `@homebridge/ciao` | mDNS/Bonjour | `hashicorp/mdns` or `grandcat/zeroconf` |
| `playwright-core` | Browser automation | `chromedp/chromedp` (CDP) or `playwright-community/playwright-go` |
| `@lydell/node-pty` | PTY/terminal | `creack/pty` |

### AI Provider Dependencies

| Node.js Dependency | Purpose | Go Equivalent |
|-------------------|---------|---------------|
| `@mariozechner/pi-*` | AI agent core | Custom implementation |
| `@agentclientprotocol/sdk` | ACP protocol | Custom implementation |
| `@aws-sdk/client-bedrock` | AWS Bedrock | `aws/aws-sdk-go-v2` |
| `@mozilla/readability` | Content extraction | `go-shiori/go-readability` |
| `node-edge-tts` | Edge TTS | Custom HTTP client |
| `linkedom` | DOM parsing | `PuerkitoBio/goquery` |

### Dev Dependencies (Not needed in Go)

The following are build/dev tools that have no Go equivalent needed:
- `tsdown`, `tsx`, `typescript`, `vitest`, `oxlint`, `oxfmt`, `rolldown`
- These are replaced by `go build`, `go test`, `go vet`, `gofmt`/`goimports`

---

## Go Rewrite Plan

### Component-by-Component Mapping

#### 1. CLI → `cmd/goclaw/` + `internal/cli/`

**Approach**: Use `cobra` for command structure, `viper` for config binding.

```go
// cmd/goclaw/main.go — entrypoint
// internal/cli/root.go — root command
// internal/cli/gateway.go — gateway start/stop/discover
// internal/cli/config.go — config get/set/apply
// internal/cli/models.go — model management
// internal/cli/channels.go — channel management
// internal/cli/cron.go — cron management
// internal/cli/daemon.go — daemon install/status
// internal/cli/browser.go — browser automation
// internal/cli/nodes.go — node management
// internal/cli/plugins.go — plugin management
// internal/cli/security.go — security audit
// internal/cli/tui.go — TUI launcher
// internal/cli/acp.go — ACP bridge
```

#### 2. Gateway Server → `internal/gateway/`

**Approach**: `net/http` for HTTP, `nhooyr.io/websocket` for WebSocket, middleware pattern.

```go
// internal/gateway/server.go — main server (HTTP + WS)
// internal/gateway/auth.go — authentication (token, password, rate limit)
// internal/gateway/handler.go — WebSocket message handler/router
// internal/gateway/broadcast.go — client broadcast
// internal/gateway/discovery.go — mDNS service discovery
// internal/gateway/openai_compat.go — OpenAI API compatibility layer
// internal/gateway/hooks.go — hook execution
// internal/gateway/methods/ — per-method handlers (chat, config, sessions, etc.)
// internal/gateway/protocol/ — protocol frame types and schemas
```

#### 3. Agents → `internal/agents/`

**Approach**: Interface-based agent runner with provider abstraction.

```go
// internal/agents/runner.go — core agent execution loop
// internal/agents/streaming.go — streaming response handler
// internal/agents/tools.go — tool registry and execution
// internal/agents/bash.go — bash/shell tool
// internal/agents/models.go — model config and resolution
// internal/agents/system_prompt.go — system prompt builder
// internal/agents/identity.go — agent identity (IDENTITY.md)
// internal/agents/context_guard.go — context window overflow protection
// internal/agents/sandbox/ — Docker sandbox management
// internal/agents/skills/ — skill loading
// internal/agents/auth/ — auth profile management
```

#### 4. Auto-Reply → `internal/autoreply/`

```go
// internal/autoreply/pipeline.go — main reply pipeline
// internal/autoreply/commands.go — inline command processing
// internal/autoreply/directives.go — model/thinking directives
// internal/autoreply/queue.go — reply queue
// internal/autoreply/streaming.go — block streaming coalescer
// internal/autoreply/dispatch.go — message dispatch
// internal/autoreply/heartbeat.go — heartbeat handling
// internal/autoreply/envelope.go — message envelope
```

#### 5. Channels → `internal/channels/`

**Approach**: Interface-based channel abstraction.

```go
// internal/channels/registry.go — channel registry
// internal/channels/types.go — channel interfaces
// internal/channels/allowlist.go — sender allowlist
// internal/channels/routing.go — message routing
// internal/channels/web/ — web channel
// internal/channels/discord/ — Discord implementation
// internal/channels/telegram/ — Telegram implementation
// internal/channels/slack/ — Slack implementation
// internal/channels/whatsapp/ — WhatsApp implementation
// internal/channels/signal/ — Signal implementation
// internal/channels/matrix/ — Matrix implementation
// internal/channels/msteams/ — MS Teams implementation
// internal/channels/line/ — LINE implementation
```

#### 6. Configuration → `internal/config/`

```go
// internal/config/config.go — main config struct and loader
// internal/config/schema.go — config validation
// internal/config/io.go — config file I/O with env preservation
// internal/config/env.go — environment variable handling
// internal/config/paths.go — config/state directory resolution
// internal/config/defaults.go — default values
// internal/config/migration.go — legacy config migration
// internal/config/types.go — config type definitions
```

#### 7. Sessions → `internal/sessions/`

```go
// internal/sessions/store.go — session store (filesystem)
// internal/sessions/transcript.go — transcript read/write
// internal/sessions/metadata.go — session metadata
// internal/sessions/paths.go — session file paths
// internal/sessions/reset.go — session reset/archive
// internal/sessions/key.go — session key resolution
```

#### 8. Plugins → `internal/plugins/`

**Approach**: Go plugin interface with dynamic loading via `plugin` package or gRPC.

```go
// internal/plugins/registry.go — plugin registry
// internal/plugins/loader.go — plugin discovery and loading
// internal/plugins/manifest.go — plugin manifest parsing
// internal/plugins/hooks.go — hook execution
// internal/plugins/runtime.go — plugin runtime
// internal/plugins/sdk.go — plugin SDK types
```

#### 9. Cron → `internal/cron/`

```go
// internal/cron/service.go — cron service (robfig/cron)
// internal/cron/schedule.go — schedule management
// internal/cron/isolated.go — isolated agent execution
// internal/cron/delivery.go — result delivery
// internal/cron/store.go — job persistence
// internal/cron/reaper.go — stale session cleanup
```

#### 10. Infrastructure → `internal/infra/`

```go
// internal/infra/env.go — environment handling
// internal/infra/dotenv.go — .env loading
// internal/infra/retry.go — retry with backoff
// internal/infra/fetch.go — HTTP client wrapper
// internal/infra/bonjour.go — mDNS discovery
// internal/infra/device.go — device identity
// internal/infra/lock.go — file locking
// internal/infra/ssrf.go — SSRF protection
// internal/infra/tls.go — TLS management
// internal/infra/system_events.go — event bus
```

#### 11. Security → `internal/security/`

```go
// internal/security/audit.go — security audit
// internal/security/fix.go — auto-fix
// internal/security/secrets.go — timing-safe comparison
// internal/security/content.go — content sanitization
// internal/security/scanner.go — skill scanning
```

#### 12. Media → `internal/media/`

```go
// internal/media/store.go — media storage
// internal/media/fetch.go — media download
// internal/media/mime.go — MIME type detection
// internal/media/image.go — image processing
// internal/media/understanding/ — vision/audio analysis
// internal/media/links/ — URL content extraction
```

#### 13. Browser → `internal/browser/`

```go
// internal/browser/server.go — browser automation server
// internal/browser/session.go — browser session management
// internal/browser/tools.go — browser tool implementations
// internal/browser/cdp.go — CDP integration
// internal/browser/chrome.go — Chrome management
```

#### 14. TTS → `internal/tts/`

```go
// internal/tts/engine.go — TTS orchestration
// internal/tts/elevenlabs.go — ElevenLabs provider
// internal/tts/edge.go — Edge TTS provider
// internal/tts/system.go — System TTS
```

#### 15. ACP → `internal/acp/`

```go
// internal/acp/server.go — stdio NDJSON server
// internal/acp/client.go — gateway WS client
// internal/acp/session.go — session mapping
// internal/acp/events.go — event translation
```

#### 16. Daemon → `internal/daemon/`

```go
// internal/daemon/service.go — daemon lifecycle
// internal/daemon/launchd.go — macOS plist
// internal/daemon/systemd.go — Linux unit files
// internal/daemon/windows.go — Windows service
```

#### 17. Logging → `internal/logging/`

```go
// internal/logging/logger.go — structured logger (slog-based)
// internal/logging/redact.go — sensitive data redaction
// internal/logging/subsystem.go — subsystem-scoped loggers
```

---

## Package Structure

```
github.com/StellariumFoundation/goclaw/
├── cmd/
│   └── goclaw/
│       └── main.go                 # Binary entrypoint
│
├── internal/
│   ├── cli/                        # CLI commands (cobra)
│   │   ├── root.go
│   │   ├── gateway.go
│   │   ├── config.go
│   │   ├── models.go
│   │   ├── channels.go
│   │   ├── cron.go
│   │   ├── daemon.go
│   │   ├── browser.go
│   │   ├── nodes.go
│   │   ├── plugins.go
│   │   ├── security.go
│   │   ├── tui.go
│   │   └── acp.go
│   │
│   ├── gateway/                    # Gateway server
│   │   ├── server.go
│   │   ├── auth.go
│   │   ├── handler.go
│   │   ├── broadcast.go
│   │   ├── discovery.go
│   │   ├── openai_compat.go
│   │   ├── hooks.go
│   │   ├── methods/
│   │   │   ├── chat.go
│   │   │   ├── config.go
│   │   │   ├── sessions.go
│   │   │   ├── agents.go
│   │   │   ├── models.go
│   │   │   ├── cron.go
│   │   │   ├── nodes.go
│   │   │   ├── send.go
│   │   │   └── ...
│   │   └── protocol/
│   │       ├── frames.go
│   │       ├── types.go
│   │       └── schema.go
│   │
│   ├── agents/                     # AI agent engine
│   │   ├── runner.go
│   │   ├── streaming.go
│   │   ├── tools.go
│   │   ├── bash.go
│   │   ├── models.go
│   │   ├── system_prompt.go
│   │   ├── identity.go
│   │   ├── context_guard.go
│   │   ├── sandbox/
│   │   ├── skills/
│   │   └── auth/
│   │
│   ├── autoreply/                  # Auto-reply pipeline
│   │   ├── pipeline.go
│   │   ├── commands.go
│   │   ├── directives.go
│   │   ├── queue.go
│   │   ├── streaming.go
│   │   ├── dispatch.go
│   │   ├── heartbeat.go
│   │   └── envelope.go
│   │
│   ├── channels/                   # Channel abstraction
│   │   ├── registry.go
│   │   ├── types.go
│   │   ├── allowlist.go
│   │   ├── routing.go
│   │   ├── web/
│   │   ├── discord/
│   │   ├── telegram/
│   │   ├── slack/
│   │   ├── whatsapp/
│   │   ├── signal/
│   │   ├── matrix/
│   │   ├── msteams/
│   │   └── line/
│   │
│   ├── config/                     # Configuration
│   │   ├── config.go
│   │   ├── schema.go
│   │   ├── io.go
│   │   ├── env.go
│   │   ├── paths.go
│   │   ├── defaults.go
│   │   ├── migration.go
│   │   └── types.go
│   │
│   ├── sessions/                   # Session management
│   │   ├── store.go
│   │   ├── transcript.go
│   │   ├── metadata.go
│   │   ├── paths.go
│   │   ├── reset.go
│   │   └── key.go
│   │
│   ├── plugins/                    # Plugin system
│   │   ├── registry.go
│   │   ├── loader.go
│   │   ├── manifest.go
│   │   ├── hooks.go
│   │   ├── runtime.go
│   │   └── sdk.go
│   │
│   ├── cron/                       # Cron scheduling
│   │   ├── service.go
│   │   ├── schedule.go
│   │   ├── isolated.go
│   │   ├── delivery.go
│   │   ├── store.go
│   │   └── reaper.go
│   │
│   ├── infra/                      # Infrastructure
│   │   ├── env.go
│   │   ├── dotenv.go
│   │   ├── retry.go
│   │   ├── fetch.go
│   │   ├── bonjour.go
│   │   ├── device.go
│   │   ├── lock.go
│   │   ├── ssrf.go
│   │   ├── tls.go
│   │   └── events.go
│   │
│   ├── security/                   # Security
│   │   ├── audit.go
│   │   ├── fix.go
│   │   ├── secrets.go
│   │   ├── content.go
│   │   └── scanner.go
│   │
│   ├── media/                      # Media handling
│   │   ├── store.go
│   │   ├── fetch.go
│   │   ├── mime.go
│   │   ├── image.go
│   │   ├── understanding/
│   │   └── links/
│   │
│   ├── browser/                    # Browser automation
│   │   ├── server.go
│   │   ├── session.go
│   │   ├── tools.go
│   │   ├── cdp.go
│   │   └── chrome.go
│   │
│   ├── tts/                        # Text-to-speech
│   │   ├── engine.go
│   │   ├── elevenlabs.go
│   │   ├── edge.go
│   │   └── system.go
│   │
│   ├── acp/                        # ACP bridge
│   │   ├── server.go
│   │   ├── client.go
│   │   ├── session.go
│   │   └── events.go
│   │
│   ├── daemon/                     # Daemon management
│   │   ├── service.go
│   │   ├── launchd.go
│   │   ├── systemd.go
│   │   └── windows.go
│   │
│   ├── logging/                    # Logging
│   │   ├── logger.go
│   │   ├── redact.go
│   │   └── subsystem.go
│   │
│   └── tui/                        # Terminal UI
│       ├── app.go
│       ├── chat.go
│       ├── commands.go
│       └── components/
│
├── pkg/                            # Public API packages
│   ├── pluginsdk/                  # Plugin SDK (public)
│   │   ├── types.go
│   │   ├── channel.go
│   │   └── hooks.go
│   └── protocol/                   # Gateway protocol (public)
│       ├── frames.go
│       └── types.go
│
├── skills/                         # Skill definitions (copied from original)
│   └── ...
│
├── assets/                         # Static assets
│
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

---

## Implementation Priority

### Phase 1: Core Foundation (Weeks 1–3)

1. **Configuration system** (`internal/config/`)
   - Config struct definitions, loading, validation
   - Environment variable handling, .env loading
   - Config file I/O with env preservation
   - Path resolution (state dir, config path)

2. **Logging** (`internal/logging/`)
   - Structured logger with slog
   - Redaction, subsystem scoping

3. **Infrastructure utilities** (`internal/infra/`)
   - Retry, backoff, HTTP client
   - File locking, SSRF protection
   - Device identity, system events

4. **Session management** (`internal/sessions/`)
   - Filesystem-based session store
   - Transcript read/write
   - Session key resolution

### Phase 2: Gateway Server (Weeks 4–6)

5. **Gateway protocol** (`internal/gateway/protocol/`)
   - Frame types, message schemas
   - Serialization/deserialization

6. **Gateway server** (`internal/gateway/`)
   - HTTP server with WebSocket upgrade
   - Authentication (token, password, rate limiting)
   - WebSocket message routing
   - Client connection management, broadcast
   - Health endpoint

7. **Gateway methods** (`internal/gateway/methods/`)
   - Session CRUD
   - Config get/set
   - Agent listing
   - Model listing

### Phase 3: Agent Engine (Weeks 7–10)

8. **Agent runner** (`internal/agents/`)
   - Core execution loop with streaming
   - Model provider abstraction (OpenAI, Anthropic, Gemini)
   - Tool registry and execution
   - System prompt construction
   - Context window management

9. **Bash/shell tools** (`internal/agents/bash.go`)
   - Command execution with approval flow
   - PTY support via `creack/pty`

10. **Auth profiles** (`internal/agents/auth/`)
    - Multi-key rotation, failover

### Phase 4: Auto-Reply Pipeline (Weeks 11–13)

11. **Auto-reply pipeline** (`internal/autoreply/`)
    - Inbound message processing
    - Command detection and execution
    - Directive parsing (model, thinking)
    - Reply queue with concurrency control
    - Block streaming coalescer
    - Heartbeat handling

### Phase 5: Channel Integrations (Weeks 14–20)

12. **Channel abstraction** (`internal/channels/`)
    - Channel interface definitions
    - Registry, allowlist, routing

13. **Telegram** (`internal/channels/telegram/`)
    - Bot creation, webhook/polling
    - Message send/receive, media, threading

14. **Discord** (`internal/channels/discord/`)
    - Bot with gateway events
    - Message handling, threads, voice messages

15. **Slack** (`internal/channels/slack/`)
    - Bolt-equivalent, Socket Mode
    - Threading, actions

16. **WhatsApp** (`internal/channels/whatsapp/`)
    - whatsmeow integration
    - QR login, session persistence

17. **Web channel** (`internal/channels/web/`)
    - HTTP/WebSocket web chat

18. **Additional channels** (Signal, Matrix, MS Teams, LINE)

### Phase 6: CLI (Weeks 21–23)

19. **CLI framework** (`internal/cli/`)
    - Cobra command structure
    - All subcommands (gateway, config, models, channels, etc.)
    - Onboarding wizard

### Phase 7: Advanced Features (Weeks 24–28)

20. **Cron service** (`internal/cron/`)
21. **Plugin system** (`internal/plugins/`)
22. **Browser automation** (`internal/browser/`)
23. **Media understanding** (`internal/media/understanding/`)
24. **TTS** (`internal/tts/`)
25. **ACP bridge** (`internal/acp/`)
26. **Daemon management** (`internal/daemon/`)
27. **TUI** (`internal/tui/`)
28. **Security audit** (`internal/security/`)

### Phase 8: Polish & Parity (Weeks 29–32)

29. **OpenAI-compatible HTTP API**
30. **mDNS/Bonjour discovery**
31. **Sandbox/Docker integration**
32. **Skill system**
33. **Comprehensive test suite**
34. **Documentation**

---

## Technical Considerations

### Concurrency Model

Go's goroutine-based concurrency maps naturally to Openclaw's architecture:

| Openclaw Pattern | Go Equivalent |
|-----------------|---------------|
| WebSocket connections | One goroutine per connection (read + write) |
| Channel monitors | Long-running goroutines with context cancellation |
| Reply queue | Buffered channels + worker goroutines |
| Cron jobs | `robfig/cron` scheduler + goroutine per job |
| Heartbeat timers | `time.Ticker` in dedicated goroutines |
| Agent execution | Goroutine with context for cancellation/timeout |
| Streaming responses | Channels for streaming chunks |
| File watchers | `fsnotify` in dedicated goroutine |
| mDNS discovery | Background goroutine |

**Key patterns**:
- Use `context.Context` throughout for cancellation propagation
- Use `sync.WaitGroup` for graceful shutdown
- Use `errgroup` for concurrent operations with error collection
- Use `sync.Map` or mutex-protected maps for shared state
- Use buffered channels for producer-consumer patterns

### Error Handling

```go
// Use Go's explicit error handling throughout
// Wrap errors with context using fmt.Errorf or errors package
func (s *SessionStore) Load(key string) (*Session, error) {
    data, err := os.ReadFile(s.pathFor(key))
    if err != nil {
        return nil, fmt.Errorf("load session %s: %w", key, err)
    }
    // ...
}

// Define sentinel errors for known conditions
var (
    ErrSessionNotFound  = errors.New("session not found")
    ErrAuthFailed       = errors.New("authentication failed")
    ErrRateLimited      = errors.New("rate limited")
    ErrContextOverflow  = errors.New("context window overflow")
)

// Use custom error types for rich error information
type ProviderError struct {
    Provider string
    Model    string
    Status   int
    Message  string
    Err      error
}
```

### Testing Strategy

```go
// Unit tests: *_test.go files alongside source
// Table-driven tests for comprehensive coverage
func TestSessionKeyParse(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    SessionKey
        wantErr bool
    }{
        {"agent main", "agent:main:main", SessionKey{Agent: "main", Session: "main"}, false},
        {"invalid", "invalid", SessionKey{}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseSessionKey(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseSessionKey() error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("ParseSessionKey() = %v, want %v", got, tt.want)
            }
        })
    }
}

// Integration tests: use build tags
//go:build integration

// E2E tests: separate test binary or test suite
// Use testcontainers-go for Docker-dependent tests
```

**Testing tools**:
- `testing` (stdlib) — primary test framework
- `testify` — assertions and mocking (optional, prefer stdlib)
- `testcontainers-go` — Docker container management for integration tests
- `httptest` — HTTP server testing
- `net/http/httptest` — request/response recording

### Build System

```makefile
# Makefile
BINARY_NAME=goclaw
MODULE=github.com/StellariumFoundation/goclaw
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

.PHONY: build run clean test lint

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/goclaw

run:
	go run ./cmd/goclaw

test:
	go test ./... -race -count=1

test-coverage:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

vet:
	go vet ./...

# Cross-compilation
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/goclaw-linux-amd64 ./cmd/goclaw
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/goclaw-linux-arm64 ./cmd/goclaw
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/goclaw-darwin-amd64 ./cmd/goclaw
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/goclaw-darwin-arm64 ./cmd/goclaw
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/goclaw-windows-amd64.exe ./cmd/goclaw

docker:
	docker build -t goclaw:$(VERSION) .
```

### Dockerfile

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /goclaw ./cmd/goclaw

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -u 1000 goclaw
USER goclaw
COPY --from=builder /goclaw /usr/local/bin/goclaw
EXPOSE 18789
ENTRYPOINT ["goclaw"]
CMD ["gateway"]
```

### Configuration Approach

Use Go structs with JSON/YAML tags and validation:

```go
type Config struct {
    Gateway   GatewayConfig   `json:"gateway" yaml:"gateway"`
    Agents    AgentsConfig    `json:"agents" yaml:"agents"`
    Channels  ChannelsConfig  `json:"channels" yaml:"channels"`
    Models    ModelsConfig    `json:"models" yaml:"models"`
    Cron      CronConfig      `json:"cron" yaml:"cron"`
    Hooks     HooksConfig     `json:"hooks" yaml:"hooks"`
    Plugins   PluginsConfig   `json:"plugins" yaml:"plugins"`
    Sandbox   SandboxConfig   `json:"sandbox" yaml:"sandbox"`
    Security  SecurityConfig  `json:"security" yaml:"security"`
    TTS       TTSConfig       `json:"tts" yaml:"tts"`
    Browser   BrowserConfig   `json:"browser" yaml:"browser"`
}

type GatewayConfig struct {
    Port     int            `json:"port" yaml:"port" default:"18789"`
    Bind     string         `json:"bind" yaml:"bind" default:"loopback"`
    Auth     GatewayAuth    `json:"auth" yaml:"auth"`
    Remote   *RemoteConfig  `json:"remote,omitempty" yaml:"remote,omitempty"`
    TLS      *TLSConfig     `json:"tls,omitempty" yaml:"tls,omitempty"`
    Tools    ToolsConfig    `json:"tools" yaml:"tools"`
}
```

### Provider Abstraction

```go
// internal/agents/provider.go
type Provider interface {
    Name() string
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamChunk, error)
    ListModels(ctx context.Context) ([]Model, error)
}

type ChatRequest struct {
    Model       string
    Messages    []Message
    Tools       []ToolDef
    MaxTokens   int
    Temperature *float64
    Stream      bool
}

// Implementations
type OpenAIProvider struct { ... }
type AnthropicProvider struct { ... }
type GeminiProvider struct { ... }
type BedrockProvider struct { ... }
type OllamaProvider struct { ... }
```

### Channel Interface

```go
// internal/channels/types.go
type Channel interface {
    ID() string
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Send(ctx context.Context, msg *OutboundMessage) error
}

type MonitorChannel interface {
    Channel
    OnMessage(handler MessageHandler)
}

type ThreadingChannel interface {
    Channel
    SendThreaded(ctx context.Context, msg *OutboundMessage, threadID string) error
}

type MediaChannel interface {
    Channel
    SendMedia(ctx context.Context, media *MediaMessage) error
    DownloadMedia(ctx context.Context, ref *MediaRef) (io.ReadCloser, error)
}
```

---

## Migration Strategy

### Incremental Approach

The rewrite should proceed incrementally, allowing the Go and TypeScript versions to coexist:

#### Stage 1: Gateway Protocol Compatibility

1. Implement the Go gateway server with **identical WebSocket protocol**
2. Ensure existing TypeScript clients (native apps, web UI) can connect to the Go gateway
3. Run both servers side-by-side for validation

#### Stage 2: Config Compatibility

1. Read the same `openclaw.json` config format
2. Use the same `~/.openclaw/` state directory structure
3. Support the same environment variables

#### Stage 3: Channel-by-Channel Migration

1. Start with the simplest channel (web) and validate end-to-end
2. Add channels one at a time, testing against the same messaging platforms
3. Use feature flags to enable/disable channels

#### Stage 4: Agent Parity

1. Implement the agent runner with identical model provider API calls
2. Validate response quality matches TypeScript version
3. Ensure tool execution produces identical results

#### Stage 5: CLI Parity

1. Implement CLI commands matching the TypeScript CLI
2. Ensure `goclaw` is a drop-in replacement for `openclaw`
3. Support the same command syntax and flags

### Compatibility Testing

```bash
# Run both servers and compare responses
openclaw gateway --port 18789 &
goclaw gateway --port 18790 &

# Send identical requests to both
curl -X POST http://localhost:18789/v1/chat/completions -d '...' > ts_response.json
curl -X POST http://localhost:18790/v1/chat/completions -d '...' > go_response.json

# Compare (ignoring timing differences)
diff <(jq -S . ts_response.json) <(jq -S . go_response.json)
```

### Feature Flags

```go
// internal/config/features.go
type FeatureFlags struct {
    EnableDiscord   bool `json:"enableDiscord" env:"GOCLAW_ENABLE_DISCORD"`
    EnableTelegram  bool `json:"enableTelegram" env:"GOCLAW_ENABLE_TELEGRAM"`
    EnableSlack     bool `json:"enableSlack" env:"GOCLAW_ENABLE_SLACK"`
    EnableWhatsApp  bool `json:"enableWhatsApp" env:"GOCLAW_ENABLE_WHATSAPP"`
    EnableBrowser   bool `json:"enableBrowser" env:"GOCLAW_ENABLE_BROWSER"`
    EnableSandbox   bool `json:"enableSandbox" env:"GOCLAW_ENABLE_SANDBOX"`
    EnableTTS       bool `json:"enableTTS" env:"GOCLAW_ENABLE_TTS"`
    EnablePlugins   bool `json:"enablePlugins" env:"GOCLAW_ENABLE_PLUGINS"`
}
```

### Data Migration

- Session transcripts: Same JSON format, read directly
- Config files: Same `openclaw.json` format
- State directory: Same `~/.openclaw/` layout
- No database migration needed (filesystem-based storage)

### Rollback Plan

- Keep the TypeScript version maintained during migration
- Use the same config/state directory so switching back is seamless
- Document any breaking changes in protocol or config format

---

## Appendix: Key Go Libraries

| Library | Version | Purpose |
|---------|---------|---------|
| `github.com/spf13/cobra` | v1.9+ | CLI framework |
| `github.com/spf13/viper` | v1.19+ | Configuration |
| `nhooyr.io/websocket` | v1.8+ | WebSocket |
| `github.com/go-chi/chi/v5` | v5.2+ | HTTP router |
| `github.com/robfig/cron/v3` | v3.0+ | Cron scheduling |
| `github.com/joho/godotenv` | v1.5+ | .env loading |
| `github.com/fatih/color` | v1.18+ | Terminal colors |
| `github.com/charmbracelet/bubbletea` | v1.3+ | TUI framework |
| `github.com/charmbracelet/lipgloss` | v1.1+ | TUI styling |
| `github.com/fsnotify/fsnotify` | v1.8+ | File watching |
| `github.com/gofrs/flock` | v0.12+ | File locking |
| `github.com/bwmarrin/discordgo` | v0.28+ | Discord |
| `github.com/go-telegram-bot-api/telegram-bot-api/v5` | v5.5+ | Telegram |
| `github.com/slack-go/slack` | v0.15+ | Slack |
| `go.mau.fi/whatsmeow` | v0.0.0+ | WhatsApp |
| `github.com/chromedp/chromedp` | v0.11+ | Browser automation |
| `github.com/creack/pty` | v1.1+ | PTY |
| `github.com/hashicorp/mdns` | v1.0+ | mDNS |
| `github.com/yuin/goldmark` | v1.7+ | Markdown |
| `github.com/skip2/go-qrcode` | v0.0.0+ | QR codes |
| `github.com/mattn/go-sqlite3` | v1.14+ | SQLite |
| `github.com/aws/aws-sdk-go-v2` | v1.36+ | AWS Bedrock |
| `github.com/PuerkitoBio/goquery` | v1.10+ | HTML parsing |
| `github.com/go-shiori/go-readability` | v0.0.0+ | Content extraction |
| `github.com/disintegration/imaging` | v1.6+ | Image processing |

---

*This document serves as the complete blueprint for rewriting Openclaw from TypeScript/Node.js to Go 1.26.0. Each section is designed to be actionable and can be used as a reference during implementation.*
