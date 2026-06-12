<script setup lang="ts">
import {
  AlertCircle,
  ArrowUp,
  Check,
  ChevronDown,
  Square,
  Copy,
  FileText,
  FolderOpen,
  Globe,
  RefreshCw,
  Search,
  Settings,
  Terminal,
  Users,
  Wrench,
} from '@lucide/vue'
import katex from 'katex'
import { marked } from 'marked'
import type { ChatItem, ChatStepEvent } from '~/composables/useChat'
import { enhanceCodeBlocks } from '~/utils/enhanceCodeBlocks'
import {
  isAgentSpawnTool,
  isToolResultFailure,
  spawnAgentIdFromInput,
  stripSpawnResultPrefix,
  toolFailureMessage,
} from '~/utils/toolDisplay'

const {
  currentSession,
  currentSessionId,
  isGenerating,
  catalogModels,
  selectedCatalogId,
  sendMessage,
  clearSessionMessages,
  stopGeneration,
  setSelectedCatalog,
  sessionPath,
} = useChat()
const { t, toolLabel } = useI18n()
const { config, load: loadSettings } = useSettings()
const toast = useToast()

const chatInput = ref('')
const chatBox = ref<HTMLElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const modelMenuRef = ref<HTMLElement | null>(null)
const modelMenuOpen = ref(false)
const expandedTools = ref<Set<string>>(new Set())
const stickToBottom = ref(true)

const selectedModelLabel = computed(() => {
  const match = catalogModels.value.find(m => m.id === selectedCatalogId.value)
  return match?.label || match?.id || t('chat.modelDefault')
})

function pickModel(id: string) {
  void setSelectedCatalog(id)
  modelMenuOpen.value = false
}

function onDocumentClick(event: MouseEvent) {
  if (!modelMenuOpen.value) return
  const target = event.target as Node | null
  if (target && modelMenuRef.value?.contains(target)) return
  modelMenuOpen.value = false
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  if (!config.value) void loadSettings()
})
onUnmounted(() => document.removeEventListener('click', onDocumentClick))

const SCROLL_STICK_THRESHOLD = 96

const suggestions = computed(() => {
  const items = [
    { id: 'remember', label: t('chat.suggestion.remember.label'), text: t('chat.suggestion.remember.text') },
    { id: 'search', label: t('chat.suggestion.search.label'), text: t('chat.suggestion.search.text') },
    { id: 'spawn', label: t('chat.suggestion.spawn.label'), text: t('chat.suggestion.spawn.text') },
    { id: 'createAgent', label: t('chat.suggestion.createAgent.label'), text: t('chat.suggestion.createAgent.text') },
    { id: 'delegate', label: t('chat.suggestion.delegate.label'), text: t('chat.suggestion.delegate.text') },
    { id: 'workspace', label: t('chat.suggestion.workspace.label'), text: t('chat.suggestion.workspace.text') },
  ]
  if (!config.value?.framework?.allow_auto_create) {
    return items.filter(item => item.id !== 'createAgent')
  }
  return items
})

const SUMMARY_KEYS = ['query', 'command', 'cmd', 'path', 'file', 'url', 'name', 'input', 'text', 'pattern', 'expression', 'message', 'prompt', 'agent_id']

type ToolGroup = {
  id: string
  toolName: string
  input?: string
  result?: string
  error?: string
}

marked.setOptions({ breaks: true, gfm: true })

function truncate(text: string, max = 72): string {
  const t = text.trim()
  if (t.length <= max) return t
  return `${t.slice(0, max)}…`
}

function isToolFailed(group: ToolGroup): boolean {
  if (group.error) return true
  return isToolResultFailure(group.result)
}

function toolDisplayLabel(group: ToolGroup): string {
  if (!isAgentSpawnTool(group.toolName)) {
    return toolLabel(group.toolName, isToolFailed(group))
  }
  if (isToolFailed(group)) {
    return toolLabel(group.toolName, true)
  }
  const agentId = spawnAgentIdFromInput(group.input)
  if (group.result?.trim()) {
    return agentId ? t('tool.ui.agent_result', agentId) : t('tool.ui.agent_result_generic')
  }
  return agentId ? t('tool.ui.agent_spawning', agentId) : toolLabel(group.toolName, false)
}

