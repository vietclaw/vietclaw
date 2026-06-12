import type { ChatResponse } from '~/types'
import { apiFetch } from '~/utils/api'

export type SpawnStatus = 'running' | 'done' | 'failed'

export type ChatStepEvent = {
  type: 'text' | 'tool_call' | 'tool_result' | 'spawn' | 'error' | 'done'
  text?: string
  toolName?: string
  toolInput?: string
  toolResult?: string
  toolEventId?: number
  agentId?: string
  spawnStatus?: string
  spawnSummary?: string
  childSessionId?: string
  parentSessionId?: string
  error?: string
}

export type CatalogModel = {
  id: string
  provider: string
  model: string
  label: string
  enabled: boolean
}

type SSEEvent = {
  event: string
  data: string
}

export type ChatItem = {
  role: 'user' | 'assistant'
  text: string
  steps: ChatStepEvent[]
  meta?: Pick<ChatResponse, 'intent' | 'provider' | 'model' | 'cost_usd'>
}

export type SessionKind = 'root' | 'spawn'

export type ChatSession = {
  id: string
  title: string
  persona: string
  messages: ChatItem[]
  sessionId: string
  createdAt: number
  kind?: SessionKind
  parentId?: string
  agentId?: string
  readOnly?: boolean
  spawnStatus?: SpawnStatus
  taskPreview?: string
}

type ApiMessage = {
  role: string
  content: string
  created_at?: string
}

type ApiToolEvent = {
  id: number
  tool_name: string
  input: string
  output: string
  ok: boolean
  error?: string
  created_at: string
}

type ApiSessionDetail = {
  session: { id: string }
  messages: ApiMessage[]
  tool_events?: ApiToolEvent[]
  run_status?: string
  run_summary?: string
}

type ApiChildSession = {
  id: string
  agent_id: string
  task_preview?: string
  run_status?: string
  has_reply?: boolean
  created_at: string
  updated_at: string
}

const STORAGE_KEY = 'vietclaw_chats'
const CONFIG_KEY = 'vietclaw_config'
const MODEL_KEY = 'vietclaw_catalog_id'

export function sessionPath(id: string) {
  return `/p/${encodeURIComponent(id)}`
}

const sessions = ref<ChatSession[]>([])
const currentSessionId = ref('')
const expandedRootId = ref('')
const isGenerating = ref(false)
const catalogModels = ref<CatalogModel[]>([])
const defaultCatalogId = ref('')
const selectedCatalogId = ref('')
let streamAbort: AbortController | null = null
let streamReader: ReadableStreamDefaultReader<Uint8Array> | null = null
let childWatchAbort: AbortController | null = null
let childWatchReader: ReadableStreamDefaultReader<Uint8Array> | null = null
let childWatchSessionId = ''

function createBackendSessionID() {
  const suffix = Math.random().toString(36).slice(2, 10)
  return `sess_${Date.now()}_${suffix}`
}

function isSpawnSessionId(id: string): boolean {
  return id.includes(':spawn:')
}

function parseParentId(id: string): string {
  const idx = id.indexOf(':spawn:')
  return idx >= 0 ? id.slice(0, idx) : id
}

function parseSpawnAgentId(id: string): string {
  const marker = ':spawn:'
  const idx = id.indexOf(marker)
  if (idx < 0) return ''
  const rest = id.slice(idx + marker.length)
  const end = rest.indexOf(':')
  return end > 0 ? rest.slice(0, end) : ''
}

