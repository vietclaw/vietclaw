<script setup lang="ts">
import type { Session } from '~/types'

defineProps<{ sessions: Session[] }>()

const channelColors: Record<string, string> = {
  discord: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20',
  telegram: 'bg-sky-500/10 text-sky-400 border-sky-500/20',
  web: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
}
</script>

<template>
  <section class="vc-bezel vc-noise overflow-hidden">
    <div class="vc-bezel-inner">
      <div v-if="sessions.length === 0" class="px-5 py-14 text-center">
        <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-[var(--bg-3)]">
          <svg class="h-5 w-5 text-[var(--fg-2)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <p class="mt-3 text-[13px] font-medium text-[var(--fg-2)]">No sessions yet.</p>
      </div>
      <NuxtLink
        v-for="session in sessions"
        :key="session.id"
        :to="`/p/${session.id}`"
        class="relative z-10 flex items-center gap-4 border-b border-[var(--border-0)] px-5 py-4 last:border-b-0 vc-transition-fast hover:bg-[var(--bg-2)]/40 vc-focus"
      >
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border text-[10px] font-bold uppercase" :class="channelColors[session.channel] || 'bg-[var(--bg-3)] text-[var(--fg-2)] border-[var(--border-0)]'">
          {{ session.channel.slice(0, 2) }}
        </div>
        <div class="min-w-0 flex-1">
          <div class="truncate text-[13px] font-semibold text-[var(--fg-0)]">{{ session.id }}</div>
          <div class="mt-0.5 text-[11px] text-[var(--fg-2)]">{{ session.channel }} · {{ session.user_id }}</div>
        </div>
        <div class="shrink-0 text-right">
          <div class="text-[11px] text-[var(--fg-2)]">{{ session.updated_at }}</div>
        </div>
        <svg class="h-3.5 w-3.5 shrink-0 text-[var(--fg-2)]/30 vc-transition-fast group-hover:translate-x-0.5 group-hover:text-[var(--fg-2)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
        </svg>
      </NuxtLink>
    </div>
  </section>
</template>
