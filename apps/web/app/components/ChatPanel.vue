<script setup lang="ts">
import {
  AlertCircle,
  ArrowUp,
  ChevronDown,
  Copy,
  FileText,
  FolderOpen,
  Globe,
  RefreshCw,
  Search,
  Sparkles,
  Terminal,
  Wrench,
} from '@lucide/vue'
import katex from 'katex'
import { marked } from 'marked'
import hljs from 'highlight.js'
import type { ChatStepEvent } from '~/composables/useChat'

const { currentSession, currentSessionId, isGenerating, sendMessage, clearSessionMessages } = useChat()
const { t, toolLabel } = useI18n()
const toast = useToast()

const chatInput = ref('')
const chatBox = ref<HTMLElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const expandedTools = ref<Set<string>>(new Set())
const stickToBottom = ref(true)

const SCROLL_STICK_THRESHOLD = 96

const suggestions = [
  { label: 'Nhớ sở thích', text: 'Nhớ giúp t: t thích deploy bằng Docker và tiết kiệm token' },
  { label: 'Tìm web', text: 'Tìm trên web VPS rẻ ở Việt Nam, tóm tắt 3 lựa chọn' },
  { label: 'Ủy researcher', text: '@researcher so sánh Redis và SQLite cho app chat nhỏ' },
  { label: 'Đọc workspace', text: 'Đọc README trong workspace và cho t biết project làm gì' },
]

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

