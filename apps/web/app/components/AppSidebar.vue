<script setup lang="ts">
import { Plus, Trash2, X } from '@lucide/vue'

defineProps<{ open: boolean }>()
defineEmits<{ close: [] }>()

const { sessions, currentSessionId, createSession, switchSession, deleteSession } = useChat()
const route = useRoute()
const { online } = useDaemon()
</script>

<template>
  <aside
    class="fixed z-30 flex h-full w-60 flex-col border-r border-vc-border bg-vc-surface transition-transform duration-200 ease-out md:relative md:translate-x-0 -translate-x-full"
    :class="open ? 'translate-x-0' : ''"
  >
    <div class="flex items-center justify-between px-4 pt-5 pb-3">
      <span class="text-[15px] font-semibold tracking-tight text-vc-text">VietClaw</span>
      <button type="button" class="vc-btn-ghost rounded-md p-1 md:hidden" @click="$emit('close')">
        <X :size="18" :stroke-width="1.75" />
      </button>
    </div>

    <div class="px-3 pb-2">
      <button type="button" class="vc-btn vc-btn-ghost w-full justify-start gap-2 px-2 py-2" @click="createSession()">
        <Plus :size="16" :stroke-width="1.75" />
        Hội thoại mới
      </button>
    </div>

    <div class="flex-1 overflow-y-auto px-2 vc-scrollbar">
      <div class="space-y-0.5 py-1">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="group flex items-center gap-1 rounded-md"
          :class="session.id === currentSessionId ? 'bg-vc-bg-subtle' : ''"
        >
          <button
            type="button"
            class="min-w-0 flex-1 truncate px-3 py-2 text-left text-sm transition-colors"
            :class="session.id === currentSessionId
              ? 'font-medium text-vc-text'
              : 'text-vc-text-secondary hover:text-vc-text'"
            @click="switchSession(session.id)"
          >
            {{ session.title }}
          </button>
          <button
            v-if="sessions.length > 1"
            type="button"
            class="vc-btn-ghost rounded-md p-1.5 opacity-0 group-hover:opacity-100"
            @click.stop="deleteSession(session.id)"
          >
            <Trash2 :size="13" :stroke-width="1.75" />
          </button>
        </div>
      </div>
    </div>

    <div class="border-t border-vc-border-subtle px-4 py-3 space-y-2">
      <NuxtLink
        to="/settings"
        class="block text-sm transition-colors"
        :class="route.path.startsWith('/settings') ? 'font-medium text-vc-text' : 'text-vc-text-secondary hover:text-vc-text'"
      >
        Cài đặt
      </NuxtLink>
      <p class="text-xs text-vc-text-muted">
        {{ online ? 'Đã kết nối daemon' : 'Chưa kết nối' }}
      </p>
    </div>
  </aside>
</template>
