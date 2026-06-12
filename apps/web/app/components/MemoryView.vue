<script setup lang="ts">
import type { MemoryRecord } from '~/types'
import { apiFetch } from '~/utils/api'
import { Search, Plus, Trash2 } from '@lucide/vue'

defineProps<{ embedded?: boolean }>()

const toast = useToast()
const { t, option } = useI18n()
const memories = ref<MemoryRecord[]>([])
const loading = ref(true)
const searchQuery = ref('')
const addMode = ref(false)
const newMemory = ref({ scope: 'global', kind: 'note', content: '', confidence: 'medium' })

async function fetchMemories() {
  loading.value = true
  try {
    memories.value = await apiFetch<MemoryRecord[]>('/api/memory')
  } catch (err) {
    toast.add(err instanceof Error ? err.message : t('memory.loadFailed'), 'error')
  } finally {
    loading.value = false
  }
}

async function addMemory() {
  if (!newMemory.value.content.trim()) return
  try {
    await apiFetch('/api/memory', { method: 'POST', body: JSON.stringify(newMemory.value) })
    toast.add(t('memory.added'), 'success')
    addMode.value = false
    newMemory.value = { scope: 'global', kind: 'note', content: '', confidence: 'medium' }
    await fetchMemories()
  } catch (err) {
    toast.add(err instanceof Error ? err.message : t('memory.addFailed'), 'error')
  }
}

async function deleteMemory(id: number) {
  try {
    await apiFetch(`/api/memory/${id}`, { method: 'DELETE' })
    toast.add(t('memory.deleted'), 'success')
    await fetchMemories()
  } catch (err) {
    toast.add(err instanceof Error ? err.message : t('memory.deleteFailed'), 'error')
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
        <h1 class="vc-display text-2xl font-medium text-vc-text">{{ t('memory.title') }}</h1>
        <p class="mt-1.5 text-sm text-vc-text-muted">{{ t('memory.count', memories.length) }}</p>
      </div>
      <button type="button" class="vc-btn vc-btn-outline text-xs" @click="addMode = !addMode">
        <Plus :size="14" :stroke-width="1.5" />
        {{ t('common.add') }}
      </button>
    </div>

    <div class="relative">
      <Search :size="14" class="absolute left-3.5 top-1/2 -translate-y-1/2 text-vc-text-muted" />
      <input
        v-model="searchQuery"
        :placeholder="t('memory.search')"
        class="vc-input rounded-full py-2 pl-9.5"
      />
    </div>

    <div v-if="addMode" class="vc-card space-y-3 p-4">
      <div class="grid grid-cols-3 gap-2">
        <VcSelect v-model="newMemory.scope" group="memory_scope" select-class="vc-input text-xs" />
        <VcSelect v-model="newMemory.kind" group="memory_kind" select-class="vc-input text-xs" />
        <VcSelect v-model="newMemory.confidence" group="memory_confidence" select-class="vc-input text-xs" />
      </div>
      <textarea
        v-model="newMemory.content"
        rows="3"
        :placeholder="t('memory.content')"
        class="vc-input resize-none"
      />
      <div class="flex justify-end gap-2">
        <button type="button" class="vc-btn vc-btn-ghost text-xs" @click="addMode = false">{{ t('common.cancel') }}</button>
        <button type="button" class="vc-btn vc-btn-primary text-xs" @click="addMemory">{{ t('common.save') }}</button>
      </div>
    </div>

    <div v-if="loading" class="space-y-2">
      <div v-for="i in 3" :key="i" class="h-16 animate-pulse rounded-2xl bg-vc-bg-subtle" :style="{ animationDelay: `${i * 0.1}s` }" />
    </div>

    <div v-else-if="filtered.length === 0" class="vc-card flex flex-col items-center px-6 py-12 text-center">
      <p class="vc-display text-lg text-vc-text">{{ t('memory.empty.title') }}</p>
      <p class="mt-1.5 max-w-xs text-sm leading-relaxed text-vc-text-muted">
        {{ t('memory.empty.desc') }}
      </p>
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="mem in filtered"
        :key="mem.id"
        class="vc-card group p-4 transition-shadow duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] hover:shadow-[var(--vc-shadow-md)]"
      >
        <div class="mb-2 flex items-center justify-between gap-2">
          <div class="flex gap-1.5">
            <span class="rounded-md bg-vc-accent-soft px-1.5 py-0.5 font-mono text-[11px] text-vc-accent">{{ option('memory_kind', mem.kind) }}</span>
            <span class="rounded-md bg-vc-bg-subtle px-1.5 py-0.5 font-mono text-[11px] text-vc-text-muted">{{ option('memory_scope', mem.scope) }}</span>
          </div>
          <button
            type="button"
            class="vc-btn-ghost rounded-full p-1 opacity-0 transition-opacity duration-200 group-hover:opacity-100 focus-visible:opacity-100"
            :aria-label="t('memory.delete')"
            @click="deleteMemory(mem.id)"
          >
            <Trash2 :size="13" :stroke-width="1.5" />
          </button>
        </div>
        <p class="text-sm leading-relaxed text-vc-text">{{ mem.content }}</p>
      </div>
    </div>
  </div>
</template>
