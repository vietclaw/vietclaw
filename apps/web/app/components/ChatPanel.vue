<script setup lang="ts">
import type { ChatResponse } from '~/types'
import { apiFetch, formatMoney } from '~/utils/api'

type ChatItem = {
  role: 'user' | 'assistant'
  text: string
  meta?: Pick<ChatResponse, 'intent' | 'provider' | 'model' | 'cost_usd'>
}

const props = withDefaults(defineProps<{ compact?: boolean }>(), { compact: false })

const message = ref('')
const mode = ref('eco')
const loading = ref(false)
const error = ref('')
const sessionId = ref('')
const chatBox = ref<HTMLElement | null>(null)

const items = ref<ChatItem[]>([
  { role: 'assistant', text: 'hỏi t gì đó. t sẽ giữ context ngắn thôi.' }
])

async function send() {
  const text = message.value.trim()
  if (!text || loading.value) return
  error.value = ''
  items.value.push({ role: 'user', text })
  message.value = ''
  loading.value = true
  await nextTick()
  scrollToBottom()
  try {
    const response = await apiFetch<ChatResponse>('/api/chat', {
      method: 'POST',
      body: JSON.stringify({
        session_id: sessionId.value || undefined,
        user_id: 'local',
        channel: 'web',
        message: text,
        mode: mode.value
      })
    })
    sessionId.value = response.session_id
    items.value.push({
      role: 'assistant',
      text: response.reply || response.error || 't chưa có câu trả lời.',
      meta: response
    })
  } catch (err) {
    const msg = err instanceof Error ? err.message : 'chat failed'
    error.value = msg
    items.value.push({ role: 'assistant', text: msg })
  } finally {
    loading.value = false
    await nextTick()
    scrollToBottom()
  }
}

function scrollToBottom() {
  if (chatBox.value) {
    chatBox.value.scrollTop = chatBox.value.scrollHeight
  }
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    void send()
  }
}
</script>

<template>
  <section class="vc-bezel vc-noise vc-ambient" :class="compact ? 'h-[520px]' : 'h-[680px]'">
    <div class="vc-bezel-inner flex flex-col overflow-hidden">
      <!-- Header -->
      <div class="relative z-10 flex items-center justify-between border-b border-[var(--border-0)] px-5 py-4">
        <div class="flex items-center gap-3">
          <div class="relative flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-[var(--accent)]/15 to-purple-500/10">
            <svg class="h-4 w-4 text-[var(--accent-light)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <div>
            <h2 class="text-[13px] font-bold text-[var(--fg-0)]">Chat</h2>
            <p class="text-[11px] text-[var(--fg-2)]">Ask VietClaw anything</p>
          </div>
        </div>
        <select v-model="mode" class="rounded-lg border border-[var(--border-1)] bg-[var(--bg-2)]/80 px-2.5 py-1.5 text-[11px] font-semibold text-[var(--fg-1)] vc-focus vc-transition-fast hover:border-[var(--border-2)]">
          <option value="eco">eco</option>
          <option value="smart" disabled>smart</option>
          <option value="max" disabled>max</option>
        </select>
      </div>

      <!-- Messages -->
      <div ref="chatBox" class="vc-scrollbar relative z-10 flex-1 overflow-y-auto px-5 py-5">
        <div class="space-y-5">
          <TransitionGroup name="msg">
            <div
              v-for="(item, index) in items"
              :key="index"
              class="flex gap-3"
              :class="item.role === 'user' ? 'flex-row-reverse' : ''"
            >
              <!-- Avatar -->
              <div
                class="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl text-[11px] font-bold vc-transition"
                :class="item.role === 'user'
                  ? 'bg-gradient-to-br from-[var(--accent)] to-[var(--accent-dim)] text-white shadow-lg shadow-[var(--accent)]/20'
                  : 'bg-[var(--bg-3)] text-[var(--fg-2)]'"
              >
                {{ item.role === 'user' ? 'U' : 'V' }}
              </div>
              <!-- Bubble -->
              <div
                class="max-w-[78%] rounded-2xl px-4 py-3 text-[13px] leading-relaxed vc-transition"
                :class="item.role === 'user'
                  ? 'bg-gradient-to-br from-[var(--accent)] to-[var(--accent-dim)] text-white rounded-tr-md shadow-lg shadow-[var(--accent)]/15'
                  : 'bg-[var(--bg-2)] text-[var(--fg-0)] rounded-tl-md border border-[var(--border-0)]'"
              >
                <p class="whitespace-pre-wrap">{{ item.text }}</p>
                <div v-if="item.meta" class="mt-2.5 flex flex-wrap gap-1.5 border-t border-white/10 pt-2.5 text-[10px] font-medium opacity-50">
                  <span>{{ item.meta.intent }}</span>
                  <span aria-hidden="true">·</span>
                  <span>{{ item.meta.provider }}/{{ item.meta.model }}</span>
                  <span aria-hidden="true">·</span>
                  <span>{{ formatMoney(item.meta.cost_usd) }}</span>
                </div>
              </div>
            </div>
          </TransitionGroup>

          <!-- Typing indicator -->
          <div v-if="loading" class="flex gap-3">
            <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-[var(--bg-3)] text-[11px] font-bold text-[var(--fg-2)]">V</div>
            <div class="flex items-center gap-1.5 rounded-2xl rounded-tl-md bg-[var(--bg-2)] border border-[var(--border-0)] px-5 py-3.5">
              <span class="h-1.5 w-1.5 animate-bounce rounded-full bg-[var(--fg-2)] [animation-delay:-0.3s]" />
              <span class="h-1.5 w-1.5 animate-bounce rounded-full bg-[var(--fg-2)] [animation-delay:-0.15s]" />
              <span class="h-1.5 w-1.5 animate-bounce rounded-full bg-[var(--fg-2)]" />
            </div>
          </div>
        </div>
      </div>

      <!-- Input -->
      <div class="relative z-10 border-t border-[var(--border-0)] p-4">
        <p v-if="error" class="mb-3 rounded-xl border border-[var(--danger)]/20 bg-[var(--danger)]/5 px-3.5 py-2.5 text-[12px] font-medium text-[var(--danger)]">{{ error }}</p>
        <div class="flex items-end gap-2.5">
          <textarea
            v-model="message"
            rows="1"
            placeholder="Type a message..."
            class="min-h-[42px] max-h-[120px] flex-1 resize-none rounded-xl border border-[var(--border-1)] bg-[var(--bg-2)]/80 px-4 py-3 text-[13px] text-[var(--fg-0)] placeholder:text-[var(--fg-2)]/40 vc-focus vc-transition-fast hover:border-[var(--border-2)] focus:border-[var(--accent)]/30 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.08)]"
            @keydown="onKeydown"
          />
          <button
            class="group flex h-[42px] w-[42px] shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-[var(--accent)] to-[var(--accent-dim)] text-white vc-transition vc-focus disabled:opacity-30 disabled:active:scale-100"
            :disabled="loading || !message.trim()"
            @click="send"
          >
            <svg class="h-4 w-4 vc-transition-fast group-hover:translate-x-0.5 group-hover:-translate-y-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 19V5m0 0l-4 4m4-4l4 4" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.msg-enter-active {
  transition: all 0.5s var(--ease-out);
}
.msg-enter-from {
  opacity: 0;
  transform: translateY(12px) blur(4px);
}
</style>
