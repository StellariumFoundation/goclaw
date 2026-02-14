# OpenClaw → GoClaw: Complete Go 1.26.0 Rewrite Plan

## Executive Summary

**OpenClaw** is a multi-channel AI gateway and digital worker platform that bridges AI language models (OpenAI, Anthropic, Google Gemini, OpenRouter, etc.) with messaging channels (Telegram, Discord, Slack, Signal, WhatsApp, iMessage, MS Teams, etc.). It provides a unified gateway server with WebSocket and HTTP APIs, a CLI for management, a TUI for interactive use, a plugin/extension system, cron scheduling, browser automation, media processing, and native macOS/iOS/Android apps.

The project is written in **TypeScript (ESM)** running on **Node.js 22+**, using **pnpm** as the package manager. It is a monorepo with ~500+ source files under `src/`, extensions under `extensions/`, skills (markdown-based agent instructions) under `skills/`, native apps under `apps/`, and npm packages under `packages/`.

**Goal of the Go rewrite**: Rewrite the entire OpenClaw backend as **GoClaw** in **Go 1.26.0**, targeting:
- Single static binary deployment (no Node.js runtime dependency)
- Superior concurrency via goroutines and channels
- Lower memory footprint and faster startup
- Type-safe, compiled codebase with excellent tooling
- Maintain full feature parity with the TypeScript implementation

---

## Original Architecture Overview

### Language & Framework Stack

| Layer | Technology |
|-------|-----------|
| Language | TypeScript 5.9+ (ESM modules) |
| Runtime | Node.js 22+ (Bun also supported for dev) |
| Build | tsdown (Rolldown-based bundler) |
| Test | Vitest with V8 coverage |
| Lint/Format | Oxlint + Oxfmt |
| Type-check | TypeScript / tsgo |
| Package Manager | pnpm 10.23 (monorepo workspaces) |
| Schema Validation | Zod 4 + @sinclair/typebox + Ajv |
| HTTP Server | Express 5 + raw Node.js `http`/`https` |
| WebSocket | ws library |
| CLI Framework | Commander.js |
| Config Format | JSON5 (openclaw.json) |
| Logging | tslog + custom subsystem logger |

### Purpose & Core Capabilities

1. **AI Gateway Server** — WebSocket + HTTP server that manages AI agent sessions, routes messages between channels and LLM providers, handles authentication, and serves a control UI
2. **Multi-Channel Messaging** — Integrates with Telegram (grammy), Discord (Carbon/discord-api-types), Slack (Bolt), Signal (signal-cli SSE), WhatsApp (Baileys), iMessage (BlueBubbles), MS Teams (extension), LINE, and more
3. **AI Agent Runtime** — Embeds the Pi agent framework (@mariozechner/pi-*) for coding agent capabilities with tool use (bash, browser, web search, file operations, etc.)
4. **Plugin/Extension System** — Runtime-loadable plugins with hooks, tools, and channel adapters
5. **Browser Automation** — Playwright-based browser control with CDP integration
6. **Cron Scheduling** — Scheduled agent tasks with delivery to channels
7. **TUI** — Terminal UI for interactive agent sessions
8. **CLI** — Comprehensive command-line interface for configuration, management, and operations
9. **Native Apps** — macOS (Swift/SwiftUI), iOS, Android companion apps connecting via gateway protocol

### Directory/Module Structure

