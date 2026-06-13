# AGENTS.md — Rules & guide for AI agents working on VietClaw

This file is for **AI coding agents** (Cursor, VietClaw harness, etc.) and human maintainers. Read it **before** changing agent loop, spawn, model catalog, or Web UI chat code.

> **Not the same as** `~/.vietclaw/agents/<id>/AGENT.md` — that defines a **runtime** agent on the user's machine. **This** file lives in the repo and documents **development conventions**.

---

## 1. Agent architecture (summary)

```text
User / Web / Discord / Telegram
        ↓
  internal/agent (Chat, loop, spawn)
        ↓
  internal/agentfs   ← ~/.vietclaw/agents/*/AGENT.md
  internal/tools     ← registry + MCP + custom tools/*.md
  internal/router    ← provider/model + budget
  internal/context   ← skills, memory, history
```

| Component | Role |
|-----------|------|
| **Main agent** (`default`) | Default agent; can spawn or create child agents |
| **AgentRegistry** | Scans `~/.vietclaw/agents/`, hot-reload |
| **RunPool** | Enforces `runtime.max_concurrent_tasks` + `framework.max_concurrent_spawns` |
| **models.catalog** | User-selectable models (Web / Discord `/models` / Telegram) |
| **Framework tools** | `agent_create`, `agent_spawn`, `agent_spawn_batch`, `agent_delegate` |

Spawn flow: parent calls tool → `RunPool.Acquire` → child `Delegate` → full agentic loop → result returned to parent.

---

## 2. Runtime agent layout (on the user's machine)

```text
~/.vietclaw/agents/<agent-id>/
  AGENT.md          # YAML frontmatter + persona (markdown body)
  skills/*.md       # name, triggers, instructions
  tools/*.md        # type: guide | custom
```

### `AGENT.md` frontmatter (canonical)

```yaml
---
id: researcher
name: Researcher
language: vi
tools: [web_search, web_fetch, agent_spawn]
providers: []          # empty = all enabled providers
model: inherit         # inherit | <catalog-id> | <provider>/<model>
memory_scope: researcher
max_steps: 0
spawnable: true
auto_create: false
---
Persona goes here (English or Vietnamese)...
```

### `tools/*.md`

| `type` | Purpose |
|--------|---------|
| `guide` | Extra prompt for a built-in tool (`tool: web_search`) |
| `custom` | New tool; `handler: mcp:<server-id>` or `handler: script:./run.sh` (requires `shell.enabled`) |

**Agent id rule:** `[a-z0-9][a-z0-9_-]*`, must not overwrite `default`.

---

## 3. Config — what changed (do not revert)

| Removed / deprecated | Replaced by |
|----------------------|-------------|
| `config.agents[]` | Filesystem `~/.vietclaw/agents/` (auto-migrate) |
| Global `agent.skill_dirs` (canonical) | Per-agent `agents/<id>/skills/*.md` |
| Model selection via router only | `models.catalog` + `ChatRequest.catalog_id` |

### New config to remember

```json
{
  "framework": {
    "max_total_agents": 20,
    "max_concurrent_spawns": 3,
    "allow_auto_create": true
  },
  "models": {
    "catalog": [{ "id", "provider", "model", "label", "enabled" }],
    "default_catalog_id": "default"
  },
  "channels": {
    "telegram": {
      "command_mode": "slash",
      "command_prefix": "/"
    }
  }
}
```

Migration flag: `agents_migrated_v1` in DB `settings`; runs on `init` / `daemon` startup.

---

## 4. Mandatory rules when changing code

### 4.1 Agent & spawn

1. **Single profile source:** read from `internal/agentfs`; do not reintroduce `config.Agents[]`.
2. **Spawn must go through RunPool** — do not call `Delegate` directly and skip the semaphore.
3. **`agent_create`** must check `allow_auto_create` and `max_total_agents`, and **`ValidateCreateRequest`** (persona ≥400 chars, ≥3 `##` sections, skills + tool_guides).
4. **Child model:** `inherit` → parent's model; or catalog id; or `provider/model`.
5. **Framework tools** are registered in `internal/tools/framework.go`; handled in `internal/agent/spawn.go`.
6. **`agent_delegate`** stays for compatibility — sync alias of `agent_spawn`.

### 4.2 Model catalog

1. Override order: `ChatRequest.catalog_id` → `sessions.preferred_catalog_id` → channel preference → router default.
2. API: `GET /api/models/catalog`, `PUT /api/sessions/{id}/model`.
3. Router: `SelectExplicit` when overridden; validate `enabled` catalog entries.

### 4.3 Web UI (`apps/web`)

