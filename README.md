# VietClaw

VietClaw is a lightweight personal **agent framework** built for **prompt-first power**: you chat naturally and the agent handles memory, tools, delegation, and channels — no setup required. Advanced controls (providers, budget, channels, config) stay available when you need them.

Phase 1 built the core daemon foundation: local configuration, SQLite storage, logging, health/status endpoints, and a tiny HTML shell.

Phase 2 adds the minimal agent runtime: rule-based intent routing, SQLite memory, provider routing, mock provider, budget checks, context building, tool policy, chat API, and CLI memory/chat commands.

Phase 3 adds chat channel adapters for Discord and Telegram. Channel adapters only normalize inbound messages, apply channel rules, call the existing agent runtime, and send replies back.

Phase 4 adds a Nuxt static web UI that gets embedded into the Go binary. Node is only needed for development and CI builds, not for running VietClaw as an end user.

## Why Go + SQLite

Go keeps the runtime small, simple to deploy, and friendly to weak VPS machines with 1-2 CPU cores and 1-2GB RAM.

SQLite keeps Phase 1 local-first with no Redis, Postgres, Docker, or external queue service required.

## Run

```sh
go run ./cmd/vietclaw version
go run ./cmd/vietclaw init
go run ./cmd/vietclaw daemon
go run ./cmd/vietclaw status
go run ./cmd/vietclaw doctor
go run ./cmd/vietclaw chat "mày là gì"
go run ./cmd/vietclaw memory add "Minh thích tiết kiệm token"
go run ./cmd/vietclaw memory search "token"
go run ./cmd/vietclaw harness create "fix failing login test"
go run ./cmd/vietclaw harness run "fix failing login test"
go run ./cmd/vietclaw harness list
go run ./cmd/vietclaw harness show <run_id>
go run ./cmd/vietclaw harness diff <run_id>
go run ./cmd/vietclaw harness cleanup --passed
go run ./cmd/vietclaw channels
```

The daemon listens on `127.0.0.1:18636` by default.

## Phase 1 Includes

- CLI commands: `version`, `init`, `daemon`, `status`, `doctor`
- Local config in the VietClaw data directory
- SQLite database initialization
- File and stdout logging
- HTTP endpoints: `/`, `/health`, `/status`, `/logs/recent`
- Embedded minimal web shell

## Phase 2 Includes

- Agent runtime for local chat requests
- Rule-based intent router for `memory_add`, `memory_query`, `chat`, and `action`
- SQLite memory add/list/search
- Provider interface with mock, OpenAI-compatible HTTP, custom HTTP, and optional OpenCode CLI providers
- Context builder with explicit character/history limits
- Budget check from `cost_events`
- Tool policy foundation with shell disabled by default and file tools limited to the workspace
- HTTP APIs: `/api/chat`, `/api/memory`, `/api/memory/search`, `/api/sessions`, `/api/costs/today`, `/api/providers`
- CLI commands: `chat`, `memory list`, `memory add`, `memory search`

## Phase 3 Includes

- Discord adapter using `discordgo`
- Telegram adapter using long polling
- Shared channel policy, mention stripping, session key builder, and in-process idempotency guard
- CLI commands: `channels`, `discord enable`, `discord disable`, `telegram enable`, `telegram disable`
- HTTP APIs: `/api/channels`, `/api/channels/discord/test`, `/api/channels/telegram/test`
- Channel audit tables: `channel_messages`, `channel_events`

## Phase 4 Includes

- Nuxt static web app in `apps/web`
- Tailwind CSS v4 through the Vite plugin
- Embedded UI dist served by Go from `internal/web/dist`
- SPA fallback for `/chat`, `/memory`, `/providers`, `/budget`, `/logs`, `/channels`, and `/sessions`
- GitHub Actions CI for web build, Go tests, and Go binary build
- Tag-based release workflow with Linux, Windows, and macOS artifacts

## Discord

1. Create a Discord bot in the Discord Developer Portal.
2. Enable Message Content Intent if the bot should read normal guild messages.
3. Set the token in the environment:

```sh
set VIETCLAW_DISCORD_TOKEN=...
go run ./cmd/vietclaw discord enable
go run ./cmd/vietclaw daemon
```

In a Discord guild, VietClaw replies only when mentioned or when a user replies to the bot:

```text
@VietClaw deploy đi
```

In Discord DM, chat normally. No slash commands are registered or required.

## Telegram

1. Create a Telegram bot with BotFather.
2. Set the token in the environment:

```sh
set VIETCLAW_TELEGRAM_TOKEN=...
go run ./cmd/vietclaw telegram enable
go run ./cmd/vietclaw daemon
```

In a Telegram group, VietClaw replies only when mentioned or when a user replies to the bot:

```text
@your_bot_username hỏi gì đó
```

In private chat, chat normally. `/ask` or other command wrappers are not required.

