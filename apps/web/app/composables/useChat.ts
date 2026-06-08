import type { ChatResponse } from '~/types'
import { apiFetch } from '~/utils/api'

export type ChatStepEvent = {
  type: 'text' | 'tool_call' | 'tool_result' | 'error' | 'done'
  text?: string
  toolName?: string
  toolInput?: string
  toolResult?: string
  error?: string
}

export type ChatItem = {
  role: 'user' | 'assistant'
  text: string
  steps: ChatStepEvent[]
  meta?: Pick<ChatResponse, 'intent' | 'provider' | 'model' | 'cost_usd'>
}

export type ChatSession = {
  id: string
  title: string
  persona: string
  messages: ChatItem[]
  sessionId: string
  createdAt: number
}

const STORAGE_KEY = 'vietclaw_chats'
const CONFIG_KEY = 'vietclaw_config'

const sessions = ref<ChatSession[]>([])
const currentSessionId = ref('')
const isGenerating = ref(false)

function loadSessions() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) sessions.value = JSON.parse(raw)
  } catch { sessions.value = [] }
  if (sessions.value.length === 0) createSession()
  else currentSessionId.value = sessions.value[0]?.id ?? ''
}

function saveSessions() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(sessions.value))
}

function createSession(persona = 'general') {
  const s: ChatSession = {
    id: `session_${Date.now()}`,
    title: 'Untitled Session',
    persona,
    messages: [],
    sessionId: '',
    createdAt: Date.now()
  }
  sessions.value.unshift(s)
  currentSessionId.value = s.id
  saveSessions()
  return s
}

function currentSession() {
  return sessions.value.find(s => s.id === currentSessionId.value)
}

function switchSession(id: string) {
  currentSessionId.value = id
}

function deleteSession(id: string) {
  sessions.value = sessions.value.filter(s => s.id !== id)
  if (currentSessionId.value === id && sessions.value.length > 0) {
    currentSessionId.value = sessions.value[0]?.id ?? ''
  }
  if (sessions.value.length === 0) createSession()
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
  try {
    const raw = localStorage.getItem(CONFIG_KEY)
    if (raw) return JSON.parse(raw)
  } catch {}
  return { apiKey: '', model: 'gemini-2.5-flash-preview-09-2025', temperature: 0.7, persona: 'general', voice: 'Zephyr' }
}

function saveConfig(cfg: Record<string, unknown>) {
  localStorage.setItem(CONFIG_KEY, JSON.stringify(cfg))
}

async function sendMessage(text: string) {
  const s = currentSession()
  if (!s || isGenerating.value) return

  s.messages.push({ role: 'user', text, steps: [] })
  if (s.title === 'Untitled Session' && s.messages.length >= 1) {
    s.title = text.slice(0, 30) + (text.length > 30 ? '...' : '')
  }
  saveSessions()

  isGenerating.value = true

  const assistantMsg: ChatItem = { role: 'assistant', text: '', steps: [] }
  s.messages.push(assistantMsg)

  try {
    const res = await fetch('/api/chat/stream', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        session_id: s.sessionId || undefined,
        user_id: 'local',
        channel: 'web',
        message: text,
        mode: 'eco'
      })
    })

    if (!res.ok) {
      throw new Error(`HTTP ${res.status}`)
    }

    const reader = res.body?.getReader()
    if (!reader) throw new Error('No response body')

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      let currentEvent = ''
      for (const line of lines) {
        if (line.startsWith('event: ')) {
          currentEvent = line.slice(7).trim()
          continue
        }
        if (line.startsWith('data: ')) {
          const data = line.slice(6).trim()
          if (currentEvent === 'done') break
          if (currentEvent === 'error') {
            const parsed = JSON.parse(data)
            assistantMsg.text += `\n\nError: ${parsed.error}`
            assistantMsg.steps.push({ type: 'error', error: parsed.error })
            break
          }
          try {
            const parsed = JSON.parse(data)
            if (currentEvent === 'tool_call') {
              assistantMsg.steps.push({
                type: 'tool_call',
                toolName: parsed.name,
                toolInput: parsed.input
              })
            } else if (currentEvent === 'tool_result') {
              assistantMsg.steps.push({
                type: 'tool_result',
                toolName: parsed.name,
                toolResult: parsed.result
              })
            } else if (parsed.text) {
              assistantMsg.text += parsed.text
              assistantMsg.steps.push({ type: 'text', text: parsed.text })
            }
          } catch {}
          currentEvent = ''
        }
      }
    }

    if (s.sessionId === '' && assistantMsg.text) {
      // Session ID should have been set by the backend
    }
  } catch (err) {
    const msg = err instanceof Error ? err.message : 'Connection failed.'
    assistantMsg.text = `⚠️ ${msg}`
    assistantMsg.steps.push({ type: 'error', error: msg })
  } finally {
    isGenerating.value = false
    saveSessions()
  }
}

export function useChat() {
  if (sessions.value.length === 0) loadSessions()

  return {
    sessions,
    currentSessionId,
    isGenerating,
    currentSession,
    createSession,
    switchSession,
    deleteSession,
    clearSessionMessages,
    sendMessage,
    loadSessions,
    saveSessions,
    loadConfig,
    saveConfig
  }
}
