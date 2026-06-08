<script setup lang="ts">
import { Menu, Edit2, Sparkles, Download } from '@lucide/vue'

defineEmits<{ toggleMobile: [] }>()

const { currentSession, sessions, currentSessionId } = useChat()

function renameSession() {
  const s = currentSession()
  if (!s) return
  const name = prompt('Update session title:', s.title)
  if (name && name.trim()) {
    s.title = name.trim()
    useChat().saveSessions()
  }
}

function exportSession() {
  const s = currentSession()
  if (!s || (s.messages?.length ?? 0) === 0) return
  const json = `data:text/json;charset=utf-8,${encodeURIComponent(JSON.stringify(s, null, 2))}`
  const link = document.createElement('a')
  link.href = json
  link.download = `${s.title.replace(/\s+/g, '_')}.json`
  document.body.appendChild(link)
  link.click()
  link.remove()
}
</script>

<template>
  <header class="h-14 border-b border-zinc-800/60 px-4 md:px-6 flex items-center justify-between bg-zinc-950/40 backdrop-blur-md z-20">
    <div class="flex items-center gap-3">
      <button class="md:hidden p-1.5 rounded hover:bg-zinc-900 text-zinc-400" @click="$emit('toggleMobile')">
        <Menu :size="16" />
      </button>
      <div class="flex items-center gap-2">
        <span class="text-xs font-semibold text-zinc-200 max-w-[140px] md:max-w-xs truncate">
          {{ currentSession()?.title || 'Untitled Session' }}
        </span>
        <button class="text-zinc-600 hover:text-zinc-400 transition-colors" @click="renameSession">
          <Edit2 :size="12" />
        </button>
      </div>
    </div>

    <div class="flex items-center gap-2">
      <button
        class="p-1.5 rounded hover:bg-zinc-900 text-zinc-500 hover:text-zinc-300 transition-colors"
        title="Export Session (JSON)"
        @click="exportSession"
      >
        <Download :size="16" />
      </button>

      <div class="flex items-center gap-1.5 px-2.5 py-1 rounded bg-zinc-900 border border-zinc-800 text-[10px] font-mono text-zinc-400">
        <span class="w-1.5 h-1.5 rounded-full bg-zinc-500" />
        <span>web</span>
      </div>
    </div>
  </header>
</template>