Do not add VietClaw to an untrusted group if dangerous tools are enabled. `shell.exec` is disabled by default.

## Web UI Development

Run the backend:

```sh
go run ./cmd/vietclaw daemon
```

Run the Nuxt dev server only while developing UI:

```sh
cd apps/web
pnpm install
pnpm dev
```

Build static UI and copy it into the Go embed path:

```sh
cd apps/web
pnpm build
```

Then build the final binary:

```sh
go build ./cmd/vietclaw
```

The final binary serves the embedded UI and does not need Node, pnpm, or npm at runtime.

## Prompt-first design

VietClaw defaults to `agent.experience: "prompt"`:

- **You:** type natural language (Vietnamese or English).
- **Agent:** auto memory, web search, files, sub-agent delegation, reflexion on errors.
- **No setup:** providers/memory/tools work out of the box with mock or your configured API keys.
- **Advanced:** memory UI, providers, budget, channels, `config.json` — hidden until you open **Công cụ nâng cao** in the web UI.

`max_steps` defaults to `0` (unlimited) so long tool chains are not cut mid-task. Set a cap in config only if you want a safety fuse.

## Phase 6 — Agent Framework

VietClaw is now an agent framework, not just a runtime:

| Capability | Description |
| --- | --- |
| Multi-agent profiles | Persona, tools, providers, memory scope, max steps per agent |
| Sub-agent delegation | `agent_delegate` tool spawns child runs with `parent_run_id` trace |
| Lifecycle hooks | `before_chat`, `after_chat`, `before_tool`, `after_tool`, `run_start`, `run_finish` |
| Extension registries | Register tools, channel adapters, inspect via `vietclaw framework list` |
| Profile enforcement | Router and tool list respect per-agent `tools` and `providers` config |

```sh
go run ./cmd/vietclaw framework list
curl http://127.0.0.1:18636/api/framework
```

Example agent profile with constrained tools:

```json
{
  "agents": [
    {
      "id": "researcher",
      "name": "Researcher",
      "persona": "Focus on research. Delegate coding to other agents.",
      "tools": ["web_search", "web_fetch", "memory_recall", "agent_delegate"],
      "providers": ["openai"],
      "max_steps": 8
    }
  ],
  "framework": {
    "enabled": true,
    "delegate_enabled": true,
    "hooks_enabled": true
  }
}
```

## Phase 5 — Research-Backed Agent Upgrades

VietClaw Phase 5 adds capabilities inspired by agent research and OpenClaw-style proactive runtime patterns:

| Feature | Research / reference | What it does |
| --- | --- | --- |
| Active memory tools (`memory_recall`, `memory_store`) | MemGPT (arXiv:2310.08560), OpenClaw Active Memory | Agent proactively searches and writes long-term memory during tool loops |
| Reflexion on tool failure | Reflexion (arXiv:2303.11366) | Failed tool calls are stored as `experience` memories for future runs |
| Tiered context | MemoryOS (EMNLP 2025) | Session summary (mid-term) + hybrid FTS/vector recall (long-term) + recent messages (short-term) |
| Heartbeat scheduler | OpenClaw heartbeat polling | Optional periodic proactive checks via `agent.heartbeat` config |
| Max agent steps | Optional safety fuse | Default `0` (unlimited, prompt-first); set e.g. `12` in config if you want a cap |

Enable heartbeat (proactive agent) in config:

```json
{
  "agent": {
    "heartbeat": {
      "enabled": true,
      "interval_seconds": 1800,
      "prompt": "Heartbeat: check reminders and pending tasks. Reply briefly if needed."
    }
  }
}
```

Disable reflexion or memory tools if you want a slimmer runtime:

```json
{
  "agent": {
    "reflexion": { "enabled": false },
    "memory_tools": { "enabled": false }
  }
}
```

## VietClaw vs OpenClaw

| | VietClaw | OpenClaw |
| --- | --- | --- |
| Runtime | Single Go binary, ~low RAM | Node/TypeScript + optional Swift |
| Database | SQLite (no Postgres/Redis) | File-based + plugins |
| Channels | Discord, Telegram | 20+ platforms via plugins |
| Harness | Built-in plan/verify/worktree runner | Community plugins |
| Security default | `shell_exec` off, workspace-scoped files | Requires manual hardening |
| Proactive agent | Heartbeat scheduler (config) | Heartbeat + Task Brain |

VietClaw wins on deploy simplicity, resource use, and built-in coding harness. OpenClaw wins on channel breadth and plugin marketplace (ClawHub).

## Next Phases

- WhatsApp and more channel adapters
- Plugin install flow (npm/ClawHub-style)
- Web UI settings for heartbeat, reflexion, and memory tools
- VietClaw Harness runner UI for plan capsules, evidence ledger, verifier loops, and worktree follow-through
