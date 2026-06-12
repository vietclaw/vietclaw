<script setup lang="ts">
import {
  AlertCircle,
  ArrowUp,
  ChevronDown,
  Square,
  Copy,
  FileText,
  FolderOpen,
  Globe,
  RefreshCw,
  Search,
  Terminal,
  Wrench,
} from '@lucide/vue'
import katex from 'katex'
import { marked } from 'marked'
import type { ChatItem, ChatStepEvent } from '~/composables/useChat'
import { enhanceCodeBlocks } from '~/utils/enhanceCodeBlocks'

const { currentSession, currentSessionId, isGenerating, sendMessage, clearSessionMessages, stopGeneration } = useChat()
const { t, toolLabel } = useI18n()
const toast = useToast()

const chatInput = ref('')
const chatBox = ref<HTMLElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const expandedTools = ref<Set<string>>(new Set())
const stickToBottom = ref(true)

const SCROLL_STICK_THRESHOLD = 96

const suggestions = computed(() => [
  { label: t('chat.suggestion.remember.label'), text: t('chat.suggestion.remember.text') },
  { label: t('chat.suggestion.search.label'), text: t('chat.suggestion.search.text') },
  { label: t('chat.suggestion.delegate.label'), text: t('chat.suggestion.delegate.text') },
  { label: t('chat.suggestion.workspace.label'), text: t('chat.suggestion.workspace.text') },
])

const SUMMARY_KEYS = ['query', 'command', 'cmd', 'path', 'file', 'url', 'name', 'input', 'text', 'pattern', 'expression', 'message', 'prompt']

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

type RenderBlock =
  | { type: 'text', text: string }
  | { type: 'tool', group: ToolGroup }

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
    return blocks
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
    } else if (step.type === 'error') {
      blocks.push({
        type: 'tool',
        group: { id: `e-${i}`, toolName: 'error', error: step.error },
      })
    }
  }
  return blocks
}

function toolIcon(name: string) {
  const n = name.toLowerCase()
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
        v-if="messages.length === 0"
        class="mx-auto flex h-full max-w-2xl flex-col justify-center px-5 pb-16 md:px-8"
      >
        <p class="vc-display vc-fade-up text-3xl font-medium text-vc-text md:text-4xl" style="text-wrap: balance">
          {{ t('chat.greeting').slice(0, -1) }}<span class="text-vc-accent">.</span>
        </p>
        <p class="vc-fade-up vc-fade-up-1 mt-3 max-w-md text-[15px] leading-relaxed text-vc-text-secondary">
          {{ t('chat.subtitle') }}
        </p>

        <div class="mt-10 grid gap-2.5 sm:grid-cols-2">
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

              <div v-else class="text-sm leading-relaxed">
                <div v-if="block.group.error" class="flex items-center gap-2 text-vc-error">
                  <AlertCircle :size="14" class="shrink-0" :stroke-width="1.75" />
                  <span>{{ block.group.error }}</span>
                </div>
                <div v-else class="flex items-start gap-2 text-vc-text-muted">
                  <component
                    :is="toolIcon(block.group.toolName)"
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
                      <span class="shrink-0 font-medium text-vc-text-secondary group-hover:text-vc-text">
                        {{ toolLabel(block.group.toolName) }}
                      </span>
                      <template v-if="toolRequestSummary(block.group.input)">
                        <span class="truncate text-vc-text-muted">- {{ toolRequestSummary(block.group.input) }}</span>
                      </template>
                      <ChevronDown
                        :size="14"
                        class="shrink-0 transition-transform"
                        :class="{ 'rotate-180': isToolExpanded(idx, block.group.id) }"
                        :stroke-width="1.75"
                      />
                    </button>
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

    <div class="shrink-0 px-4 pb-4 pt-2 md:px-8 md:pb-6">
      <div class="mx-auto max-w-2xl">
        <div class="vc-composer flex items-center gap-1 py-2 pl-4 pr-2">
          <textarea
            ref="textareaRef"
            v-model="chatInput"
            rows="1"
            :placeholder="t('chat.placeholder')"
            class="vc-composer-input max-h-32 min-h-[36px] flex-1 resize-none bg-transparent py-1.5 text-[15px] leading-snug text-vc-text placeholder:text-vc-text-muted focus:outline-none"
            @input="autoResize($event.target as HTMLTextAreaElement)"
            @keydown="onKeydown"
          />
          <div class="flex shrink-0 items-center gap-0.5">
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