function normalizeSession(value: unknown): ChatSession | null {
  if (!value || typeof value !== 'object') return null
  const raw = value as Partial<ChatSession>
  const sessionId = typeof raw.sessionId === 'string' && raw.sessionId
    ? raw.sessionId
    : typeof raw.id === 'string' && raw.id.startsWith('sess_')
      ? raw.id
      : createBackendSessionID()
  const kind = raw.kind === 'spawn' || isSpawnSessionId(sessionId) ? 'spawn' : 'root'
  const parentId = typeof raw.parentId === 'string'
    ? raw.parentId
    : kind === 'spawn' ? parseParentId(sessionId) : undefined
  const agentId = typeof raw.agentId === 'string'
    ? raw.agentId
    : kind === 'spawn' ? parseSpawnAgentId(sessionId) : undefined
  return {
    id: sessionId,
    title: typeof raw.title === 'string' && raw.title ? raw.title : (agentId || 'Untitled Session'),
    persona: typeof raw.persona === 'string' && raw.persona ? raw.persona : 'general',
    messages: Array.isArray(raw.messages)
      ? raw.messages.map(message => ({
          role: message.role === 'assistant' ? 'assistant' : 'user',
          text: typeof message.text === 'string' ? message.text : '',
          steps: Array.isArray(message.steps) ? message.steps : [],
          meta: message.meta
        }))
      : [],
    sessionId,
    createdAt: typeof raw.createdAt === 'number' ? raw.createdAt : Date.now(),
    kind,
    parentId,
    agentId,
    readOnly: kind === 'spawn' ? true : raw.readOnly,
    spawnStatus: raw.spawnStatus,
    taskPreview: typeof raw.taskPreview === 'string' ? raw.taskPreview : undefined
  }
}

function truncateTask(text: string, max = 48): string {
  const t = text.trim()
  if (t.length <= max) return t
  return `${t.slice(0, max)}…`
}

function mapRunStatus(status?: string): SpawnStatus {
  if (status === 'completed') return 'done'
  if (status === 'failed' || status === 'blocked' || status === 'needs_approval') return 'failed'
  if (status === 'running') return 'running'
  return 'done'
}

function isPendingChildId(id: string): boolean {
  return id.includes(':pending')
}

function removePendingPlaceholders(parentId: string, agentId: string) {
  sessions.value = sessions.value.filter(s =>
    !(s.kind === 'spawn' && s.parentId === parentId && s.agentId === agentId && isPendingChildId(s.id))
  )
}

function normalizeSessions(value: unknown): ChatSession[] {
  if (!Array.isArray(value)) return []
  return value.map(normalizeSession).filter((session): session is ChatSession => Boolean(session))
}

function loadSessions() {
  if (import.meta.client) {
    try {
      const raw = localStorage.getItem(STORAGE_KEY)
      if (raw) {
        sessions.value = normalizeSessions(JSON.parse(raw))
          .filter(s => !(s.kind === 'spawn' && isPendingChildId(s.id)))
      }
    } catch { sessions.value = [] }
  }
  if (sessions.value.length === 0) createSession()
  else currentSessionId.value = sessions.value[0]?.id ?? ''
}

function saveSessions() {
  if (!import.meta.client) return
  localStorage.setItem(STORAGE_KEY, JSON.stringify(sessions.value))
}

function createSession(persona = 'general') {
  const sessionId = createBackendSessionID()
  const s: ChatSession = {
    id: sessionId,
    title: 'Untitled Session',
    persona,
    messages: [],
    sessionId,
    createdAt: Date.now(),
    kind: 'root'
  }
  sessions.value.unshift(s)
  currentSessionId.value = s.id
  expandedRootId.value = s.id
  saveSessions()
  return s
}

const rootSessions = computed(() =>
  sessions.value.filter(s => s.kind !== 'spawn')
)

function childrenOf(parentId: string) {
  return sessions.value
    .filter(s => s.kind === 'spawn' && s.parentId === parentId)
    .sort((a, b) => a.createdAt - b.createdAt)
}

const activeRootId = computed(() => {
  const current = sessions.value.find(s => s.id === currentSessionId.value)
  if (!current) return expandedRootId.value
  return current.kind === 'spawn' ? (current.parentId ?? expandedRootId.value) : current.id
})

function setExpandedRoot(id: string) {
  expandedRootId.value = id
}

function registerChildSession(opts: {
  parentId: string
  childSessionId: string
  agentId: string
  spawnStatus?: SpawnStatus
  taskPreview?: string
}) {
  if (!isPendingChildId(opts.childSessionId)) {
    removePendingPlaceholders(opts.parentId, opts.agentId)
  }

  const existing = sessions.value.find(s => s.id === opts.childSessionId)
  if (existing) {
    if (opts.spawnStatus) existing.spawnStatus = opts.spawnStatus
    if (opts.taskPreview) existing.taskPreview = opts.taskPreview
    saveSessions()
    return existing
  }
  const child: ChatSession = {
    id: opts.childSessionId,
    sessionId: opts.childSessionId,
    title: opts.agentId,
    persona: opts.agentId,
    messages: [],
    createdAt: Date.now(),
    kind: 'spawn',
    parentId: opts.parentId,
    agentId: opts.agentId,
    readOnly: true,
    spawnStatus: opts.spawnStatus ?? 'running',
    taskPreview: opts.taskPreview
  }
  sessions.value.push(child)
  saveSessions()
  return child
}