1. **i18n required** — every user-facing string in both `locales/en.json` and `locales/vi.json`.
2. Tool labels: `tool.ui.<tool_name>` (`.` → `_` in names).
3. Empty-chat suggestions: keep `chat.suggestion.*.label` + `.text` pattern.
4. Spawn SSE: `event: spawn` → render in `ChatPanel.vue` (`spawn` step type).
5. Model picker: **inline in the composer** (`vc-composer-model-*`), not a separate row above.
6. Model catalog settings: `/settings/models` + entry in `settingsNav.ts`.
7. Types mirror Go: `apps/web/app/types/config.ts`.

### 4.4 Channels

1. `/models` is handled in `internal/channels/commands.go` **before** the agent loop.
2. Discord: slash command registered on `Start()` — **daemon restart** required after code changes.
3. Telegram: `command_mode` = `slash` | `prefix`.

### 4.5 Code style (repo)

- Go: focused diffs, match existing conventions, no over-engineering.
- Do not commit `internal/web/dist` (except CI stub).
- Do not commit secrets, `.env`, or local databases.
- Tests: `go test ./...` (CI builds web first).

---

## 5. Checklist when adding or changing agent features

Use this before opening a PR:

- [ ] `internal/agentfs` — parse / registry / migrate still consistent?
- [ ] `profiles.go`, `context/builder.go`, `tools/registry.go` use the registry?
- [ ] New tool → `framework.go` + spawn handler if it's a framework tool?
- [ ] `ChatRequest` / API handlers / channels inject `catalog_id`?
- [ ] DB migration in `internal/db/migrate.go` if schema changed?
- [ ] `config/types.go` + `defaults.go` + `merge.go` + `validate.go`?
- [ ] Web: `config.ts`, settings page, i18n en+vi, chat UI (tool labels + spawn)?
- [ ] Tests: `tests/agentfs/`, `tests/agent/`, `tests/channels/`?
- [ ] README / CONTRIBUTING if user-facing behavior changed?

---

## 6. Important file map

| Area | Files |
|------|-------|
| Agent loop | `internal/agent/loop.go`, `chat.go`, `delegate.go`, `spawn.go`, `pool.go`, `model_select.go` |
| Filesystem agents | `internal/agentfs/*` |
| Tools | `internal/tools/registry.go`, `framework.go`, `custom.go` |
| Config | `internal/config/types.go`, `defaults.go`, `models.go` |
| HTTP API | `internal/web/agent_handlers.go`, `models_handlers.go`, `chat_handlers.go` |
| Channels | `internal/channels/commands.go`, `discord/discord.go`, `handler.go` |
| Web chat | `apps/web/app/components/ChatPanel.vue`, `composables/useChat.ts` |
| Web settings | `apps/web/app/components/ModelsView.vue`, `SettingsView.vue` |
| Daemon boot | `cmd/vietclaw/daemon.go`, `init.go` |

---

## 7. API endpoints (agent & model)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/agents` | List filesystem agents |
| `GET` | `/api/agents/{id}` | Agent detail |
| `POST` | `/api/agents/reload` | Rescan registry |
| `GET` | `/api/models/catalog` | Enabled catalog |
| `PUT` | `/api/sessions/{id}/model` | Set `preferred_catalog_id` |
| `POST` | `/api/chat/stream` | Accepts `catalog_id` in body |

---

## 8. Security (do not violate)

- `shell_exec` is **off by default** — do not enable in default config.
- File tools: `workspace_only: true` by default.
- Custom tool `script:` is **blocked** when `shell.enabled = false`.
- `agent_create` must not overwrite `default`; strict id validation.
- Budget checks apply to every child run (`internal/router/budget.go`).
- Do not bypass workspace/shell policy via tool guides.

---

## 9. Guidance for AI agents in this repo

1. **Read related files before editing** — match existing naming and patterns.
2. **Minimal diffs** — one purpose per PR; no unrelated refactors.
3. **Bilingual UI** — default agent language is `vi`; UI strings always en + vi.
4. **Do not edit plan files** in `.cursor/plans/` unless the user asks.
5. **Harness** reads this file via `importantDocs()` — keep it **concise and current**.
6. When adding agent tools or spawn behavior → update **sections 4, 5, and 6** in the same PR.

---

## 10. Architecture history (avoid regressions)

| Era | Change |
|-----|--------|
| Before 2026-06 | `config.agents[]`, sync-only `agent_delegate`, router-only model |
| Current | Filesystem agents, parallel spawn, model catalog, channel `/models` |

If code or docs still treat `config.agents[]` as the primary source, that is a **bug / tech debt** — migrate or remove it.

---

*Last updated: 2026-06 — after merging agentfs + spawn + model catalog.*
