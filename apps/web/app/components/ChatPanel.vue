<script setup lang="ts">
import { ArrowUp, RefreshCw, Copy, Terminal, Wrench, CheckCircle2, AlertCircle } from '@lucide/vue'
import { marked } from 'marked'
import hljs from 'highlight.js'

const { currentSession, isGenerating, sendMessage, clearSessionMessages } = useChat()
const toast = useToast()

const chatInput = ref('')
const chatBox = ref<HTMLElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)

marked.setOptions({ breaks: true, gfm: true })

function renderMarkdown(text: string): string {
  try { return marked.parse(text) as string } catch { return text }
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

async function handleSend() {
  const text = chatInput.value.trim()
  if (!text || isGenerating.value) return
  chatInput.value = ''
  if (textareaRef.value) textareaRef.value.style.height = 'auto'
  await sendMessage(text)
  await nextTick()
  scrollToBottom()
}

function scrollToBottom() {
  if (chatBox.value) {
    chatBox.value.scrollTop = chatBox.value.scrollHeight
  }
}

function highlightCode(el: Element) {
  el.querySelectorAll('pre code').forEach((block) => {
    hljs.highlightElement(block as HTMLElement)
  })
}

async function copyMessage(text: string) {
  await window.navigator.clipboard.writeText(text)
  toast.add('Copied', 'success')
}

const session = computed(() => currentSession())
const messages = computed(() => session.value?.messages || [])
</script>

<template>
  <div class="flex-1 flex flex-col h-full">
    <!-- Chat Messages -->
    <div
      ref="chatBox"
      class="flex-1 overflow-y-auto p-4 md:p-6 vc-scrollbar"
      @vue:mounted="(el: any) => { if (el?.$el) highlightCode(el.$el) }"
    >
      <!-- Empty State -->
      <div
        v-if="messages.length === 0"
        class="max-w-xl mx-auto h-full flex flex-col justify-center py-20 px-4 text-center"
      >
        <div class="w-10 h-10 rounded border border-zinc-800 bg-zinc-900/40 flex items-center justify-center mx-auto mb-4">
          <Terminal :size="20" class="text-zinc-400" />
        </div>
        <h3 class="text-sm font-semibold tracking-tight text-zinc-200 mb-1">VietClaw Workspace</h3>
        <p class="text-zinc-500 text-xs max-w-sm mx-auto mb-6">Lightweight agent workspace. Ask anything or use tools.</p>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-2 max-w-md mx-auto text-left">
          <button
            class="p-3 rounded border border-zinc-900 bg-zinc-950/40 hover:border-zinc-700 cursor-pointer transition-all text-left"
            @click="chatInput = 'Giải thích về kiến trúc microservices'; autoResize($refs.textareaRef as any)"
          >
            <h4 class="text-xs font-mono text-zinc-300 flex items-center gap-1.5">
              <Terminal :size="14" class="text-zinc-500" /> microservices
            </h4>
            <p class="text-[10px] text-zinc-500 mt-1">Giải thích kiến trúc microservices...</p>
          </button>
          <button
            class="p-3 rounded border border-zinc-900 bg-zinc-950/40 hover:border-zinc-700 cursor-pointer transition-all text-left"
            @click="chatInput = 'Tìm thông tin về Go 1.25 release'; autoResize($refs.textareaRef as any)"
          >
            <h4 class="text-xs font-mono text-zinc-300 flex items-center gap-1.5">
              <Terminal :size="14" class="text-zinc-500" /> go_release
            </h4>
            <p class="text-[10px] text-zinc-500 mt-1">Tìm thông tin Go 1.25...</p>
          </button>
          <button
            class="p-3 rounded border border-zinc-900 bg-zinc-950/40 hover:border-zinc-700 cursor-pointer transition-all text-left"
            @click="chatInput = 'Đọc file config.json trong workspace'; autoResize($refs.textareaRef as any)"
          >
            <h4 class="text-xs font-mono text-zinc-300 flex items-center gap-1.5">
              <Terminal :size="14" class="text-zinc-500" /> read_file
            </h4>
            <p class="text-[10px] text-zinc-500 mt-1">Đọc file từ workspace...</p>
          </button>
          <button
            class="p-3 rounded border border-zinc-900 bg-zinc-950/40 hover:border-zinc-700 cursor-pointer transition-all text-left"
            @click="chatInput = 'Tìm file có chứa từ khóa error trong workspace'; autoResize($refs.textareaRef as any)"
          >
            <h4 class="text-xs font-mono text-zinc-300 flex items-center gap-1.5">
              <Terminal :size="14" class="text-zinc-500" /> grep_files
            </h4>
            <p class="text-[10px] text-zinc-500 mt-1">Tìm kiếm trong workspace...</p>
          </button>
        </div>
      </div>

      <!-- Messages -->
      <div class="space-y-6 max-w-3xl mx-auto">
        <div
          v-for="(msg, idx) in messages"
          :key="idx"
          class="flex gap-4"
          :class="msg.role === 'user' ? 'justify-end' : 'justify-start'"
        >
          <!-- AI Avatar -->
          <div
            v-if="msg.role === 'assistant'"
            class="w-7 h-7 rounded bg-zinc-100 flex items-center justify-center shrink-0 text-zinc-950"
          >
            <span class="text-[9px] font-bold">AI</span>
          </div>

          <!-- User Bubble -->
          <div
            v-if="msg.role === 'user'"
            class="px-3.5 py-2.5 rounded bg-zinc-900 border border-zinc-800 text-zinc-200 text-sm max-w-[85%] leading-relaxed"
          >
            <p class="whitespace-pre-wrap">{{ msg.text }}</p>
          </div>

          <!-- Assistant Bubble -->
          <div
            v-else
            class="max-w-[85%] space-y-2"
          >
            <!-- Step-by-step execution -->
            <div v-if="msg.steps.length > 0" class="space-y-1.5">
              <template v-for="(step, si) in msg.steps" :key="si">
                <!-- Tool Call -->
                <div
                  v-if="step.type === 'tool_call'"
                  class="flex items-center gap-2 px-3 py-1.5 rounded bg-amber-950/20 border border-amber-900/20 text-[11px]"
                >
                  <Wrench :size="12" class="text-amber-400 shrink-0" />
                  <span class="text-amber-300 font-mono font-medium">{{ step.toolName }}</span>
                  <span class="text-zinc-500 truncate">{{ step.toolInput }}</span>
                </div>

                <!-- Tool Result -->
                <div
                  v-else-if="step.type === 'tool_result'"
                  class="flex items-start gap-2 px-3 py-1.5 rounded bg-emerald-950/20 border border-emerald-900/20 text-[11px]"
                >
                  <CheckCircle2 :size="12" class="text-emerald-400 shrink-0 mt-0.5" />
                  <span class="text-emerald-300/80 font-mono whitespace-pre-wrap break-all max-h-32 overflow-y-auto">{{ step.toolResult }}</span>
                </div>

                <!-- Error -->
                <div
                  v-else-if="step.type === 'error'"
                  class="flex items-center gap-2 px-3 py-1.5 rounded bg-rose-950/20 border border-rose-900/20 text-[11px]"
                >
                  <AlertCircle :size="12" class="text-rose-400 shrink-0" />
                  <span class="text-rose-300">{{ step.error }}</span>
                </div>

                <!-- Text (streamed) -->
                <div
                  v-else-if="step.type === 'text' && step.text"
                  class="text-sm text-zinc-300 leading-relaxed prose prose-invert"
                  v-html="renderMarkdown(step.text)"
                  v-html-hook="highlightCode"
                />
              </template>
            </div>

            <!-- Final text response (if no steps or text accumulated outside steps) -->
            <div
              v-if="msg.text && msg.steps.filter(s => s.type === 'text').length === 0"
              class="px-4 py-3 rounded bg-zinc-950/40 border border-zinc-900 text-sm prose prose-invert"
              v-html="renderMarkdown(msg.text)"
              v-html-hook="highlightCode"
            />

            <!-- Meta line -->
            <div
              v-if="msg.text || msg.steps.length > 0"
              class="flex items-center gap-3.5 pt-1 text-[10px] text-zinc-500"
            >
              <button
                class="flex items-center gap-1 hover:text-zinc-300 transition-colors"
                @click="copyMessage(msg.text)"
              >
                <Copy :size="12" /> [COPY]
              </button>
            </div>
          </div>

          <!-- User Avatar -->
          <div
            v-if="msg.role === 'user'"
            class="w-7 h-7 rounded border border-zinc-800 bg-zinc-900 flex items-center justify-center shrink-0"
          >
            <span class="text-[9px] font-semibold text-zinc-500">USR</span>
          </div>
        </div>

        <!-- Typing Indicator -->
        <div v-if="isGenerating" class="flex gap-4 justify-start items-center">
          <div class="w-7 h-7 rounded border border-zinc-800 bg-zinc-950 flex items-center justify-center shrink-0">
            <span class="text-[9px] text-zinc-500">SYS</span>
          </div>
          <div class="px-3.5 py-2 rounded bg-zinc-950/20 border border-zinc-900 text-zinc-500 text-xs flex items-center gap-2">
            <div class="flex space-x-1">
              <span class="w-1.5 h-1.5 bg-zinc-500 rounded-full animate-bounce" style="animation-delay: 0.1s" />
              <span class="w-1.5 h-1.5 bg-zinc-500 rounded-full animate-bounce" style="animation-delay: 0.2s" />
              <span class="w-1.5 h-1.5 bg-zinc-500 rounded-full animate-bounce" style="animation-delay: 0.3s" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Input Area -->
    <div class="p-4 md:p-6 bg-gradient-to-t from-zinc-950/80 to-transparent border-t border-zinc-800/10 z-20">
      <div class="max-w-3xl mx-auto relative">
        <div class="rounded-lg border border-zinc-800 bg-zinc-900/30 focus-within:border-zinc-700 transition-colors p-2 flex flex-col">
          <textarea
            ref="textareaRef"
            v-model="chatInput"
            rows="1"
            placeholder="Type a message..."
            class="w-full bg-transparent text-zinc-200 placeholder-zinc-600 focus:outline-none resize-none max-h-48 min-h-[32px] px-2 py-1 text-sm leading-relaxed"
            @input="autoResize($event.target as HTMLTextAreaElement)"
            @keydown="onKeydown"
          />

          <div class="flex items-center justify-between pt-2 px-1 border-t border-zinc-800/40 mt-1">
            <div class="flex items-center gap-1" />

            <div class="flex items-center gap-2">
              <button
                class="p-1.5 rounded hover:bg-zinc-800 text-zinc-500 hover:text-zinc-300 transition-colors"
                title="Reset Session"
                @click="clearSessionMessages()"
              >
                <RefreshCw :size="14" />
              </button>
              <button
                class="flex items-center justify-center w-7 h-7 rounded bg-zinc-100 hover:bg-zinc-200 text-zinc-950 transition-colors disabled:opacity-30"
                :disabled="isGenerating || !chatInput.trim()"
                @click="handleSend"
              >
                <ArrowUp :size="16" />
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
      mounted(el: HTMLElement, binding: any) {
        if (binding.value) binding.value(el)
      },
      updated(el: HTMLElement, binding: any) {
        if (binding.value) binding.value(el)
      }
    }
  }
}
</script>
