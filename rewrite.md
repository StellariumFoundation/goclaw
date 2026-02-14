# Openclaw → Go 1.26.0 Rewrite Blueprint

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Original Architecture Overview](#2-original-architecture-overview)
3. [Module-by-Module Mapping](#3-module-by-module-mapping)
4. [Proposed Go Project Structure](#4-proposed-go-project-structure)
5. [Dependency Mapping](#5-dependency-mapping)
6. [Data Models in Go](#6-data-models-in-go)
7. [API Layer Rewrite Plan](#7-api-layer-rewrite-plan)
8. [Concurrency and Performance](#8-concurrency-and-performance)
9. [Configuration Management](#9-configuration-management)
10. [Error Handling Strategy](#10-error-handling-strategy)
11. [Testing Strategy](#11-testing-strategy)
12. [Migration Phases with Milestones](#12-migration-phases-with-milestones)
13. [Risk Assessment](#13-risk-assessment)

---

## 1. Executive Summary

**Openclaw** is a large-scale, multi-channel AI gateway platform written in TypeScript (Node.js ≥22.12). It orchestrates AI agent interactions across 10+ messaging channels (Discord, Telegram, Slack, WhatsApp, Signal, iMessage, LINE, Mattermost, MS Teams, IRC, web), provides a WebSocket-based gateway server, manages cron jobs, plugins, hooks, browser automation, TTS, memory/embeddings (SQLite + vector search), and exposes both a CLI and a TUI.

The codebase comprises **~1,754 non-test TypeScript source files** across **50+ source directories**, with native mobile apps (iOS/Swift, Android/Kotlin), a macOS desktop app, and a shared Swift SDK (`OpenClawKit`). Extensions are loaded as plugins with a dedicated plugin SDK.

### Why Go?

| Concern | TypeScript (Current) | Go 1.26.0 (Target) |
|---|---|---|
| Startup time | ~500ms–2s (Node.js cold start) | ~10ms (compiled binary) |
| Memory footprint | 150–400 MB typical | 30–80 MB typical |
| Concurrency | Single-threaded event loop + worker threads | Native goroutines + channels |
| Deployment | Requires Node.js runtime | Single static binary |
| Type safety | Compile-time (TS) + runtime (Zod) | Compile-time (strong typing) |
| Cross-compilation | Complex (pkg/nexe) | `GOOS=linux GOARCH=amd64 go build` |

### Scope

This rewrite targets the **server-side core**: gateway, CLI, channels, agents, cron, plugins, hooks, memory, media, config, and infrastructure. Native mobile apps (iOS, Android, macOS) and the browser-based Control UI remain in their current languages but will communicate with the Go gateway via the same WebSocket/HTTP protocol.

---

## 2. Original Architecture Overview

### 2.1 Language & Runtime

- **Language**: TypeScript 5.9+ (ESM modules, `"type": "module"`)
- **Runtime**: Node.js ≥22.12.0
- **Package Manager**: pnpm 10.23.0
- **Build Tool**: tsdown (Rolldown-based bundler)
- **Linting**: oxlint (with type-aware mode)
- **Formatting**: oxfmt
- **Testing**: Vitest 4.x (unit, e2e, live, gateway configs)
- **Schema Validation**: Zod 4.x + AJV 8.x + TypeBox 0.34.x

### 2.2 Entry Points

| Entry | File | Purpose |
|---|---|---|
| CLI entry | `src/entry.ts` | Process bootstrap, respawn for Node flags, delegates to `cli/run-main.ts` |
| Library index | `src/index.ts` | Public API exports |
| Plugin SDK | `src/plugin-sdk/index.ts` | Plugin development API |
| Extension API | `src/extensionAPI.ts` | Extension interface |
| Hook handlers | `src/hooks/bundled/*/handler.ts` | Built-in lifecycle hooks |
| Warning filter | `src/infra/warning-filter.ts` | Node.js warning suppression |

### 2.3 Core Modules (50+ directories)

#### Gateway Server (`src/gateway/`)
The heart of Openclaw. A WebSocket + HTTP server that:
- Authenticates clients (token/password, rate-limited)
- Manages agent sessions and chat
- Routes messages between channels and AI providers
- Serves the Control UI (web dashboard)
- Exposes an OpenAI-compatible HTTP API (`openresponses-http.ts`)
- Handles node registration, discovery (mDNS/Bonjour, wide-area DNS)
- Manages cron jobs, plugins, hooks, exec approvals
- Broadcasts events to connected clients

Key sub-modules:
- `gateway/protocol/` — JSON Schema-based RPC protocol with AJV validation
- `gateway/server-methods/` — RPC method handlers (agents, chat, config, cron, devices, health, logs, models, nodes, sessions, skills, system, talk, tts, usage, voicewake, web, wizard)
- `gateway/server/` — HTTP listener, WebSocket connection management, TLS, health state

#### Agents (`src/agents/`)
AI agent orchestration:
- `pi-embedded-runner.ts` — Core agent execution engine (uses `@mariozechner/pi-agent-core`)
- `pi-embedded-subscribe.ts` — Streaming subscription for agent responses
- `pi-embedded-helpers.ts` — Error classification, sanitization, turn ordering
- `pi-embedded-messaging.ts` — Agent messaging tools
- `model-catalog.ts` / `model-selection.ts` / `model-fallback.ts` — Model discovery and fallback chains
- `auth-profiles/` — Multi-provider auth profile management
- `skills/` — Workspace skill loading and resolution
- `tools/` — 30+ built-in tools (browser, canvas, cron, discord-actions, gateway, image, memory, message, nodes, sessions, slack-actions, telegram-actions, tts, web-fetch, web-search, whatsapp-actions)
- `sandbox/` — Docker-based sandboxed execution
- `schema/` — Tool schema adapters (TypeBox, Gemini clean)
- `pi-extensions/` — Compaction safeguard, context pruning

#### Auto-Reply Engine (`src/auto-reply/`)
Message processing pipeline:
- `reply/` — Core reply dispatcher with 80+ files covering: agent runner, block streaming, command handling, directive parsing, memory flush, model selection, queue management, session management, typing indicators
- `commands-registry.ts` — Native command registration (!/slash commands)
- `dispatch.ts` — Message routing and dispatch
- `heartbeat.ts` — Periodic heartbeat messages
- `skill-commands.ts` — Skill-triggered commands

#### Channels (`src/channels/`)
Channel abstraction layer:
- `registry.ts` — Channel plugin registry
- `plugins/` — Per-channel adapters with normalize, onboarding, outbound, status-issues, actions, agent-tools sub-modules
- `web/` — Web channel (Control UI chat)
- Allowlists, mention gating, command gating, typing indicators

#### Channel Implementations
Each channel has its own top-level directory:

| Directory | Channel | Key Files |
|---|---|---|
| `src/discord/` | Discord | bot monitor, send, voice, presence, threading, guild admin |
| `src/telegram/` | Telegram | Grammy bot, webhooks, inline buttons, draft streaming, sticker cache |
| `src/slack/` | Slack | Bolt framework, Socket Mode, slash commands, threading |
| `src/signal/` | Signal | SSE reconnect, daemon management, reaction levels |
| `src/imessage/` | iMessage | BlueBubbles client, monitor, send |
| `src/line/` | LINE | Flex templates, rich menus, markdown conversion |
| `src/web/` | WhatsApp Web | Baileys client, QR login, auto-reply, media handling |
| `src/whatsapp/` | WhatsApp (shared) | Normalization utilities |

#### Configuration (`src/config/`)
Comprehensive configuration system:
- `config.ts` — Main config loader (JSON5, dotenv, env substitution)
- `schema.ts` / `zod-schema.ts` — Zod-based validation schemas
- `types.ts` — 30+ type definition files (one per domain: agents, auth, channels, cron, discord, gateway, hooks, models, plugins, sandbox, sessions, skills, telegram, tools, tts, whatsapp, etc.)
- `sessions/` — Session store with file-based persistence, locking, pruning, transcript management
- `legacy.ts` / `legacy-migrate.ts` — Config migration from older versions
- `includes.ts` — Config file inclusion/composition
- `env-preserve.ts` / `env-substitution.ts` — Environment variable handling

#### Infrastructure (`src/infra/`)
Cross-cutting concerns:
- `env.ts` — Environment normalization
- `fetch.ts` — HTTP client with retry, proxy support
- `heartbeat-runner.ts` — Heartbeat scheduling and execution
- `restart-sentinel.ts` — Graceful restart coordination
- `update-check.ts` / `update-runner.ts` — Self-update mechanism
- `bonjour-discovery.ts` / `bonjour-ciao.ts` — mDNS service discovery
- `device-auth-store.ts` / `device-identity.ts` / `device-pairing.ts` — Device authentication
- `exec-approvals.ts` — Command execution approval system
- `ssh-config.ts` — SSH configuration parsing
- `net/` — SSRF protection, fetch guards
- `tls/` — TLS fingerprinting, gateway TLS

#### CLI (`src/cli/`)
Commander.js-based CLI with 40+ commands:
- `program.ts` / `program/` — Command tree registration
- `gateway-cli.ts` — Gateway start/stop/discover
- `daemon-cli.ts` — Daemon management (launchd, systemd, schtasks)
- `browser-cli.ts` — Browser automation commands
- `nodes-cli.ts` — Node management (camera, canvas, screen, pairing)
- `plugins-cli.ts` — Plugin install/uninstall/update
- `cron-cli.ts` — Cron job management
- `memory-cli.ts` — Memory/embeddings management
- `skills-cli.ts` — Skills management
- `update-cli.ts` — Self-update commands

#### Cron (`src/cron/`)
Cron job scheduling:
- `service.ts` — Core cron service with timer management
- `schedule.ts` — Cron expression parsing (uses `croner`)
- `store.ts` — Persistent cron job storage
- `isolated-agent/` — Isolated agent execution for cron jobs
- `session-reaper.ts` — Automatic session cleanup

#### Plugins (`src/plugins/`)
Plugin system:
- `loader.ts` / `discovery.ts` — Plugin loading and discovery
- `registry.ts` — Plugin registry
- `manifest.ts` — Plugin manifest parsing
- `runtime/` — Plugin runtime environment
- `services.ts` — Plugin service lifecycle
- `hooks.ts` — Plugin hook integration
- `tools.ts` — Plugin tool registration
- `slots.ts` — Plugin slot system

#### Hooks (`src/hooks/`)
Lifecycle hook system:
- `hooks.ts` — Hook execution engine
- `loader.ts` — Hook file loading
- `install.ts` — Hook installation
- `bundled/` — Built-in hooks (boot-md, bootstrap-extra-files, command-logger, session-memory)
- `gmail.ts` / `gmail-watcher.ts` — Gmail integration hooks

#### Memory & Embeddings (`src/memory/`)
Vector memory system:
- `manager.ts` — Memory manager (sync, search, embedding ops)
- `embeddings.ts` — Embedding generation (OpenAI, Gemini, Voyage, node-llama)
- `sqlite.ts` / `sqlite-vec.ts` — SQLite + sqlite-vec vector storage
- `hybrid.ts` — Hybrid search (vector + keyword)
- `qmd-manager.ts` — QMD (query metadata) management
- `search-manager.ts` — Search orchestration

#### Media (`src/media/`)
Media handling:
- `store.ts` — Media file storage with redirect support
- `server.ts` — Media HTTP server
- `fetch.ts` — Media download
- `image-ops.ts` — Image processing (sharp)
- `audio.ts` / `audio-tags.ts` — Audio processing
- `mime.ts` — MIME type detection

#### Browser Automation (`src/browser/`)
Playwright-based browser control:
- `pw-session.ts` — Playwright session management
- `pw-ai.ts` — AI-assisted browser interaction
- `pw-tools-core.ts` — Browser tool implementations
- `server.ts` — Browser control HTTP server
- `cdp.ts` — Chrome DevTools Protocol integration
- `chrome.ts` — Chrome executable detection and profile management
- `routes/` — HTTP route handlers for browser operations

#### Logging (`src/logging/`)
Structured logging:
- `logger.ts` — Core logger (tslog-based)
- `subsystem.ts` — Per-subsystem loggers
- `redact.ts` — Sensitive data redaction
- `console.ts` — Console output formatting
- `diagnostic.ts` — Diagnostic heartbeat logging

#### Security (`src/security/`)
Security utilities:
- `audit.ts` — Security audit logging
- `secret-equal.ts` — Timing-safe comparison
- `skill-scanner.ts` — Skill file security scanning
- `external-content.ts` — External content sanitization
- `windows-acl.ts` — Windows ACL management

#### TTS (`src/tts/`)
Text-to-speech:
- `tts.ts` — TTS orchestration
- `tts-core.ts` — Core TTS engine (ElevenLabs, Edge TTS)

#### Daemon (`src/daemon/`)
System service management:
- `launchd.ts` — macOS launchd integration
- `systemd.ts` — Linux systemd integration
- `schtasks.ts` — Windows Task Scheduler integration
- `service.ts` — Cross-platform service abstraction

#### Wizard (`src/wizard/`)
Interactive onboarding:
- `onboarding.ts` — Onboarding wizard flow
- `session.ts` — Wizard session management
- `prompts.ts` — Interactive prompts (clack)

#### Other Modules

| Directory | Purpose |
|---|---|
| `src/acp/` | Agent Client Protocol (ACP) server/client |
| `src/canvas-host/` | Canvas hosting server (A2UI) |
| `src/commands/` | Standalone command implementations (doctor, setup, dashboard) |
| `src/compat/` | Legacy name compatibility |
| `src/docs/` | Documentation utilities |
| `src/link-understanding/` | URL/link content extraction |
| `src/macos/` | macOS-specific utilities |
| `src/markdown/` | Markdown processing |
| `src/media-understanding/` | Media content analysis |
| `src/node-host/` | Node host runner (invoke commands on mobile/desktop nodes) |
| `src/pairing/` | Device pairing protocol |
| `src/plugin-sdk/` | Plugin SDK for extension developers |
| `src/process/` | Process management (exec, spawn, command queue, lanes) |
| `src/routing/` | Message routing and session key resolution |
| `src/scripts/` | Build/utility scripts |
| `src/sessions/` | Session utilities (send policy, transcript events) |
| `src/shared/` | Shared utilities (reasoning tag parsing) |
| `src/terminal/` | Terminal UI utilities (ANSI, tables, themes) |
| `src/tui/` | Terminal UI (interactive mode) |
| `src/types/` | Shared type definitions |
| `src/utils/` | General utilities |

### 2.4 Extensions (`extensions/`)

| Extension | Purpose |
|---|---|
| `copilot-proxy/` | GitHub Copilot proxy integration |
| `matrix/` | Matrix messaging protocol |
| `mattermost/` | Mattermost integration |
| `msteams/` | Microsoft Teams integration |
| `open-prose/` | Open Prose DSL for agent workflows |

### 2.5 Native Apps (`apps/`)

| App | Language | Purpose |
|---|---|---|
| `apps/android/` | Kotlin | Android node client |
| `apps/ios/` | Swift | iOS node client |
| `apps/macos/` | Swift | macOS desktop app |
| `apps/shared/OpenClawKit/` | Swift | Shared Swift SDK |

### 2.6 Packages (`packages/`)

| Package | Purpose |
|---|---|
| `clawdbot/` | CLI wrapper package |
| `moltbot/` | Alternative CLI wrapper |

---

## 3. Module-by-Module Mapping

### 3.1 Source Directory → Go Package Mapping

| TypeScript Module | Go Package | Notes |
|---|---|---|
| `src/entry.ts` | `cmd/openclaw/main.go` | CLI entry point |
| `src/index.ts` | `pkg/openclaw/openclaw.go` | Public library API |
| `src/extensionAPI.ts` | `pkg/extensionapi/api.go` | Extension interface |
| `src/acp/` | `internal/acp/` | Agent Client Protocol |
| `src/agents/` | `internal/agents/` | Agent orchestration (split into sub-packages) |
| `src/agents/tools/` | `internal/agents/tools/` | Built-in agent tools |
| `src/agents/auth-profiles/` | `internal/agents/authprofiles/` | Auth profile management |
| `src/agents/skills/` | `internal/agents/skills/` | Skills system |
| `src/agents/sandbox/` | `internal/agents/sandbox/` | Docker sandbox |
| `src/agents/schema/` | `internal/agents/schema/` | Schema adapters |
| `src/agents/pi-embedded-runner/` | `internal/agents/runner/` | Agent execution engine |
| `src/agents/pi-extensions/` | `internal/agents/extensions/` | Agent extensions |
| `src/auto-reply/` | `internal/autoreply/` | Auto-reply engine |
| `src/auto-reply/reply/` | `internal/autoreply/reply/` | Reply pipeline |
| `src/auto-reply/reply/queue/` | `internal/autoreply/queue/` | Reply queue |
| `src/browser/` | `internal/browser/` | Browser automation |
| `src/browser/routes/` | `internal/browser/routes/` | Browser HTTP routes |
| `src/canvas-host/` | `internal/canvashost/` | Canvas hosting |
| `src/channels/` | `internal/channels/` | Channel abstraction |
| `src/channels/plugins/` | `internal/channels/plugins/` | Channel plugin adapters |
| `src/channels/web/` | `internal/channels/web/` | Web channel |
| `src/cli/` | `internal/cli/` | CLI commands |
| `src/cli/program/` | `internal/cli/program/` | Command registration |
| `src/cli/daemon-cli/` | `internal/cli/daemon/` | Daemon commands |
| `src/cli/gateway-cli/` | `internal/cli/gatewaycmd/` | Gateway commands |
| `src/cli/nodes-cli/` | `internal/cli/nodes/` | Node commands |
| `src/cli/update-cli/` | `internal/cli/update/` | Update commands |
| `src/commands/` | `internal/commands/` | Standalone commands |
| `src/compat/` | `internal/compat/` | Legacy compatibility |
| `src/config/` | `internal/config/` | Configuration system |
| `src/config/sessions/` | `internal/config/sessions/` | Session config/store |
| `src/cron/` | `internal/cron/` | Cron scheduling |
| `src/cron/service/` | `internal/cron/service/` | Cron service internals |
| `src/cron/isolated-agent/` | `internal/cron/isolatedagent/` | Isolated cron agents |
| `src/daemon/` | `internal/daemon/` | System service management |
| `src/discord/` | `internal/discord/` | Discord channel |
| `src/discord/monitor/` | `internal/discord/monitor/` | Discord event monitoring |
| `src/docs/` | `internal/docs/` | Documentation utilities |
| `src/gateway/` | `internal/gateway/` | Gateway server (core) |
| `src/gateway/protocol/` | `internal/gateway/protocol/` | RPC protocol |
| `src/gateway/protocol/schema/` | `internal/gateway/protocol/schema/` | Protocol schemas |
| `src/gateway/server/` | `internal/gateway/server/` | HTTP/WS server |
| `src/gateway/server-methods/` | `internal/gateway/methods/` | RPC method handlers |
| `src/hooks/` | `internal/hooks/` | Hook system |
| `src/hooks/bundled/` | `internal/hooks/bundled/` | Built-in hooks |
| `src/imessage/` | `internal/imessage/` | iMessage channel |
| `src/infra/` | `internal/infra/` | Infrastructure utilities |
| `src/infra/net/` | `internal/infra/net/` | Network security (SSRF) |
| `src/infra/tls/` | `internal/infra/tls/` | TLS utilities |
| `src/line/` | `internal/line/` | LINE channel |
| `src/link-understanding/` | `internal/linkunderstanding/` | Link content extraction |
| `src/logging/` | `internal/logging/` | Logging system |
| `src/macos/` | `internal/macos/` | macOS utilities |
| `src/markdown/` | `internal/markdown/` | Markdown processing |
| `src/media/` | `internal/media/` | Media handling |
| `src/media-understanding/` | `internal/mediaunderstanding/` | Media analysis |
| `src/memory/` | `internal/memory/` | Memory/embeddings |
| `src/node-host/` | `internal/nodehost/` | Node host runner |
| `src/pairing/` | `internal/pairing/` | Device pairing |
| `src/plugins/` | `internal/plugins/` | Plugin system |
| `src/plugins/runtime/` | `internal/plugins/runtime/` | Plugin runtime |
| `src/plugin-sdk/` | `pkg/pluginsdk/` | Plugin SDK (public) |
| `src/process/` | `internal/process/` | Process management |
| `src/providers/` | `internal/providers/` | Provider auth (Copilot, Qwen) |
| `src/routing/` | `internal/routing/` | Message routing |
| `src/security/` | `internal/security/` | Security utilities |
| `src/sessions/` | `internal/sessions/` | Session utilities |
| `src/shared/` | `internal/shared/` | Shared utilities |
| `src/signal/` | `internal/signal/` | Signal channel |
| `src/slack/` | `internal/slack/` | Slack channel |
| `src/slack/http/` | `internal/slack/http/` | Slack HTTP mode |
| `src/slack/monitor/` | `internal/slack/monitor/` | Slack event monitoring |
| `src/telegram/` | `internal/telegram/` | Telegram channel |
| `src/telegram/bot/` | `internal/telegram/bot/` | Telegram bot helpers |
| `src/terminal/` | `internal/terminal/` | Terminal UI utilities |
| `src/tts/` | `internal/tts/` | Text-to-speech |
| `src/tui/` | `internal/tui/` | Terminal UI |
| `src/types/` | `internal/types/` | Shared types |
| `src/utils/` | `internal/utils/` | General utilities |
| `src/web/` | `internal/web/` | WhatsApp Web channel |
| `src/web/auto-reply/` | `internal/web/autoreply/` | WhatsApp auto-reply |
| `src/web/inbound/` | `internal/web/inbound/` | WhatsApp inbound |
| `src/whatsapp/` | `internal/whatsapp/` | WhatsApp shared |
| `src/wizard/` | `internal/wizard/` | Onboarding wizard |

### 3.2 Extension → Go Package Mapping

| Extension | Go Package | Notes |
|---|---|---|
| `extensions/copilot-proxy/` | `internal/extensions/copilotproxy/` | Copilot proxy |
| `extensions/matrix/` | `internal/extensions/matrix/` | Matrix protocol |
| `extensions/mattermost/` | `internal/extensions/mattermost/` | Mattermost |
| `extensions/msteams/` | `internal/extensions/msteams/` | MS Teams |
| `extensions/open-prose/` | `internal/extensions/openprose/` | Open Prose DSL |

---

## 4. Proposed Go Project Structure

```
goclaw/
├── cmd/
│   └── openclaw/
│       └── main.go                    # CLI entry point
├── internal/
│   ├── acp/                           # Agent Client Protocol
│   │   ├── client.go
│   │   ├── server.go
│   │   ├── session.go
│   │   ├── translator.go
│   │   └── types.go
│   ├── agents/
│   │   ├── runner/                    # Agent execution engine
│   │   │   ├── embedded.go
│   │   │   ├── subscribe.go
│   │   │   ├── helpers.go
│   │   │   ├── messaging.go
│   │   │   └── utils.go
│   │   ├── authprofiles/              # Auth profile management
│   │   │   ├── store.go
│   │   │   ├── profiles.go
│   │   │   ├── oauth.go
│   │   │   ├── order.go
│   │   │   └── types.go
│   │   ├── extensions/                # Agent extensions
│   │   │   ├── compaction.go
│   │   │   └── contextpruning.go
│   │   ├── sandbox/                   # Docker sandbox
│   │   │   ├── sandbox.go
│   │   │   ├── create.go
│   │   │   ├── merge.go
│   │   │   └── paths.go
│   │   ├── schema/                    # Schema adapters
│   │   │   ├── typebox.go
│   │   │   └── gemini.go
│   │   ├── skills/                    # Skills system
│   │   │   ├── skills.go
│   │   │   ├── install.go
│   │   │   ├── status.go
│   │   │   └── refresh.go
│   │   ├── tools/                     # Built-in tools (30+)
│   │   │   ├── browser.go
│   │   │   ├── canvas.go
│   │   │   ├── cron.go
│   │   │   ├── discord_actions.go
│   │   │   ├── gateway.go
│   │   │   ├── image.go
│   │   │   ├── memory.go
│   │   │   ├── message.go
│   │   │   ├── nodes.go
│   │   │   ├── sessions.go
│   │   │   ├── slack_actions.go
│   │   │   ├── telegram_actions.go
│   │   │   ├── tts.go
│   │   │   ├── web_fetch.go
│   │   │   ├── web_search.go
│   │   │   └── whatsapp_actions.go
│   │   ├── agent_scope.go
│   │   ├── context.go
│   │   ├── identity.go
│   │   ├── model_catalog.go
│   │   ├── model_fallback.go
│   │   ├── model_selection.go
│   │   ├── models_config.go
│   │   ├── system_prompt.go
│   │   └── subagent_registry.go
│   ├── autoreply/
│   │   ├── reply/                     # Reply pipeline
│   │   │   ├── agent_runner.go
│   │   │   ├── block_streaming.go
│   │   │   ├── commands.go
│   │   │   ├── directive.go
│   │   │   ├── dispatcher.go
│   │   │   ├── formatting.go
│   │   │   ├── history.go
│   │   │   ├── memory_flush.go
│   │   │   ├── model_selection.go
│   │   │   ├── normalize.go
│   │   │   ├── session.go
│   │   │   └── typing.go
│   │   ├── queue/                     # Reply queue
│   │   │   ├── enqueue.go
│   │   │   ├── drain.go
│   │   │   ├── state.go
│   │   │   └── types.go
│   │   ├── commands_registry.go
│   │   ├── dispatch.go
│   │   ├── heartbeat.go
│   │   ├── model.go
│   │   └── status.go
│   ├── browser/
│   │   ├── routes/
│   │   │   ├── agent.go
│   │   │   ├── basic.go
│   │   │   ├── dispatcher.go
│   │   │   └── tabs.go
│   │   ├── cdp.go
│   │   ├── chrome.go
│   │   ├── client.go
│   │   ├── config.go
│   │   ├── profiles.go
│   │   ├── pw_session.go
│   │   ├── pw_tools.go
│   │   ├── screenshot.go
│   │   └── server.go
│   ├── canvashost/
│   │   ├── a2ui.go
│   │   └── server.go
│   ├── channels/
│   │   ├── plugins/
│   │   │   ├── catalog.go
│   │   │   ├── config.go
│   │   │   ├── loader.go
│   │   │   ├── normalize.go
│   │   │   ├── onboarding.go
│   │   │   ├── outbound.go
│   │   │   ├── status.go
│   │   │   └── types.go
│   │   ├── web/
│   │   │   └── web.go
│   │   ├── registry.go
│   │   ├── config.go
│   │   ├── mention_gating.go
│   │   ├── command_gating.go
│   │   ├── sender_identity.go
│   │   ├── session.go
│   │   ├── targets.go
│   │   └── typing.go
│   ├── cli/
│   │   ├── program/
│   │   │   ├── build.go
│   │   │   ├── registry.go
│   │   │   ├── context.go
│   │   │   └── help.go
│   │   ├── daemon/
│   │   │   ├── install.go
│   │   │   ├── lifecycle.go
│   │   │   ├── probe.go
│   │   │   └── status.go
│   │   ├── gatewaycmd/
│   │   │   ├── run.go
│   │   │   ├── discover.go
│   │   │   └── call.go
│   │   ├── nodes/
│   │   │   ├── camera.go
│   │   │   ├── canvas.go
│   │   │   ├── screen.go
│   │   │   └── pairing.go
│   │   ├── update/
│   │   │   ├── update.go
│   │   │   ├── status.go
│   │   │   └── wizard.go
│   │   ├── browser_cli.go
│   │   ├── channels_cli.go
│   │   ├── config_cli.go
│   │   ├── cron_cli.go
│   │   ├── memory_cli.go
│   │   ├── models_cli.go
│   │   ├── plugins_cli.go
│   │   ├── profile.go
│   │   ├── run_main.go
│   │   └── skills_cli.go
│   ├── commands/
│   │   ├── doctor.go
│   │   ├── setup.go
│   │   ├── dashboard.go
│   │   └── onboard.go
│   ├── config/
│   │   ├── sessions/
│   │   │   ├── store.go
│   │   │   ├── transcript.go
│   │   │   ├── metadata.go
│   │   │   ├── paths.go
│   │   │   └── types.go
│   │   ├── config.go
│   │   ├── defaults.go
│   │   ├── env_preserve.go
│   │   ├── env_substitution.go
│   │   ├── env_vars.go
│   │   ├── includes.go
│   │   ├── io.go
│   │   ├── legacy_migrate.go
│   │   ├── merge.go
│   │   ├── normalize_paths.go
│   │   ├── paths.go
│   │   ├── plugin_auto_enable.go
│   │   ├── schema.go
│   │   ├── types.go
│   │   └── validation.go
│   ├── cron/
│   │   ├── service/
│   │   │   ├── jobs.go
│   │   │   ├── locked.go
│   │   │   ├── ops.go
│   │   │   ├── state.go
│   │   │   ├── store.go
│   │   │   └── timer.go
│   │   ├── isolatedagent/
│   │   │   ├── delivery.go
│   │   │   ├── helpers.go
│   │   │   ├── run.go
│   │   │   └── session.go
│   │   ├── delivery.go
│   │   ├── normalize.go
│   │   ├── parse.go
│   │   ├── schedule.go
│   │   ├── service.go
│   │   ├── session_reaper.go
│   │   ├── store.go
│   │   └── types.go
│   ├── daemon/
│   │   ├── launchd.go
│   │   ├── systemd.go
│   │   ├── schtasks.go
│   │   ├── service.go
│   │   ├── paths.go
│   │   └── constants.go
│   ├── discord/
│   │   ├── monitor/
│   │   │   ├── components.go
│   │   │   ├── allowlist.go
│   │   │   ├── exec_approvals.go
│   │   │   ├── handler.go
│   │   │   ├── presence.go
│   │   │   ├── provider.go
│   │   │   ├── threading.go
│   │   │   └── typing.go
│   │   ├── accounts.go
│   │   ├── api.go
│   │   ├── audit.go
│   │   ├── bot.go
│   │   ├── chunk.go
│   │   ├── monitor.go
│   │   ├── probe.go
│   │   ├── send.go
│   │   ├── targets.go
│   │   └── token.go
│   ├── extensions/
│   │   ├── copilotproxy/
│   │   ├── matrix/
│   │   ├── mattermost/
│   │   ├── msteams/
│   │   └── openprose/
│   ├── gateway/
│   │   ├── protocol/
│   │   │   ├── schema/
│   │   │   │   ├── agent.go
│   │   │   │   ├── channels.go
│   │   │   │   ├── config.go
│   │   │   │   ├── cron.go
│   │   │   │   ├── devices.go
│   │   │   │   ├── error_codes.go
│   │   │   │   ├── exec_approvals.go
│   │   │   │   ├── frames.go
│   │   │   │   ├── logs.go
│   │   │   │   ├── nodes.go
│   │   │   │   ├── primitives.go
│   │   │   │   ├── sessions.go
│   │   │   │   ├── snapshot.go
│   │   │   │   ├── types.go
│   │   │   │   └── wizard.go
│   │   │   ├── protocol.go
│   │   │   ├── client_info.go
│   │   │   └── validator.go
│   │   ├── server/
│   │   │   ├── http_listen.go
│   │   │   ├── ws_connection.go
│   │   │   ├── tls.go
│   │   │   ├── health.go
│   │   │   ├── hooks.go
│   │   │   └── plugins_http.go
│   │   ├── methods/
│   │   │   ├── agent.go
│   │   │   ├── agents.go
│   │   │   ├── browser.go
│   │   │   ├── channels.go
│   │   │   ├── chat.go
│   │   │   ├── config.go
│   │   │   ├── connect.go
│   │   │   ├── cron.go
│   │   │   ├── devices.go
│   │   │   ├── exec_approval.go
│   │   │   ├── health.go
│   │   │   ├── logs.go
│   │   │   ├── models.go
│   │   │   ├── nodes.go
│   │   │   ├── send.go
│   │   │   ├── sessions.go
│   │   │   ├── skills.go
│   │   │   ├── system.go
│   │   │   ├── talk.go
│   │   │   ├── tts.go
│   │   │   ├── usage.go
│   │   │   ├── voicewake.go
│   │   │   ├── web.go
│   │   │   └── wizard.go
│   │   ├── auth.go
│   │   ├── auth_rate_limit.go
│   │   ├── boot.go
│   │   ├── call.go
│   │   ├── chat_attachments.go
│   │   ├── chat_sanitize.go
│   │   ├── client.go
│   │   ├── config_reload.go
│   │   ├── control_ui.go
│   │   ├── device_auth.go
│   │   ├── exec_approval_manager.go
│   │   ├── hooks_mapping.go
│   │   ├── hooks.go
│   │   ├── net.go
│   │   ├── node_registry.go
│   │   ├── openresponses_http.go
│   │   ├── origin_check.go
│   │   ├── probe.go
│   │   ├── server_broadcast.go
│   │   ├── server_channels.go
│   │   ├── server_chat.go
│   │   ├── server_close.go
│   │   ├── server_cron.go
│   │   ├── server_discovery.go
│   │   ├── server_lanes.go
│   │   ├── server_maintenance.go
│   │   ├── server_methods.go
│   │   ├── server_plugins.go
│   │   ├── server_reload.go
│   │   ├── server_runtime_config.go
│   │   ├── server_runtime_state.go
│   │   ├── server_session_key.go
│   │   ├── server_startup.go
│   │   ├── server_tailscale.go
│   │   ├── server_wizard_sessions.go
│   │   ├── server.go
│   │   └── session_utils.go
│   ├── hooks/
│   │   ├── bundled/
│   │   │   ├── bootmd/
│   │   │   ├── bootstrapfiles/
│   │   │   ├── commandlogger/
│   │   │   └── sessionmemory/
│   │   ├── hooks.go
│   │   ├── loader.go
│   │   ├── install.go
│   │   ├── gmail.go
│   │   └── types.go
│   ├── imessage/
│   │   ├── monitor/
│   │   ├── accounts.go
│   │   ├── client.go
│   │   ├── monitor.go
│   │   ├── probe.go
│   │   ├── send.go
│   │   └── targets.go
│   ├── infra/
│   │   ├── net/
│   │   │   ├── fetch_guard.go
│   │   │   └── ssrf.go
│   │   ├── tls/
│   │   │   ├── fingerprint.go
│   │   │   └── gateway.go
│   │   ├── bonjour.go
│   │   ├── device_auth.go
│   │   ├── device_identity.go
│   │   ├── device_pairing.go
│   │   ├── env.go
│   │   ├── errors.go
│   │   ├── exec_approvals.go
│   │   ├── fetch.go
│   │   ├── fs_safe.go
│   │   ├── heartbeat_runner.go
│   │   ├── home_dir.go
│   │   ├── http_body.go
│   │   ├── path_env.go
│   │   ├── ports.go
│   │   ├── provider_usage.go
│   │   ├── restart_sentinel.go
│   │   ├── retry.go
│   │   ├── runtime_guard.go
│   │   ├── session_cost.go
│   │   ├── shell_env.go
│   │   ├── skills_remote.go
│   │   ├── ssh_config.go
│   │   ├── state_migrations.go
│   │   ├── system_events.go
│   │   ├── system_presence.go
│   │   ├── update_check.go
│   │   ├── update_runner.go
│   │   └── ws.go
│   ├── line/
│   │   ├── flex_templates/
│   │   ├── accounts.go
│   │   ├── bot.go
│   │   ├── config_schema.go
│   │   ├── download.go
│   │   ├── flex_templates.go
│   │   ├── markdown_to_line.go
│   │   ├── monitor.go
│   │   ├── probe.go
│   │   ├── reply_chunks.go
│   │   ├── rich_menu.go
│   │   └── send.go
│   ├── logging/
│   │   ├── config.go
│   │   ├── console.go
│   │   ├── diagnostic.go
│   │   ├── levels.go
│   │   ├── logger.go
│   │   ├── redact.go
│   │   ├── state.go
│   │   └── subsystem.go
│   ├── markdown/
│   │   └── markdown.go
│   ├── media/
│   │   ├── audio.go
│   │   ├── constants.go
│   │   ├── fetch.go
│   │   ├── host.go
│   │   ├── image_ops.go
│   │   ├── input_files.go
│   │   ├── mime.go
│   │   ├── parse.go
│   │   ├── server.go
│   │   └── store.go
│   ├── memory/
│   │   ├── embeddings.go
│   │   ├── hybrid.go
│   │   ├── manager.go
│   │   ├── qmd.go
│   │   ├── search.go
│   │   ├── session_files.go
│   │   ├── sqlite.go
│   │   ├── sqlite_vec.go
│   │   ├── sync.go
│   │   └── types.go
│   ├── nodehost/
│   │   ├── config.go
│   │   ├── invoke.go
│   │   ├── runner.go
│   │   └── timeout.go
│   ├── pairing/
│   │   └── pairing.go
│   ├── plugins/
│   │   ├── runtime/
│   │   │   ├── runtime.go
│   │   │   └── types.go
│   │   ├── config.go
│   │   ├── discovery.go
│   │   ├── hooks.go
│   │   ├── install.go
│   │   ├── loader.go
│   │   ├── manifest.go
│   │   ├── registry.go
│   │   ├── services.go
│   │   ├── slots.go
│   │   ├── tools.go
│   │   └── types.go
│   ├── process/
│   │   ├── command_queue.go
│   │   ├── exec.go
│   │   ├── lanes.go
│   │   ├── restart_recovery.go
│   │   └── spawn.go
│   ├── providers/
│   │   ├── copilot.go
│   │   └── qwen.go
│   ├── routing/
│   │   ├── bindings.go
│   │   ├── resolve_route.go
│   │   └── session_key.go
│   ├── security/
│   │   ├── audit.go
│   │   ├── external_content.go
│   │   ├── secret_equal.go
│   │   ├── skill_scanner.go
│   │   └── windows_acl.go
│   ├── sessions/
│   │   ├── send_policy.go
│   │   ├── session_key.go
│   │   └── transcript_events.go
│   ├── signal/
│   │   ├── monitor/
│   │   ├── accounts.go
│   │   ├── client.go
│   │   ├── daemon.go
│   │   ├── format.go
│   │   ├── identity.go
│   │   ├── monitor.go
│   │   ├── probe.go
│   │   ├── send.go
│   │   └── sse_reconnect.go
│   ├── slack/
│   │   ├── http/
│   │   ├── monitor/
│   │   │   ├── events/
│   │   │   ├── message_handler/
│   │   │   ├── allowlist.go
│   │   │   ├── auth.go
│   │   │   ├── channel_config.go
│   │   │   ├── commands.go
│   │   │   ├── context.go
│   │   │   ├── events.go
│   │   │   ├── media.go
│   │   │   ├── policy.go
│   │   │   ├── provider.go
│   │   │   ├── replies.go
│   │   │   ├── slash.go
│   │   │   └── thread_resolution.go
│   │   ├── accounts.go
│   │   ├── actions.go
│   │   ├── client.go
│   │   ├── format.go
│   │   ├── monitor.go
│   │   ├── probe.go
│   │   ├── resolve_channels.go
│   │   ├── send.go
│   │   ├── targets.go
│   │   ├── threading.go
│   │   └── token.go
│   ├── telegram/
│   │   ├── bot/
│   │   ├── accounts.go
│   │   ├── bot.go
│   │   ├── bot_handlers.go
│   │   ├── bot_message.go
│   │   ├── bot_native_commands.go
│   │   ├── download.go
│   │   ├── draft_streaming.go
│   │   ├── format.go
│   │   ├── inline_buttons.go
│   │   ├── monitor.go
│   │   ├── network_config.go
│   │   ├── probe.go
│   │   ├── proxy.go
│   │   ├── send.go
│   │   ├── sticker_cache.go
│   │   ├── targets.go
│   │   ├── token.go
│   │   ├── voice.go
│   │   └── webhook.go
│   ├── terminal/
│   │   ├── ansi.go
│   │   ├── links.go
│   │   ├── palette.go
│   │   ├── progress.go
│   │   ├── table.go
│   │   └── theme.go
│   ├── tts/
│   │   ├── tts.go
│   │   └── core.go
│   ├── tui/
│   │   └── tui.go
│   ├── web/
│   │   ├── autoreply/
│   │   ├── inbound/
│   │   ├── accounts.go
│   │   ├── auto_reply.go
│   │   ├── inbound.go
│   │   ├── login.go
│   │   ├── media.go
│   │   ├── monitor.go
│   │   ├── outbound.go
│   │   ├── reconnect.go
│   │   └── session.go
│   ├── whatsapp/
│   │   └── normalize.go
│   └── wizard/
│       ├── onboarding.go
│       ├── prompts.go
│       └── session.go
├── pkg/
│   ├── openclaw/
│   │   └── openclaw.go                # Public library API
│   ├── pluginsdk/
│   │   └── sdk.go                     # Plugin SDK (public)
│   └── extensionapi/
│       └── api.go                     # Extension API (public)
├── assets/                            # Static assets (copied from original)
├── docs/                              # Documentation
├── extensions/                        # Extension skill files (prose, etc.)
├── skills/                            # Skill definitions
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── .env.example
└── README.md
```

---

## 5. Dependency Mapping

### 5.1 Core Dependencies

| TypeScript Package | Go Equivalent | Notes |
|---|---|---|
| `express` 5.x | `net/http` (stdlib) + `github.com/go-chi/chi/v5` | HTTP routing; chi is lightweight and idiomatic |
| `ws` 8.x | `github.com/gorilla/websocket` | WebSocket server/client |
| `commander` 14.x | `github.com/spf13/cobra` | CLI framework |
| `zod` 4.x | Go struct tags + `github.com/go-playground/validator/v10` | Validation |
| `ajv` 8.x | `github.com/santhosh-tekuri/jsonschema/v6` | JSON Schema validation |
| `@sinclair/typebox` | Go struct definitions | Static typing replaces runtime schema |
| `dotenv` 17.x | `github.com/joho/godotenv` | .env file loading |
| `chalk` 5.x | `github.com/fatih/color` | Terminal colors |
| `yaml` 2.x | `gopkg.in/yaml.v3` | YAML parsing |
| `json5` 2.x | `github.com/yosuke-furukawa/json5` or custom parser | JSON5 config support |
| `tslog` 4.x | `github.com/rs/zerolog` or `log/slog` (stdlib) | Structured logging |
| `croner` 10.x | `github.com/robfig/cron/v3` | Cron scheduling |
| `chokidar` 5.x | `github.com/fsnotify/fsnotify` | File watching |
| `undici` 7.x | `net/http` (stdlib) | HTTP client |
| `proper-lockfile` 4.x | `github.com/gofrs/flock` | File locking |
| `tar` 7.x | `archive/tar` (stdlib) | Tar archives |
| `jszip` 3.x | `archive/zip` (stdlib) | ZIP archives |
| `sharp` 0.34.x | `github.com/disintegration/imaging` or CGo bindings | Image processing |
| `markdown-it` 14.x | `github.com/yuin/goldmark` | Markdown parsing |
| `linkedom` 0.18.x | `golang.org/x/net/html` | HTML parsing |
| `@mozilla/readability` | `github.com/nicholasgasior/goinern` or port | Content extraction |
| `osc-progress` | `github.com/schollz/progressbar/v3` | Progress bars |
| `cli-highlight` | `github.com/alecthomas/chroma/v2` | Syntax highlighting |
| `@clack/prompts` | `github.com/charmbracelet/huh` | Interactive prompts |
| `long` 5.x | `math/big` (stdlib) | Big integers |

### 5.2 AI/ML Dependencies

| TypeScript Package | Go Equivalent | Notes |
|---|---|---|
| `@mariozechner/pi-agent-core` | Custom Go implementation | Core agent loop — must be reimplemented |
| `@mariozechner/pi-ai` | Custom Go implementation | AI provider abstraction |
| `@mariozechner/pi-coding-agent` | Custom Go implementation | Coding agent capabilities |
| `@mariozechner/pi-tui` | `github.com/charmbracelet/bubbletea` | TUI framework |
| `ollama` | `github.com/ollama/ollama/api` | Ollama client |
| `pdfjs-dist` | `github.com/ledongthuc/pdf` or `github.com/unidoc/unipdf` | PDF parsing |

### 5.3 Channel-Specific Dependencies

| TypeScript Package | Go Equivalent | Notes |
|---|---|---|
| `grammy` 1.x | `github.com/go-telegram-bot-api/telegram-bot-api/v5` or `github.com/gotd/td` | Telegram bot |
| `@grammyjs/runner` | Goroutine-based runner | Native concurrency |
| `@grammyjs/transformer-throttler` | Custom rate limiter (`golang.org/x/time/rate`) | Rate limiting |
| `discord-api-types` | `github.com/bwmarrin/discordgo` | Discord API |
| `@slack/bolt` 4.x | `github.com/slack-go/slack` | Slack API |
| `@slack/web-api` 7.x | `github.com/slack-go/slack` | Slack Web API |
| `@whiskeysockets/baileys` 7.x | Custom Go WhatsApp client or `github.com/nicholasgasior/gowsp` | WhatsApp Web |
| `@line/bot-sdk` 10.x | `github.com/line/line-bot-sdk-go/v8` | LINE bot |
| `@larksuiteoapi/node-sdk` | `github.com/larksuite/oapi-sdk-go/v3` | Lark/Feishu |
| `@buape/carbon` | Custom MS Teams implementation | MS Teams |
| `signal-utils` | Custom Signal CLI wrapper | Signal messenger |

### 5.4 Infrastructure Dependencies

| TypeScript Package | Go Equivalent | Notes |
|---|---|---|
| `@homebridge/ciao` | `github.com/hashicorp/mdns` or `github.com/grandcat/zeroconf` | mDNS/Bonjour |
| `playwright-core` 1.x | `github.com/chromedp/chromedp` | Browser automation (CDP) |
| `sqlite-vec` | `github.com/asg017/sqlite-vec` (CGo) or pure Go alternative | Vector search |
| `better-sqlite3` (implicit) | `github.com/mattn/go-sqlite3` or `modernc.org/sqlite` | SQLite |
| `@lydell/node-pty` | `github.com/creack/pty` | PTY management |
| `https-proxy-agent` | `net/http` with proxy transport | HTTP proxy |
| `qrcode-terminal` | `github.com/mdp/qrterminal/v3` | QR code display |
| `file-type` 21.x | `net/http.DetectContentType` + `github.com/h2non/filetype` | File type detection |
| `node-edge-tts` | Custom HTTP client to Edge TTS API | Edge TTS |
| `jiti` 2.x | Not needed (Go compiles) | Dynamic imports |

### 5.5 AWS/Cloud Dependencies

| TypeScript Package | Go Equivalent | Notes |
|---|---|---|
| `@aws-sdk/client-bedrock` | `github.com/aws/aws-sdk-go-v2/service/bedrock` | AWS Bedrock |
| `@agentclientprotocol/sdk` | Custom Go ACP implementation | Agent Client Protocol |

---

## 6. Data Models in Go

### 6.1 Configuration Types

```go
// internal/config/types.go

package config

// OpenClawConfig is the root configuration structure
type OpenClawConfig struct {
    Gateway    GatewayConfig    `json:"gateway,omitempty"`
    Agents     AgentsConfig     `json:"agents,omitempty"`
    Channels   ChannelsConfig   `json:"channels,omitempty"`
    Cron       CronConfig       `json:"cron,omitempty"`
    Hooks      HooksConfig      `json:"hooks,omitempty"`
    Plugins    PluginsConfig    `json:"plugins,omitempty"`
    Models     ModelsConfig     `json:"models,omitempty"`
    Skills     SkillsConfig     `json:"skills,omitempty"`
    Tools      ToolsConfig      `json:"tools,omitempty"`
    Sessions   SessionConfig    `json:"sessions,omitempty"`
    Memory     MemoryConfig     `json:"memory,omitempty"`
    TTS        TTSConfig        `json:"tts,omitempty"`
    Sandbox    SandboxConfig    `json:"sandbox,omitempty"`
    Browser    BrowserConfig    `json:"browser,omitempty"`
    Env        map[string]string `json:"env,omitempty"`
    Includes   []string         `json:"includes,omitempty"`
}

type GatewayConfig struct {
    Port          int                  `json:"port,omitempty"`
    Bind          GatewayBindMode      `json:"bind,omitempty"`
    Auth          GatewayAuthConfig    `json:"auth,omitempty"`
    TLS           GatewayTLSConfig     `json:"tls,omitempty"`
    Discovery     DiscoveryConfig      `json:"discovery,omitempty"`
    ControlUI     GatewayControlUIConfig `json:"controlUi,omitempty"`
    CanvasHost    CanvasHostConfig     `json:"canvasHost,omitempty"`
    Talk          TalkConfig           `json:"talk,omitempty"`
    MaxPayload    int                  `json:"maxPayload,omitempty"`
    Tailscale     *TailscaleConfig     `json:"tailscale,omitempty"`
}

type GatewayBindMode string

const (
    BindAuto     GatewayBindMode = "auto"
    BindLAN      GatewayBindMode = "lan"
    BindLoopback GatewayBindMode = "loopback"
    BindCustom   GatewayBindMode = "custom"
    BindTailnet  GatewayBindMode = "tailnet"
)

type GatewayAuthConfig struct {
    Mode     GatewayAuthMode `json:"mode,omitempty"`
    Token    string          `json:"token,omitempty"`
    Password string          `json:"password,omitempty"`
}

type GatewayAuthMode string

const (
    AuthToken    GatewayAuthMode = "token"
    AuthPassword GatewayAuthMode = "password"
)

type GatewayTLSConfig struct {
    Enabled      bool   `json:"enabled,omitempty"`
    AutoGenerate bool   `json:"autoGenerate,omitempty"`
    CertPath     string `json:"certPath,omitempty"`
    KeyPath      string `json:"keyPath,omitempty"`
    CAPath       string `json:"caPath,omitempty"`
}

type DiscoveryConfig struct {
    WideArea *WideAreaDiscoveryConfig `json:"wideArea,omitempty"`
    MDNS     *MDNSDiscoveryConfig     `json:"mdns,omitempty"`
}

type WideAreaDiscoveryConfig struct {
    Enabled bool   `json:"enabled,omitempty"`
    Domain  string `json:"domain,omitempty"`
}

type MDNSDiscoveryConfig struct {
    Mode MDNSDiscoveryMode `json:"mode,omitempty"`
}

type MDNSDiscoveryMode string

const (
    MDNSOff     MDNSDiscoveryMode = "off"
    MDNSMinimal MDNSDiscoveryMode = "minimal"
    MDNSFull    MDNSDiscoveryMode = "full"
)

type GatewayControlUIConfig struct {
    Enabled                      bool     `json:"enabled,omitempty"`
    BasePath                     string   `json:"basePath,omitempty"`
    Root                         string   `json:"root,omitempty"`
    AllowedOrigins               []string `json:"allowedOrigins,omitempty"`
    AllowInsecureAuth            bool     `json:"allowInsecureAuth,omitempty"`
    DangerouslyDisableDeviceAuth bool     `json:"dangerouslyDisableDeviceAuth,omitempty"`
}

type CanvasHostConfig struct {
    Enabled    bool   `json:"enabled,omitempty"`
    Root       string `json:"root,omitempty"`
    Port       int    `json:"port,omitempty"`
    LiveReload bool   `json:"liveReload,omitempty"`
}

type TalkConfig struct {
    VoiceID          string            `json:"voiceId,omitempty"`
    VoiceAliases     map[string]string `json:"voiceAliases,omitempty"`
    ModelID          string            `json:"modelId,omitempty"`
    OutputFormat     string            `json:"outputFormat,omitempty"`
    APIKey           string            `json:"apiKey,omitempty"`
    InterruptOnSpeech bool            `json:"interruptOnSpeech,omitempty"`
}
```

### 6.2 Agent Types

```go
// internal/config/types_agents.go

type AgentsConfig struct {
    Defaults *AgentDefaultsConfig `json:"defaults,omitempty"`
    List     []AgentConfig        `json:"list,omitempty"`
}

type AgentConfig struct {
    ID          string              `json:"id"`
    Default     bool                `json:"default,omitempty"`
    Name        string              `json:"name,omitempty"`
    Workspace   string              `json:"workspace,omitempty"`
    AgentDir    string              `json:"agentDir,omitempty"`
    Model       *AgentModelConfig   `json:"model,omitempty"`
    Skills      []string            `json:"skills,omitempty"`
    MemorySearch *MemorySearchConfig `json:"memorySearch,omitempty"`
    HumanDelay  *HumanDelayConfig   `json:"humanDelay,omitempty"`
    Heartbeat   *HeartbeatConfig    `json:"heartbeat,omitempty"`
    Identity    *IdentityConfig     `json:"identity,omitempty"`
    GroupChat   *GroupChatConfig    `json:"groupChat,omitempty"`
    Subagents   *SubagentsConfig    `json:"subagents,omitempty"`
    Sandbox     *AgentSandboxConfig `json:"sandbox,omitempty"`
    Tools       *AgentToolsConfig   `json:"tools,omitempty"`
}

type AgentModelConfig struct {
    Primary   string   `json:"primary,omitempty"`
    Fallbacks []string `json:"fallbacks,omitempty"`
}

type AgentDefaultsConfig struct {
    Model       *AgentModelConfig `json:"model,omitempty"`
    Heartbeat   *HeartbeatConfig  `json:"heartbeat,omitempty"`
    HumanDelay  *HumanDelayConfig `json:"humanDelay,omitempty"`
    Identity    *IdentityConfig   `json:"identity,omitempty"`
}

type AgentBinding struct {
    AgentID string             `json:"agentId"`
    Match   AgentBindingMatch  `json:"match"`
}

type AgentBindingMatch struct {
    Channel   string `json:"channel"`
    AccountID string `json:"accountId,omitempty"`
    Peer      *PeerMatch `json:"peer,omitempty"`
    GuildID   string `json:"guildId,omitempty"`
    TeamID    string `json:"teamId,omitempty"`
}

type PeerMatch struct {
    Kind ChatType `json:"kind"`
    ID   string   `json:"id"`
}

type AgentSandboxConfig struct {
    Mode                  string                `json:"mode,omitempty"`
    WorkspaceAccess       string                `json:"workspaceAccess,omitempty"`
    SessionToolsVisibility string               `json:"sessionToolsVisibility,omitempty"`
    Scope                 string                `json:"scope,omitempty"`
    PerSession            bool                  `json:"perSession,omitempty"`
    WorkspaceRoot         string                `json:"workspaceRoot,omitempty"`
    Docker                *SandboxDockerSettings `json:"docker,omitempty"`
    Browser               *SandboxBrowserSettings `json:"browser,omitempty"`
    Prune                 *SandboxPruneSettings  `json:"prune,omitempty"`
}

type SubagentsConfig struct {
    AllowAgents []string          `json:"allowAgents,omitempty"`
    Model       *AgentModelConfig `json:"model,omitempty"`
}
```

### 6.3 Session Types

```go
// internal/config/types_sessions.go

type SessionScope string

const (
    SessionScopePerSender SessionScope = "per-sender"
    SessionScopeGlobal    SessionScope = "global"
)

type DmScope string

const (
    DmScopeMain              DmScope = "main"
    DmScopePerPeer           DmScope = "per-peer"
    DmScopePerChannelPeer    DmScope = "per-channel-peer"
    DmScopePerAccountChannelPeer DmScope = "per-account-channel-peer"
)

type SessionConfig struct {
    Scope                SessionScope              `json:"scope,omitempty"`
    DmScope              DmScope                   `json:"dmScope,omitempty"`
    IdentityLinks        map[string][]string       `json:"identityLinks,omitempty"`
    ResetTriggers        []string                  `json:"resetTriggers,omitempty"`
    IdleMinutes          int                       `json:"idleMinutes,omitempty"`
    Reset                *SessionResetConfig       `json:"reset,omitempty"`
    ResetByType          *SessionResetByTypeConfig `json:"resetByType,omitempty"`
    ResetByChannel       map[string]SessionResetConfig `json:"resetByChannel,omitempty"`
    Store                string                    `json:"store,omitempty"`
    TypingIntervalSeconds int                      `json:"typingIntervalSeconds,omitempty"`
    TypingMode           TypingMode                `json:"typingMode,omitempty"`
    MainKey              string                    `json:"mainKey,omitempty"`
    SendPolicy           *SessionSendPolicyConfig  `json:"sendPolicy,omitempty"`
    AgentToAgent         *AgentToAgentConfig       `json:"agentToAgent,omitempty"`
}

type SessionResetMode string

const (
    ResetDaily SessionResetMode = "daily"
    ResetIdle  SessionResetMode = "idle"
)

type SessionResetConfig struct {
    Mode        SessionResetMode `json:"mode,omitempty"`
    AtHour      int              `json:"atHour,omitempty"`
    IdleMinutes int              `json:"idleMinutes,omitempty"`
}

type SessionResetByTypeConfig struct {
    Direct *SessionResetConfig `json:"direct,omitempty"`
    DM     *SessionResetConfig `json:"dm,omitempty"` // deprecated
    Group  *SessionResetConfig `json:"group,omitempty"`
    Thread *SessionResetConfig `json:"thread,omitempty"`
}

type TypingMode string

const (
    TypingNever    TypingMode = "never"
    TypingInstant  TypingMode = "instant"
    TypingThinking TypingMode = "thinking"
    TypingMessage  TypingMode = "message"
)

type SessionSendPolicyConfig struct {
    Default string                  `json:"default,omitempty"`
    Rules   []SessionSendPolicyRule `json:"rules,omitempty"`
}

type SessionSendPolicyRule struct {
    Action string                  `json:"action"`
    Match  *SessionSendPolicyMatch `json:"match,omitempty"`
}

type SessionSendPolicyMatch struct {
    Channel   string   `json:"channel,omitempty"`
    ChatType  ChatType `json:"chatType,omitempty"`
    KeyPrefix string   `json:"keyPrefix,omitempty"`
}

type AgentToAgentConfig struct {
    MaxPingPongTurns int `json:"maxPingPongTurns,omitempty"`
}
```

### 6.4 Channel Types

```go
// internal/config/types_channels.go

type ChatType string

const (
    ChatTypeDirect  ChatType = "direct"
    ChatTypeGroup   ChatType = "group"
    ChatTypeThread  ChatType = "thread"
    ChatTypeChannel ChatType = "channel"
)

type ReplyMode string

const (
    ReplyModeText    ReplyMode = "text"
    ReplyModeCommand ReplyMode = "command"
)

type GroupPolicy string

const (
    GroupPolicyOpen      GroupPolicy = "open"
    GroupPolicyDisabled  GroupPolicy = "disabled"
    GroupPolicyAllowlist GroupPolicy = "allowlist"
)

type DmPolicy string

const (
    DmPolicyPairing   DmPolicy = "pairing"
    DmPolicyAllowlist DmPolicy = "allowlist"
    DmPolicyOpen      DmPolicy = "open"
    DmPolicyDisabled  DmPolicy = "disabled"
)

type ChannelsConfig struct {
    Discord    *DiscordConfig    `json:"discord,omitempty"`
    Telegram   *TelegramConfig   `json:"telegram,omitempty"`
    Slack      *SlackConfig      `json:"slack,omitempty"`
    Signal     *SignalConfig     `json:"signal,omitempty"`
    IMessage   *IMessageConfig   `json:"imessage,omitempty"`
    WhatsApp   *WhatsAppConfig   `json:"whatsapp,omitempty"`
    Line       *LineConfig       `json:"line,omitempty"`
    IRC        *IRCConfig        `json:"irc,omitempty"`
    GoogleChat *GoogleChatConfig `json:"googlechat,omitempty"`
    MSTeams    *MSTeamsConfig    `json:"msteams,omitempty"`
}

type DiscordConfig struct {
    BotToken       string      `json:"botToken,omitempty"`
    GroupPolicy    GroupPolicy `json:"groupPolicy,omitempty"`
    DmPolicy       DmPolicy    `json:"dmPolicy,omitempty"`
    AllowFrom      []string    `json:"allowFrom,omitempty"`
    MentionPatterns []string   `json:"mentionPatterns,omitempty"`
    Presence       *DiscordPresenceConfig `json:"presence,omitempty"`
}

type TelegramConfig struct {
    BotToken        string      `json:"botToken,omitempty"`
    GroupPolicy     GroupPolicy `json:"groupPolicy,omitempty"`
    DmPolicy        DmPolicy    `json:"dmPolicy,omitempty"`
    AllowFrom       []string    `json:"allowFrom,omitempty"`
    MentionPatterns []string    `json:"mentionPatterns,omitempty"`
    CustomCommands  []TelegramCustomCommand `json:"customCommands,omitempty"`
    WebhookSecret   string      `json:"webhookSecret,omitempty"`
}

type SlackConfig struct {
    BotToken    string      `json:"botToken,omitempty"`
    AppToken    string      `json:"appToken,omitempty"`
    GroupPolicy GroupPolicy `json:"groupPolicy,omitempty"`
    DmPolicy    DmPolicy    `json:"dmPolicy,omitempty"`
    AllowFrom   []string    `json:"allowFrom,omitempty"`
}

type WhatsAppConfig struct {
    GroupPolicy     GroupPolicy `json:"groupPolicy,omitempty"`
    DmPolicy        DmPolicy    `json:"dmPolicy,omitempty"`
    AllowFrom       []string    `json:"allowFrom,omitempty"`
    MentionPatterns []string    `json:"mentionPatterns,omitempty"`
}
```

### 6.5 Cron Types

```go
// internal/cron/types.go

type CronJob struct {
    ID          string            `json:"id"`
    Name        string            `json:"name,omitempty"`
    Schedule    string            `json:"schedule"`
    AgentID     string            `json:"agentId,omitempty"`
    Text        string            `json:"text,omitempty"`
    Enabled     bool              `json:"enabled"`
    OneShot     bool              `json:"oneShot,omitempty"`
    DeliveryTargets []DeliveryTarget `json:"deliveryTargets,omitempty"`
    SystemEvent string            `json:"systemEvent,omitempty"`
    LastRun     *time.Time        `json:"lastRun,omitempty"`
    NextRun     *time.Time        `json:"nextRun,omitempty"`
    RunCount    int               `json:"runCount,omitempty"`
    CreatedAt   time.Time         `json:"createdAt"`
    UpdatedAt   time.Time         `json:"updatedAt"`
}

type DeliveryTarget struct {
    Channel   string `json:"channel"`
    AccountID string `json:"accountId,omitempty"`
    PeerID    string `json:"peerId,omitempty"`
    ThreadID  string `json:"threadId,omitempty"`
}

type CronRunLog struct {
    JobID     string    `json:"jobId"`
    RunAt     time.Time `json:"runAt"`
    Duration  int64     `json:"durationMs"`
    Status    string    `json:"status"`
    Error     string    `json:"error,omitempty"`
    AgentText string    `json:"agentText,omitempty"`
}
```

### 6.6 Gateway Protocol Types

```go
// internal/gateway/protocol/schema/types.go

type ChatEvent struct {
    Type      string          `json:"type"`
    SessionKey string         `json:"sessionKey,omitempty"`
    AgentID   string          `json:"agentId,omitempty"`
    Text      string          `json:"text,omitempty"`
    Role      string          `json:"role,omitempty"`
    Metadata  json.RawMessage `json:"metadata,omitempty"`
    Timestamp time.Time       `json:"timestamp"`
}

type AgentEvent struct {
    Type      string          `json:"type"`
    AgentID   string          `json:"agentId"`
    SessionKey string         `json:"sessionKey,omitempty"`
    Data      json.RawMessage `json:"data,omitempty"`
    Timestamp time.Time       `json:"timestamp"`
}

type AgentSummary struct {
    ID        string `json:"id"`
    Name      string `json:"name,omitempty"`
    IsDefault bool   `json:"isDefault,omitempty"`
    Model     string `json:"model,omitempty"`
    Status    string `json:"status,omitempty"`
}

type SessionSummary struct {
    Key        string    `json:"key"`
    AgentID    string    `json:"agentId,omitempty"`
    Channel    string    `json:"channel,omitempty"`
    ChatType   ChatType  `json:"chatType,omitempty"`
    CreatedAt  time.Time `json:"createdAt"`
    UpdatedAt  time.Time `json:"updatedAt"`
    TurnCount  int       `json:"turnCount"`
    TokensUsed int64     `json:"tokensUsed,omitempty"`
}

type HealthResult struct {
    Status    string            `json:"status"`
    Version   string            `json:"version"`
    Uptime    int64             `json:"uptimeMs"`
    Channels  map[string]string `json:"channels,omitempty"`
    Agents    []AgentSummary    `json:"agents,omitempty"`
    Memory    *MemoryHealth     `json:"memory,omitempty"`
}

type MemoryHealth struct {
    Backend    string `json:"backend"`
    Documents  int    `json:"documents"`
    Embeddings int    `json:"embeddings"`
}

// WebSocket frame types
type WSFrame struct {
    ID     string          `json:"id,omitempty"`
    Method string          `json:"method"`
    Params json.RawMessage `json:"params,omitempty"`
}

type WSResponse struct {
    ID     string          `json:"id,omitempty"`
    Result json.RawMessage `json:"result,omitempty"`
    Error  *WSError        `json:"error,omitempty"`
}

type WSError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}

type WSEvent struct {
    Event string          `json:"event"`
    Data  json.RawMessage `json:"data,omitempty"`
}
```

### 6.7 Memory/Embeddings Types

```go
// internal/memory/types.go

type EmbeddingProvider string

const (
    EmbeddingOpenAI  EmbeddingProvider = "openai"
    EmbeddingGemini  EmbeddingProvider = "gemini"
    EmbeddingVoyage  EmbeddingProvider = "voyage"
    EmbeddingOllama  EmbeddingProvider = "ollama"
    EmbeddingLlama   EmbeddingProvider = "llama"
)

type MemoryDocument struct {
    ID        string            `json:"id"`
    Content   string            `json:"content"`
    Metadata  map[string]string `json:"metadata,omitempty"`
    Embedding []float32         `json:"embedding,omitempty"`
    CreatedAt time.Time         `json:"createdAt"`
    UpdatedAt time.Time         `json:"updatedAt"`
}

type SearchResult struct {
    Document   MemoryDocument `json:"document"`
    Score      float64        `json:"score"`
    Highlights []string       `json:"highlights,omitempty"`
}

type SearchQuery struct {
    Text      string            `json:"text"`
    Limit     int               `json:"limit,omitempty"`
    Threshold float64           `json:"threshold,omitempty"`
    Filters   map[string]string `json:"filters,omitempty"`
    AgentID   string            `json:"agentId,omitempty"`
    Scope     string            `json:"scope,omitempty"`
}

type EmbeddingBatch struct {
    Provider EmbeddingProvider `json:"provider"`
    Model    string            `json:"model"`
    Inputs   []string          `json:"inputs"`
}

type EmbeddingResult struct {
    Vectors    [][]float32 `json:"vectors"`
    TokensUsed int         `json:"tokensUsed"`
}
```

### 6.8 Plugin Types

```go
// internal/plugins/types.go

type PluginManifest struct {
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Description string            `json:"description,omitempty"`
    Author      string            `json:"author,omitempty"`
    Entry       string            `json:"entry"`
    Hooks       []string          `json:"hooks,omitempty"`
    Tools       []PluginTool      `json:"tools,omitempty"`
    Config      map[string]any    `json:"config,omitempty"`
    Channels    []string          `json:"channels,omitempty"`
}

type PluginTool struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Parameters  map[string]any `json:"parameters,omitempty"`
}

type PluginInstance struct {
    Manifest  PluginManifest `json:"manifest"`
    Enabled   bool           `json:"enabled"`
    Path      string         `json:"path"`
    Status    string         `json:"status"`
}
```

### 6.9 Hook Types

```go
// internal/hooks/types.go

type HookType string

const (
    HookBeforeToolCall  HookType = "before-tool-call"
    HookAfterToolCall   HookType = "after-tool-call"
    HookOnMessage       HookType = "on-message"
    HookOnSession       HookType = "on-session"
    HookOnCompaction    HookType = "on-compaction"
    HookOnGatewayStart  HookType = "on-gateway-start"
    HookOnBoot          HookType = "on-boot"
)

type HookDefinition struct {
    Name     string   `json:"name"`
    Type     HookType `json:"type"`
    Path     string   `json:"path"`
    Priority int      `json:"priority,omitempty"`
}

type HookContext struct {
    AgentID    string          `json:"agentId,omitempty"`
    SessionKey string          `json:"sessionKey,omitempty"`
    Channel    string          `json:"channel,omitempty"`
    Data       json.RawMessage `json:"data,omitempty"`
}

type HookResult struct {
    Modified bool            `json:"modified"`
    Data     json.RawMessage `json:"data,omitempty"`
    Error    string          `json:"error,omitempty"`
}
```

---

## 7. API Layer Rewrite Plan

### 7.1 WebSocket RPC Protocol

The gateway exposes a JSON-RPC-like WebSocket protocol. In Go:

```go
// internal/gateway/server/ws_handler.go

type RPCHandler func(ctx context.Context, conn *WSConn, params json.RawMessage) (any, error)

type RPCRouter struct {
    mu       sync.RWMutex
    handlers map[string]RPCHandler
}

func (r *RPCRouter) Register(method string, handler RPCHandler) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.handlers[method] = handler
}

func (r *RPCRouter) Handle(ctx context.Context, conn *WSConn, frame WSFrame) {
    r.mu.RLock()
    handler, ok := r.handlers[frame.Method]
    r.mu.RUnlock()
    if !ok {
        conn.SendError(frame.ID, ErrMethodNotFound)
        return
    }
    result, err := handler(ctx, conn, frame.Params)
    if err != nil {
        conn.SendError(frame.ID, err)
        return
    }
    conn.SendResult(frame.ID, result)
}
```

### 7.2 RPC Methods to Implement

All methods from `src/gateway/server-methods/`:

| Method Group | Methods | Go Handler File |
|---|---|---|
| Agent | `agent.get`, `agent.wait`, `agent.identity`, `agent.event` | `methods/agent.go` |
| Agents | `agents.list`, `agents.create`, `agents.update`, `agents.delete`, `agents.files.*` | `methods/agents.go` |
| Browser | `browser.*` | `methods/browser.go` |
| Channels | `channels.status`, `channels.logout` | `methods/channels.go` |
| Chat | `chat.send`, `chat.inject`, `chat.abort`, `chat.history` | `methods/chat.go` |
| Config | `config.get`, `config.patch`, `config.apply` | `methods/config.go` |
| Connect | `connect` | `methods/connect.go` |
| Cron | `cron.list`, `cron.add`, `cron.edit`, `cron.delete`, `cron.run` | `methods/cron.go` |
| Devices | `devices.list`, `devices.approve`, `devices.revoke` | `methods/devices.go` |
| Exec Approval | `exec.approve`, `exec.deny`, `exec.list` | `methods/exec_approval.go` |
| Health | `health` | `methods/health.go` |
| Logs | `logs.chat`, `logs.system` | `methods/logs.go` |
| Models | `models.list`, `models.catalog` | `methods/models.go` |
| Nodes | `nodes.list`, `nodes.invoke`, `nodes.pair` | `methods/nodes.go` |
| Send | `send` | `methods/send.go` |
| Sessions | `sessions.list`, `sessions.get`, `sessions.delete`, `sessions.patch`, `sessions.usage` | `methods/sessions.go` |
| Skills | `skills.list`, `skills.update` | `methods/skills.go` |
| System | `system.info`, `system.restart`, `system.shutdown` | `methods/system.go` |
| Talk | `talk.config` | `methods/talk.go` |
| TTS | `tts.synthesize` | `methods/tts.go` |
| Usage | `usage.summary`, `usage.sessions` | `methods/usage.go` |
| VoiceWake | `voicewake.*` | `methods/voicewake.go` |
| Web | `web.*` | `methods/web.go` |
| Wizard | `wizard.*` | `methods/wizard.go` |

### 7.3 HTTP API Endpoints

The gateway also exposes REST-like HTTP endpoints:

| Endpoint | Purpose | Go Implementation |
|---|---|---|
| `GET /health` | Health check | `chi` route |
| `POST /v1/chat/completions` | OpenAI-compatible chat API | `openresponses_http.go` |
| `POST /v1/responses` | OpenAI Responses API | `openresponses_http.go` |
| `GET /v1/models` | Model listing | `openresponses_http.go` |
| `POST /hooks/:name` | Webhook receiver | `server/hooks.go` |
| `POST /tools/invoke` | Tool invocation HTTP API | `tools_invoke_http.go` |
| `GET /` | Control UI (SPA) | Static file serving |
| Plugin HTTP routes | Per-plugin HTTP endpoints | `server/plugins_http.go` |

### 7.4 Authentication Middleware

```go
// internal/gateway/auth.go

type AuthMiddleware struct {
    config      *config.GatewayAuthConfig
    rateLimiter *RateLimiter
    deviceAuth  *DeviceAuthStore
}

func (a *AuthMiddleware) Authenticate(r *http.Request) (*AuthResult, error) {
    // Token from header, query param, or cookie
    // Rate limiting per IP
    // Device identity verification
}

func (a *AuthMiddleware) AuthenticateWS(r *http.Request) (*AuthResult, error) {
    // WebSocket upgrade authentication
    // Origin check
}
```

---

## 8. Concurrency and Performance

### 8.1 Goroutine Architecture

```
main goroutine
├── Gateway HTTP/WS Server (net/http.Server)
│   ├── Per-connection goroutine (WebSocket read loop)
│   ├── Per-connection goroutine (WebSocket write loop)
│   └── Per-request goroutine (HTTP handlers)
├── Channel Monitors (one goroutine per active channel)
│   ├── Discord monitor goroutine
│   ├── Telegram polling/webhook goroutine
│   ├── Slack Socket Mode goroutine
│   ├── Signal SSE goroutine
│   ├── WhatsApp Web goroutine
│   ├── LINE webhook goroutine
│   └── iMessage monitor goroutine
├── Cron Service goroutine
│   ├── Timer management goroutine
│   └── Per-job execution goroutines (bounded by semaphore)
├── Heartbeat Runner goroutine
├── Auto-Reply Queue goroutines (worker pool)
│   └── Per-reply agent execution goroutine
├── Memory Manager goroutine
│   ├── Embedding batch goroutine
│   └── Sync goroutine
├── Plugin Service goroutines (one per plugin)
├── mDNS Discovery goroutine
├── Update Check goroutine
├── Config Reload watcher goroutine
├── Media Server goroutine
├── Canvas Host goroutine
└── Browser Control Server goroutine
```

### 8.2 Channel Patterns

```go
// Reply queue with bounded concurrency
type ReplyQueue struct {
    ch        chan ReplyRequest
    workers   int
    wg        sync.WaitGroup
    ctx       context.Context
    cancel    context.CancelFunc
}

func NewReplyQueue(workers int) *ReplyQueue {
    ctx, cancel := context.WithCancel(context.Background())
    q := &ReplyQueue{
        ch:      make(chan ReplyRequest, 1000),
        workers: workers,
        ctx:     ctx,
        cancel:  cancel,
    }
    for i := 0; i < workers; i++ {
        q.wg.Add(1)
        go q.worker()
    }
    return q
}

func (q *ReplyQueue) worker() {
    defer q.wg.Done()
    for {
        select {
        case <-q.ctx.Done():
            return
        case req := <-q.ch:
            q.processReply(req)
        }
    }
}
```

### 8.3 Event Broadcasting

```go
// Pub/sub for gateway events
type EventBus struct {
    mu          sync.RWMutex
    subscribers map[string][]chan Event
}

func (eb *EventBus) Subscribe(event string) <-chan Event {
    ch := make(chan Event, 100)
    eb.mu.Lock()
    eb.subscribers[event] = append(eb.subscribers[event], ch)
    eb.mu.Unlock()
    return ch
}

func (eb *EventBus) Publish(event string, data Event) {
    eb.mu.RLock()
    subs := eb.subscribers[event]
    eb.mu.RUnlock()
    for _, ch := range subs {
        select {
        case ch <- data:
        default:
            // Drop if subscriber is slow
        }
    }
}
```

### 8.4 Concurrency Lanes

The original uses "lanes" for concurrency control per agent/session:

```go
// internal/process/lanes.go

type LaneManager struct {
    mu    sync.Mutex
    lanes map[string]*Lane
}

type Lane struct {
    sem     chan struct{} // Semaphore
    pending int64
}

func (lm *LaneManager) Acquire(ctx context.Context, key string) error {
    lm.mu.Lock()
    lane, ok := lm.lanes[key]
    if !ok {
        lane = &Lane{sem: make(chan struct{}, 1)}
        lm.lanes[key] = lane
    }
    atomic.AddInt64(&lane.pending, 1)
    lm.mu.Unlock()

    select {
    case lane.sem <- struct{}{}:
        return nil
    case <-ctx.Done():
        atomic.AddInt64(&lane.pending, -1)
        return ctx.Err()
    }
}

func (lm *LaneManager) Release(key string) {
    lm.mu.Lock()
    lane := lm.lanes[key]
    lm.mu.Unlock()
    atomic.AddInt64(&lane.pending, -1)
    <-lane.sem
}
```

### 8.5 Graceful Shutdown

```go
// cmd/openclaw/main.go

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    server := gateway.NewServer(cfg)
    
    g, gCtx := errgroup.WithContext(ctx)
    
    g.Go(func() error { return server.Start(gCtx) })
    g.Go(func() error { return cronService.Run(gCtx) })
    g.Go(func() error { return channelManager.Run(gCtx) })
    // ... other services
    
    g.Go(func() error {
        <-gCtx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        return server.Shutdown(shutdownCtx)
    })
    
    if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
        log.Fatal(err)
    }
}
```

---

## 9. Configuration Management

### 9.1 Config Loading Order

1. Default values (compiled into binary)
2. `~/.openclaw/openclaw.json` (JSON5 format)
3. `./.env` → `~/.openclaw/.env` (dotenv)
4. `openclaw.json` `env` block
5. Process environment variables (highest priority)

### 9.2 Go Implementation

```go
// internal/config/config.go

type ConfigLoader struct {
    paths      ConfigPaths
    envLoader  *EnvLoader
    validator  *ConfigValidator
    watcher    *fsnotify.Watcher
    onChange   []func(old, new *OpenClawConfig)
}

func (cl *ConfigLoader) Load() (*OpenClawConfig, error) {
    // 1. Load defaults
    cfg := DefaultConfig()
    
    // 2. Load config file (JSON5)
    if data, err := os.ReadFile(cl.paths.ConfigPath); err == nil {
        if err := json5.Unmarshal(data, cfg); err != nil {
            return nil, fmt.Errorf("parse config: %w", err)
        }
    }
    
    // 3. Process includes
    if err := cl.processIncludes(cfg); err != nil {
        return nil, err
    }
    
    // 4. Load env files
    cl.envLoader.Load()
    
    // 5. Apply env substitution
    cl.applyEnvSubstitution(cfg)
    
    // 6. Validate
    if err := cl.validator.Validate(cfg); err != nil {
        return nil, err
    }
    
    // 7. Apply legacy migrations
    cl.migrateLegacy(cfg)
    
    return cfg, nil
}

func (cl *ConfigLoader) Watch(ctx context.Context) error {
    // Watch config file for changes using fsnotify
    for {
        select {
        case <-ctx.Done():
            return nil
        case event := <-cl.watcher.Events:
            if event.Op&fsnotify.Write != 0 {
                newCfg, err := cl.Load()
                if err != nil {
                    log.Warn().Err(err).Msg("config reload failed")
                    continue
                }
                for _, fn := range cl.onChange {
                    fn(cl.current, newCfg)
                }
                cl.current = newCfg
            }
        }
    }
}
```

### 9.3 Config Paths

```go
// internal/config/paths.go

type ConfigPaths struct {
    StateDir   string // ~/.openclaw
    ConfigPath string // ~/.openclaw/openclaw.json
    HomeDir    string // ~
    DataDir    string // ~/.openclaw/data
    LogDir     string // ~/.openclaw/logs
    PluginDir  string // ~/.openclaw/plugins
    SkillsDir  string // ~/.openclaw/skills
    SessionDir string // ~/.openclaw/sessions
}

func ResolveConfigPaths() ConfigPaths {
    stateDir := os.Getenv("OPENCLAW_STATE_DIR")
    if stateDir == "" {
        home, _ := os.UserHomeDir()
        stateDir = filepath.Join(home, ".openclaw")
    }
    return ConfigPaths{
        StateDir:   stateDir,
        ConfigPath: envOrDefault("OPENCLAW_CONFIG_PATH", filepath.Join(stateDir, "openclaw.json")),
        HomeDir:    envOrDefault("OPENCLAW_HOME", mustUserHomeDir()),
        DataDir:    filepath.Join(stateDir, "data"),
        LogDir:     filepath.Join(stateDir, "logs"),
        PluginDir:  filepath.Join(stateDir, "plugins"),
        SkillsDir:  filepath.Join(stateDir, "skills"),
        SessionDir: filepath.Join(stateDir, "sessions"),
    }
}
```

---

## 10. Error Handling Strategy

### 10.1 Error Types

```go
// internal/infra/errors.go

// Sentinel errors
var (
    ErrNotFound         = errors.New("not found")
    ErrUnauthorized     = errors.New("unauthorized")
    ErrForbidden        = errors.New("forbidden")
    ErrRateLimited      = errors.New("rate limited")
    ErrConfigInvalid    = errors.New("invalid configuration")
    ErrSessionNotFound  = errors.New("session not found")
    ErrAgentNotFound    = errors.New("agent not found")
    ErrChannelOffline   = errors.New("channel offline")
    ErrProviderError    = errors.New("provider error")
    ErrContextOverflow  = errors.New("context window overflow")
    ErrBillingError     = errors.New("billing error")
    ErrAuthError        = errors.New("authentication error")
    ErrSandboxError     = errors.New("sandbox error")
    ErrPluginError      = errors.New("plugin error")
    ErrTimeout          = errors.New("timeout")
)

// Structured error with code
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Cause   error  `json:"-"`
    Details any    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Cause
}

// Error codes matching the original protocol
const (
    CodeParseError     = -32700
    CodeInvalidRequest = -32600
    CodeMethodNotFound = -32601
    CodeInvalidParams  = -32602
    CodeInternalError  = -32603
    CodeAuthError      = -32000
    CodeRateLimited    = -32001
    CodeNotFound       = -32002
    CodeForbidden      = -32003
)
```

### 10.2 Error Wrapping Pattern

```go
// Consistent error wrapping throughout the codebase
func (s *SessionStore) Get(key string) (*Session, error) {
    data, err := os.ReadFile(s.pathFor(key))
    if err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("session %q: %w", key, ErrSessionNotFound)
        }
        return nil, fmt.Errorf("read session %q: %w", key, err)
    }
    var session Session
    if err := json.Unmarshal(data, &session); err != nil {
        return nil, fmt.Errorf("parse session %q: %w", key, err)
    }
    return &session, nil
}
```

### 10.3 Recovery Middleware

```go
// internal/gateway/server/recovery.go

func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                stack := debug.Stack()
                log.Error().
                    Interface("panic", rec).
                    Str("stack", string(stack)).
                    Str("path", r.URL.Path).
                    Msg("panic recovered")
                http.Error(w, "internal server error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

---

## 11. Testing Strategy

### 11.1 Test Organization

```
goclaw/
├── internal/
│   ├── gateway/
│   │   ├── server_test.go          # Unit tests
│   │   ├── server_e2e_test.go      # E2E tests (build tag: e2e)
│   │   └── server_live_test.go     # Live tests (build tag: live)
│   ├── config/
│   │   ├── config_test.go
│   │   └── schema_test.go
│   └── ...
├── test/
│   ├── integration/                 # Cross-module integration tests
│   ├── fixtures/                    # Test data
│   └── helpers/                     # Shared test utilities
```

### 11.2 Test Categories

| Category | Build Tag | Original Equivalent | Description |
|---|---|---|---|
| Unit | (none) | `vitest run` | Fast, isolated tests |
| E2E | `e2e` | `vitest.e2e.config.ts` | End-to-end with real gateway |
| Live | `live` | `vitest.live.config.ts` | Tests against live AI providers |
| Gateway | `gateway` | `vitest.gateway.config.ts` | Gateway-specific integration |
| Extension | `extension` | `vitest.extensions.config.ts` | Extension tests |

### 11.3 Test Helpers

```go
// test/helpers/gateway.go

type TestGateway struct {
    Server  *gateway.Server
    URL     string
    WSConn  *websocket.Conn
    Token   string
    cleanup func()
}

func NewTestGateway(t *testing.T, opts ...TestOption) *TestGateway {
    t.Helper()
    cfg := testConfig(opts...)
    srv := gateway.NewServer(cfg)
    
    listener, err := net.Listen("tcp", "127.0.0.1:0")
    require.NoError(t, err)
    
    go srv.Serve(listener)
    
    t.Cleanup(func() {
        srv.Shutdown(context.Background())
        listener.Close()
    })
    
    return &TestGateway{
        Server: srv,
        URL:    fmt.Sprintf("http://%s", listener.Addr()),
        Token:  cfg.Gateway.Auth.Token,
    }
}

func (tg *TestGateway) ConnectWS(t *testing.T) *websocket.Conn {
    t.Helper()
    wsURL := strings.Replace(tg.URL, "http://", "ws://", 1) + "/ws"
    header := http.Header{"Authorization": []string{"Bearer " + tg.Token}}
    conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
    require.NoError(t, err)
    t.Cleanup(func() { conn.Close() })
    return conn
}
```

### 11.4 Mocking Strategy

```go
// Use interfaces for all external dependencies
type AIProvider interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    Stream(ctx context.Context, req ChatRequest) (<-chan ChatEvent, error)
}

type ChannelSender interface {
    Send(ctx context.Context, target Target, message Message) error
}

// Mock implementations for testing
type MockAIProvider struct {
    mock.Mock
}

func (m *MockAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*ChatResponse), args.Error(1)
}
```

### 11.5 Benchmark Tests

```go
// internal/gateway/server_bench_test.go

func BenchmarkWSMessageRouting(b *testing.B) {
    gw := setupBenchGateway(b)
    conn := gw.ConnectWS(b)
    
    msg := WSFrame{Method: "chat.send", Params: json.RawMessage(`{"text":"hello"}`)}
    data, _ := json.Marshal(msg)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        conn.WriteMessage(websocket.TextMessage, data)
        _, _, _ = conn.ReadMessage()
    }
}
```

---

## 12. Migration Phases with Milestones

### Phase 1: Foundation (Weeks 1–4)

**Goal**: Core infrastructure, config, and CLI skeleton

| Milestone | Deliverable | Files |
|---|---|---|
| 1.1 | Go module init, project structure | `go.mod`, directory tree |
| 1.2 | Configuration system | `internal/config/` (all files) |
| 1.3 | Logging system | `internal/logging/` |
| 1.4 | Infrastructure utilities | `internal/infra/` (env, paths, errors, retry, fs) |
| 1.5 | CLI skeleton (cobra) | `cmd/openclaw/`, `internal/cli/program/` |
| 1.6 | Basic `openclaw --version`, `openclaw doctor` | CLI commands |

**Exit Criteria**: `go build` produces a binary that loads config and runs `doctor`.

### Phase 2: Gateway Core (Weeks 5–10)

**Goal**: WebSocket gateway with auth, health, and basic RPC

| Milestone | Deliverable | Files |
|---|---|---|
| 2.1 | HTTP server with chi router | `internal/gateway/server/` |
| 2.2 | WebSocket handler + frame parsing | `internal/gateway/server/ws_connection.go` |
| 2.3 | Authentication (token/password) | `internal/gateway/auth.go`, `auth_rate_limit.go` |
| 2.4 | RPC router + method dispatch | `internal/gateway/protocol/` |
| 2.5 | Health endpoint | `internal/gateway/methods/health.go` |
| 2.6 | Config methods | `internal/gateway/methods/config.go` |
| 2.7 | Session store (file-based) | `internal/config/sessions/` |
| 2.8 | Session methods | `internal/gateway/methods/sessions.go` |
| 2.9 | TLS support | `internal/gateway/server/tls.go` |
| 2.10 | Origin check + CORS | `internal/gateway/origin_check.go` |

**Exit Criteria**: Gateway starts, accepts WS connections, authenticates, responds to health/config/session RPCs.

### Phase 3: Agent Engine (Weeks 11–18)

**Goal**: AI agent execution with model providers

| Milestone | Deliverable | Files |
|---|---|---|
| 3.1 | Model catalog + selection | `internal/agents/model_catalog.go`, `model_selection.go` |
| 3.2 | Auth profiles | `internal/agents/authprofiles/` |
| 3.3 | Agent execution engine | `internal/agents/runner/` |
| 3.4 | Streaming response handling | `internal/agents/runner/subscribe.go` |
| 3.5 | System prompt builder | `internal/agents/system_prompt.go` |
| 3.6 | Basic tools (web_search, web_fetch) | `internal/agents/tools/web_*.go` |
| 3.7 | Chat RPC methods | `internal/gateway/methods/chat.go` |
| 3.8 | Agent RPC methods | `internal/gateway/methods/agent.go`, `agents.go` |
| 3.9 | Model fallback chain | `internal/agents/model_fallback.go` |
| 3.10 | Context pruning + compaction | `internal/agents/extensions/` |

**Exit Criteria**: Can send a chat message via WS, get an AI response streamed back.

### Phase 4: Auto-Reply Pipeline (Weeks 19–24)

**Goal**: Full message processing pipeline

| Milestone | Deliverable | Files |
|---|---|---|
| 4.1 | Reply queue + worker pool | `internal/autoreply/queue/` |
| 4.2 | Command registry + parsing | `internal/autoreply/commands_registry.go` |
| 4.3 | Directive handling | `internal/autoreply/reply/directive.go` |
| 4.4 | Block streaming | `internal/autoreply/reply/block_streaming.go` |
| 4.5 | Session management | `internal/autoreply/reply/session.go` |
| 4.6 | Typing indicators | `internal/autoreply/reply/typing.go` |
| 4.7 | Memory flush | `internal/autoreply/reply/memory_flush.go` |
| 4.8 | Heartbeat system | `internal/autoreply/heartbeat.go`, `internal/infra/heartbeat_runner.go` |

**Exit Criteria**: Full reply pipeline works end-to-end with commands, directives, streaming.

### Phase 5: Channel Integrations (Weeks 25–36)

**Goal**: All messaging channels operational

| Milestone | Deliverable | Priority |
|---|---|---|
| 5.1 | Channel abstraction layer | `internal/channels/` | P0 |
| 5.2 | Discord integration | `internal/discord/` | P0 |
| 5.3 | Telegram integration | `internal/telegram/` | P0 |
| 5.4 | Slack integration | `internal/slack/` | P0 |
| 5.5 | WhatsApp Web integration | `internal/web/` | P1 |
| 5.6 | Signal integration | `internal/signal/` | P1 |
| 5.7 | iMessage integration | `internal/imessage/` | P2 |
| 5.8 | LINE integration | `internal/line/` | P2 |
| 5.9 | Web channel (Control UI chat) | `internal/channels/web/` | P0 |
| 5.10 | Routing + session key resolution | `internal/routing/` | P0 |

**Exit Criteria**: All P0 channels send/receive messages with full feature parity.

### Phase 6: Advanced Features (Weeks 37–44)

**Goal**: Cron, plugins, hooks, memory, browser, tools

| Milestone | Deliverable | Files |
|---|---|---|
| 6.1 | Cron service | `internal/cron/` |
| 6.2 | Plugin system | `internal/plugins/` |
| 6.3 | Hook system | `internal/hooks/` |
| 6.4 | Memory/embeddings | `internal/memory/` |
| 6.5 | Browser automation | `internal/browser/` |
| 6.6 | All agent tools (30+) | `internal/agents/tools/` |
| 6.7 | Media server | `internal/media/` |
| 6.8 | TTS | `internal/tts/` |
| 6.9 | Canvas host | `internal/canvashost/` |
| 6.10 | Node host | `internal/nodehost/` |

**Exit Criteria**: Feature parity with TypeScript version for all advanced features.

### Phase 7: CLI & Operations (Weeks 45–48)

**Goal**: Full CLI, daemon management, updates

| Milestone | Deliverable | Files |
|---|---|---|
| 7.1 | All CLI commands | `internal/cli/` |
| 7.2 | Daemon management | `internal/daemon/` |
| 7.3 | mDNS discovery | `internal/infra/bonjour.go` |
| 7.4 | Device pairing | `internal/pairing/` |
| 7.5 | Self-update mechanism | `internal/cli/update/` |
| 7.6 | Onboarding wizard | `internal/wizard/` |
| 7.7 | TUI | `internal/tui/` |
| 7.8 | Security audit | `internal/security/` |

**Exit Criteria**: Full CLI feature parity, daemon install/uninstall works on all platforms.

### Phase 8: Extensions & Polish (Weeks 49–52)

**Goal**: Extensions, OpenAI-compatible API, performance tuning

| Milestone | Deliverable | Files |
|---|---|---|
| 8.1 | Plugin SDK (public) | `pkg/pluginsdk/` |
| 8.2 | Extension API | `pkg/extensionapi/` |
| 8.3 | Matrix extension | `internal/extensions/matrix/` |
| 8.4 | Mattermost extension | `internal/extensions/mattermost/` |
| 8.5 | MS Teams extension | `internal/extensions/msteams/` |
| 8.6 | OpenAI-compatible HTTP API | `internal/gateway/openresponses_http.go` |
| 8.7 | ACP server | `internal/acp/` |
| 8.8 | Performance benchmarks + tuning | Benchmarks |
| 8.9 | Documentation | `docs/` |
| 8.10 | Release packaging (Docker, installers) | `Dockerfile`, scripts |

**Exit Criteria**: Full feature parity, all tests passing, performance meets targets.

---

## 13. Risk Assessment

### 13.1 High Risk

| Risk | Impact | Mitigation |
|---|---|---|
| **Agent execution engine reimplementation** | The `@mariozechner/pi-*` packages are complex AI orchestration libraries. Reimplementing in Go is the single largest effort. | Start with a simplified agent loop; iterate. Consider wrapping the Node.js agent core via subprocess initially. |
| **WhatsApp Web protocol** | `@whiskeysockets/baileys` implements the WhatsApp Web protocol (reverse-engineered). No mature Go equivalent exists. | Evaluate `whatsmeow` (Go WhatsApp library) or maintain a Node.js sidecar for WhatsApp. |
| **Browser automation** | Playwright has no Go equivalent with the same feature set. `chromedp` is CDP-only (no Firefox/WebKit). | Use `chromedp` for Chrome; accept reduced browser coverage. Or use `rod` (Go browser automation). |
| **Plugin ecosystem compatibility** | Existing TypeScript plugins won't work in Go. | Provide a Go plugin SDK + support running TS plugins via subprocess bridge. |

### 13.2 Medium Risk

| Risk | Impact | Mitigation |
|---|---|---|
| **Protocol compatibility** | Native apps (iOS, Android, macOS) depend on the exact WebSocket protocol. Any deviation breaks them. | Generate protocol schemas from a shared source; comprehensive protocol conformance tests. |
| **Config format compatibility** | Existing `openclaw.json` files must work unchanged. | Extensive config migration tests; JSON5 parser compatibility. |
| **SQLite + sqlite-vec** | CGo dependency for SQLite vector search. Complicates cross-compilation. | Use `modernc.org/sqlite` (pure Go) where possible; CGo only for sqlite-vec. |
| **Edge TTS / ElevenLabs** | TTS APIs are HTTP-based but have streaming requirements. | Straightforward HTTP client implementation; test with real APIs. |
| **Signal CLI integration** | Signal uses a Java-based CLI daemon. Integration is via SSE/HTTP. | HTTP client integration is straightforward in Go. |

### 13.3 Low Risk

| Risk | Impact | Mitigation |
|---|---|---|
| **Discord/Telegram/Slack APIs** | Mature Go libraries exist for all three. | Use well-maintained Go SDKs. |
| **Cron scheduling** | `robfig/cron` is battle-tested. | Direct mapping from `croner`. |
| **Configuration loading** | Go has excellent JSON/YAML/env support. | Standard libraries + `godotenv`. |
| **HTTP/WebSocket server** | Go's `net/http` + `gorilla/websocket` are production-grade. | Standard approach. |
| **CLI framework** | `cobra` is the de facto Go CLI framework. | Direct mapping from `commander`. |
| **File-based session store** | Simple file I/O with locking. | `gofrs/flock` for cross-platform locking. |

### 13.4 Performance Targets

| Metric | TypeScript (Current) | Go (Target) |
|---|---|---|
| Cold start | 500ms–2s | <50ms |
| Memory (idle) | 150–400 MB | 30–80 MB |
| WS message latency (p99) | 5–15ms | 1–3ms |
| Concurrent WS connections | ~1,000 | ~10,000 |
| Chat throughput | ~100 req/s | ~1,000 req/s |
| Binary size | ~50 MB (bundled) | ~30 MB (static) |

### 13.5 Compatibility Matrix

| Component | Backward Compatible | Notes |
|---|---|---|
| `openclaw.json` config | ✅ Yes | Must parse identically |
| WebSocket protocol | ✅ Yes | Must be wire-compatible |
| HTTP API | ✅ Yes | Same endpoints, same responses |
| Session files | ✅ Yes | Same JSON format on disk |
| Plugin manifests | ✅ Yes | Same `openclaw.plugin.json` format |
| TypeScript plugins | ⚠️ Partial | Need subprocess bridge |
| CLI commands | ✅ Yes | Same command names and flags |
| Environment variables | ✅ Yes | Same `OPENCLAW_*` vars |
| Daemon integration | ✅ Yes | Same launchd/systemd/schtasks |
| Native app protocol | ✅ Yes | Same WebSocket frames |

---

*This document covers the complete Openclaw codebase: 1,754+ TypeScript source files across 50+ modules, mapped to a Go 1.26.0 project structure with 80+ Go packages, comprehensive data models, API specifications, concurrency patterns, and a 52-week migration plan.*