```
Openclaw/
├── src/                          # Main TypeScript source
│   ├── index.ts                  # Library entry point
│   ├── entry.ts                  # CLI entry point
│   ├── runtime.ts                # Runtime environment singleton
│   ├── logger.ts                 # Global logger
│   ├── logging.ts                # Console capture
│   │
│   ├── cli/                      # CLI command registration & execution
│   │   ├── program.ts            # Main CLI program builder
│   │   ├── run-main.ts           # CLI bootstrap (dotenv, routing, argv)
│   │   ├── program/              # Command registration modules
│   │   ├── gateway-cli.ts        # Gateway subcommands
│   │   ├── channels-cli.ts       # Channel management commands
│   │   ├── browser-cli.ts        # Browser automation commands
│   │   ├── config-cli.ts         # Config get/set commands
│   │   ├── cron-cli.ts           # Cron job commands
│   │   ├── daemon-cli.ts         # Daemon install/status commands
│   │   ├── models-cli.ts         # Model listing/selection
│   │   ├── plugins-cli.ts        # Plugin management
│   │   ├── skills-cli.ts         # Skills management
│   │   ├── nodes-cli.ts          # Mobile node commands
│   │   ├── tui-cli.ts            # TUI launcher
│   │   └── ...                   # Many more CLI modules
│   │
│   ├── commands/                 # Command implementations (business logic)
│   │   ├── agent.ts              # Agent run command
│   │   ├── configure.ts          # Interactive configuration wizard
│   │   ├── onboard.ts            # Onboarding flow
│   │   ├── doctor.ts             # Diagnostic/health checks
│   │   ├── channels.ts           # Channel add/remove/status
│   │   ├── models.ts             # Model management
│   │   ├── sandbox.ts            # Sandbox management
│   │   ├── health.ts             # Health check command
│   │   ├── sessions.ts           # Session management
│   │   └── ...                   # Auth, OAuth, status, etc.
│   │
│   ├── gateway/                  # Gateway server (core)
│   │   ├── server.impl.ts        # Main gateway server startup
│   │   ├── server-http.ts        # HTTP request routing (hooks, OpenAI compat, control UI)
│   │   ├── server-ws-runtime.ts  # WebSocket connection handling
│   │   ├── server-methods.ts     # Gateway RPC method handlers
│   │   ├── server-methods/       # Individual method implementations
│   │   │   ├── agents.ts         # Agent CRUD
│   │   │   ├── chat.ts           # Chat send/inject/abort
│   │   │   ├── config.ts         # Config get/set/patch
│   │   │   ├── sessions.ts       # Session management
│   │   │   ├── models.ts         # Model listing
│   │   │   ├── cron.ts           # Cron operations
│   │   │   ├── health.ts         # Health snapshots
│   │   │   ├── nodes.ts          # Mobile node management
│   │   │   ├── logs.ts           # Log tailing
│   │   │   └── ...               # Devices, exec-approvals, etc.
│   │   ├── server-channels.ts    # Channel lifecycle management
│   │   ├── server-chat.ts        # Agent event handling
│   │   ├── server-cron.ts        # Cron service integration
│   │   ├── server-discovery.ts   # Bonjour/mDNS discovery
│   │   ├── server-plugins.ts     # Plugin loading
│   │   ├── server-tailscale.ts   # Tailscale exposure
│   │   ├── openai-http.ts        # OpenAI-compatible HTTP API
│   │   ├── openresponses-http.ts # OpenAI Responses API compat
│   │   ├── auth.ts               # Gateway authentication
│   │   ├── client.ts             # Gateway WebSocket client
│   │   ├── protocol/             # Protocol schema definitions (TypeBox)
│   │   │   ├── schema/           # Individual schema modules
│   │   │   └── index.ts          # Ajv validators
│   │   ├── server/               # Server infrastructure
│   │   │   ├── ws-connection.ts  # WebSocket connection management
│   │   │   ├── health-state.ts   # Health state cache
│   │   │   ├── tls.ts            # TLS configuration
│   │   │   └── http-listen.ts    # HTTP listener
│   │   └── ...
│   │
│   ├── agents/                   # AI agent runtime
│   │   ├── pi-embedded-runner/   # Embedded Pi agent runner
│   │   │   ├── run.ts            # Main agent run loop
│   │   │   ├── model.ts          # Model selection
│   │   │   ├── compact.ts        # Context compaction
│   │   │   ├── history.ts        # Session history
│   │   │   └── ...
│   │   ├── pi-embedded-subscribe.ts  # Agent event subscription
│   │   ├── pi-embedded-helpers.ts    # Agent helper utilities
│   │   ├── pi-tools.ts           # Tool definitions for agents
│   │   ├── tools/                # Individual tool implementations
│   │   │   ├── bash-tool.ts      # Shell execution
│   │   │   ├── browser-tool.ts   # Browser automation
│   │   │   ├── web-fetch.ts      # Web fetching
│   │   │   ├── web-search.ts     # Web search
│   │   │   ├── image-tool.ts     # Image generation
│   │   │   ├── memory-tool.ts    # Memory/RAG
│   │   │   ├── message-tool.ts   # Messaging
│   │   │   ├── sessions-*.ts     # Session management tools
│   │   │   ├── discord-actions.ts # Discord-specific actions
│   │   │   ├── slack-actions.ts  # Slack-specific actions
│   │   │   ├── telegram-actions.ts # Telegram-specific actions
│   │   │   └── ...
│   │   ├── bash-tools.ts         # Bash/shell execution engine
│   │   ├── sandbox/              # Docker sandbox management
│   │   │   ├── docker.ts         # Docker container lifecycle
│   │   │   ├── config.ts         # Sandbox configuration
│   │   │   ├── fs-bridge.ts      # Filesystem bridge
│   │   │   └── ...
│   │   ├── skills/               # Skills system
│   │   ├── auth-profiles/        # Auth profile management
│   │   ├── models-config.ts      # Model configuration
│   │   ├── system-prompt.ts      # System prompt generation
│   │   └── ...
│   │
│   ├── config/                   # Configuration system
│   │   ├── config.ts             # Config loading/writing barrel
│   │   ├── io.ts                 # Config file I/O
│   │   ├── schema.ts             # JSON Schema (TypeBox)
│   │   ├── zod-schema.ts         # Zod validation schemas
│   │   ├── types.ts              # Type definitions barrel
│   │   ├── types.*.ts            # Per-domain type definitions
│   │   ├── validation.ts         # Config validation
│   │   ├── legacy-migrate.ts     # Legacy config migration
│   │   ├── sessions/             # Session config & storage
│   │   └── ...
│   │
│   ├── auto-reply/               # Auto-reply engine
│   │   ├── reply/                # Reply pipeline
│   │   │   ├── agent-runner.ts   # Agent execution for replies
│   │   │   ├── commands.ts       # Inline command handling
│   │   │   ├── directive-handling.ts # Reply directives
│   │   │   ├── queue.ts          # Reply queue management
│   │   │   ├── dispatcher-registry.ts # Reply dispatchers
│   │   │   └── ...
│   │   ├── commands-registry.ts  # Command registry
│   │   ├── heartbeat.ts          # Heartbeat/scheduled replies
│   │   └── ...
│   │
│   ├── channels/                 # Channel abstraction layer
│   │   ├── registry.ts           # Channel registry
│   │   ├── plugins/              # Channel plugin system
│   │   │   ├── index.ts          # Plugin loading & catalog
│   │   │   ├── types.ts          # Channel plugin interfaces
│   │   │   ├── onboarding/       # Per-channel onboarding
│   │   │   ├── outbound/         # Per-channel outbound adapters
│   │   │   ├── normalize/        # Per-channel message normalization
│   │   │   └── actions/          # Per-channel action handlers
│   │   └── ...
│   │
│   ├── telegram/                 # Telegram channel implementation
│   │   ├── bot.ts                # Grammy bot setup
│   │   ├── monitor.ts            # Message monitoring
│   │   ├── send.ts               # Message sending
│   │   └── ...
│   │
│   ├── discord/                  # Discord channel implementation
│   │   ├── monitor.ts            # Discord event monitoring
│   │   ├── send.ts               # Message sending
│   │   └── ...
│   │
│   ├── slack/                    # Slack channel implementation
│   │   ├── monitor.ts            # Slack event monitoring
│   │   ├── send.ts               # Message sending
│   │   ├── http/                 # Slack HTTP mode
│   │   └── ...
│   │
│   ├── signal/                   # Signal channel implementation
│   ├── web/ (WhatsApp)           # WhatsApp Web implementation (Baileys)
│   ├── imessage/                 # iMessage (BlueBubbles) implementation
│   │
│   ├── browser/                  # Browser automation
│   │   ├── server.ts             # Browser control server
│   │   ├── pw-session.ts         # Playwright session management
│   │   ├── pw-tools-core.ts      # Browser tool implementations
│   │   ├── cdp.ts                # Chrome DevTools Protocol
│   │   ├── routes/               # HTTP routes for browser control
│   │   └── ...
│   │
│   ├── infra/                    # Infrastructure utilities
│   │   ├── env.ts                # Environment variable handling
│   │   ├── dotenv.ts             # .env file loading
│   │   ├── errors.ts             # Error formatting
│   │   ├── fetch.ts              # HTTP fetch utilities
│   │   ├── retry.ts              # Retry logic
│   │   ├── backoff.ts            # Exponential backoff
│   │   ├── heartbeat-runner.ts   # Heartbeat scheduling
│   │   ├── bonjour.ts            # mDNS/Bonjour discovery
│   │   ├── device-identity.ts    # Device identity management
│   │   ├── device-pairing.ts     # Device pairing
│   │   ├── exec-approvals.ts     # Execution approval system
│   │   ├── ssh-tunnel.ts         # SSH tunneling
│   │   ├── tailscale.ts          # Tailscale integration
│   │   ├── update-check.ts       # Update checking
│   │   ├── provider-usage.ts     # Provider usage tracking
│   │   ├── system-events.ts      # System event bus
│   │   ├── net/                  # Network security (SSRF protection)
│   │   ├── tls/                  # TLS fingerprinting
│   │   └── ...
│   │
│   ├── plugins/                  # Plugin system
│   │   ├── loader.ts             # Plugin loading
│   │   ├── registry.ts           # Plugin registry
│   │   ├── runtime.ts            # Plugin runtime
│   │   ├── hooks.ts              # Hook system
│   │   ├── services.ts           # Plugin services
│   │   ├── tools.ts              # Plugin tool registration
│   │   └── ...
│   │
│   ├── hooks/                    # Hook system
│   │   ├── hooks.ts              # Hook execution
│   │   ├── loader.ts             # Hook loading
│   │   ├── bundled/              # Built-in hooks
│   │   └── ...
│   │
│   ├── cron/                     # Cron scheduling
│   │   ├── service.ts            # Cron service
│   │   ├── store.ts              # Cron job storage
│   │   ├── schedule.ts           # Schedule parsing
│   │   ├── isolated-agent/       # Isolated agent execution
│   │   └── ...
│   │
│   ├── media/                    # Media processing
│   │   ├── server.ts             # Media server
│   │   ├── store.ts              # Media storage
│   │   ├── fetch.ts              # Media fetching
│   │   ├── image-ops.ts          # Image operations (sharp)
│   │   └── ...
│   │
│   ├── tui/                      # Terminal UI
│   │   ├── tui.ts                # Main TUI
│   │   ├── components/           # TUI components
│   │   └── ...
│   │
│   ├── logging/                  # Logging subsystem
│   │   ├── subsystem.ts          # Subsystem logger
│   │   ├── console.ts            # Console capture
│   │   ├── redact.ts             # Secret redaction
│   │   └── ...
│   │
│   ├── security/                 # Security utilities
│   │   ├── audit.ts              # Security auditing
│   │   ├── secret-equal.ts       # Timing-safe comparison
│   │   ├── skill-scanner.ts      # Skill security scanning
│   │   └── ...
│   │
│   ├── process/                  # Process management
│   │   ├── command-queue.ts      # Command queue
│   │   ├── exec.ts               # Process execution
│   │   └── ...
│   │
│   ├── daemon/                   # System daemon management
│   │   ├── launchd.ts            # macOS launchd
│   │   ├── systemd.ts            # Linux systemd
│   │   ├── schtasks.ts           # Windows scheduled tasks
│   │   └── ...
│   │
│   ├── tts/                      # Text-to-speech
│   ├── acp/                      # Agent Client Protocol
│   ├── canvas-host/              # Canvas/A2UI hosting
│   ├── node-host/                # Mobile node hosting
│   ├── wizard/                   # Onboarding wizard
│   ├── routing/                  # Message routing
│   ├── sessions/                 # Session utilities
│   ├── terminal/                 # Terminal formatting
│   ├── markdown/                 # Markdown processing
│   ├── pairing/                  # Device pairing
│   ├── plugin-sdk/               # Plugin SDK
│   └── shared/                   # Shared utilities
│
├── extensions/                   # Extension plugins
│   ├── msteams/                  # Microsoft Teams
│   ├── whatsapp/                 # WhatsApp (extension)
│   ├── open-prose/               # Open Prose language
│   ├── llm-task/                 # LLM task tool
│   ├── memory-core/              # Memory/RAG core
│   ├── copilot-proxy/            # GitHub Copilot proxy
│   ├── thread-ownership/         # Thread ownership
│   └── phone-control/            # Phone control
│
├── skills/                       # Skill definitions (markdown)
│   ├── github/                   # GitHub skills
│   ├── slack/                    # Slack skills
│   ├── discord/                  # Discord skills
│   ├── coding-agent/             # Coding agent skills
│   └── ...                       # 50+ skill directories
│
├── apps/                         # Native applications
│   ├── macos/                    # macOS app (Swift)
│   ├── shared/OpenClawKit/       # Shared Swift framework
│   └── ...
│
├── packages/                     # npm packages
│   ├── clawdbot/                 # Legacy compat package
│   └── moltbot/                  # Legacy compat package
│
├── docs/                         # Documentation (Mintlify)
├── scripts/                      # Build/dev scripts
└── Swabble/                      # Additional tooling
```

### Entry Points and Main Execution Flow

1. **CLI Entry** (`openclaw.mjs` → `src/entry.ts` → `src/cli/run-main.ts`):
   - Loads `.env` files, normalizes environment
   - Parses argv, routes to appropriate subcommand
   - Builds Commander.js program with all registered commands

2. **Gateway Server** (`openclaw gateway run` → `src/gateway/server.impl.ts`):
   - Loads config from `~/.openclaw/openclaw.json`
   - Starts HTTP/HTTPS + WebSocket server
   - Initializes channel managers (Telegram, Discord, Slack, etc.)
   - Starts cron service, plugin system, discovery (Bonjour)
   - Handles WebSocket connections with JSON-RPC-like protocol
   - Serves OpenAI-compatible HTTP API
   - Serves control UI (web dashboard)

3. **Agent Execution** (`openclaw agent` → `src/commands/agent.ts`):
   - Initializes Pi embedded agent runner
   - Manages session state, tool registration
   - Streams responses back to channels

4. **TUI** (`openclaw tui` → `src/tui/tui.ts`):
   - Interactive terminal interface using Pi TUI framework

### Core Components & Responsibilities

| Component | Responsibility |
|-----------|---------------|
| `gateway/` | WebSocket/HTTP server, protocol handling, client management |
| `agents/` | AI agent runtime, tool execution, session management |
| `auto-reply/` | Message processing pipeline, reply generation, command handling |
| `channels/` | Channel abstraction, plugin system, message normalization |
| `config/` | Configuration loading, validation, schema, migration |
| `cli/` | CLI command registration and argument parsing |
| `commands/` | Business logic for CLI commands |
| `infra/` | Cross-cutting infrastructure (env, fetch, retry, discovery) |
| `plugins/` | Plugin loading, registry, hook system |
| `browser/` | Playwright browser automation |
| `cron/` | Scheduled task execution |
| `media/` | Media processing and storage |
| `security/` | Audit, secret handling, SSRF protection |
| `logging/` | Structured logging with subsystems |
| `tui/` | Terminal user interface |
| `daemon/` | System service management (launchd/systemd) |

### Data Models and Schemas

Configuration is defined via dual schema systems:
- **TypeBox schemas** (`@sinclair/typebox`) for the gateway protocol (JSON Schema generation + Ajv validation)
- **Zod schemas** (`zod`) for config file validation

