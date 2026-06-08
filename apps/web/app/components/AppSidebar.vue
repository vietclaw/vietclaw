<script setup lang="ts">
import { Terminal, Plus, Settings, ChevronRight, Trash2 } from '@lucide/vue'

defineProps<{ open: boolean }>()
defineEmits<{ close: [] }>()

const { sessions, currentSessionId, createSession, switchSession, deleteSession } = useChat()

const settingsOpen = useState('settingsOpen', () => false)

function toggleSettings() {
  settingsOpen.value = !settingsOpen.value
}
</script>

<template>
  <aside
    class="w-72 bg-zinc-950/80 border-r border-zinc-800/80 flex flex-col h-full transition-transform duration-200 md:translate-x-0 -translate-x-full fixed md:relative z-30 backdrop-blur-md"
    :class="open ? 'translate-x-0' : ''"
  >
    <!-- Logo Header -->
    <div class="p-4 border-b border-zinc-800/60 flex items-center justify-between">
      <div class="flex items-center gap-2.5">
        <div class="w-7 h-7 rounded bg-zinc-100 flex items-center justify-center text-zinc-950">
          <Terminal :size="16" />
        </div>
        <div>
          <h1 class="text-sm font-semibold tracking-tight text-zinc-100">vietclaw.console</h1>
          <p class="text-[9px] text-zinc-500 font-mono">v0.1.0</p>
        </div>
      </div>
      <button class="md:hidden p-1 rounded hover:bg-zinc-900 text-zinc-400" @click="$emit('close')">
        <X :size="16" />
      </button>
    </div>

    <!-- Action Button -->
    <div class="p-3">
      <button
        class="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-md bg-zinc-100 hover:bg-zinc-200 text-zinc-950 font-medium text-xs transition-colors"
        @click="createSession()"
      >
        <Plus :size="14" />
        <span>New Session</span>
      </button>
    </div>

    <!-- History Feed -->
    <div class="flex-1 overflow-y-auto px-2 py-3 space-y-1 vc-scrollbar">
      <div class="flex items-center justify-between px-2 mb-2">
        <span class="text-[10px] font-medium text-zinc-400 uppercase tracking-wider">Active Sessions</span>
      </div>
      <div class="space-y-0.5">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="group flex items-center justify-between p-2 rounded cursor-pointer transition-all"
          :class="session.id === currentSessionId
            ? 'bg-zinc-900 border border-zinc-800'
            : 'hover:bg-zinc-950 border border-transparent'"
          @click="switchSession(session.id)"
        >
          <div class="flex items-center gap-2 overflow-hidden flex-1">
            <Terminal
              :size="14"
              :class="session.id === currentSessionId ? 'text-zinc-100' : 'text-zinc-600'"
              class="shrink-0"
            />
            <span
              class="text-xs truncate"
              :class="session.id === currentSessionId ? 'text-zinc-200 font-medium' : 'text-zinc-400'"
            >{{ session.title }}</span>
          </div>
          <button
            v-if="sessions.length > 1"
            class="p-0.5 rounded hover:bg-zinc-800 text-zinc-600 hover:text-rose-400 opacity-0 group-hover:opacity-100 transition-opacity"
            @click.stop="deleteSession(session.id)"
          >
            <Trash2 :size="14" />
          </button>
        </div>
      </div>
    </div>

    <!-- Control Panel Footer -->
    <div class="p-3 border-t border-zinc-800/60 bg-zinc-950/40">
      <button
        class="w-full flex items-center justify-between px-2.5 py-2 rounded-md hover:bg-zinc-900 text-zinc-400 hover:text-zinc-200 transition-colors"
        @click="toggleSettings"
      >
        <div class="flex items-center gap-2">
          <Settings :size="14" />
          <span class="text-xs">Preferences</span>
        </div>
        <ChevronRight :size="12" class="text-zinc-600" />
      </button>
    </div>
  </aside>
</template>