function updateChildSession(
  childId: string,
  patch: Partial<Pick<ChatSession, 'spawnStatus' | 'title' | 'taskPreview'>>,
) {
  const child = sessions.value.find(s => s.id === childId)
  if (!child) return
  if (patch.spawnStatus) child.spawnStatus = patch.spawnStatus
  if (patch.title) child.title = patch.title
  if (patch.taskPreview) child.taskPreview = patch.taskPreview
  saveSessions()
}

function buildStepsFromHistory(
  assistantMsgs: ApiMessage[],
  toolEvents: ApiToolEvent[],
): ChatStepEvent[] {
  type TimelineItem =
    | { kind: 'text', at: string, content: string }
    | { kind: 'tool_call', at: string, event: ApiToolEvent }
    | { kind: 'tool_result', at: string, event: ApiToolEvent }

  const timeline: TimelineItem[] = []
  for (const msg of assistantMsgs) {
    const content = msg.content?.trim() ?? ''
    if (!content) continue
    timeline.push({ kind: 'text', at: msg.created_at ?? '', content })
  }
  for (const event of toolEvents) {
    timeline.push({ kind: 'tool_call', at: event.created_at, event })
    timeline.push({ kind: 'tool_result', at: event.created_at, event })
  }
  timeline.sort((a, b) => a.at.localeCompare(b.at))

  const steps: ChatStepEvent[] = []
  let prevAssistant = ''
  for (const item of timeline) {
    if (item.kind === 'text') {
      const delta = item.content.startsWith(prevAssistant)
        ? item.content.slice(prevAssistant.length)
        : item.content
      if (delta.trim()) steps.push({ type: 'text', text: delta })
      prevAssistant = item.content
      continue
    }
    if (item.kind === 'tool_call') {
      steps.push({
        type: 'tool_call',
        toolName: item.event.tool_name,
        toolInput: item.event.input
      })
      continue
    }
    const step: ChatStepEvent = {
      type: 'tool_result',
      toolName: item.event.tool_name,
      toolResult: item.event.output,
      toolEventId: item.event.id
    }
    if (!item.event.ok && item.event.error) {
      step.toolResult = item.event.error
    }
    steps.push(step)
  }
  return steps
}

function buildSessionMessages(
  messages: ApiMessage[],
  toolEvents: ApiToolEvent[] = [],
): ChatItem[] {
  const visible = messages.filter(msg => msg.role !== 'system')
  const items: ChatItem[] = []

  for (const msg of visible) {
    if (msg.role !== 'user') continue
    const content = msg.content?.trim() ?? ''
    if (!content) continue
    items.push({ role: 'user', text: content, steps: [] })
  }

  const assistantMsgs = visible.filter(msg => msg.role === 'assistant')
  const steps = buildStepsFromHistory(assistantMsgs, toolEvents)
  const finalText = assistantMsgs[assistantMsgs.length - 1]?.content?.trim() ?? ''
  if (finalText || steps.length > 0) {
    items.push({ role: 'assistant', text: finalText, steps })
  }
  return items
}

function maxToolEventId(toolEvents: ApiToolEvent[] = []): number {
  return toolEvents.reduce((max, event) => Math.max(max, event.id || 0), 0)
}