Key data models:
- `OpenClawConfig` — Main configuration object with nested channel, agent, model, gateway, plugin configs
- `SessionConfig` — Session scoping, reset, and policy
- `AgentConfig` — Agent identity, model, tools, skills
- `GatewayProtocol` — WebSocket frames (connect, hello-ok, request/response, events)
- `ChatEvent` — Streaming chat events (text, tool calls, errors)
- `CronJob` — Scheduled job definitions
- `ChannelPlugin` — Channel adapter interface

### API Endpoints and Protocols

**WebSocket Protocol** (primary):
- JSON-RPC-like request/response over WebSocket
- Methods: `chat.send`, `chat.inject`, `chat.abort`, `config.get`, `config.set`, `config.patch`, `config.apply`, `sessions.list`, `sessions.patch`, `sessions.reset`, `models.list`, `agents.list`, `agents.create`, `agents.update`, `agents.delete`, `cron.list`, `cron.add`, `cron.remove`, `cron.run`, `channels.status`, `channels.logout`, `health`, `logs.tail`, `nodes.*`, `devices.*`, `exec-approvals.*`, `update.run`, `skills.*`, `browser.*`, `tts.*`
- Events: `chat`, `agent-event`, `tick`, `shutdown`, `snapshot`

**HTTP API**:
- `/v1/chat/completions` — OpenAI-compatible chat completions
- `/v1/responses` — OpenAI Responses API compatible
- `/api/hooks/wake` — Webhook to wake agent
- `/api/hooks/message` — Webhook to send message to agent
- `/api/slack/*` — Slack HTTP mode endpoints
- `/api/plugins/*` — Plugin HTTP endpoints
- `/api/tools/invoke` — Tool invocation endpoint
- `/control/*` — Control UI (web dashboard)
- `/canvas/*` — Canvas/A2UI hosting
- `/media/*` — Media serving

### Configuration and Environment Management

- Primary config: `~/.openclaw/openclaw.json` (JSON5)
- Environment: `.env` files (process env > `./.env` > `~/.openclaw/.env` > config `env` block)
- State directory: `~/.openclaw/` (sessions, credentials, agent data)
- Config includes: `$include` directive for splitting config files
- Runtime overrides via environment variables (`OPENCLAW_*`)

### External Dependencies & Integrations

| Category | Dependencies |
|----------|-------------|
| LLM Providers | OpenAI, Anthropic, Google Gemini, OpenRouter, AWS Bedrock, Ollama, vLLM, Minimax, HuggingFace, Qwen, xAI, Chutes |
| Messaging | grammy (Telegram), @buape/carbon (Discord), @slack/bolt (Slack), @whiskeysockets/baileys (WhatsApp), @line/bot-sdk (LINE) |
| Browser | playwright-core |
| Database | SQLite (sqlite-vec for vector search) |
| Media | sharp (image processing), pdfjs-dist (PDF), node-edge-tts (TTS) |
| Discovery | @homebridge/ciao (mDNS/Bonjour) |
| Network | undici, https-proxy-agent |
| Schema | zod, @sinclair/typebox, ajv |
| CLI | commander, @clack/prompts |
| Misc | chokidar (file watching), croner (cron), ws (WebSocket), yaml, json5, jszip |

### Concurrency Model

- **Node.js event loop** with async/await throughout
- **AbortController/AbortSignal** for cancellation
- **EventEmitter** patterns for pub/sub
- **Process spawning** for bash tool execution (node-pty for PTY)
- **Worker pool** via Vitest forks for testing
- **Queue-based** reply processing with concurrency limits (lanes)

### Error Handling Patterns

- Try/catch with typed error classes
- Custom error types (e.g., `GatewayError`, `NodeError`)
- Error codes in protocol responses
- Retry with exponential backoff for transient errors
- Unhandled rejection handler for crash prevention

### Testing Approach

- **Vitest** with V8 coverage (70% threshold)
- Colocated `*.test.ts` files
- E2E tests (`*.e2e.test.ts`) for integration scenarios
- Live tests (`*.live.test.ts`) for real API testing
- Docker-based E2E tests for full system testing
- Mock helpers for gateway, agents, channels

---

## Module-by-Module Mapping

### `src/cli/` → `cmd/openclaw/` + `internal/cli/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `run-main.ts` | `cmd/openclaw/main.go` | CLI entry point, env loading, command routing |
| `program.ts`, `program/build-program.ts` | `internal/cli/root.go` | Root command builder (cobra) |
| `program/register.agent.ts` | `internal/cli/cmd_agent.go` | Agent command registration |
| `program/register.configure.ts` | `internal/cli/cmd_configure.go` | Configure command |
| `program/register.onboard.ts` | `internal/cli/cmd_onboard.go` | Onboard command |
| `program/register.status-health-sessions.ts` | `internal/cli/cmd_status.go` | Status/health/sessions commands |
| `program/register.message.ts` | `internal/cli/cmd_message.go` | Message commands |
| `program/register.subclis.ts` | `internal/cli/subclis.go` | Sub-CLI registration |
| `gateway-cli.ts`, `gateway-cli/` | `internal/cli/cmd_gateway.go` | Gateway subcommands |
| `channels-cli.ts` | `internal/cli/cmd_channels.go` | Channel management |
| `browser-cli.ts` | `internal/cli/cmd_browser.go` | Browser commands |
| `config-cli.ts` | `internal/cli/cmd_config.go` | Config commands |
| `cron-cli.ts`, `cron-cli/` | `internal/cli/cmd_cron.go` | Cron commands |
| `daemon-cli.ts`, `daemon-cli/` | `internal/cli/cmd_daemon.go` | Daemon commands |
| `models-cli.ts` | `internal/cli/cmd_models.go` | Model commands |
| `plugins-cli.ts` | `internal/cli/cmd_plugins.go` | Plugin commands |
| `skills-cli.ts` | `internal/cli/cmd_skills.go` | Skills commands |
| `nodes-cli.ts`, `nodes-cli/` | `internal/cli/cmd_nodes.go` | Node commands |
| `tui-cli.ts` | `internal/cli/cmd_tui.go` | TUI launcher |
| `update-cli.ts`, `update-cli/` | `internal/cli/cmd_update.go` | Update commands |
| `deps.ts` | `internal/cli/deps.go` | Dependency injection |
| `argv.ts` | `internal/cli/argv.go` | Argument parsing helpers |
| `progress.ts` | `internal/cli/progress.go` | Progress indicators |
| `prompt.ts` | `internal/cli/prompt.go` | Interactive prompts |

### `src/commands/` → `internal/commands/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `agent.ts` | `internal/commands/agent.go` | Agent run command |
| `configure.ts` | `internal/commands/configure.go` | Configuration wizard |
| `onboard.ts`, `onboard-*.ts` | `internal/commands/onboard.go` | Onboarding flow |
| `doctor.ts`, `doctor-*.ts` | `internal/commands/doctor.go` | Diagnostic checks |
| `channels.ts`, `channels/` | `internal/commands/channels.go` | Channel operations |
| `models.ts`, `models/` | `internal/commands/models.go` | Model management |
| `sandbox.ts` | `internal/commands/sandbox.go` | Sandbox management |
| `health.ts` | `internal/commands/health.go` | Health checks |
| `sessions.ts` | `internal/commands/sessions.go` | Session management |
| `auth-choice.ts`, `auth-choice.*.ts` | `internal/commands/auth.go` | Auth provider selection |
| `dashboard.ts` | `internal/commands/dashboard.go` | Dashboard command |
| `status*.ts` | `internal/commands/status.go` | Status reporting |
| `reset.ts` | `internal/commands/reset.go` | Reset command |
| `setup.ts` | `internal/commands/setup.go` | Setup command |

### `src/gateway/` → `internal/gateway/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `server.impl.ts` | `internal/gateway/server.go` | Gateway server startup & lifecycle |
| `server-http.ts` | `internal/gateway/http.go` | HTTP request routing |
| `server-ws-runtime.ts` | `internal/gateway/ws.go` | WebSocket handler |
| `server-methods.ts` | `internal/gateway/methods.go` | RPC method dispatch |
| `server-methods/*.ts` | `internal/gateway/methods/` | Individual method handlers |
| `server-channels.ts` | `internal/gateway/channels.go` | Channel lifecycle |
| `server-chat.ts` | `internal/gateway/chat.go` | Chat event handling |
| `server-cron.ts` | `internal/gateway/cron.go` | Cron integration |
| `server-discovery.ts` | `internal/gateway/discovery.go` | mDNS discovery |
| `server-plugins.ts` | `internal/gateway/plugins.go` | Plugin loading |
| `server-tailscale.ts` | `internal/gateway/tailscale.go` | Tailscale integration |
| `openai-http.ts` | `internal/gateway/openai_compat.go` | OpenAI API compatibility |
| `openresponses-http.ts` | `internal/gateway/openresponses.go` | Responses API compat |
| `auth.ts` | `internal/gateway/auth.go` | Authentication |
| `client.ts` | `internal/gateway/client.go` | WebSocket client |
| `protocol/` | `internal/gateway/protocol/` | Protocol definitions |
| `server/` | `internal/gateway/server/` | Server infrastructure |
| `config-reload.ts` | `internal/gateway/reload.go` | Config hot-reload |
| `hooks.ts`, `hooks-mapping.ts` | `internal/gateway/hooks.go` | Webhook handling |
| `control-ui.ts` | `internal/gateway/controlui.go` | Control UI serving |
| `session-utils.ts` | `internal/gateway/sessions.go` | Session utilities |
| `node-registry.ts` | `internal/gateway/noderegistry.go` | Mobile node registry |
| `exec-approval-manager.ts` | `internal/gateway/execapproval.go` | Exec approval management |

