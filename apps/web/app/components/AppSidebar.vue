<script setup lang="ts">
defineProps<{ open: boolean }>()
defineEmits<{ close: [] }>()

const route = useRoute()

const items = [
  { label: 'Overview', to: '/', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-4 0a1 1 0 01-1-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 01-1 1h-2' },
  { label: 'Chat', to: '/chat', icon: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z' },
  { label: 'Memory', to: '/memory', icon: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547' },
  { label: 'Sessions', to: '/sessions', icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z' },
  { label: 'Providers', to: '/providers', icon: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4' },
  { label: 'Budget', to: '/budget', icon: 'M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
  { label: 'Channels', to: '/channels', icon: 'M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1' },
  { label: 'Logs', to: '/logs', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' }
]
</script>

<template>
  <aside
    class="fixed inset-y-0 left-0 z-50 flex w-[240px] flex-col transition-all duration-500 lg:translate-x-0"
    :class="open ? 'translate-x-0' : '-translate-x-full'"
  >
    <!-- Glass background -->
    <div class="absolute inset-0 bg-[var(--bg-1)]/80 backdrop-blur-2xl border-r border-[var(--border-0)]" />

    <div class="relative z-10 flex h-full flex-col">
      <!-- Logo -->
      <div class="flex h-16 items-center gap-3 px-5">
        <div class="relative flex h-8 w-8 items-center justify-center rounded-xl bg-gradient-to-br from-[var(--accent)] to-purple-500 shadow-lg shadow-[var(--accent)]/20">
          <svg class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </div>
        <div>
          <span class="text-[14px] font-bold tracking-tight text-[var(--fg-0)]">VietClaw</span>
          <span class="ml-1.5 inline-flex items-center rounded-md bg-[var(--accent)]/10 px-1.5 py-0.5 text-[9px] font-semibold text-[var(--accent-light)]">v0.1</span>
        </div>
      </div>

      <!-- Divider -->
      <div class="mx-4 h-px bg-gradient-to-r from-transparent via-[var(--border-1)] to-transparent" />

      <!-- Nav -->
      <nav class="flex-1 overflow-y-auto px-3 py-4 vc-scrollbar">
        <NuxtLink
          v-for="item in items"
          :key="item.to"
          :to="item.to"
          class="group mb-1 flex items-center gap-3 rounded-xl px-3 py-2.5 text-[13px] font-medium text-[var(--fg-2)] vc-transition-fast vc-focus"
          :class="route.path === item.to
            ? 'bg-[var(--accent)]/10 text-[var(--accent-light)] shadow-sm shadow-[var(--accent)]/5'
            : 'hover:bg-[var(--bg-3)]/60 hover:text-[var(--fg-1)]'"
        >
          <svg
            class="h-[18px] w-[18px] shrink-0 vc-transition-fast"
            :class="route.path === item.to ? 'text-[var(--accent-light)]' : 'text-[var(--fg-2)] group-hover:text-[var(--fg-1)]'"
            fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75"
          >
            <path stroke-linecap="round" stroke-linejoin="round" :d="item.icon" />
          </svg>
          {{ item.label }}
          <div
            v-if="route.path === item.to"
            class="ml-auto h-1.5 w-1.5 rounded-full bg-[var(--accent)] vc-pulse-dot"
          />
        </NuxtLink>
      </nav>

      <!-- Footer -->
      <div class="border-t border-[var(--border-0)] p-3">
        <div class="rounded-xl bg-[var(--bg-2)]/60 px-3.5 py-3">
          <div class="flex items-center gap-2">
            <div class="relative">
              <div class="h-2 w-2 rounded-full bg-[var(--success)]" />
              <div class="absolute inset-0 h-2 w-2 animate-ping rounded-full bg-[var(--success)] opacity-40" />
            </div>
            <span class="text-[11px] font-semibold text-[var(--fg-1)]">System online</span>
          </div>
          <div class="mt-1.5 text-[10px] font-medium text-[var(--fg-2)]">Go + SQLite</div>
        </div>
      </div>
    </div>
  </aside>
</template>
