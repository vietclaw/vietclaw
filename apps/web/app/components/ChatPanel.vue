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
  <section class="flex min-h-[520px] flex-col rounded-xl border border-white/[0.08] bg-[var(--vc-panel)]">
    <div class="flex items-center justify-between gap-4 border-b border-white/[0.07] px-4 py-3">
      <div>
        <h2 class="text-sm font-medium text-white">Chat</h2>
        <p class="text-xs text-[var(--vc-muted)]">Ask VietClaw...</p>
      </div>
      <select v-model="mode" class="rounded-lg border border-white/[0.08] bg-white/[0.04] px-3 py-2 text-sm text-white vc-focus">
        <option value="eco">eco</option>
        <option value="smart" disabled>smart</option>
        <option value="max" disabled>max</option>
      </select>
    </div>

    <div class="vc-scrollbar flex-1 space-y-4 overflow-auto p-4" :class="compact ? 'max-h-[420px]' : 'max-h-[62vh]'">
      <div
        v-for="(item, index) in items"
        :key="index"
        class="flex"
        :class="item.role === 'user' ? 'justify-end' : 'justify-start'"
      >
        <div
          class="max-w-[780px] rounded-2xl px-4 py-3 text-sm leading-6"
          :class="item.role === 'user' ? 'bg-[var(--vc-accent)] text-[#07101b]' : 'bg-white/[0.055] text-white'"
        >
          <p class="whitespace-pre-wrap">{{ item.text }}</p>
          <div v-if="item.meta" class="mt-3 flex flex-wrap gap-2 text-[11px] opacity-75">
            <span>{{ item.meta.intent }}</span>
            <span>{{ item.meta.provider }}/{{ item.meta.model }}</span>
            <span>{{ formatMoney(item.meta.cost_usd) }}</span>
          </div>
        </div>
      </div>
      <div v-if="loading" class="text-sm text-[var(--vc-muted)]">Thinking...</div>
    </div>

    <div class="border-t border-white/[0.07] p-4">
      <p v-if="error" class="mb-3 text-sm text-[var(--vc-bad)]">{{ error }}</p>
      <div class="flex gap-3">
        <textarea
          v-model="message"
          rows="2"
          placeholder="Ask VietClaw..."
          class="min-h-[52px] flex-1 resize-none rounded-xl border border-white/[0.08] bg-[#0b0d12] px-4 py-3 text-sm text-white placeholder:text-[var(--vc-subtle)] vc-focus"
          @keydown="onKeydown"
        />
        <button
          class="rounded-xl bg-white px-5 py-3 text-sm font-medium text-[#0b0d12] transition hover:bg-[#dfe8f5] vc-focus"
          :disabled="loading || !message.trim()"
          @click="send"
        >
          Send
        </button>
      </div>
    </div>
  </section>
</template>