function dedupeSpawnBlocks(blocks: RenderBlock[]): RenderBlock[] {
  const completedAgents = new Set<string>()
  for (const block of blocks) {
    if (block.type !== 'tool' || !isAgentSpawnTool(block.group.toolName)) continue
    if (!block.group.result || isToolFailed(block.group)) continue
    const agentId = spawnAgentIdFromInput(block.group.input)
    if (agentId) completedAgents.add(agentId)
  }
  if (completedAgents.size === 0) return blocks
  return blocks.filter(block => {
    if (block.type !== 'spawn') return true
    if (block.spawn.status !== 'done') return true
    return !completedAgents.has(block.spawn.agentId)
  })
}

function toolRequestSummary(input?: string): string {
  if (!input?.trim()) return ''
  try {
    const obj = JSON.parse(input.trim())
    if (typeof obj === 'string') return truncate(obj)
    for (const key of SUMMARY_KEYS) {
      const val = obj[key]
      if (typeof val === 'string' && val.trim()) return truncate(val)
    }
    const first = Object.values(obj).find(v => typeof v === 'string' && (v as string).trim())
    if (first) return truncate(String(first))
    return truncate(JSON.stringify(obj))
  } catch {
    return truncate(input)
  }
}

type SpawnGroup = {
  id: string
  agentId: string
  status: string
  summary?: string
  childSessionId?: string
}

type RenderBlock =
  | { type: 'text', text: string }
  | { type: 'tool', group: ToolGroup }
  | { type: 'spawn', spawn: SpawnGroup }

function upsertSpawnBlock(blocks: RenderBlock[], step: ChatStepEvent, index: number) {
  const agentId = step.agentId ?? 'agent'
  const existing = blocks.findIndex(
    block => block.type === 'spawn' && block.spawn.agentId === agentId,
  )
  const next: SpawnGroup = {
    id: `spawn-${agentId}`,
    agentId,
    status: step.spawnStatus ?? 'running',
    summary: step.spawnSummary,
    childSessionId: step.childSessionId,
  }
  if (existing >= 0) {
    const prev = blocks[existing]
    if (prev?.type === 'spawn') {
      blocks[existing] = {
        type: 'spawn',
        spawn: { ...prev.spawn, ...next },
      }
    }
    return
  }
  blocks.push({ type: 'spawn', spawn: next })
}

function spawnStatusLabel(status: string): string {
  const key = `chat.spawn.${status}`
  const label = t(key)
  return label === key ? status : label
}

function spawnStatusClass(status: string): string {
  if (status === 'done') return 'text-vc-success'
  if (status === 'failed') return 'text-vc-error'
  return 'text-vc-accent'
}

function buildToolGroups(steps: ChatStepEvent[]): ToolGroup[] {
  const groups: ToolGroup[] = []
  for (const step of steps) {
    if (step.type === 'tool_call') {
      groups.push({
        id: `c-${groups.length}`,
        toolName: step.toolName ?? 'tool',
        input: step.toolInput,
      })
    } else if (step.type === 'tool_result') {
      const last = groups[groups.length - 1]
      if (last && last.toolName === step.toolName && !last.result) {
        last.result = step.toolResult
      } else {
        groups.push({
          id: `r-${groups.length}`,
          toolName: step.toolName ?? 'tool',
          result: step.toolResult,
        })
      }
    } else if (step.type === 'spawn') {
      groups.push({
        id: `s-${groups.length}`,
        toolName: 'agent_spawn',
        input: JSON.stringify({ agent_id: step.agentId, status: step.spawnStatus }),
        result: step.spawnSummary,
      })
    } else if (step.type === 'error') {
      groups.push({ id: `e-${groups.length}`, toolName: 'error', error: step.error })
    }
  }
  return groups
}

function splitLegacyTextAroundTools(text: string, toolCount: number): [string, string] | null {
  if (!text.trim() || toolCount === 0) return null
  const trimmed = text.trim()

  const patterns = [
    /\n\n+(?=(Xong|Done|OK|Đã |Hoàn |Finished|Successfully))/i,
    /(?<=[.!?…])\s+(?=(Xong|Done|OK|Đã |Hoàn |Finished))/i,
    /(?<=[😎✅👍])\s+(?=(Xong|Done|OK|Đã |Hoàn |Finished))/i,
  ]
  for (const pattern of patterns) {
    const match = pattern.exec(trimmed)
    if (match && match.index > 0) {
      const before = trimmed.slice(0, match.index).trim()
      const after = trimmed.slice(match.index).trim()
      if (before && after) return [before, after]
    }
  }
  return null
}

