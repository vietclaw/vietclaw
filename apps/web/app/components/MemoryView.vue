<script setup lang="ts">
import type { MemoryRecord } from '~/types'
import { apiFetch } from '~/utils/api'
import { Search, Plus, Trash2 } from '@lucide/vue'

defineProps<{ embedded?: boolean }>()

const toast = useToast()
const memories = ref<MemoryRecord[]>([])
const loading = ref(true)
const searchQuery = ref('')
const addMode = ref(false)
const newMemory = ref({ scope: 'global', kind: 'note', content: '', confidence: 'medium' })

const fieldClass = 'rounded-md border border-vc-border bg-vc-bg px-2 py-1.5 text-xs text-vc-text focus:border-vc-accent focus:outline-none'

async function fetchMemories() {
  loading.value = true
  try {
    memories.value = await apiFetch<MemoryRecord[]>('/api/memory')
  } catch (err) {
    toast.add(err instanceof Error ? err.message : 'Failed to load memories', 'error')
  } finally {
    loading.value = false
  }
}

async function addMemory() {
  if (!newMemory.value.content.trim()) return
  try {
    await apiFetch('/api/memory', { method: 'POST', body: JSON.stringify(newMemory.value) })
    toast.add('Đã thêm memory', 'success')
    addMode.value = false
    newMemory.value = { scope: 'global', kind: 'note', content: '', confidence: 'medium' }
    await fetchMemories()
  } catch (err) {
    toast.add(err instanceof Error ? err.message : 'Failed to add', 'error')
  }
}

async function deleteMemory(id: number) {
  try {
    await apiFetch(`/api/memory/${id}`, { method: 'DELETE' })
    toast.add('Đã xóa', 'success')
    await fetchMemories()
  } catch (err) {
    toast.add(err instanceof Error ? err.message : 'Failed to delete', 'error')
  }
}

const filtered = computed(() => {
  if (!searchQuery.value) return memories.value
  const q = searchQuery.value.toLowerCase()
  return memories.value.filter(m =>
    m.content.toLowerCase().includes(q) || m.kind.toLowerCase().includes(q) || m.scope.toLowerCase().includes(q)
  )
})

onMounted(fetchMemories)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h1 class="text-lg font-semibold tracking-tight text-vc-text">Memory</h1>
        <p class="mt-1 text-sm text-vc-text-muted">{{ memories.length }} bản ghi</p>
      </div>
      <button type="button" class="vc-btn vc-btn-ghost text-xs" @click="addMode = !addMode">
        <Plus :size="14" :stroke-width="1.75" />
        Thêm
      </button>
    </div>

    <div class="relative">
      <Search :size="14" class="absolute left-3 top-1/2 -translate-y-1/2 text-vc-text-muted" />
      <input
        v-model="searchQuery"
        placeholder="Tìm memory..."
        class="w-full rounded-lg border border-vc-border bg-vc-surface py-2 pl-9 pr-3 text-sm text-vc-text placeholder:text-vc-text-muted focus:border-vc-accent focus:outline-none"
      />
    </div>

    <div v-if="addMode" class="rounded-lg border border-vc-border bg-vc-surface p-4 space-y-3">
      <div class="grid grid-cols-3 gap-2">
        <select v-model="newMemory.scope" :class="fieldClass">
          <option value="global">global</option>
          <option value="user">user</option>
          <option value="session">session</option>
        </select>
        <select v-model="newMemory.kind" :class="fieldClass">
          <option value="note">note</option>
          <option value="fact">fact</option>
          <option value="rule">rule</option>
          <option value="preference">preference</option>
        </select>
        <select v-model="newMemory.confidence" :class="fieldClass">
          <option value="high">high</option>
          <option value="medium">medium</option>
          <option value="low">low</option>
        </select>
      </div>
      <textarea
        v-model="newMemory.content"
        rows="3"
        placeholder="Nội dung..."
        class="w-full resize-none rounded-md border border-vc-border bg-vc-bg px-3 py-2 text-sm text-vc-text focus:border-vc-accent focus:outline-none"
      />
      <div class="flex justify-end gap-2">
        <button type="button" class="vc-btn vc-btn-ghost text-xs" @click="addMode = false">Hủy</button>
        <button type="button" class="vc-btn vc-btn-primary text-xs" @click="addMemory">Lưu</button>
      </div>
    </div>

    <div v-if="loading" class="space-y-2">
      <div v-for="i in 3" :key="i" class="h-16 rounded-lg bg-vc-bg-subtle animate-pulse" />
    </div>

    <p v-else-if="filtered.length === 0" class="text-sm text-vc-text-muted">Chưa có memory.</p>

    <div v-else class="space-y-2">
      <div
        v-for="mem in filtered"
        :key="mem.id"
        class="group rounded-lg border border-vc-border bg-vc-surface p-4"
      >
        <div class="flex items-center justify-between gap-2 mb-2">
          <div class="flex gap-2 text-xs font-mono text-vc-text-muted">
            <span>{{ mem.kind }}</span>
            <span>{{ mem.scope }}</span>
          </div>
          <button
            type="button"
            class="vc-btn-ghost rounded p-1 opacity-0 group-hover:opacity-100"
            @click="deleteMemory(mem.id)"
          >
            <Trash2 :size="13" :stroke-width="1.75" />
          </button>
        </div>
        <p class="text-sm leading-relaxed text-vc-text">{{ mem.content }}</p>
      </div>
    </div>
  </div>
</template>
