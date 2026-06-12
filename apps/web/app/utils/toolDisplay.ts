export type ToolDisplayView =
  | { mode: 'code', path?: string, content: string, lang: string }
  | { mode: 'command', command: string }
  | { mode: 'path', path: string }
  | { mode: 'failure', error: string, output?: string }
  | { mode: 'raw', text: string }
  | { mode: 'empty' }

const TOOL_OUTPUT_MARKERS = [
  '--- stdout/stderr ---',
  '--- output ---',
]

const TOOL_FAILURE_PREFIXES = [
  'Lệnh thất bại:',
  'Command failed:',
  'Lỗi thực thi công cụ:',
  'Tool execution error:',
]

const LANG_BY_EXT: Record<string, string> = {
  py: 'python',
  js: 'javascript',
  mjs: 'javascript',
  cjs: 'javascript',
  ts: 'typescript',
  tsx: 'typescript',
  jsx: 'javascript',
  go: 'go',
  rs: 'rust',
  java: 'java',
  kt: 'kotlin',
  cs: 'csharp',
  cpp: 'cpp',
  cc: 'cpp',
  c: 'c',
  h: 'c',
  hpp: 'cpp',
  rb: 'ruby',
  php: 'php',
  swift: 'swift',
  sh: 'bash',
  bash: 'bash',
  zsh: 'bash',
  ps1: 'powershell',
  sql: 'sql',
  json: 'json',
  yaml: 'yaml',
  yml: 'yaml',
  toml: 'toml',
  xml: 'xml',
  html: 'html',
  htm: 'html',
  css: 'css',
  scss: 'scss',
  md: 'markdown',
  vue: 'vue',
  dockerfile: 'docker',
}

export function langFromPath(path?: string): string {
  if (!path) return 'plaintext'
  const base = path.split(/[/\\]/).pop() || path
  if (base.toLowerCase() === 'dockerfile') return 'docker'
  const dot = base.lastIndexOf('.')
  if (dot === -1) return 'plaintext'
  return LANG_BY_EXT[base.slice(dot + 1).toLowerCase()] || 'plaintext'
}

function normalizeToolName(name: string): string {
  return name.trim().toLowerCase().replace(/\./g, '_')
}

function tryParseJson(raw: string): Record<string, unknown> | null {
  try {
    const parsed = JSON.parse(raw.trim())
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>
    }
  } catch {
    /* not json */
  }
  return null
}

function pickString(obj: Record<string, unknown>, keys: string[]): string | undefined {
  for (const key of keys) {
    const val = obj[key]
    if (typeof val === 'string' && val.trim()) return val
  }
  return undefined
}

function isTrivialResult(text: string): boolean {
  const t = text.trim().toLowerCase()
  return t === '' || t === 'ok' || t === 'success' || t === 'done'
}

function looksLikeCode(text: string): boolean {
  return text.length > 24 && (text.includes('\n') || text.includes('{') || text.includes('def ') || text.includes('function '))
}

export function parseToolInputDisplay(toolName: string, raw?: string): ToolDisplayView {
  if (!raw?.trim()) return { mode: 'empty' }
  const name = normalizeToolName(toolName)
  const obj = tryParseJson(raw)

  if (obj) {
    const path = pickString(obj, ['path', 'file', 'filepath', 'filename'])
    const content = pickString(obj, ['content', 'text', 'body', 'data', 'html'])
    const command = pickString(obj, ['command', 'cmd', 'shell'])

    if (name.includes('file_write') && path && content) {
      return { mode: 'code', path, content, lang: langFromPath(path) }
    }
    if (name.includes('file_read') && path) {
      return { mode: 'path', path }
    }
    if ((name.includes('shell') || name.includes('exec')) && command) {
      return { mode: 'command', command }
    }
    if (path && content && looksLikeCode(content)) {
      return { mode: 'code', path, content, lang: langFromPath(path) }
    }
    if (content && looksLikeCode(content)) {
      return { mode: 'code', content, lang: 'plaintext' }
    }
    if (command) {
      return { mode: 'command', command }
    }
    if (path) {
      return { mode: 'path', path }
    }
  }

  if ((name.includes('shell') || name.includes('exec')) && !raw.trim().startsWith('{')) {
    return { mode: 'command', command: raw.trim() }
  }

  return { mode: 'raw', text: formatToolJson(raw) }
}