async function hydrateSessionFromAPI(id: string): Promise<ChatSession | null> {
  try {
    const detail = await apiFetch<ApiSessionDetail>(`/api/sessions/${encodeURIComponent(id)}`)
    const agentId = parseSpawnAgentId(id)
    const parentId = parseParentId(id)
    const userMsg = detail.messages?.find(m => m.role === 'user' && m.content?.trim())
    const taskPreview = userMsg?.content ? truncateTask(userMsg.content) : undefined
    const spawnStatus = mapRunStatus(detail.run_status)
    let session = sessions.value.find(s => s.id === id)
    if (!session) {
      session = registerChildSession({
        parentId,
        childSessionId: id,
        agentId,
        spawnStatus,
        taskPreview
      })
    }
    session.messages = buildSessionMessages(detail.messages || [], detail.tool_events || [])
    session.readOnly = true
    session.kind = 'spawn'
    session.parentId = parentId
    session.agentId = agentId
    session.spawnStatus = spawnStatus
    if (taskPreview) session.taskPreview = taskPreview
    saveSessions()

    if (spawnStatus === 'running' && currentSessionId.value === id) {
      void watchChildSession(id, maxToolEventId(detail.tool_events))
    }
    return session
  } catch {
    return null
  }
}

async function loadChildrenForParent(parentId: string) {
  try {
    const children = await apiFetch<ApiChildSession[]>(
      `/api/sessions/${encodeURIComponent(parentId)}/children`
    )
    for (const item of children) {
      const spawnStatus = mapRunStatus(item.run_status)
      const taskPreview = item.task_preview ? truncateTask(item.task_preview) : undefined
      const existing = sessions.value.find(s => s.id === item.id)
      if (existing) {
        if (!existing.agentId) existing.agentId = item.agent_id
        existing.spawnStatus = spawnStatus
        if (taskPreview) existing.taskPreview = taskPreview
        continue
      }
      registerChildSession({
        parentId,
        childSessionId: item.id,
        agentId: item.agent_id,
        spawnStatus,
        taskPreview
      })
    }
    saveSessions()
  } catch {
    /* daemon offline */
  }
}

function handleSpawnSSE(
  parentSession: ChatSession,
  parsed: {
    agent_id?: string
    status?: string
    summary?: string
    child_session_id?: string
    parent_session_id?: string
  }
) {
  const agentId = parsed.agent_id ?? 'agent'
  const status = (parsed.status ?? 'running') as SpawnStatus
  const childSessionId = parsed.child_session_id
  const parentId = parsed.parent_session_id || parentSession.sessionId
  const taskPreview = parsed.summary && status === 'running'
    ? truncateTask(parsed.summary)
    : undefined

  if (childSessionId) {
    registerChildSession({
      parentId,
      childSessionId,
      agentId,
      spawnStatus: status,
      taskPreview
    })
    if (status === 'running' && childSessionId && currentSessionId.value === childSessionId) {
      void watchChildSession(childSessionId)
    }
    if (status === 'done' || status === 'failed') {
      updateChildSession(childSessionId, { spawnStatus: status })
      if (currentSessionId.value === childSessionId) {
        stopChildWatch()
      }
      void hydrateSessionFromAPI(childSessionId)
    }
  } else {
    const fallbackId = `${parentId}:spawn:${agentId}:pending`
    registerChildSession({
      parentId,
      childSessionId: fallbackId,
      agentId,
      spawnStatus: status,
      taskPreview
    })
  }
}

function currentSession() {
  return sessions.value.find(s => s.id === currentSessionId.value)
}

function switchSession(id: string) {
  if (currentSessionId.value !== id) {
    stopChildWatch()
  }
  currentSessionId.value = id
  const session = sessions.value.find(s => s.id === id)
  if (session?.kind === 'spawn' && session.parentId) {
    expandedRootId.value = session.parentId
  } else if (session?.kind !== 'spawn') {
    expandedRootId.value = id
  }
  if (session?.kind === 'spawn' && session.spawnStatus === 'running') {
    const lastToolId = maxToolEventIdFromSteps(session.messages)
    void watchChildSession(id, lastToolId)
  }
}

function maxToolEventIdFromSteps(messages: ChatItem[]): number {
  let max = 0
  for (const msg of messages) {
    for (const step of msg.steps) {
      if (typeof step.toolEventId === 'number') max = Math.max(max, step.toolEventId)
    }
  }
  return max
}

function deleteSession(id: string) {
  const wasCurrent = currentSessionId.value === id
  sessions.value = sessions.value.filter(s => s.id !== id && s.parentId !== id)
  if (wasCurrent) {
    if (sessions.value.length > 0) {
      currentSessionId.value = sessions.value[0]?.id ?? ''
      if (import.meta.client) {
        void navigateTo(sessionPath(currentSessionId.value), { replace: true })
      }
    } else {
      const created = createSession()
      if (import.meta.client) {
        void navigateTo(sessionPath(created.id), { replace: true })
      }
    }
  }
  saveSessions()
}

