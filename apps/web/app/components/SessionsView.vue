<script setup lang="ts">
import { MessageSquare, Plus, Trash2 } from '@lucide/vue'

const { sessions, currentSessionId, createSession, switchSession, deleteSession } = useChat()
</script>

<template>
  <section class="space-y-4">
    <div class="flex items-center justify-between gap-3">
      <div>
        <h2 class="text-sm font-semibold text-zinc-100">Sessions</h2>
        <p class="mt-1 text-xs text-zinc-500">Local chat workspace history.</p>
      </div>
      <button
        class="flex items-center gap-2 rounded-md bg-zinc-100 px-3 py-2 text-xs font-medium text-zinc-950 hover:bg-zinc-200"
        @click="createSession()"
      >
        <Plus :size="14" />
        New
      </button>
    </div>

    <div class="overflow-hidden rounded-lg border border-zinc-800 bg-zinc-950/40">
      <button
        v-for="session in sessions"
        :key="session.id"
        class="group flex w-full items-center justify-between gap-3 border-b border-zinc-900 px-4 py-3 text-left last:border-b-0 hover:bg-zinc-900/50"
        :class="session.id === currentSessionId ? 'bg-zinc-900/80' : ''"
        @click="switchSession(session.id)"
      >
        <div class="flex min-w-0 items-center gap-3">
          <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded border border-zinc-800 bg-zinc-950 text-zinc-500">
            <MessageSquare :size="14" />
          </div>
          <div class="min-w-0">
            <div class="truncate text-sm font-medium text-zinc-200">{{ session.title }}</div>
            <div class="mt-0.5 text-[11px] text-zinc-500">
              {{ session.messages?.length ?? 0 }} messages
            </div>
          </div>
        </div>
        <button
          v-if="sessions.length > 1"
          class="rounded p-1 text-zinc-600 opacity-0 transition hover:bg-zinc-800 hover:text-rose-400 group-hover:opacity-100"
          @click.stop="deleteSession(session.id)"
        >
          <Trash2 :size="14" />
        </button>
      </button>
    </div>
  </section>
</template>
