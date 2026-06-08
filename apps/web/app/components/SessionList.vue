<script setup lang="ts">
import type { Session } from '~/types'

defineProps<{ sessions: Session[] }>()
</script>

<template>
  <section class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)]">
    <NuxtLink
      v-for="session in sessions"
      :key="session.id"
      :to="`/sessions/${session.id}`"
      class="block border-b border-white/[0.06] px-5 py-4 last:border-b-0 hover:bg-white/[0.035] vc-focus"
    >
      <div class="flex items-center justify-between gap-4">
        <div class="min-w-0">
          <div class="truncate text-sm font-medium text-white">{{ session.id }}</div>
          <div class="mt-1 text-xs text-[var(--vc-muted)]">{{ session.channel }} · {{ session.user_id }}</div>
        </div>
        <div class="shrink-0 text-xs text-[var(--vc-subtle)]">{{ session.updated_at }}</div>
      </div>
    </NuxtLink>
    <div v-if="sessions.length === 0" class="px-5 py-8 text-sm text-[var(--vc-muted)]">No sessions yet.</div>
  </section>
</template>