### `src/agents/` → `internal/agents/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `pi-embedded-runner/` | `internal/agents/runner/` | Agent execution engine |
| `pi-embedded-subscribe.ts` | `internal/agents/subscribe.go` | Event subscription |
| `pi-embedded-helpers.ts` | `internal/agents/helpers.go` | Agent helpers |
| `pi-tools.ts` | `internal/agents/tools.go` | Tool definitions |
| `tools/*.ts` | `internal/agents/tools/` | Individual tool implementations |
| `bash-tools.ts` | `internal/agents/bash.go` | Shell execution |
| `sandbox/` | `internal/agents/sandbox/` | Docker sandbox |
| `skills/` | `internal/agents/skills/` | Skills system |
| `auth-profiles/` | `internal/agents/authprofiles/` | Auth profiles |
| `models-config.ts` | `internal/agents/models.go` | Model configuration |
| `system-prompt.ts` | `internal/agents/systemprompt.go` | System prompt generation |
| `identity-file.ts` | `internal/agents/identity.go` | Agent identity |
| `session-slug.ts` | `internal/agents/sessionslug.go` | Session slug generation |

### `src/config/` → `internal/config/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `io.ts` | `internal/config/io.go` | Config file I/O |
| `schema.ts` | `internal/config/schema.go` | Config schema |
| `zod-schema.ts` | `internal/config/validate.go` | Config validation |
| `types.*.ts` | `internal/config/types.go` | Config type definitions (Go structs) |
| `validation.ts` | `internal/config/validate.go` | Validation logic |
| `legacy-migrate.ts` | `internal/config/migrate.go` | Legacy migration |
| `paths.ts` | `internal/config/paths.go` | Config paths |
| `env-vars.ts` | `internal/config/env.go` | Environment variables |
| `sessions/` | `internal/config/sessions/` | Session config |

### `src/auto-reply/` → `internal/autoreply/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `reply/agent-runner.ts` | `internal/autoreply/runner.go` | Agent execution for replies |
| `reply/commands.ts` | `internal/autoreply/commands.go` | Inline command handling |
| `reply/directive-handling.ts` | `internal/autoreply/directives.go` | Reply directives |
| `reply/queue.ts` | `internal/autoreply/queue.go` | Reply queue |
| `reply/dispatcher-registry.ts` | `internal/autoreply/dispatch.go` | Reply dispatchers |
| `commands-registry.ts` | `internal/autoreply/registry.go` | Command registry |
| `heartbeat.ts` | `internal/autoreply/heartbeat.go` | Heartbeat replies |
| `reply/history.ts` | `internal/autoreply/history.go` | Reply history |
| `reply/mentions.ts` | `internal/autoreply/mentions.go` | Mention handling |
| `reply/session.ts` | `internal/autoreply/session.go` | Session management |

### `src/channels/` → `internal/channels/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `registry.ts` | `internal/channels/registry.go` | Channel registry |
| `plugins/index.ts` | `internal/channels/plugins.go` | Plugin loading |
| `plugins/types.ts` | `internal/channels/types.go` | Channel interfaces |
| `plugins/onboarding/` | `internal/channels/onboarding/` | Onboarding per channel |
| `plugins/outbound/` | `internal/channels/outbound/` | Outbound adapters |
| `plugins/normalize/` | `internal/channels/normalize/` | Message normalization |
| `plugins/actions/` | `internal/channels/actions/` | Channel actions |
| `allowlists/` | `internal/channels/allowlists/` | Allowlist matching |

### `src/telegram/` → `internal/channels/telegram/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `bot.ts` | `internal/channels/telegram/bot.go` | Bot setup |
| `monitor.ts` | `internal/channels/telegram/monitor.go` | Message monitoring |
| `send.ts` | `internal/channels/telegram/send.go` | Message sending |
| `format.ts` | `internal/channels/telegram/format.go` | Message formatting |
| `download.ts` | `internal/channels/telegram/download.go` | Media download |
| `webhook.ts` | `internal/channels/telegram/webhook.go` | Webhook handling |
| `probe.ts` | `internal/channels/telegram/probe.go` | Connection probing |

### `src/discord/` → `internal/channels/discord/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `monitor.ts` | `internal/channels/discord/monitor.go` | Event monitoring |
| `send.ts` | `internal/channels/discord/send.go` | Message sending |
| `api.ts` | `internal/channels/discord/api.go` | Discord API client |
| `resolve-channels.ts` | `internal/channels/discord/resolve.go` | Channel resolution |

### `src/slack/` → `internal/channels/slack/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `monitor.ts` | `internal/channels/slack/monitor.go` | Event monitoring |
| `send.ts` | `internal/channels/slack/send.go` | Message sending |
| `client.ts` | `internal/channels/slack/client.go` | Slack API client |
| `http/` | `internal/channels/slack/http/` | HTTP mode |

### `src/signal/` → `internal/channels/signal/`
### `src/web/` (WhatsApp) → `internal/channels/whatsapp/`
### `src/imessage/` → `internal/channels/imessage/`

### `src/browser/` → `internal/browser/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `server.ts` | `internal/browser/server.go` | Browser control server |
| `pw-session.ts` | `internal/browser/session.go` | Browser session |
| `pw-tools-core.ts` | `internal/browser/tools.go` | Browser tools |
| `cdp.ts` | `internal/browser/cdp.go` | CDP client |
| `routes/` | `internal/browser/routes/` | HTTP routes |

### `src/infra/` → `internal/infra/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `env.ts` | `internal/infra/env.go` | Environment handling |
| `dotenv.ts` | `internal/infra/dotenv.go` | .env loading |
| `errors.ts` | `internal/infra/errors.go` | Error utilities |
| `fetch.ts` | `internal/infra/fetch.go` | HTTP client |
| `retry.ts` | `internal/infra/retry.go` | Retry logic |
| `backoff.ts` | `internal/infra/backoff.go` | Exponential backoff |
| `heartbeat-runner.ts` | `internal/infra/heartbeat.go` | Heartbeat scheduling |
| `bonjour.ts` | `internal/infra/bonjour.go` | mDNS discovery |
| `device-identity.ts` | `internal/infra/deviceid.go` | Device identity |
| `device-pairing.ts` | `internal/infra/pairing.go` | Device pairing |
| `exec-approvals.ts` | `internal/infra/execapproval.go` | Exec approvals |
| `ssh-tunnel.ts` | `internal/infra/ssh.go` | SSH tunneling |
| `tailscale.ts` | `internal/infra/tailscale.go` | Tailscale |
| `update-check.ts` | `internal/infra/update.go` | Update checking |
| `provider-usage.ts` | `internal/infra/usage.go` | Usage tracking |
| `system-events.ts` | `internal/infra/events.go` | Event bus |
| `net/ssrf.ts` | `internal/infra/ssrf.go` | SSRF protection |
| `tls/fingerprint.ts` | `internal/infra/tlspin.go` | TLS pinning |

### `src/plugins/` → `internal/plugins/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `loader.ts` | `internal/plugins/loader.go` | Plugin loading |
| `registry.ts` | `internal/plugins/registry.go` | Plugin registry |
| `runtime.ts` | `internal/plugins/runtime.go` | Plugin runtime |
| `hooks.ts` | `internal/plugins/hooks.go` | Hook system |
| `services.ts` | `internal/plugins/services.go` | Plugin services |
| `tools.ts` | `internal/plugins/tools.go` | Tool registration |

### `src/cron/` → `internal/cron/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `service.ts` | `internal/cron/service.go` | Cron service |
| `store.ts` | `internal/cron/store.go` | Job storage |
| `schedule.ts` | `internal/cron/schedule.go` | Schedule parsing |
| `isolated-agent/` | `internal/cron/agent.go` | Isolated agent execution |

### `src/media/` → `internal/media/`

| Original File | Go Package | Description |
|--------------|------------|-------------|
| `server.ts` | `internal/media/server.go` | Media server |
| `store.ts` | `internal/media/store.go` | Media storage |
| `fetch.ts` | `internal/media/fetch.go` | Media fetching |
| `image-ops.ts` | `internal/media/image.go` | Image processing |

### `src/logging/` → `internal/logging/`
### `src/security/` → `internal/security/`
### `src/process/` → `internal/process/`
### `src/daemon/` → `internal/daemon/`
### `src/tts/` → `internal/tts/`
### `src/acp/` → `internal/acp/`
### `src/tui/` → `internal/tui/`
### `src/wizard/` → `internal/wizard/`
### `src/routing/` → `internal/routing/`
### `src/terminal/` → `internal/terminal/`
### `src/markdown/` → `internal/markdown/`
### `src/plugin-sdk/` → `pkg/pluginsdk/`

### `extensions/` → `internal/extensions/` (built-in) or `pkg/extensions/`

| Extension | Go Package | Description |
|-----------|-----------|-------------|
| `msteams/` | `internal/extensions/msteams/` | MS Teams integration |
| `whatsapp/` | `internal/extensions/whatsapp/` | WhatsApp extension |
| `open-prose/` | `internal/extensions/openprose/` | Open Prose VM |
| `llm-task/` | `internal/extensions/llmtask/` | LLM task tool |
| `memory-core/` | `internal/extensions/memory/` | Memory/RAG |
| `copilot-proxy/` | `internal/extensions/copilotproxy/` | Copilot proxy |
| `thread-ownership/` | `internal/extensions/threadown/` | Thread ownership |

---

## Go Project Structure