function buildRenderBlocks(msg: ChatItem): RenderBlock[] {
  const hasTextSteps = msg.steps.some(step => step.type === 'text')
  const toolGroups = buildToolGroups(msg.steps)

  if (!hasTextSteps) {
    const split = splitLegacyTextAroundTools(msg.text, toolGroups.length)
    if (split) {
      return [
        { type: 'text', text: split[0] },
        ...toolGroups.map(group => ({ type: 'tool' as const, group })),
        { type: 'text', text: split[1] },
      ]
    }
    const blocks: RenderBlock[] = []
    if (msg.text.trim()) blocks.push({ type: 'text', text: msg.text })
    for (const group of toolGroups) blocks.push({ type: 'tool', group })
    return dedupeSpawnBlocks(blocks)
  }

  const blocks: RenderBlock[] = []
  for (let i = 0; i < msg.steps.length; i++) {
    const step = msg.steps[i]
    if (!step) continue
    if (step.type === 'text' && step.text) {
      blocks.push({ type: 'text', text: step.text })
    } else if (step.type === 'tool_call') {
      const group: ToolGroup = {
        id: `c-${i}`,
        toolName: step.toolName ?? 'tool',
        input: step.toolInput,
      }
      const next = msg.steps[i + 1]
      if (next?.type === 'tool_result' && next.toolName === step.toolName) {
        group.result = next.toolResult
        i++
      }
      blocks.push({ type: 'tool', group })
    } else if (step.type === 'tool_result') {
      blocks.push({
        type: 'tool',
        group: {
          id: `r-${i}`,
          toolName: step.toolName ?? 'tool',
          result: step.toolResult,
        },
      })
    } else if (step.type === 'spawn') {
      upsertSpawnBlock(blocks, step, i)
    } else if (step.type === 'error') {
      blocks.push({
        type: 'tool',
        group: { id: `e-${i}`, toolName: 'error', error: step.error },
      })
    }
  }
  return dedupeSpawnBlocks(blocks)
}

function toolIcon(name: string) {
  const n = name.toLowerCase()
  if (n.includes('agent') || n.startsWith('spawn:')) return Users
  if (n.includes('web') || n.includes('search') || n.includes('fetch') || n.includes('http')) return Globe
  if (n.includes('shell') || n.includes('exec') || n.includes('cmd')) return Terminal
  if (n.includes('file') || n.includes('read') || n.includes('grep')) return FileText
  if (n.includes('dir') || n.includes('folder')) return FolderOpen
  if (n.includes('find')) return Search
  return Wrench
}

function toolExpandKey(msgIdx: number, groupId: string) {
  return `${msgIdx}:${groupId}`
}

function isToolExpanded(msgIdx: number, groupId: string) {
  return expandedTools.value.has(toolExpandKey(msgIdx, groupId))
}