export function parseToolFailure(raw: string): { error: string, output?: string } | null {
  for (const marker of TOOL_OUTPUT_MARKERS) {
    const idx = raw.indexOf(marker)
    if (idx === -1) continue
    const error = raw.slice(0, idx).trim()
    const output = raw.slice(idx + marker.length).trim()
    return { error, output: output || undefined }
  }

  const trimmed = raw.trim()
  for (const prefix of TOOL_FAILURE_PREFIXES) {
    if (trimmed.startsWith(prefix)) {
      return { error: trimmed }
    }
  }
  return null
}

export function isToolResultFailure(raw?: string): boolean {
  if (!raw?.trim()) return false
  return parseToolFailure(raw) !== null
}

export function toolFailureMessage(raw?: string): string {
  if (!raw?.trim()) return ''
  const failure = parseToolFailure(raw)
  return failure?.error ?? ''
}

const AGENT_SPAWN_TOOLS = new Set(['agent_spawn', 'agent_delegate', 'agent_spawn_batch'])

export function isAgentSpawnTool(toolName: string): boolean {
  const base = toolName.split(':')[0] ?? toolName
  return AGENT_SPAWN_TOOLS.has(normalizeToolName(base))
}

export function spawnAgentIdFromInput(input?: string): string {
  if (!input?.trim()) return ''
  const obj = tryParseJson(input.trim())
  if (!obj) return ''
  return pickString(obj, ['agent_id']) ?? ''
}

export function stripSpawnResultPrefix(raw: string): string {
  const trimmed = raw.trim()
  const single = trimmed.match(/^Spawned\s+([^\s:]+):\s*([\s\S]*)$/i)
  if (single?.[2]) return single[2].trim()
  return trimmed
}

export function parseToolResultDisplay(
  toolName: string,
  raw?: string,
  inputRaw?: string,
): ToolDisplayView {
  if (!raw?.trim()) return { mode: 'empty' }
  const name = normalizeToolName(toolName)
  let displayRaw = raw
  if (isAgentSpawnTool(toolName)) {
    displayRaw = stripSpawnResultPrefix(raw)
  }

  const failure = parseToolFailure(raw)
  if (failure) {
    return { mode: 'failure', error: failure.error, output: failure.output }
  }

  if (isTrivialResult(displayRaw) && (name.includes('file_write') || name.includes('write'))) {
    return { mode: 'empty' }
  }

  if ((name.includes('shell') || name.includes('exec')) && displayRaw.includes('\n')) {
    return { mode: 'failure', error: '', output: displayRaw.trim() }
  }

  let path: string | undefined
  const inputObj = inputRaw ? tryParseJson(inputRaw) : null
  if (inputObj) {
    path = pickString(inputObj, ['path', 'file', 'filepath', 'filename'])
  }

  const obj = tryParseJson(displayRaw)
  if (obj) {
    const content = pickString(obj, ['content', 'text', 'body', 'data', 'output', 'result'])
    if (content && looksLikeCode(content)) {
      return { mode: 'code', path, content, lang: langFromPath(path) }
    }
  }

  if (
    name.includes('file_read')
    || name.includes('file_head')
    || name.includes('file_tail')
    || name.includes('file_grep')
    || name.includes('grep')
  ) {
    if (looksLikeCode(displayRaw) || displayRaw.includes('\n')) {
      return { mode: 'code', path, content: displayRaw, lang: langFromPath(path) }
    }
  }

  if (isAgentSpawnTool(toolName) && displayRaw.includes('\n')) {
    return { mode: 'code', content: displayRaw, lang: 'markdown' }
  }

  if (looksLikeCode(displayRaw)) {
    return { mode: 'code', path, content: displayRaw, lang: langFromPath(path) }
  }

  return { mode: 'raw', text: formatToolJson(displayRaw) }
}

export function formatToolJson(raw: string, max = 8000): string {
  const trimmed = raw.trim()
  if (!trimmed) return ''
  try {
    const pretty = JSON.stringify(JSON.parse(trimmed), null, 2)
    return pretty.length > max ? `${pretty.slice(0, max)}…` : pretty
  } catch {
    return trimmed.length > max ? `${trimmed.slice(0, max)}…` : trimmed
  }
}