function formatToolBody(raw?: string, max = 8000): string {
  if (!raw) return ''
  const trimmed = raw.trim()
  if (!trimmed) return ''
  try {
    const pretty = JSON.stringify(JSON.parse(trimmed), null, 2)
    return pretty.length > max ? `${pretty.slice(0, max)}…` : pretty
  } catch {
    return trimmed.length > max ? `${trimmed.slice(0, max)}…` : trimmed
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

function highlightCode(el: Element) {
  el.querySelectorAll('pre code').forEach((block) => {
    hljs.highlightElement(block as HTMLElement)
  })
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

watch(
  () => messages.value.map(msg => `${msg.role}:${msg.text.length}:${msg.steps.length}`).join('|'),
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
  <div class="flex h-full flex-1 flex-col">
    <div ref="chatBox" class="flex-1 overflow-y-auto p-4 md:p-8 vc-scrollbar" @scroll="onChatScroll">
      <div
        v-if="messages.length === 0"
        class="mx-auto flex h-full max-w-2xl flex-col justify-center px-2 py-12"
      >
        <div class="mb-6 flex items-center gap-3">
          <div class="flex h-11 w-11 items-center justify-center rounded-xl border border-zinc-800 bg-zinc-900/50">
            <Sparkles :size="22" class="text-zinc-300" />
          </div>
          <div>
            <h1 class="text-lg font-semibold tracking-tight text-zinc-100">Chỉ cần prompt</h1>
            <p class="text-sm text-zinc-500">Nhắn bình thường — agent tự nhớ, tìm, chạy tool, ủy agent khác.</p>
          </div>
        </div>

        <div class="grid gap-2 sm:grid-cols-2">
          <button
            v-for="item in suggestions"
            :key="item.label"
            type="button"
            class="group rounded-lg border border-zinc-800/80 bg-zinc-950/50 p-3.5 text-left transition-colors hover:border-zinc-600 hover:bg-zinc-900/40"
            @click="applySuggestion(item.text)"
          >
            <span class="text-[10px] font-medium uppercase tracking-wider text-zinc-500">{{ item.label }}</span>
            <p class="mt-1.5 text-xs leading-relaxed text-zinc-300 group-hover:text-zinc-100">{{ item.text }}</p>
          </button>
        </div>

        <p class="mt-8 text-center text-[11px] text-zinc-600">
          Không cần cấu hình trước. Cần memory / providers / budget → mở <strong class="text-zinc-500">Công cụ nâng cao</strong> ở sidebar.
        </p>
      </div>

      <div class="mx-auto max-w-3xl space-y-6">
        <template v-for="(msg, idx) in messages" :key="idx">
          <!-- User bubble -->
          <div v-if="msg.role === 'user'" class="flex justify-end">
            <div class="max-w-[88%] rounded-2xl rounded-br-md border border-zinc-800 bg-zinc-900/80 px-4 py-2.5 text-sm leading-relaxed text-zinc-100 shadow-sm shadow-black/20">
              <p class="whitespace-pre-wrap">{{ msg.text }}</p>
            </div>
          </div>

          <!-- Assistant: plain markdown + compact tool lines -->
          <div v-else class="space-y-3">
            <p
              v-if="isStreamingMessage(idx) && !msg.text && msg.steps.length === 0"
              class="text-xs text-zinc-500"
            >
              <span class="inline-flex items-center gap-2">
                <span class="h-1 w-1 animate-pulse rounded-full bg-zinc-500" />
                {{ t('chat.thinking') }}
              </span>
            </p>

            <div v-if="buildToolGroups(msg.steps).length > 0" class="space-y-1">
              <div
                v-for="group in buildToolGroups(msg.steps)"
                :key="group.id"
                class="text-xs leading-relaxed"
              >
                <div v-if="group.error" class="flex items-center gap-2 text-rose-400/90">
                  <AlertCircle :size="13" class="shrink-0" />
                  <span>{{ group.error }}</span>
                </div>
                <div v-else class="flex items-start gap-2 text-zinc-500">
                  <component
                    :is="toolIcon(group.toolName)"
                    :size="13"
                    class="mt-0.5 shrink-0 text-zinc-600"
                  />
                  <div class="min-w-0 flex-1">
                    <button
                      type="button"
                      class="group flex max-w-full items-center gap-1.5 text-left transition-colors hover:text-zinc-300"
                      @click="toggleToolExpand(idx, group.id)"
                    >
                      <span class="shrink-0 text-zinc-400 group-hover:text-zinc-300">{{ toolLabel(group.toolName) }}</span>
                      <template v-if="toolRequestSummary(group.input)">
                        <span class="shrink-0 text-zinc-600">:</span>
                        <span class="truncate text-zinc-500 group-hover:text-zinc-400">{{ toolRequestSummary(group.input) }}</span>
                      </template>
                      <ChevronDown
                        :size="13"
                        class="shrink-0 text-zinc-600 transition-transform group-hover:text-zinc-400"
                        :class="{ 'rotate-180': isToolExpanded(idx, group.id) }"
                      />
                    </button>
                    <div
                      v-if="isToolExpanded(idx, group.id)"
                      class="mt-2 space-y-2 border-l border-zinc-800 pl-3"
                    >
                      <div v-if="group.input">
                        <p class="mb-1 text-[10px] uppercase tracking-wider text-zinc-600">{{ t('tool.call_detail') }}</p>
                        <pre class="max-h-52 overflow-auto font-mono text-[10px] leading-relaxed whitespace-pre-wrap text-zinc-400">{{ formatToolBody(group.input) }}</pre>
                      </div>
                      <div v-if="group.result">
                        <p class="mb-1 text-[10px] uppercase tracking-wider text-zinc-600">{{ t('tool.result_detail') }}</p>
                        <pre class="max-h-64 overflow-auto font-mono text-[10px] leading-relaxed whitespace-pre-wrap text-zinc-400">{{ formatToolBody(group.result) }}</pre>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="msg.text" class="relative">
              <div
                v-if="isStreamingMessage(idx)"
                class="text-sm leading-relaxed whitespace-pre-wrap text-zinc-200"
              >
                {{ msg.text }}<span class="ml-0.5 inline-block h-4 w-0.5 animate-pulse bg-zinc-500 align-middle" />
              </div>
              <div
                v-else
                class="prose prose-invert max-w-none text-sm"
                v-html="renderMarkdown(msg.text)"
                v-html-hook="highlightCode"
              />
              <button
                v-if="!isStreamingMessage(idx)"
                type="button"
                class="mt-2 flex items-center gap-1 text-[10px] text-zinc-600 transition-colors hover:text-zinc-400"
                @click="copyMessage(msg.text)"
              >
                <Copy :size="11" /> {{ t('chat.copy') }}
              </button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <div class="border-t border-zinc-800/60 bg-zinc-950/80 px-4 py-4 md:px-8">
      <div class="mx-auto max-w-3xl">
        <div class="flex flex-col rounded-xl border border-zinc-800 bg-zinc-900/40 p-2 shadow-lg shadow-black/20 focus-within:border-zinc-600">
          <textarea
            ref="textareaRef"
            v-model="chatInput"
            rows="1"
            placeholder="Nhắn gì cũng được — nhớ, tìm, code, ủy agent…"
            class="max-h-48 min-h-[40px] w-full resize-none bg-transparent px-3 py-2 text-sm leading-relaxed text-zinc-100 placeholder-zinc-600 focus:outline-none"
            @input="autoResize($event.target as HTMLTextAreaElement)"
            @keydown="onKeydown"
          />
          <div class="flex items-center justify-between px-2 pt-1">
            <span class="text-[10px] text-zinc-600">Enter gửi · Shift+Enter xuống dòng</span>
            <div class="flex items-center gap-1">
              <button
                type="button"
                class="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-zinc-800 hover:text-zinc-300"
                title="Xóa tin nhắn phiên này"
                @click="clearSessionMessages()"
              >
                <RefreshCw :size="15" />
              </button>
              <button
                type="button"
                class="flex h-9 w-9 items-center justify-center rounded-lg bg-zinc-100 text-zinc-950 transition-colors hover:bg-white disabled:opacity-30"
                :disabled="isGenerating || !chatInput.trim()"
                @click="handleSend"
              >
                <ArrowUp :size="18" />
              </button>
            </div>
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