function clearSessionMessages() {
  const s = currentSession()
  if (s) {
    s.messages = []
    saveSessions()
  }
}

function loadConfig() {
  if (!import.meta.client) {
    return { apiKey: '', model: 'gemini-2.5-flash-preview-09-2025', temperature: 0.7, persona: 'general', voice: 'Zephyr' }
  }
  try {
    const raw = localStorage.getItem(CONFIG_KEY)
    if (raw) return JSON.parse(raw)
  } catch {}
  return { apiKey: '', model: 'gemini-2.5-flash-preview-09-2025', temperature: 0.7, persona: 'general', voice: 'Zephyr' }
}

function saveConfig(cfg: Record<string, unknown>) {
  if (!import.meta.client) return
  localStorage.setItem(CONFIG_KEY, JSON.stringify(cfg))
}

function parseSSEBlock(block: string): SSEEvent | null {
  const lines = block.split(/\r?\n/)
  let event = 'message'
  const data: string[] = []
  for (const line of lines) {
    if (line.startsWith('event:')) {
      event = line.slice(6).trim()
    } else if (line.startsWith('data:')) {
      data.push(line.slice(5).trimStart())
    }
  }
  if (data.length === 0) return null
  return { event, data: data.join('\n') }
}

function applySSEEvent(event: SSEEvent, session: ChatSession, msgIndex: number): boolean {
  const assistantMsg = session.messages[msgIndex]
  if (!assistantMsg) return true

  if (event.event === 'done') return true
  if (event.event === 'error') {
    const parsed = JSON.parse(event.data)
    assistantMsg.text += `\n\nError: ${parsed.error}`
    assistantMsg.steps.push({ type: 'error', error: parsed.error })
    return true
  }

  const parsed = JSON.parse(event.data)
  if (event.event === 'session') {
    const nextSessionID = typeof parsed.session_id === 'string' ? parsed.session_id : ''
    if (nextSessionID && nextSessionID !== session.sessionId) {
      session.id = nextSessionID
      session.sessionId = nextSessionID
      currentSessionId.value = nextSessionID
      if (import.meta.client) {
        void navigateTo(sessionPath(nextSessionID), { replace: true })
      }
    }
  } else if (event.event === 'tool_call') {
    assistantMsg.steps.push({
      type: 'tool_call',
      toolName: parsed.name,
      toolInput: parsed.input
    })
  } else if (event.event === 'tool_result') {
    assistantMsg.steps.push({
      type: 'tool_result',
      toolName: parsed.name,
      toolResult: parsed.result
    })
  } else if (event.event === 'spawn') {
    assistantMsg.steps.push({
      type: 'spawn',
      agentId: parsed.agent_id,
      spawnStatus: parsed.status,
      spawnSummary: parsed.summary,
      childSessionId: parsed.child_session_id,
      parentSessionId: parsed.parent_session_id
    })
    handleSpawnSSE(session, parsed)
  } else if (event.event === 'text' || parsed.text) {
    assistantMsg.text += parsed.text
    const last = assistantMsg.steps[assistantMsg.steps.length - 1]
    if (last?.type === 'text') {
      last.text = (last.text || '') + parsed.text
    } else {
      assistantMsg.steps.push({ type: 'text', text: parsed.text })
    }
  }
  return false
}

function yieldToUI() {
  return new Promise<void>(resolve => {
    if (typeof window !== 'undefined' && window.requestAnimationFrame) {
      window.requestAnimationFrame(() => resolve())
      return
    }
    setTimeout(resolve, 0)
  })
}

function stopGeneration() {
  if (streamAbort) {
    streamAbort.abort()
  }
  if (streamReader) {
    void streamReader.cancel()
  }
}

function stopChildWatch() {
  if (childWatchAbort) {
    childWatchAbort.abort()
  }
  if (childWatchReader) {
    void childWatchReader.cancel()
  }
  childWatchAbort = null
  childWatchReader = null
  childWatchSessionId = ''
  if (isGenerating.value && currentSession()?.kind === 'spawn') {
    isGenerating.value = false
  }
}

