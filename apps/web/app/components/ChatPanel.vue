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
    <div ref="chatBox" class="flex-1 overflow-y-auto vc-scrollbar" @scroll="onChatScroll">
      <div
        v-if="messages.length === 0"
        class="mx-auto max-w-2xl px-4 pt-12 pb-8 md:px-8 md:pt-16"
      >
        <p class="text-lg font-semibold tracking-tight text-vc-text">
          Xin chào.
        </p>
        <p class="mt-2 max-w-md text-[15px] text-vc-text-secondary">
          Gõ bất cứ gì. Agent sẽ nhớ, tìm, chạy lệnh hoặc ủy việc khi cần.
        </p>

        <ul class="mt-8 space-y-1 border-t border-vc-border-subtle pt-6">
          <li v-for="item in suggestions" :key="item.label">
            <button
              type="button"
              class="w-full rounded-md px-2 py-2.5 text-left text-sm text-vc-text-secondary transition-colors hover:bg-vc-bg-subtle hover:text-vc-text"
              @click="applySuggestion(item.text)"
            >
              {{ item.text }}
            </button>
          </li>
        </ul>
      </div>

      <div class="mx-auto max-w-2xl space-y-10 px-4 py-6 md:px-8">
        <template v-for="(msg, idx) in messages" :key="idx">
          <div v-if="msg.role === 'user'" class="flex justify-end">
            <div class="max-w-[85%] rounded-xl bg-vc-user px-4 py-2.5 text-[15px] leading-relaxed text-vc-text">
              <p class="whitespace-pre-wrap">{{ msg.text }}</p>
            </div>
          </div>

          <div v-else class="space-y-3">
            <p
              v-if="isStreamingMessage(idx) && !msg.text && msg.steps.length === 0"
              class="text-sm text-vc-text-muted"
            >
              {{ t('chat.thinking') }}
            </p>

            <div v-if="buildToolGroups(msg.steps).length > 0" class="space-y-1.5">
              <div
                v-for="group in buildToolGroups(msg.steps)"
                :key="group.id"
                class="text-sm leading-relaxed"
              >
                <div v-if="group.error" class="flex items-center gap-2 text-vc-error">
                  <AlertCircle :size="14" class="shrink-0" :stroke-width="1.75" />
                  <span>{{ group.error }}</span>
                </div>
                <div v-else class="flex items-start gap-2 text-vc-text-muted">
                  <component
                    :is="toolIcon(group.toolName)"
                    :size="14"
                    class="mt-0.5 shrink-0"
                    :stroke-width="1.75"
                  />
                  <div class="min-w-0 flex-1">
                    <button
                      type="button"
                      class="group flex max-w-full items-center gap-1.5 text-left transition-colors hover:text-vc-text-secondary"
                      @click="toggleToolExpand(idx, group.id)"
                    >
                      <span class="shrink-0 font-medium text-vc-text-secondary group-hover:text-vc-text">
                        {{ toolLabel(group.toolName) }}
                      </span>
                      <template v-if="toolRequestSummary(group.input)">
                        <span class="truncate text-vc-text-muted">- {{ toolRequestSummary(group.input) }}</span>
                      </template>
                      <ChevronDown
                        :size="14"
                        class="shrink-0 transition-transform"
                        :class="{ 'rotate-180': isToolExpanded(idx, group.id) }"
                        :stroke-width="1.75"
                      />
                    </button>
                    <div
                      v-if="isToolExpanded(idx, group.id)"
                      class="mt-2 space-y-2 border-l-2 border-vc-border-subtle pl-3"
                    >
                      <div v-if="group.input">
                        <p class="mb-1 text-[11px] font-medium text-vc-text-muted">{{ t('tool.call_detail') }}</p>
                        <pre class="max-h-52 overflow-auto font-mono text-[11px] leading-relaxed whitespace-pre-wrap text-vc-text-secondary">{{ formatToolBody(group.input) }}</pre>
                      </div>
                      <div v-if="group.result">
                        <p class="mb-1 text-[11px] font-medium text-vc-text-muted">{{ t('tool.result_detail') }}</p>
                        <pre class="max-h-64 overflow-auto font-mono text-[11px] leading-relaxed whitespace-pre-wrap text-vc-text-secondary">{{ formatToolBody(group.result) }}</pre>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="msg.text" class="relative">
              <div
                v-if="isStreamingMessage(idx)"
                class="text-[15px] leading-relaxed whitespace-pre-wrap text-vc-text"
              >
                {{ msg.text }}<span class="ml-0.5 inline-block h-4 w-0.5 animate-pulse bg-vc-accent align-middle" />
              </div>
              <div
                v-else
                class="prose max-w-none"
                v-html="renderMarkdown(msg.text)"
                v-html-hook="highlightCode"
              />
              <button
                v-if="!isStreamingMessage(idx)"
                type="button"
                class="mt-2 flex items-center gap-1.5 text-xs text-vc-text-muted transition-colors hover:text-vc-text-secondary"
                @click="copyMessage(msg.text)"
              >
                <Copy :size="12" :stroke-width="1.75" /> {{ t('chat.copy') }}
              </button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <div class="shrink-0 px-4 pb-4 pt-2 md:px-8 md:pb-6">
      <div class="mx-auto max-w-2xl">
        <div class="vc-composer flex items-center gap-1 pl-3 pr-1.5 py-1.5">
          <textarea
            ref="textareaRef"
            v-model="chatInput"
            rows="1"
            placeholder="Nhắn tin nhắn..."
            class="vc-composer-input max-h-32 min-h-[36px] flex-1 resize-none bg-transparent py-1.5 text-[15px] leading-snug text-vc-text placeholder:text-vc-text-muted focus:outline-none"
            @input="autoResize($event.target as HTMLTextAreaElement)"
            @keydown="onKeydown"
          />
          <div class="flex shrink-0 items-center gap-0.5">
            <button
              v-if="messages.length > 0"
              type="button"
              class="vc-composer-btn"
              title="Xóa tin nhắn trong phiên"
              @click="clearSessionMessages()"
            >
              <RefreshCw :size="15" :stroke-width="1.75" />
            </button>
            <button
              type="button"
              class="vc-composer-send"
              :disabled="isGenerating || !chatInput.trim()"
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