```
goclaw/
├── cmd/
│   └── openclaw/
│       └── main.go                    # CLI entry point
│
├── internal/
│   ├── cli/                           # CLI command definitions (cobra)
│   │   ├── root.go                    # Root command
│   │   ├── cmd_agent.go              # Agent commands
│   │   ├── cmd_gateway.go            # Gateway commands
│   │   ├── cmd_channels.go           # Channel commands
│   │   ├── cmd_config.go             # Config commands
│   │   ├── cmd_cron.go               # Cron commands
│   │   ├── cmd_daemon.go             # Daemon commands
│   │   ├── cmd_models.go             # Model commands
│   │   ├── cmd_browser.go            # Browser commands
│   │   ├── cmd_plugins.go            # Plugin commands
│   │   ├── cmd_skills.go             # Skills commands
│   │   ├── cmd_nodes.go              # Node commands
│   │   ├── cmd_tui.go                # TUI command
│   │   ├── cmd_update.go             # Update commands
│   │   ├── cmd_doctor.go             # Doctor command
│   │   ├── cmd_onboard.go            # Onboard command
│   │   ├── cmd_status.go             # Status command
│   │   ├── cmd_message.go            # Message commands
│   │   ├── cmd_sessions.go           # Sessions commands
│   │   ├── deps.go                   # Dependency injection
│   │   ├── progress.go               # Progress indicators
│   │   └── prompt.go                 # Interactive prompts
│   │
│   ├── gateway/                       # Gateway server
│   │   ├── server.go                 # Server lifecycle
│   │   ├── http.go                   # HTTP routing
│   │   ├── ws.go                     # WebSocket handling
│   │   ├── auth.go                   # Authentication
│   │   ├── client.go                 # WS client
│   │   ├── channels.go              # Channel management
│   │   ├── chat.go                   # Chat handling
│   │   ├── cron.go                   # Cron integration
│   │   ├── discovery.go             # mDNS discovery
│   │   ├── hooks.go                  # Webhook handling
│   │   ├── controlui.go             # Control UI
│   │   ├── openai_compat.go         # OpenAI API compat
│   │   ├── openresponses.go         # Responses API
│   │   ├── reload.go                # Config reload
│   │   ├── tailscale.go             # Tailscale
│   │   ├── noderegistry.go          # Node registry
│   │   ├── execapproval.go          # Exec approvals
│   │   ├── sessions.go              # Session utils
│   │   ├── protocol/                # Protocol definitions
│   │   │   ├── frames.go
│   │   │   ├── schema.go
│   │   │   ├── types.go
│   │   │   └── validate.go
│   │   ├── methods/                  # RPC method handlers
│   │   │   ├── agents.go
│   │   │   ├── chat.go
│   │   │   ├── config.go
│   │   │   ├── sessions.go
│   │   │   ├── models.go
│   │   │   ├── cron.go
│   │   │   ├── health.go
│   │   │   ├── nodes.go
│   │   │   ├── logs.go
│   │   │   ├── devices.go
│   │   │   └── execapproval.go
│   │   └── server/                   # Server infrastructure
│   │       ├── wsconn.go
│   │       ├── health.go
│   │       ├── tls.go
│   │       └── listen.go
│   │
│   ├── agents/                        # AI agent runtime
│   │   ├── runner/                   # Agent execution engine
│   │   │   ├── run.go
│   │   │   ├── model.go
│   │   │   ├── compact.go
│   │   │   ├── history.go
│   │   │   └── params.go
│   │   ├── tools/                    # Tool implementations
│   │   │   ├── bash.go
│   │   │   ├── browser.go
│   │   │   ├── webfetch.go
│   │   │   ├── websearch.go
│   │   │   ├── image.go
│   │   │   ├── memory.go
│   │   │   ├── message.go
│   │   │   ├── sessions.go
│   │   │   ├── canvas.go
│   │   │   ├── cron.go
│   │   │   ├── gateway.go
│   │   │   └── tts.go
│   │   ├── sandbox/                  # Docker sandbox
│   │   │   ├── docker.go
│   │   │   ├── config.go
│   │   │   ├── fsbridge.go
│   │   │   └── workspace.go
│   │   ├── skills/                   # Skills system
│   │   ├── authprofiles/             # Auth profiles
│   │   ├── subscribe.go
│   │   ├── helpers.go
│   │   ├── models.go
│   │   ├── systemprompt.go
│   │   ├── identity.go
│   │   └── bash.go
│   │
│   ├── config/                        # Configuration
│   │   ├── config.go                 # Main config types
│   │   ├── io.go                     # File I/O
│   │   ├── validate.go              # Validation
│   │   ├── schema.go                # Schema generation
│   │   ├── paths.go                 # Config paths
│   │   ├── env.go                   # Env var handling
│   │   ├── migrate.go               # Legacy migration
│   │   ├── defaults.go              # Default values
│   │   └── sessions/                # Session config
│   │
│   ├── autoreply/                     # Auto-reply engine
│   │   ├── runner.go
│   │   ├── commands.go
│   │   ├── directives.go
│   │   ├── queue.go
│   │   ├── dispatch.go
│   │   ├── registry.go
│   │   ├── heartbeat.go
│   │   ├── history.go
│   │   ├── mentions.go
│   │   └── session.go
│   │
│   ├── channels/                      # Channel system
│   │   ├── registry.go
│   │   ├── types.go
│   │   ├── plugins.go
│   │   ├── telegram/                 # Telegram
│   │   ├── discord/                  # Discord
│   │   ├── slack/                    # Slack
│   │   ├── signal/                   # Signal
│   │   ├── whatsapp/                 # WhatsApp
│   │   ├── imessage/                 # iMessage
│   │   ├── onboarding/
│   │   ├── outbound/
│   │   ├── normalize/
│   │   └── actions/
│   │
│   ├── browser/                       # Browser automation
│   │   ├── server.go
│   │   ├── session.go
│   │   ├── tools.go
│   │   ├── cdp.go
│   │   └── routes/
│   │
│   ├── infra/                         # Infrastructure
│   │   ├── env.go
│   │   ├── dotenv.go
│   │   ├── errors.go
│   │   ├── fetch.go
│   │   ├── retry.go
│   │   ├── backoff.go
│   │   ├── heartbeat.go
│   │   ├── bonjour.go
│   │   ├── deviceid.go
│   │   ├── pairing.go
│   │   ├── execapproval.go
│   │   ├── ssh.go
│   │   ├── tailscale.go
│   │   ├── update.go
│   │   ├── usage.go
│   │   ├── events.go
│   │   ├── ssrf.go
│   │   └── tlspin.go
│   │
│   ├── plugins/                       # Plugin system
│   │   ├── loader.go
│   │   ├── registry.go
│   │   ├── runtime.go
│   │   ├── hooks.go
│   │   ├── services.go
│   │   └── tools.go
│   │
│   ├── cron/                          # Cron scheduling
│   │   ├── service.go
│   │   ├── store.go
│   │   ├── schedule.go
│   │   └── agent.go
│   │
│   ├── media/                         # Media processing
│   │   ├── server.go
│   │   ├── store.go
│   │   ├── fetch.go
│   │   └── image.go
│   │
│   ├── logging/                       # Logging
│   │   ├── logger.go
│   │   ├── subsystem.go
│   │   ├── redact.go
│   │   └── console.go
│   │
│   ├── security/                      # Security
│   │   ├── audit.go
│   │   ├── secret.go
│   │   └── scanner.go
│   │
│   ├── process/                       # Process management
│   │   ├── exec.go
│   │   ├── queue.go
│   │   └── spawn.go
│   │
│   ├── daemon/                        # System daemon
│   │   ├── launchd.go
│   │   ├── systemd.go
│   │   └── schtasks.go
│   │
│   ├── tts/                           # Text-to-speech
│   │   └── tts.go
│   │
│   ├── acp/                           # Agent Client Protocol
│   │   ├── client.go
│   │   ├── server.go
│   │   └── session.go
│   │
│   ├── tui/                           # Terminal UI
│   │   ├── tui.go
│   │   └── components/
│   │
│   ├── wizard/                        # Onboarding wizard
│   │   └── onboarding.go
│   │
│   ├── routing/                       # Message routing
│   │   └── route.go
│   │
│   ├── terminal/                      # Terminal formatting
│   │   ├── table.go
│   │   ├── ansi.go
│   │   └── palette.go
│   │
│   ├── markdown/                      # Markdown processing
│   │   └── markdown.go
│   │
│   └── extensions/                    # Built-in extensions
│       ├── msteams/
│       ├── openprose/
│       ├── llmtask/
│       ├── memory/
│       └── copilotproxy/
│
├── pkg/                               # Public packages
│   ├── pluginsdk/                    # Plugin SDK
│   │   └── sdk.go
│   └── protocol/                     # Gateway protocol types
│       └── protocol.go
│
├── api/                               # API definitions
│   └── openapi.yaml                  # OpenAPI spec
│
├── configs/                           # Default configs
│   └── default.json
│
├── skills/                            # Skill definitions (copied from original)
│
├── web/                               # Control UI (embedded)
│   └── dist/                         # Built web assets
│
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

---

## Dependency Mapping

### Core Framework

| TypeScript Dependency | Go Equivalent | Notes |
|----------------------|---------------|-------|
| `commander` (CLI) | `github.com/spf13/cobra` | Industry-standard Go CLI framework |
| `express` (HTTP) | `net/http` + `github.com/go-chi/chi/v5` | Chi for routing, stdlib for server |
| `ws` (WebSocket) | `github.com/gorilla/websocket` or `nhooyr.io/websocket` | gorilla/websocket is battle-tested |
| `@sinclair/typebox` + `ajv` (schema) | Go structs + `github.com/go-playground/validator` | Native struct validation |
| `zod` (validation) | `github.com/go-playground/validator/v10` | Struct tag validation |
| `tslog` (logging) | `log/slog` (stdlib) or `go.uber.org/zap` | slog is Go 1.21+ stdlib |
| `dotenv` | `github.com/joho/godotenv` | .env file loading |
| `json5` | `github.com/yosuke-furukawa/json5` or custom parser | JSON5 config support |
| `yaml` | `gopkg.in/yaml.v3` | YAML parsing |
| `chalk` (colors) | `github.com/fatih/color` | Terminal colors |
| `chokidar` (file watch) | `github.com/fsnotify/fsnotify` | File system notifications |
| `croner` (cron) | `github.com/robfig/cron/v3` | Cron scheduling |
| `undici` / `node:fetch` | `net/http` (stdlib) | Go's HTTP client is excellent |
| `sharp` (image) | `github.com/disintegration/imaging` or `github.com/h2non/bimg` | Image processing |
| `pdfjs-dist` (PDF) | `github.com/pdfcpu/pdfcpu` | PDF processing |
| `jszip` | `archive/zip` (stdlib) | ZIP handling |
| `tar` | `archive/tar` (stdlib) | TAR handling |
| `proper-lockfile` | `github.com/gofrs/flock` | File locking |
| `markdown-it` | `github.com/yuin/goldmark` | Markdown parsing |
| `@clack/prompts` | `github.com/charmbracelet/huh` | Interactive prompts |
| `osc-progress` | `github.com/charmbracelet/bubbles` | Progress bars |
| `qrcode-terminal` | `github.com/mdp/qrterminal` | QR code in terminal |
| `signal-utils` | Go channels + sync primitives | Native Go concurrency |

### Messaging Channel SDKs

| TypeScript Dependency | Go Equivalent | Notes |
|----------------------|---------------|-------|
| `grammy` (Telegram) | `github.com/go-telegram-bot-api/telegram-bot-api/v5` or `github.com/gotd/td` | Telegram bot API |
| `@buape/carbon` (Discord) | `github.com/bwmarrin/discordgo` | Discord API |
| `@slack/bolt` | `github.com/slack-go/slack` | Slack API |
| `@whiskeysockets/baileys` (WhatsApp) | `github.com/nickstenning/go-whatsapp` or custom | WhatsApp Web protocol |
| `@line/bot-sdk` | `github.com/line/line-bot-sdk-go` | LINE API |
| `@larksuiteoapi/node-sdk` | `github.com/larksuite/oapi-sdk-go` | Lark/Feishu API |

### AI/LLM

| TypeScript Dependency | Go Equivalent | Notes |
|----------------------|---------------|-------|
| `@mariozechner/pi-*` (Pi agent) | Custom Go implementation | Core agent runtime rewrite |
| `@aws-sdk/client-bedrock` | `github.com/aws/aws-sdk-go-v2/service/bedrockruntime` | AWS Bedrock |
| `ollama` | `github.com/ollama/ollama/api` | Ollama client |
| `node-llama-cpp` | CGo bindings or HTTP API | Local LLM inference |

### Browser Automation

| TypeScript Dependency | Go Equivalent | Notes |
|----------------------|---------------|-------|
| `playwright-core` | `github.com/playwright-community/playwright-go` | Playwright for Go |

### Database

| TypeScript Dependency | Go Equivalent | Notes |
|----------------------|---------------|-------|
| `better-sqlite3` / `sqlite-vec` | `github.com/mattn/go-sqlite3` + `github.com/asg017/sqlite-vec` | SQLite with vector extensions |

### Network/Discovery

| TypeScript Dependency | Go Equivalent | Notes |
|----------------------|---------------|-------|
| `@homebridge/ciao` (mDNS) | `github.com/hashicorp/mdns` | mDNS/Bonjour |
| `https-proxy-agent` | `net/http` with proxy transport | Built-in proxy support |

### TTS

| TypeScript Dependency | Go Equivalent | Notes |
|----------------------|---------------|-------|
| `node-edge-tts` | Custom HTTP client to Edge TTS API | Direct API calls |

### TUI

| TypeScript Dependency | Go Equivalent | Notes |
|----------------------|---------------|-------|
| `@mariozechner/pi-tui` | `github.com/charmbracelet/bubbletea` | Elm-architecture TUI framework |
| `cli-highlight` | `github.com/alecthomas/chroma` | Syntax highlighting |

---

## Data Models in Go

### Core Configuration

```go
// internal/config/config.go