function ensureChildAssistantMessage(session: ChatSession): number {
  let msgIndex = session.messages.findIndex(msg => msg.role === 'assistant')
  if (msgIndex < 0) {
    session.messages.push({ role: 'assistant', text: '', steps: [] })
    msgIndex = session.messages.length - 1
  }
  return msgIndex
}

function applyChildWatchEvent(
  session: ChatSession,
  msgIndex: number,
  event: SSEEvent,
): boolean {
  const assistantMsg = session.messages[msgIndex]
  if (!assistantMsg) return true

  if (event.event === 'done') return true
  if (event.event === 'error') {
    const parsed = JSON.parse(event.data) as { error?: string }
    assistantMsg.steps.push({ type: 'error', error: parsed.error ?? 'Unknown error' })
    session.spawnStatus = 'failed'
    return true
  }

  const parsed = JSON.parse(event.data) as Record<string, string>
  if (event.event === 'run_status') {
    session.spawnStatus = mapRunStatus(parsed.status)
    if (parsed.summary?.trim() && session.spawnStatus === 'done') {
      assistantMsg.text = parsed.summary
    }
    return false
  }
  if (event.event === 'tool_call') {
    assistantMsg.steps.push({
      type: 'tool_call',
      toolName: parsed.name,
      toolInput: parsed.input
    })
  } else if (event.event === 'tool_result') {
    assistantMsg.steps.push({
      type: 'tool_result',
      toolName: parsed.name,
      toolResult: parsed.result
    })
  } else if (event.event === 'text' || parsed.text) {
    assistantMsg.text += parsed.text ?? ''
    const last = assistantMsg.steps[assistantMsg.steps.length - 1]
    if (last?.type === 'text') {
      last.text = (last.text || '') + (parsed.text ?? '')
    } else {
      assistantMsg.steps.push({ type: 'text', text: parsed.text })
    }
  }
  return false
}

async function watchChildSession(sessionId: string, afterToolEventId = 0) {
  const session = sessions.value.find(s => s.id === sessionId)
  if (!session || session.kind !== 'spawn' || session.spawnStatus !== 'running') return
  if (childWatchSessionId === sessionId && childWatchAbort) return

  stopChildWatch()
  childWatchSessionId = sessionId
  childWatchAbort = new AbortController()
  isGenerating.value = true

  const msgIndex = ensureChildAssistantMessage(session)

  try {
    const query = afterToolEventId > 0 ? `?after=${afterToolEventId}` : ''
    const res = await fetch(`/api/sessions/${encodeURIComponent(sessionId)}/watch${query}`, {
      signal: childWatchAbort.signal
    })
    if (!res.ok) {
      throw new Error(`HTTP ${res.status}`)
    }

    const reader = res.body?.getReader()
    if (!reader) throw new Error('No response body')
    childWatchReader = reader

    const decoder = new TextDecoder()
    let buffer = ''
    let stop = false

    while (!stop) {
      const { done, value } = await reader.read()
      if (done) {
        buffer += decoder.decode()
      } else {
        buffer += decoder.decode(value, { stream: true })
      }

      const blocks = buffer.split(/\r?\n\r?\n/)
      buffer = blocks.pop() || ''
      for (const block of blocks) {
        const event = parseSSEBlock(block)
        if (!event) continue
        try {
          stop = applyChildWatchEvent(session, msgIndex, event)
          saveSessions()
          await yieldToUI()
        } catch {
          assistantMsgSafe(session, msgIndex)?.steps.push({ type: 'error', error: 'Invalid stream event' })
          stop = true
        }
        if (stop) break
      }
      if (done) break
    }
  } catch (err) {
    if (!(err instanceof Error && err.name === 'AbortError')) {
      const assistant = session.messages[msgIndex]
      if (assistant) {
        assistant.steps.push({
          type: 'error',
          error: err instanceof Error ? err.message : 'Connection failed.'
        })
      }
    }
  } finally {
    if (childWatchSessionId === sessionId) {
      childWatchAbort = null
      childWatchReader = null
      childWatchSessionId = ''
      isGenerating.value = false
      saveSessions()
      void hydrateSessionFromAPI(sessionId)
    }
  }
}

function assistantMsgSafe(session: ChatSession, msgIndex: number) {
  return session.messages[msgIndex]
}

