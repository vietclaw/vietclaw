<script setup lang="ts">
import { Sparkles, Plus, PanelRight, Trash2, X } from '@lucide/vue'

defineProps<{ open: boolean }>()
defineEmits<{ close: [] }>()

const { sessions, currentSessionId, createSession, switchSession, deleteSession } = useChat()
const advancedOpen = useState('advancedConsoleOpen', () => false)
const { online } = useDaemon()
</script>

<template>
  <aside
    class="fixed z-30 flex h-full w-72 flex-col border-r border-zinc-800/80 bg-zinc-950/95 backdrop-blur-md transition-transform duration-200 md:relative md:translate-x-0 -translate-x-full"
    :class="open ? 'translate-x-0' : ''"
  >
    <div class="flex items-center justify-between border-b border-zinc-800/60 p-4">
      <div class="flex items-center gap-2.5">
        <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-zinc-100 text-zinc-950">
          <Sparkles :size="16" />
        </div>
        <div>
          <h1 class="text-sm font-semibold text-zinc-100">VietClaw</h1>
          <p class="text-[10px] text-zinc-500">prompt-first agent</p>
        </div>
      </div>
      <button type="button" class="rounded p-1 text-zinc-500 hover:bg-zinc-900 md:hidden" @click="$emit('close')">
        <X :size="16" />
      </button>
    </div>

    <div class="p-3 space-y-2">
      <button
        type="button"
        class="flex w-full items-center justify-center gap-2 rounded-lg bg-zinc-100 py-2.5 text-xs font-semibold text-zinc-950 transition-colors hover:bg-white"
        @click="createSession()"
      >
        <Plus :size="14" />
        Hội thoại mới
      </button>
      <button
        type="button"
        class="flex w-full items-center justify-center gap-2 rounded-lg border border-zinc-800 py-2 text-xs text-zinc-400 transition-colors hover:border-zinc-600 hover:text-zinc-200"
        @click="advancedOpen = true"
      >
        <PanelRight :size="14" />
        Công cụ nâng cao
      </button>
    </div>

    <div class="flex-1 overflow-y-auto px-2 py-2 vc-scrollbar">
      <p class="px-2 mb-2 text-[10px] font-medium uppercase tracking-wider text-zinc-600">Gần đây</p>
      <div class="space-y-0.5">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="group flex items-center justify-between rounded-lg p-2 cursor-pointer transition-colors"
          :class="session.id === currentSessionId ? 'bg-zinc-900 border border-zinc-800' : 'hover:bg-zinc-900/50 border border-transparent'"
          @click="switchSession(session.id)"
        >
          <span
            class="truncate text-xs"
            :class="session.id === currentSessionId ? 'text-zinc-200 font-medium' : 'text-zinc-500'"
          >{{ session.title }}</span>
          <button
            v-if="sessions.length > 1"
            type="button"
            class="rounded p-1 text-zinc-600 opacity-0 transition-opacity hover:text-rose-400 group-hover:opacity-100"
            @click.stop="deleteSession(session.id)"
          >
            <Trash2 :size="13" />
          </button>
        </div>
      </div>
    </div>

    <div class="border-t border-zinc-800/60 p-3">
      <div class="flex items-center gap-2 text-[10px] text-zinc-500">
        <span class="h-1.5 w-1.5 rounded-full" :class="online ? 'bg-emerald-500' : 'bg-zinc-600'" />
        {{ online ? 'Đã kết nối' : 'Chưa kết nối daemon' }}
      </div>
    </div>
  </aside>
</template>