type OpenClawConfig struct {
    Gateway    GatewayConfig            `json:"gateway,omitempty"`
    Agents     map[string]AgentConfig   `json:"agents,omitempty"`
    Models     ModelsConfig             `json:"models,omitempty"`
    Channels   ChannelsConfig           `json:"channels,omitempty"`
    Plugins    map[string]PluginConfig  `json:"plugins,omitempty"`
    Hooks      []HookConfig             `json:"hooks,omitempty"`
    Cron       CronConfig               `json:"cron,omitempty"`
    Browser    BrowserConfig            `json:"browser,omitempty"`
    Sandbox    SandboxConfig            `json:"sandbox,omitempty"`
    Tools      ToolsConfig              `json:"tools,omitempty"`
    Skills     SkillsConfig             `json:"skills,omitempty"`
    TTS        TTSConfig                `json:"tts,omitempty"`
    Memory     MemoryConfig             `json:"memory,omitempty"`
    Env        map[string]string        `json:"env,omitempty"`
}

type GatewayConfig struct {
    Mode           string          `json:"mode,omitempty"`       // "local" | "remote"
    Port           int             `json:"port,omitempty"`
    Bind           string          `json:"bind,omitempty"`
    Auth           GatewayAuth     `json:"auth,omitempty"`
    TLS            *TLSConfig      `json:"tls,omitempty"`
    Discovery      *DiscoveryConfig `json:"discovery,omitempty"`
    Tailscale      *TailscaleConfig `json:"tailscale,omitempty"`
    RemoteURL      string          `json:"remoteUrl,omitempty"`
    RemoteToken    string          `json:"remoteToken,omitempty"`
}

type GatewayAuth struct {
    Token    string `json:"token,omitempty"`
    Password string `json:"password,omitempty"`
}

type AgentConfig struct {
    Identity       IdentityConfig    `json:"identity,omitempty"`
    Model          string            `json:"model,omitempty"`
    ThinkingModel  string            `json:"thinkingModel,omitempty"`
    SystemPrompt   string            `json:"systemPrompt,omitempty"`
    Tools          ToolsConfig       `json:"tools,omitempty"`
    Skills         []string          `json:"skills,omitempty"`
    Session        SessionConfig     `json:"session,omitempty"`
    Sandbox        *SandboxConfig    `json:"sandbox,omitempty"`
    Concurrency    *ConcurrencyConfig `json:"concurrency,omitempty"`
}

type IdentityConfig struct {
    Name   string `json:"name,omitempty"`
    Avatar string `json:"avatar,omitempty"`
}

type SessionConfig struct {
    Scope          string              `json:"scope,omitempty"`
    DmScope        string              `json:"dmScope,omitempty"`
    IdleMinutes    int                 `json:"idleMinutes,omitempty"`
    Reset          *SessionResetConfig `json:"reset,omitempty"`
    ResetByType    *SessionResetByType `json:"resetByType,omitempty"`
    TypingMode     string              `json:"typingMode,omitempty"`
    MainKey        string              `json:"mainKey,omitempty"`
    SendPolicy     *SendPolicyConfig   `json:"sendPolicy,omitempty"`
}

type SessionResetConfig struct {
    Mode        string `json:"mode,omitempty"`
    AtHour      int    `json:"atHour,omitempty"`
    IdleMinutes int    `json:"idleMinutes,omitempty"`
}
```

### Gateway Protocol

```go
// internal/gateway/protocol/types.go

type ConnectParams struct {
    MinProtocol int           `json:"minProtocol"`
    MaxProtocol int           `json:"maxProtocol"`
    Client      ClientInfo    `json:"client"`
    Caps        []string      `json:"caps,omitempty"`
    Commands    []string      `json:"commands,omitempty"`
    Auth        *ConnectAuth  `json:"auth,omitempty"`
    Device      *DeviceAuth   `json:"device,omitempty"`
}

type ClientInfo struct {
    ID              string `json:"id"`
    DisplayName     string `json:"displayName,omitempty"`
    Version         string `json:"version"`
    Platform        string `json:"platform"`
    DeviceFamily    string `json:"deviceFamily,omitempty"`
    Mode            string `json:"mode"`
    InstanceID      string `json:"instanceId,omitempty"`
}

type HelloOk struct {
    Type     string       `json:"type"`
    Protocol int          `json:"protocol"`
    Server   ServerInfo   `json:"server"`
    Features Features     `json:"features"`
    Snapshot Snapshot     `json:"snapshot"`
}

type GatewayFrame struct {
    Type    string          `json:"type"`
    ID      string          `json:"id,omitempty"`
    Method  string          `json:"method,omitempty"`
    Params  json.RawMessage `json:"params,omitempty"`
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *ErrorShape     `json:"error,omitempty"`
    Event   string          `json:"event,omitempty"`
    Data    json.RawMessage `json:"data,omitempty"`
}

type RequestFrame struct {
    Type   string          `json:"type"` // "request"
    ID     string          `json:"id"`
    Method string          `json:"method"`
    Params json.RawMessage `json:"params"`
}

type ResponseFrame struct {
    Type   string          `json:"type"` // "response"
    ID     string          `json:"id"`
    Result json.RawMessage `json:"result,omitempty"`
    Error  *ErrorShape     `json:"error,omitempty"`
}

type EventFrame struct {
    Type  string          `json:"type"` // "event"
    Event string          `json:"event"`
    Data  json.RawMessage `json:"data"`
}

type ErrorShape struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

type ChatEvent struct {
    SessionKey string `json:"sessionKey"`
    AgentID    string `json:"agentId,omitempty"`
    Type       string `json:"type"` // "text", "tool-call", "tool-result", "error", "done"
    Text       string `json:"text,omitempty"`
    ToolName   string `json:"toolName,omitempty"`
    ToolCallID string `json:"toolCallId,omitempty"`
}
```

### Channel Types

```go
// internal/channels/types.go

type ChannelID string

type ChannelPlugin interface {
    ID() ChannelID
    DisplayName() string
    Start(ctx context.Context, config ChannelConfig) error
    Stop(ctx context.Context) error
    Send(ctx context.Context, msg OutboundMessage) error
    Probe(ctx context.Context) (*ProbeResult, error)
}

type InboundMessage struct {
    ChannelID   ChannelID
    SenderID    string
    SenderName  string
    Text        string
    ChatType    ChatType // "dm" | "group" | "thread"
    ChatID      string
    ThreadID    string
    MediaURLs   []string
    ReplyToID   string
    Timestamp   time.Time
}

type OutboundMessage struct {
    ChannelID   ChannelID
    ChatID      string
    ThreadID    string
    Text        string
    MediaURLs   []string
    ReplyToID   string
}