function toggleToolExpand(msgIdx: number, groupId: string) {
  const key = toolExpandKey(msgIdx, groupId)
  const next = new Set(expandedTools.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  expandedTools.value = next
}

function renderMath(html: string): string {
  let out = html.replace(/\$\$([\s\S]+?)\$\$/g, (_, tex) => {
    try {
      return katex.renderToString(tex.trim(), { displayMode: true, throwOnError: false })
    } catch {
      return `$$${tex}$$`
    }
  })
  out = out.replace(/\$([^$\n]+?)\$/g, (_, tex) => {
    try {
      return katex.renderToString(tex.trim(), { displayMode: false, throwOnError: false })
    } catch {
      return `$${tex}$`
    }
  })
  return out
}

function renderMarkdown(text: string): string {
  try {
    return renderMath(marked.parse(text) as string)
  } catch {
    return text
  }
}

function autoResize(el: HTMLTextAreaElement) {
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 192) + 'px'
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

function applySuggestion(text: string) {
  chatInput.value = text
  if (textareaRef.value) autoResize(textareaRef.value)
  textareaRef.value?.focus()
}

async function handleSend() {
  const text = chatInput.value.trim()
  if (!text || isGenerating.value) return
  chatInput.value = ''
  if (textareaRef.value) textareaRef.value.style.height = 'auto'
  stickToBottom.value = true
  void sendMessage(text)
  await nextTick()
  scrollToBottom(true)
}

function isNearBottom(): boolean {
  const el = chatBox.value
  if (!el) return true
  return el.scrollHeight - el.scrollTop - el.clientHeight <= SCROLL_STICK_THRESHOLD
}

function onChatScroll() {
  stickToBottom.value = isNearBottom()
}

function scrollToBottom(force = false) {
  if (!force && !stickToBottom.value) return
  const el = chatBox.value
  if (!el) return
  el.scrollTop = el.scrollHeight
}

function enhanceMarkdownCode(el: Element) {
  enhanceCodeBlocks(el, async (text) => {
    await window.navigator.clipboard.writeText(text)
    toast.add(t('chat.copied'), 'success')
  }, t('chat.copy'))
}

async function copyMessage(text: string) {
  await window.navigator.clipboard.writeText(text)
  toast.add(t('chat.copied'), 'success')
}

const session = computed(() => currentSession())
const messages = computed(() => session.value?.messages || [])
const isReadOnly = computed(() => session.value?.readOnly === true)

function isStreamingMessage(idx: number) {
  return isGenerating.value
    && idx === messages.value.length - 1
    && messages.value[idx]?.role === 'assistant'
}

function messageBlocks(msg: ChatItem): RenderBlock[] {
  return buildRenderBlocks(msg)
}

watch(
  () => messages.value.map(msg =>
    `${msg.role}:${msg.text.length}:${msg.steps.map(s => `${s.type}:${s.text?.length ?? 0}`).join(',')}`
  ).join('|'),
  async () => {
    await nextTick()
    scrollToBottom()
  },
  { flush: 'post' }
)

watch(currentSessionId, () => {
  stickToBottom.value = true
})
</script>

<template>
  <div class="flex h-full min-h-0 flex-col">
    <div ref="chatBox" class="min-h-0 flex-1 overflow-y-auto vc-scrollbar" @scroll="onChatScroll">
      <div
        v-if="messages.length === 0 && isReadOnly"
        class="mx-auto flex h-full max-w-2xl flex-col justify-center px-5 pb-16 md:px-8"
      >
        <p class="text-sm text-vc-text-muted">{{ t('nav.subAgentEmpty') }}</p>
      </div>

      <div
        v-else-if="messages.length === 0"
        class="mx-auto flex h-full max-w-2xl flex-col justify-center px-5 pb-16 md:px-8"
      >
        <p class="vc-display vc-fade-up text-3xl font-medium text-vc-text md:text-4xl" style="text-wrap: balance">
          {{ t('chat.greeting').slice(0, -1) }}<span class="text-vc-accent">.</span>
        </p>
        <p class="vc-fade-up vc-fade-up-1 mt-3 max-w-md text-[15px] leading-relaxed text-vc-text-secondary">
          {{ t('chat.subtitle') }}
        </p>

        <div class="mt-10 grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
          <button
            v-for="(item, i) in suggestions"
            :key="item.label"
            type="button"
            class="vc-fade-up group rounded-2xl border border-vc-border-subtle bg-vc-surface p-4 text-left transition-all duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] hover:-translate-y-0.5 hover:border-vc-border hover:shadow-[var(--vc-shadow-md)] active:scale-[0.98]"
            :class="`vc-fade-up-${i + 1}`"
            @click="applySuggestion(item.text)"
          >
            <span class="vc-eyebrow block transition-colors duration-300 group-hover:text-vc-accent">
              {{ item.label }}
            </span>
            <span class="mt-1.5 block text-sm leading-relaxed text-vc-text-secondary transition-colors duration-300 group-hover:text-vc-text">
              {{ item.text }}
            </span>
          </button>
        </div>
      </div>

      <div v-else class="mx-auto max-w-2xl space-y-10 px-4 py-6 md:px-8">
        <template v-for="(msg, idx) in messages" :key="idx">
          <div v-if="msg.role === 'user'" class="flex justify-end">
            <div class="max-w-[85%] rounded-2xl rounded-br-md bg-vc-user px-4 py-2.5 text-[15px] leading-relaxed text-vc-text shadow-[var(--vc-shadow-sm)]">
              <p class="whitespace-pre-wrap">{{ msg.text }}</p>
            </div>
          </div>

          <div v-else class="space-y-3">
            <p
              v-if="isStreamingMessage(idx) && messageBlocks(msg).length === 0"
              class="flex items-center gap-2.5 text-sm text-vc-text-muted"
            >
              <span class="vc-thinking" aria-hidden="true"><span /><span /><span /></span>
              {{ t('chat.thinking') }}
            </p>

            <template v-for="(block, bi) in messageBlocks(msg)" :key="`${idx}-${bi}`">
              <div v-if="block.type === 'text'" class="relative">
                <div
                  v-if="isStreamingMessage(idx) && bi === messageBlocks(msg).length - 1"
                  class="text-[15px] leading-relaxed whitespace-pre-wrap text-vc-text"
                >
                  {{ block.text }}<span class="ml-0.5 inline-block h-4 w-0.5 animate-pulse bg-vc-accent align-middle" />
                </div>
                <div
                  v-else
                  class="prose max-w-none"
                  v-html="renderMarkdown(block.text)"
                  v-html-hook="enhanceMarkdownCode"
                />
              </div>

              <div v-else-if="block.type === 'spawn'" class="text-sm leading-relaxed">
                <div class="flex items-start gap-2 text-vc-text-muted">
                  <Users :size="14" class="mt-0.5 shrink-0" :stroke-width="1.75" />
                  <div class="min-w-0 flex-1">
                    <div class="flex flex-wrap items-center gap-1.5">
                      <NuxtLink
                        v-if="block.spawn.childSessionId"
                        :to="sessionPath(block.spawn.childSessionId)"
                        class="font-medium text-vc-text-secondary transition-colors hover:text-vc-accent"
                      >
                        {{ block.spawn.agentId }}
                      </NuxtLink>
                      <span v-else class="font-medium text-vc-text-secondary">{{ block.spawn.agentId }}</span>
                      <span class="text-xs" :class="spawnStatusClass(block.spawn.status)">
                        {{ spawnStatusLabel(block.spawn.status) }}
                      </span>
                    </div>
                    <p v-if="block.spawn.summary" class="mt-1 text-vc-text-secondary">
                      {{ truncate(block.spawn.summary, 160) }}
                    </p>
                  </div>
                </div>
              </div>

              <div v-else-if="block.type === 'tool'" class="text-sm leading-relaxed">
                <div v-if="block.group.error" class="flex items-center gap-2 text-vc-error">
                  <AlertCircle :size="14" class="shrink-0" :stroke-width="1.75" />
                  <span>{{ block.group.error }}</span>
                </div>
                <div v-else class="flex items-start gap-2" :class="isToolFailed(block.group) ? 'text-vc-error' : 'text-vc-text-muted'">
                  <component
                    :is="isToolFailed(block.group) ? AlertCircle : toolIcon(block.group.toolName)"
                    :size="14"
                    class="mt-0.5 shrink-0"
                    :stroke-width="1.75"
                  />
                  <div class="min-w-0 flex-1">
                    <button
                      type="button"
                      class="group flex max-w-full items-center gap-1.5 text-left transition-colors hover:text-vc-text-secondary"
                      @click="toggleToolExpand(idx, block.group.id)"
                    >
                      <span
                        class="shrink-0 font-medium group-hover:text-vc-text"
                        :class="isToolFailed(block.group) ? 'text-vc-error' : 'text-vc-text-secondary'"
                      >
                        {{ toolDisplayLabel(block.group) }}
                      </span>
                      <template v-if="!block.group.result && toolRequestSummary(block.group.input)">
                        <span class="truncate" :class="isToolFailed(block.group) ? 'text-vc-error/80' : 'text-vc-text-muted'">
                          - {{ toolRequestSummary(block.group.input) }}
                        </span>
                      </template>
                      <template v-else-if="block.group.result && isAgentSpawnTool(block.group.toolName) && !isToolFailed(block.group)">
                        <span class="truncate text-vc-text-muted">
                          - {{ truncate(stripSpawnResultPrefix(block.group.result), 72) }}
                        </span>
                      </template>
                      <ChevronDown
                        :size="14"
                        class="shrink-0 transition-transform"
                        :class="{ 'rotate-180': isToolExpanded(idx, block.group.id) }"
                        :stroke-width="1.75"
                      />
                    </button>
                    <p
                      v-if="isToolFailed(block.group) && toolFailureMessage(block.group.result)"
                      class="mt-1 text-xs leading-relaxed text-vc-error"
                    >
                      {{ truncate(toolFailureMessage(block.group.result), 200) }}
                    </p>
                    <div
                      v-if="isToolExpanded(idx, block.group.id)"
                      class="mt-2 space-y-3 border-l-2 border-vc-accent/30 pl-3"
                    >
                      <ToolDetailBody
                        v-if="block.group.input"
                        :tool-name="block.group.toolName"
                        :raw="block.group.input"
                        side="input"
                        :label="t('tool.call_detail')"
                      />
                      <ToolDetailBody
                        v-if="block.group.result"
                        :tool-name="block.group.toolName"
                        :raw="block.group.result"
                        :input-raw="block.group.input"
                        side="result"
                        :label="t('tool.result_detail')"
                      />
                    </div>
                  </div>
                </div>
              </div>
            </template>

            <button
              v-if="msg.text && !isStreamingMessage(idx)"
              type="button"
              class="flex items-center gap-1.5 text-xs text-vc-text-muted transition-colors hover:text-vc-text-secondary"
              @click="copyMessage(msg.text)"
            >
              <Copy :size="12" :stroke-width="1.75" /> {{ t('chat.copy') }}
            </button>
          </div>
        </template>
      </div>
    </div>

    <div v-if="!isReadOnly" class="shrink-0 px-4 pb-4 pt-2 md:px-8 md:pb-6">
      <div class="mx-auto max-w-2xl">
        <div class="vc-composer flex items-end gap-1 py-2 pl-4 pr-2">
          <textarea
            ref="textareaRef"
            v-model="chatInput"
            rows="1"
            :placeholder="t('chat.placeholder')"
            class="vc-composer-input max-h-32 min-h-[36px] flex-1 resize-none bg-transparent py-1.5 text-[15px] leading-snug text-vc-text placeholder:text-vc-text-muted focus:outline-none"
            @input="autoResize($event.target as HTMLTextAreaElement)"
            @keydown="onKeydown"
          />
          <div class="flex shrink-0 items-center gap-1 pb-0.5">
            <div
              v-if="catalogModels.length"
              ref="modelMenuRef"
              class="vc-composer-model"
            >
              <button
                type="button"
                class="vc-composer-model-btn"
                :aria-expanded="modelMenuOpen"
                :aria-label="t('chat.model')"
                @click.stop="modelMenuOpen = !modelMenuOpen"
              >
                <span class="max-w-[7.5rem] truncate">{{ selectedModelLabel }}</span>
                <ChevronDown
                  :size="14"
                  :stroke-width="2"
                  class="shrink-0 transition-transform duration-200"
                  :class="{ 'rotate-180': modelMenuOpen }"
                />
              </button>
              <Transition name="vc-model-menu">
                <div v-if="modelMenuOpen" class="vc-composer-model-menu">
                  <button
                    v-for="m in catalogModels"
                    :key="m.id"
                    type="button"
                    class="vc-composer-model-item"
                    :class="{ 'is-active': m.id === selectedCatalogId }"
                    @click="pickModel(m.id)"
                  >
                    <span class="truncate">{{ m.label || m.id }}</span>
                    <Check v-if="m.id === selectedCatalogId" :size="14" :stroke-width="2.25" class="shrink-0 opacity-70" />
                  </button>
                  <div class="vc-composer-model-divider" />
                  <NuxtLink
                    to="/settings/models"
                    class="vc-composer-model-item vc-composer-model-config"
                    @click="modelMenuOpen = false"
                  >
                    <Settings :size="14" :stroke-width="1.75" class="shrink-0 opacity-60" />
                    <span>{{ t('chat.modelConfigure') }}</span>
                  </NuxtLink>
                </div>
              </Transition>
            </div>
            <button
              v-if="messages.length > 0"
              type="button"
              class="vc-composer-btn"
              :title="t('chat.clearSession')"
              @click="clearSessionMessages()"
            >
              <RefreshCw :size="15" :stroke-width="1.75" />
            </button>
            <button
              v-if="isGenerating"
              type="button"
              class="vc-composer-stop"
              :title="t('chat.stop')"
              @click="stopGeneration()"
            >
              <Square :size="14" :stroke-width="0" fill="currentColor" />
            </button>
            <button
              v-else
              type="button"
              class="vc-composer-send"
              :disabled="!chatInput.trim()"
              @click="handleSend"
            >
              <ArrowUp :size="16" :stroke-width="2.25" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
export default {
  directives: {
    htmlHook: {
      mounted(el: HTMLElement, binding: { value?: (el: Element) => void }) {
        if (binding.value) binding.value(el)
      },
      updated(el: HTMLElement, binding: { value?: (el: Element) => void }) {
        if (binding.value) binding.value(el)
      }
    }
  }
}
</script>
