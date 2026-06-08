export type DaemonStatus = {
  version: string
  commit: string
  uptime: string
  db_ok: boolean
  mode: string
  max_concurrent_tasks: number
}

export type ChatResponse = {
  ok: boolean
  session_id: string
  intent: string
  reply: string
  provider: string
  model: string
  cost_usd: number
  error?: string
}

export type MemoryRecord = {
  id: number
  scope: string
  kind: string
  content: string
  confidence: string
  created_at: string
  updated_at: string
}

export type ProviderConfig = {
  id: string
  type: string
  enabled: boolean
  default_model: string
  base_url?: string
  api_key_env?: string
}

export type ChannelStatus = {
  name: string
  enabled: boolean
  running: boolean
  error?: string
}

export type BudgetStatus = {
  total_cost_usd: number
  daily_usd_limit: number
  require_approval_above_usd: number
  cheap_first: boolean
  allow_escalation: boolean
}

export type Session = {
  id: string
  channel: string
  user_id: string
  created_at: string
  updated_at: string
}

export type SessionMessage = {
  id: number
  session_id: string
  role: string
  content: string
  created_at: string
}

export type SessionDetail = {
  session: Session
  messages: SessionMessage[]
}