type ChatType string

const (
    ChatTypeDM     ChatType = "dm"
    ChatTypeGroup  ChatType = "group"
    ChatTypeThread ChatType = "thread"
)
```

### Cron Types

```go
// internal/cron/types.go

type CronJob struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Schedule    string    `json:"schedule"`
    Message     string    `json:"message"`
    AgentID     string    `json:"agentId,omitempty"`
    Deliver     *Delivery `json:"deliver,omitempty"`
    Enabled     bool      `json:"enabled"`
    OneShot     bool      `json:"oneShot,omitempty"`
    CreatedAt   time.Time `json:"createdAt"`
    LastRunAt   time.Time `json:"lastRunAt,omitempty"`
}

type Delivery struct {
    Channel string `json:"channel"`
    To      string `json:"to"`
}

type CronRunLog struct {
    JobID     string    `json:"jobId"`
    StartedAt time.Time `json:"startedAt"`
    Duration  int64     `json:"durationMs"`
    Status    string    `json:"status"`
    Output    string    `json:"output,omitempty"`
    Error     string    `json:"error,omitempty"`
}
```

### Agent Types

```go
// internal/agents/types.go

type AgentRun struct {
    SessionKey    string
    AgentID       string
    Model         string
    SystemPrompt  string
    Tools         []ToolDefinition
    History       []Message
    AbortCh       <-chan struct{}
}

type Message struct {
    Role    string          `json:"role"` // "user", "assistant", "system", "tool"
    Content json.RawMessage `json:"content"`
}

type ToolDefinition struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  json.RawMessage `json:"parameters"`
}

type ToolCall struct {
    ID        string          `json:"id"`
    Name      string          `json:"name"`
    Arguments json.RawMessage `json:"arguments"`
}

type ToolResult struct {
    ToolCallID string `json:"toolCallId"`
    Content    string `json:"content"`
    IsError    bool   `json:"isError,omitempty"`
}
```

---

## API Layer Rewrite Plan

### WebSocket Gateway

The WebSocket gateway will use `gorilla/websocket` with a custom JSON-RPC-like protocol handler:

```go
// internal/gateway/ws.go

type WSHandler struct {
    upgrader websocket.Upgrader
    auth     *AuthManager
    methods  *MethodRegistry
    clients  *ClientRegistry
}

func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
    conn, err := h.upgrader.Upgrade(w, r, nil)
    if err != nil { return }
    
    client := NewClient(conn)
    defer client.Close()
    
    // Authentication handshake
    if err := h.authenticate(client); err != nil { return }
    
    // Send hello-ok with snapshot
    h.sendHello(client)
    
    // Message loop with goroutines for read/write
    go client.WritePump()
    client.ReadPump(h.methods)
}
```

### HTTP API Endpoints

Using Chi router:

```go
// internal/gateway/http.go

func NewHTTPRouter(gw *Gateway) http.Handler {
    r := chi.NewRouter()
    
    // OpenAI-compatible API
    r.Route("/v1", func(r chi.Router) {
        r.Post("/chat/completions", gw.HandleChatCompletions)
        r.Post("/responses", gw.HandleResponses)
    })
    
    // Webhook hooks
    r.Route("/api", func(r chi.Router) {
        r.Post("/hooks/wake", gw.HandleWakeHook)
        r.Post("/hooks/message", gw.HandleMessageHook)
        r.Route("/plugins", func(r chi.Router) {
            r.HandleFunc("/*", gw.HandlePluginHTTP)
        })
        r.Post("/tools/invoke", gw.HandleToolInvoke)
    })
    
    // Slack HTTP mode
    r.Route("/api/slack", func(r chi.Router) {
        r.Post("/events", gw.HandleSlackEvents)
        r.Post("/interactions", gw.HandleSlackInteractions)
    })
    
    // Control UI
    r.Route("/control", func(r chi.Router) {
        r.Handle("/*", gw.ControlUIHandler())
    })
    
    // Media serving
    r.Route("/media", func(r chi.Router) {
        r.Get("/{id}", gw.HandleMediaServe)
    })
    
    // Canvas/A2UI
    r.Route("/canvas", func(r chi.Router) {
        r.Handle("/*", gw.CanvasHandler())
    })
    
    return r
}
```

### Method Handler Pattern

Each gateway method maps to a handler function:

```go
// internal/gateway/methods/chat.go

func (m *Methods) ChatSend(ctx context.Context, client *Client, params ChatSendParams) (*ChatSendResult, error) {
    // Validate params
    // Resolve session key
    // Start agent run in goroutine
    // Stream events back to client
    return &ChatSendResult{SessionKey: key}, nil
}
```

---

## Concurrency & Performance

### Goroutines Replace async/await

| TypeScript Pattern | Go Pattern |
|-------------------|------------|
| `async function` / `await` | Goroutines + channels |
| `Promise.all()` | `errgroup.Group` or `sync.WaitGroup` |
| `EventEmitter` | Channels or `sync.Cond` |
| `AbortController` / `AbortSignal` | `context.Context` with cancellation |
| `setTimeout` / `setInterval` | `time.After` / `time.Ticker` |
| `process.nextTick` | `runtime.Gosched()` (rarely needed) |
| Callback-based APIs | Channel-based or sync APIs |
| `AsyncLocalStorage` | `context.Context` value propagation |

### Key Concurrency Patterns

```go
// Agent execution with cancellation
func (r *Runner) RunAgent(ctx context.Context, params AgentRun) error {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    g, ctx := errgroup.WithContext(ctx)
    
    // Stream events to client
    eventCh := make(chan ChatEvent, 100)
    g.Go(func() error {
        return r.streamEvents(ctx, eventCh)
    })
    
    // Run agent loop
    g.Go(func() error {
        defer close(eventCh)
        return r.executeLoop(ctx, params, eventCh)
    })
    
    return g.Wait()
}

// Reply queue with worker pool
type ReplyQueue struct {
    jobs    chan ReplyJob
    workers int
}

func (q *ReplyQueue) Start(ctx context.Context) {
    for i := 0; i < q.workers; i++ {
        go q.worker(ctx)
    }
}

func (q *ReplyQueue) worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case job := <-q.jobs:
            q.processReply(ctx, job)
        }
    }
}

// WebSocket client with read/write pumps
type Client struct {
    conn    *websocket.Conn
    send    chan []byte
    done    chan struct{}
}

func (c *Client) WritePump() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case msg, ok := <-c.send:
            if !ok { return }
            c.conn.WriteMessage(websocket.TextMessage, msg)
        case <-ticker.C:
            c.conn.WriteMessage(websocket.PingMessage, nil)
        case <-c.done:
            return
        }
    }
}
```

### Performance Advantages

1. **Single binary** — No Node.js runtime overhead, instant startup
2. **Goroutine pool** — Thousands of concurrent connections with minimal memory
3. **Zero-copy I/O** — Efficient buffer management
4. **Compiled code** — No JIT warmup, consistent performance
5. **Memory efficiency** — Go's garbage collector is optimized for low-latency
6. **Static linking** — No dependency resolution at runtime

---

## Configuration Management

### Using Viper + JSON5

```go
// internal/config/config.go

import (
    "github.com/spf13/viper"
    "github.com/joho/godotenv"
)

type ConfigManager struct {
    v       *viper.Viper
    config  *OpenClawConfig
    path    string
    mu      sync.RWMutex
    watchers []func(*OpenClawConfig)
}

func NewConfigManager() *ConfigManager {
    v := viper.New()
    v.SetConfigName("openclaw")
    v.SetConfigType("json")
    v.AddConfigPath("$HOME/.openclaw")
    v.AddConfigPath(".")
    
    // Environment variable binding
    v.SetEnvPrefix("OPENCLAW")
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    return &ConfigManager{v: v}
}

func (cm *ConfigManager) Load() (*OpenClawConfig, error) {
    // Load .env files
    godotenv.Load(".env")
    godotenv.Load(filepath.Join(cm.StateDir(), ".env"))
    
    // Read config file (JSON5 support via custom decoder)
    if err := cm.v.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }
    
    var config OpenClawConfig
    if err := cm.v.Unmarshal(&config); err != nil {
        return nil, err
    }
    
    // Validate
    if err := config.Validate(); err != nil {
        return nil, err
    }
    
    cm.mu.Lock()
    cm.config = &config
    cm.mu.Unlock()
    
    return &config, nil
}

// Hot-reload support
func (cm *ConfigManager) Watch(ctx context.Context) {
    cm.v.WatchConfig()
    cm.v.OnConfigChange(func(e fsnotify.Event) {
        config, err := cm.reload()
        if err != nil { return }
        for _, fn := range cm.watchers {
            fn(config)
        }
    })
}
```

### Config Paths

```go
func StateDir() string {
    if dir := os.Getenv("OPENCLAW_STATE_DIR"); dir != "" {
        return dir
    }
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".openclaw")
}

func ConfigPath() string {
    if p := os.Getenv("OPENCLAW_CONFIG_PATH"); p != "" {
        return p
    }
    return filepath.Join(StateDir(), "openclaw.json")
}
```

---

## Error Handling Strategy

### Go Idiomatic Error Handling

```go
// internal/infra/errors.go

// Sentinel errors for common cases
var (
    ErrNotFound       = errors.New("not found")
    ErrUnauthorized   = errors.New("unauthorized")
    ErrRateLimited    = errors.New("rate limited")
    ErrConfigInvalid  = errors.New("invalid configuration")
    ErrSessionExpired = errors.New("session expired")
)

// Typed errors with context
type GatewayError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Cause   error  `json:"-"`
}

