<script setup lang="ts">
import { Menu, Edit2, Download } from '@lucide/vue'

defineEmits<{ toggleMobile: [] }>()

const { currentSession } = useChat()
const { status, online } = useDaemon()

function renameSession() {
  const s = currentSession()
  if (!s) return
  const name = prompt('Tên hội thoại:', s.title)
  if (name?.trim()) {
    s.title = name.trim()
    useChat().saveSessions()
  }
}

function exportSession() {
  const s = currentSession()
  if (!s?.messages?.length) return
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
  <header class="z-20 flex h-12 items-center justify-between border-b border-zinc-800/60 bg-zinc-950/50 px-4 md:px-6 backdrop-blur-md">
    <div class="flex min-w-0 items-center gap-2">
      <button type="button" class="rounded-lg p-1.5 text-zinc-400 hover:bg-zinc-900 md:hidden" @click="$emit('toggleMobile')">
        <Menu :size="18" />
      </button>
      <span class="truncate text-sm font-medium text-zinc-200 max-w-[200px] md:max-w-md">
        {{ currentSession()?.title || 'Hội thoại' }}
      </span>
      <button type="button" class="text-zinc-600 hover:text-zinc-400" @click="renameSession">
        <Edit2 :size="13" />
      </button>
    </div>

    <div class="flex items-center gap-2">
      <button
        type="button"
        class="rounded-lg p-1.5 text-zinc-500 hover:bg-zinc-900 hover:text-zinc-300"
        title="Export JSON"
        @click="exportSession"
      >
        <Download :size="16" />
      </button>
      <div class="hidden items-center gap-1.5 rounded-md border border-zinc-800 px-2 py-1 text-[10px] font-mono text-zinc-500 sm:flex">
        <span class="h-1.5 w-1.5 rounded-full" :class="online ? 'bg-emerald-500' : 'bg-zinc-600'" />
        <span>{{ status?.version || 'offline' }}</span>
      </div>
    </div>
  </header>
</template>
