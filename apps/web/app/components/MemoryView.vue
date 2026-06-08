<script setup lang="ts">
import type { MemoryRecord } from '~/types'
import { apiFetch } from '~/utils/api'
import { Search, Plus, Trash2, Database } from '@lucide/vue'
import { marked } from 'marked'

const toast = useToast()
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
    toast.add(err instanceof Error ? err.message : 'Failed to load memories', 'error')
  } finally {
    loading.value = false
  }
}

async function addMemory() {
  if (!newMemory.value.content.trim()) return
  try {
    await apiFetch('/api/memory', {
      method: 'POST',
      body: JSON.stringify(newMemory.value)
    })
    toast.add('Memory added', 'success')
    addMode.value = false
    newMemory.value = { scope: 'global', kind: 'note', content: '', confidence: 'medium' }
    await fetchMemories()
  } catch (err) {
    toast.add(err instanceof Error ? err.message : 'Failed to add memory', 'error')
  }
}

async function deleteMemory(id: number) {
  try {
    await apiFetch(`/api/memory/${id}`, { method: 'DELETE' })
    toast.add('Memory deleted', 'success')
    await fetchMemories()
  } catch (err) {
    toast.add(err instanceof Error ? err.message : 'Failed to delete memory', 'error')
  }
}

const filtered = computed(() => {
  if (!searchQuery.value) return memories.value
  const q = searchQuery.value.toLowerCase()
  return memories.value.filter(m => m.content.toLowerCase().includes(q) || m.kind.toLowerCase().includes(q) || m.scope.toLowerCase().includes(q))
})

onMounted(fetchMemories)
</script>

<template>
  <div class="max-w-3xl mx-auto space-y-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <Database :size="16" class="text-zinc-400" />
        <h2 class="text-sm font-semibold text-zinc-200">Memory</h2>
        <span class="text-[10px] text-zinc-500 font-mono">{{ memories.length }} records</span>
      </div>
      <button
        class="flex items-center gap-1.5 px-2.5 py-1.5 rounded border border-zinc-800 hover:border-zinc-700 bg-zinc-900/60 text-zinc-300 text-xs transition-colors"
        @click="addMode = !addMode"
      >
        <Plus :size="12" />
        Add
      </button>
    </div>

    <!-- Search -->
    <div class="relative">
      <Search :size="14" class="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
      <input
        v-model="searchQuery"
        placeholder="Search memories..."
        class="w-full bg-zinc-900/60 border border-zinc-800 rounded-lg pl-9 pr-3 py-2 text-xs text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 transition-colors"
      >
    </div>

    <!-- Add Form -->
    <div v-if="addMode" class="rounded-lg border border-zinc-800 bg-zinc-900/40 p-4 space-y-3">
      <div class="grid grid-cols-3 gap-2">
        <select v-model="newMemory.scope" class="bg-zinc-900 border border-zinc-800 rounded px-2 py-1.5 text-xs text-zinc-300 focus:outline-none">
          <option value="global">global</option>
          <option value="user">user</option>
          <option value="session">session</option>
        </select>
        <select v-model="newMemory.kind" class="bg-zinc-900 border border-zinc-800 rounded px-2 py-1.5 text-xs text-zinc-300 focus:outline-none">
          <option value="note">note</option>
          <option value="fact">fact</option>
          <option value="rule">rule</option>
          <option value="preference">preference</option>
        </select>
        <select v-model="newMemory.confidence" class="bg-zinc-900 border border-zinc-800 rounded px-2 py-1.5 text-xs text-zinc-300 focus:outline-none">
          <option value="high">high</option>
          <option value="medium">medium</option>
          <option value="low">low</option>
        </select>
      </div>
      <textarea
        v-model="newMemory.content"
        rows="3"
        placeholder="Memory content..."
        class="w-full bg-zinc-900 border border-zinc-800 rounded px-3 py-2 text-xs text-zinc-200 placeholder-zinc-600 focus:outline-none focus:border-zinc-700 resize-none"
      />
      <div class="flex justify-end gap-2">
        <button class="px-3 py-1.5 rounded border border-zinc-800 hover:bg-zinc-900 text-zinc-400 text-xs transition-colors" @click="addMode = false">Cancel</button>
        <button class="px-3 py-1.5 rounded bg-zinc-100 hover:bg-zinc-200 text-zinc-950 text-xs font-semibold transition-colors" @click="addMemory">Save</button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <div v-for="i in 3" :key="i" class="h-20 rounded-lg bg-zinc-900/40 animate-pulse" />
    </div>

    <!-- Empty -->
    <div v-else-if="filtered.length === 0" class="text-center py-16">
      <Database :size="32" class="mx-auto text-zinc-700 mb-3" />
      <p class="text-xs text-zinc-500">No memories found.</p>
    </div>

    <!-- List -->
    <div v-else class="space-y-2">
      <div
        v-for="mem in filtered"
        :key="mem.id"
        class="rounded-lg border border-zinc-900 bg-zinc-950/40 p-4 group hover:border-zinc-800 transition-colors"
      >
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center gap-1.5">
            <span class="px-1.5 py-0.5 rounded text-[9px] font-mono font-medium bg-zinc-900 text-zinc-400 border border-zinc-800">{{ mem.kind }}</span>
            <span class="px-1.5 py-0.5 rounded text-[9px] font-mono font-medium bg-zinc-900 text-zinc-500 border border-zinc-800">{{ mem.confidence }}</span>
          </div>
          <button
            class="p-1 rounded hover:bg-zinc-800 text-zinc-600 hover:text-rose-400 opacity-0 group-hover:opacity-100 transition-all"
            @click="deleteMemory(mem.id)"
          >
            <Trash2 :size="12" />
          </button>
        </div>
        <p class="text-xs text-zinc-300 leading-relaxed">{{ mem.content }}</p>
        <div class="flex items-center justify-between mt-2.5 pt-2 border-t border-zinc-900/60">
          <span class="text-[10px] text-zinc-600 font-mono">{{ mem.scope }}</span>
          <span class="text-[10px] text-zinc-600">{{ mem.created_at }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