func (e *GatewayError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("gateway error %d: %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("gateway error %d: %s", e.Code, e.Message)
}

func (e *GatewayError) Unwrap() error { return e.Cause }

// Error wrapping pattern
func (s *Server) handleRequest(ctx context.Context, req *Request) (*Response, error) {
    result, err := s.methods.Dispatch(ctx, req.Method, req.Params)
    if err != nil {
        var gwErr *GatewayError
        if errors.As(err, &gwErr) {
            return &Response{Error: gwErr}, nil
        }
        return nil, fmt.Errorf("dispatch %s: %w", req.Method, err)
    }
    return &Response{Result: result}, nil
}

// Retry with backoff
func RetryWithBackoff(ctx context.Context, opts RetryOpts, fn func() error) error {
    var lastErr error
    for attempt := 0; attempt < opts.MaxAttempts; attempt++ {
        if err := fn(); err != nil {
            lastErr = err
            if !isTransient(err) {
                return err
            }
            delay := opts.BaseDelay * time.Duration(1<<attempt)
            if delay > opts.MaxDelay {
                delay = opts.MaxDelay
            }
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(delay):
            }
            continue
        }
        return nil
    }
    return fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

---

## Testing Strategy

### Go Testing Approach

```go
// Unit tests: *_test.go colocated with source
// Integration tests: *_integration_test.go with build tag
// E2E tests: test/e2e/ directory

// Example unit test
// internal/config/config_test.go
func TestLoadConfig(t *testing.T) {
    t.Run("loads valid config", func(t *testing.T) {
        dir := t.TempDir()
        configPath := filepath.Join(dir, "openclaw.json")
        os.WriteFile(configPath, []byte(`{"gateway":{"port":8080}}`), 0644)
        
        cm := NewConfigManager()
        cm.SetConfigPath(configPath)
        
        config, err := cm.Load()
        require.NoError(t, err)
        assert.Equal(t, 8080, config.Gateway.Port)
    })
    
    t.Run("returns error for invalid config", func(t *testing.T) {
        // ...
    })
}

// Example integration test
// internal/gateway/server_integration_test.go
//go:build integration

func TestGatewayServer(t *testing.T) {
    srv := startTestServer(t)
    defer srv.Close()
    
    // Connect WebSocket client
    conn, _, err := websocket.DefaultDialer.Dial(srv.WSURL(), nil)
    require.NoError(t, err)
    defer conn.Close()
    
    // Send connect
    // Verify hello-ok
    // Send chat request
    // Verify response events
}
```

### Testing Tools

| Purpose | Go Package |
|---------|-----------|
| Test framework | `testing` (stdlib) |
| Assertions | `github.com/stretchr/testify` |
| Mocking | `github.com/stretchr/testify/mock` or `go.uber.org/mock` |
| HTTP testing | `net/http/httptest` (stdlib) |
| Table-driven tests | Native Go pattern |
| Test fixtures | `testdata/` directories |
| Coverage | `go test -cover` |
| Benchmarks | `testing.B` |
| Fuzzing | `testing.F` (Go 1.18+) |

### Coverage Targets

Maintain the same 70% threshold as the TypeScript codebase:
- Lines: 70%
- Functions: 70%
- Branches: 55%

---

## Migration Phases

### Phase 1: Foundation (Weeks 1-3)

**Goal**: Core infrastructure and configuration

- [ ] Go module setup with dependency management
- [ ] Configuration system (`internal/config/`)
  - Config loading (JSON5), validation, paths
  - Environment variable handling
  - Legacy config migration
- [ ] Logging subsystem (`internal/logging/`)
  - Structured logging with slog
  - Subsystem loggers
  - Secret redaction
- [ ] Infrastructure utilities (`internal/infra/`)
  - Error types, retry, backoff
  - HTTP client with SSRF protection
  - File locking, dotenv loading
- [ ] CLI skeleton (`cmd/openclaw/`, `internal/cli/`)
  - Cobra root command
  - Config get/set commands
  - Version/help

**Milestone**: `openclaw config get gateway.port` works

### Phase 2: Gateway Core (Weeks 4-7)

**Goal**: WebSocket gateway server with protocol handling

- [ ] HTTP server with Chi router (`internal/gateway/`)
- [ ] WebSocket server with gorilla/websocket
- [ ] Gateway protocol implementation
  - Frame parsing/serialization
  - Authentication (token, password, device)
  - Connect/hello-ok handshake
  - Request/response dispatch
  - Event broadcasting
- [ ] Method handlers
  - `config.*` methods
  - `health` method
  - `sessions.*` methods
  - `models.list` method
- [ ] Control UI serving (embedded static files)
- [ ] Config hot-reload with fsnotify

**Milestone**: Gateway starts, accepts WebSocket connections, serves health

### Phase 3: Agent Runtime (Weeks 8-12)

**Goal**: AI agent execution engine

- [ ] LLM provider clients
  - OpenAI API client
  - Anthropic API client
  - Google Gemini client
  - OpenRouter client
  - Ollama client
- [ ] Agent runner (`internal/agents/runner/`)
  - Session management
  - Message history
  - Context compaction
  - Streaming responses
- [ ] Tool system
  - Tool definition and registration
  - Bash tool (os/exec)
  - Web fetch tool
  - Web search tool
  - File operations
- [ ] System prompt generation
- [ ] OpenAI-compatible HTTP API (`/v1/chat/completions`)

**Milestone**: `openclaw agent` runs an interactive session with tool use

### Phase 4: Channel Integrations (Weeks 13-18)

**Goal**: Messaging channel support

- [ ] Channel plugin interface and registry
- [ ] Telegram integration (go-telegram-bot-api)
- [ ] Discord integration (discordgo)
- [ ] Slack integration (slack-go)
- [ ] Auto-reply engine (`internal/autoreply/`)
  - Reply queue with goroutine workers
  - Command handling
  - Directive processing
  - Mention gating
  - Block streaming
- [ ] Message routing
- [ ] Outbound message delivery

**Milestone**: Bot responds to messages on Telegram, Discord, and Slack

### Phase 5: Advanced Features (Weeks 19-24)

**Goal**: Full feature parity

- [ ] Browser automation (playwright-go)
- [ ] Cron scheduling (robfig/cron)
- [ ] Docker sandbox management
- [ ] Media processing (image, PDF, audio)
- [ ] TTS (text-to-speech)
- [ ] Webhook hooks system
- [ ] Plugin system (Go plugin or Wasm)
- [ ] Device pairing and mDNS discovery
- [ ] Tailscale integration
- [ ] Signal channel
- [ ] WhatsApp channel
- [ ] iMessage channel

**Milestone**: All channels and features operational

### Phase 6: CLI & TUI (Weeks 25-28)

**Goal**: Complete CLI and TUI

- [ ] All CLI commands ported
  - Doctor, onboard, configure
  - Channels add/remove/status
  - Models list/set
  - Sessions, cron, plugins, skills
  - Browser, sandbox, daemon
  - Update, security
- [ ] TUI with bubbletea
- [ ] Daemon management (launchd, systemd, schtasks)
- [ ] Onboarding wizard

**Milestone**: Full CLI parity

### Phase 7: Extensions & Polish (Weeks 29-32)

**Goal**: Extensions, testing, documentation

- [ ] MS Teams extension
- [ ] Open Prose VM
- [ ] Memory/RAG extension
- [ ] LLM Task extension
- [ ] Copilot proxy extension
- [ ] Comprehensive test suite (70% coverage)
- [ ] Performance benchmarks
- [ ] Documentation
- [ ] Docker image
- [ ] CI/CD pipeline

**Milestone**: Production-ready release

---

## Risk Assessment

### High Risk

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Pi agent framework rewrite** | The embedded Pi agent runtime (@mariozechner/pi-*) is complex and tightly integrated. Rewriting it in Go is the largest single effort. | Start with a simplified agent runner that calls LLM APIs directly. Incrementally add features (compaction, tool use, streaming). Consider using the Pi agent as a subprocess initially. |
| **WhatsApp Web protocol** | Baileys (WhatsApp Web) is a complex reverse-engineered protocol. No mature Go equivalent exists. | Use a Node.js sidecar for WhatsApp initially, or contribute to/fork a Go WhatsApp library. |
| **Plugin system architecture** | TypeScript plugins use dynamic `require`/`import`. Go's plugin system is limited (Linux only, same Go version). | Use HashiCorp go-plugin (gRPC-based) or Wasm (wazero) for plugin isolation. Alternatively, compile extensions into the binary. |

### Medium Risk

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Browser automation** | playwright-go is less mature than playwright-core for Node.js. | Evaluate playwright-go capabilities early. Fall back to CDP direct integration if needed. |
| **Feature parity gap** | The TypeScript codebase has 500+ files with extensive edge cases. | Prioritize core flows first. Use integration tests from the original project as acceptance criteria. |
| **Config compatibility** | Must read existing `openclaw.json` files without breaking changes. | Implement JSON5 parsing and validate against the same schema. Run config compatibility tests against real configs. |
| **Native app protocol** | macOS/iOS apps communicate via the gateway WebSocket protocol. | Implement protocol conformance tests. Test against the existing Swift client. |

### Low Risk

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Telegram/Discord/Slack SDKs** | Mature Go libraries exist for all three. | Use well-maintained libraries (go-telegram-bot-api, discordgo, slack-go). |
| **HTTP/WebSocket server** | Go excels at network servers. | Use stdlib + gorilla/websocket. Well-understood patterns. |
| **Configuration** | Viper is battle-tested for Go config management. | Standard Viper + godotenv setup. |
| **CLI** | Cobra is the de facto Go CLI framework. | Straightforward port of Commander.js commands to Cobra. |
| **Concurrency** | Go's goroutine model is simpler than Node.js async. | Natural mapping from async/await to goroutines. |

### Key Success Criteria

1. **Protocol compatibility** — Existing macOS/iOS/Android apps must connect without changes
2. **Config compatibility** — Existing `openclaw.json` files must work unchanged
3. **API compatibility** — OpenAI-compatible API must pass existing integration tests
4. **Channel parity** — All messaging channels must function identically
5. **Performance** — Lower memory usage, faster startup, higher connection capacity
6. **Single binary** — No runtime dependencies (except optional: Docker, Playwright, signal-cli)
