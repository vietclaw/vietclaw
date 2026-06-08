<script setup lang="ts">
import type { MemoryRecord } from '~/types'
import { apiFetch } from '~/utils/api'

definePageMeta({ title: 'Memory' })

const query = ref('')
const content = ref('')
const kind = ref('note')
const loading = ref(false)
const error = ref('')
const memories = ref<MemoryRecord[]>([])

const kinds = ['profile', 'preference', 'project', 'workflow', 'decision', 'connection', 'note']

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    const path = query.value.trim() ? `/api/memory/search?q=${encodeURIComponent(query.value.trim())}` : '/api/memory'
    memories.value = await apiFetch<MemoryRecord[]>(path)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'memory failed'
  } finally {
    loading.value = false
  }
}

async function addMemory() {
  const text = content.value.trim()
  if (!text) return
  loading.value = true
  error.value = ''
  try {
    await apiFetch('/api/memory', {
      method: 'POST',
      body: JSON.stringify({ scope: 'user:local', kind: kind.value, content: text, confidence: 'confirmed' })
    })
    content.value = ''
    await refresh()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'add memory failed'
  } finally {
    loading.value = false
  }
}

onMounted(refresh)
</script>

<template>
  <div class="space-y-5">
    <!-- Add form -->
    <section class="vc-bezel vc-noise vc-ambient vc-stagger">
      <div class="vc-bezel-inner relative p-5">
        <div class="relative z-10 flex items-center gap-3 mb-4">
          <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-[var(--accent)]/10">
            <svg class="h-4 w-4 text-[var(--accent-light)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
          </div>
          <div>
            <h2 class="text-[13px] font-bold text-[var(--fg-0)]">Add memory</h2>
            <p class="text-[11px] text-[var(--fg-2)]">Store knowledge for the agent</p>
          </div>
        </div>
        <div class="relative z-10 flex gap-2.5">
          <textarea v-model="content" rows="2" placeholder="Type memory content..." class="min-h-[44px] flex-1 resize-none rounded-xl border border-[var(--border-1)] bg-[var(--bg-2)]/80 px-4 py-3 text-[13px] text-[var(--fg-0)] placeholder:text-[var(--fg-2)]/40 vc-focus vc-transition-fast hover:border-[var(--border-2)] focus:border-[var(--accent)]/30 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.08)]" />
          <button
            class="group flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-[var(--accent)] to-[var(--accent-dim)] text-white vc-transition vc-focus disabled:opacity-30 disabled:active:scale-100"
            :disabled="loading || !content.trim()"
            @click="addMemory"
          >
            <svg class="h-4 w-4 vc-transition-fast group-hover:rotate-90" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
          </button>
        </div>
        <p v-if="error" class="relative z-10 mt-3 text-[12px] font-medium text-[var(--danger)]">{{ error }}</p>
      </div>
    </section>

    <!-- Search + filter -->
    <section class="vc-bezel vc-noise vc-stagger" style="animation-delay: 80ms;">
      <div class="vc-bezel-inner relative p-5">
        <div class="relative z-10 flex items-center gap-3 mb-4">
          <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-[var(--bg-3)]">
            <svg class="h-4 w-4 text-[var(--fg-2)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
          <div>
            <h2 class="text-[13px] font-bold text-[var(--fg-0)]">Search memory</h2>
            <p class="text-[11px] text-[var(--fg-2)]">{{ memories.length }} records</p>
          </div>
        </div>
        <div class="relative z-10 flex gap-2.5">
          <input v-model="query" placeholder="Search..." class="flex-1 rounded-xl border border-[var(--border-1)] bg-[var(--bg-2)]/80 px-4 py-3 text-[13px] text-[var(--fg-0)] placeholder:text-[var(--fg-2)]/40 vc-focus vc-transition-fast hover:border-[var(--border-2)] focus:border-[var(--accent)]/30 focus:shadow-[0_0_0_3px_rgba(99,102,241,0.08)]" @keydown.enter="refresh">
          <select v-model="kind" class="rounded-xl border border-[var(--border-1)] bg-[var(--bg-2)]/80 px-3.5 py-3 text-[13px] font-medium text-[var(--fg-1)] vc-focus vc-transition-fast hover:border-[var(--border-2)]">
            <option v-for="item in kinds" :key="item" :value="item">{{ item }}</option>
          </select>
          <button class="vc-btn-magnetic border border-[var(--border-1)] bg-[var(--bg-2)]/80 text-[var(--fg-1)] hover:bg-[var(--bg-3)] hover:text-[var(--fg-0)]" @click="refresh">Search</button>
        </div>
      </div>
    </section>

    <!-- Results grid -->
    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3 vc-stagger" style="animation-delay: 160ms;">
      <MemoryCard v-for="item in memories" :key="item.id" :memory="item" />
    </div>

    <!-- Empty state -->
    <div v-if="!loading && memories.length === 0" class="vc-bezel vc-noise vc-stagger" style="animation-delay: 240ms;">
      <div class="vc-bezel-inner px-5 py-14 text-center">
        <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-[var(--bg-3)]">
          <svg class="h-5 w-5 text-[var(--fg-2)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547" />
          </svg>
        </div>
        <p class="mt-3 text-[13px] font-medium text-[var(--fg-2)]">No memory found.</p>
      </div>
    </div>
  </div>
</template>
