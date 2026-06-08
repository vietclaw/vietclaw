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
    <section class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)] p-5">
      <div class="grid gap-3 lg:grid-cols-[1fr_180px_auto]">
        <input v-model="query" placeholder="Search memory..." class="rounded-xl border border-white/[0.08] bg-[#0b0d12] px-4 py-3 text-sm text-white vc-focus" @keydown.enter="refresh">
        <select v-model="kind" class="rounded-xl border border-white/[0.08] bg-[#0b0d12] px-4 py-3 text-sm text-white vc-focus">
          <option v-for="item in kinds" :key="item" :value="item">{{ item }}</option>
        </select>
        <button class="rounded-xl bg-white px-5 py-3 text-sm font-medium text-[#0b0d12]" @click="refresh">Refresh</button>
      </div>
      <div class="mt-3 flex gap-3">
        <textarea v-model="content" rows="2" placeholder="Add memory..." class="min-h-[52px] flex-1 resize-none rounded-xl border border-white/[0.08] bg-[#0b0d12] px-4 py-3 text-sm text-white vc-focus" />
        <button class="rounded-xl bg-[var(--vc-accent)] px-5 py-3 text-sm font-medium text-[#07101b]" :disabled="loading || !content.trim()" @click="addMemory">Add</button>
      </div>
      <p v-if="error" class="mt-3 text-sm text-[var(--vc-bad)]">{{ error }}</p>
    </section>

    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <MemoryCard v-for="item in memories" :key="item.id" :memory="item" />
    </div>
    <div v-if="!loading && memories.length === 0" class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)] p-8 text-sm text-[var(--vc-muted)]">
      No memory found.
    </div>
  </div>
</template>