async function loadCatalogModels() {
  try {
    const res = await apiFetch<{ catalog: CatalogModel[], default_catalog_id: string }>('/api/models/catalog')
    catalogModels.value = (res.catalog || []).filter(m => m.enabled)
    defaultCatalogId.value = res.default_catalog_id || catalogModels.value[0]?.id || ''
    if (!selectedCatalogId.value) {
      selectedCatalogId.value = import.meta.client
        ? localStorage.getItem(MODEL_KEY) || defaultCatalogId.value
        : defaultCatalogId.value
    }
  } catch {
    catalogModels.value = []
  }
}

async function setSelectedCatalog(id: string) {
  selectedCatalogId.value = id
  if (import.meta.client) {
    localStorage.setItem(MODEL_KEY, id)
  }
  const s = currentSession()
  if (s?.sessionId) {
    try {
      await apiFetch(`/api/sessions/${encodeURIComponent(s.sessionId)}/model`, {
        method: 'PUT',
        body: JSON.stringify({ catalog_id: id })
      })
    } catch {}
  }
}

async function sendMessage(text: string) {
  const s = currentSession()
  if (!s || s.readOnly || isGenerating.value) return

  s.messages.push({ role: 'user', text, steps: [] })
  if (s.title === 'Untitled Session' && s.messages.length >= 1) {
    s.title = text.slice(0, 30) + (text.length > 30 ? '...' : '')
  }
  saveSessions()

  isGenerating.value = true
  streamAbort = new AbortController()

  const msgIndex = s.messages.length
  s.messages.push({ role: 'assistant', text: '', steps: [] })

  try {
    const res = await fetch('/api/chat/stream', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      signal: streamAbort.signal,
      body: JSON.stringify({
        session_id: s.sessionId,
        user_id: 'local',
        channel: 'web',
        message: text,
        mode: 'eco',
        catalog_id: selectedCatalogId.value || undefined
      })
    })

    if (!res.ok) {
      throw new Error(`HTTP ${res.status}`)
    }

    const reader = res.body?.getReader()
    if (!reader) throw new Error('No response body')
    streamReader = reader

    const decoder = new TextDecoder()
    let buffer = ''
    let stop = false

    while (!stop) {
      const { done, value } = await reader.read()
      if (done) {
        buffer += decoder.decode()
      } else {
        buffer += decoder.decode(value, { stream: true })
      }

      const blocks = buffer.split(/\r?\n\r?\n/)
      buffer = blocks.pop() || ''
      for (const block of blocks) {
        const event = parseSSEBlock(block)
        if (!event) continue
        try {
          stop = applySSEEvent(event, s, msgIndex)
          await yieldToUI()
        } catch {
          const msg = s.messages[msgIndex]
          if (msg) {
            msg.steps.push({ type: 'error', error: 'Invalid stream event' })
          }
          stop = true
        }
        if (stop) break
      }

      if (done) break
    }

  } catch (err) {
    const assistant = s.messages[msgIndex]
    if (!assistant) return

    if (err instanceof Error && err.name === 'AbortError') {
      return
    }

    const msg = err instanceof Error ? err.message : 'Connection failed.'
    assistant.text = `⚠️ ${msg}`
    assistant.steps.push({ type: 'error', error: msg })
  } finally {
    streamAbort = null
    streamReader = null
    isGenerating.value = false
    saveSessions()
  }
}

export function useChat() {
  if (sessions.value.length === 0) loadSessions()
  if (catalogModels.value.length === 0) void loadCatalogModels()

  return {
    sessions,
    rootSessions,
    currentSessionId,
    expandedRootId,
    activeRootId,
    isGenerating,
    catalogModels,
    selectedCatalogId,
    currentSession,
    childrenOf,
    createSession,
    switchSession,
    deleteSession,
    clearSessionMessages,
    sendMessage,
    stopGeneration,
    loadCatalogModels,
    setSelectedCatalog,
    sessionPath,
    loadSessions,
    saveSessions,
    loadConfig,
    saveConfig,
    setExpandedRoot,
    registerChildSession,
    updateChildSession,
    hydrateSessionFromAPI,
    loadChildrenForParent,
    watchChildSession,
    stopChildWatch,
    isSpawnSessionId,
    parseParentId
  }
}
