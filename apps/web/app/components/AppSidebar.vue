<script setup lang="ts">
const items = [
  { label: 'Overview', to: '/' },
  { label: 'Chat', to: '/chat' },
  { label: 'Memory', to: '/memory' },
  { label: 'Sessions', to: '/sessions' },
  { label: 'Providers', to: '/providers' },
  { label: 'Budget', to: '/budget' },
  { label: 'Channels', to: '/channels' },
  { label: 'Logs', to: '/logs' }
]
</script>

<template>
  <aside class="sticky top-0 hidden h-screen flex-col border-r border-white/[0.07] px-4 py-5 lg:flex">
    <NuxtLink to="/" class="mb-8 block rounded-lg px-3 py-2 vc-focus">
      <div class="text-lg font-semibold tracking-tight">VietClaw</div>
      <div class="mt-1 text-xs text-[var(--vc-muted)]">personal agent runtime</div>
    </NuxtLink>
    <nav class="space-y-1">
      <NuxtLink
        v-for="item in items"
        :key="item.to"
        :to="item.to"
        class="block rounded-lg px-3 py-2 text-sm text-[var(--vc-muted)] transition hover:bg-white/[0.06] hover:text-white vc-focus"
        active-class="bg-white/[0.08] text-white"
      >
        {{ item.label }}
      </NuxtLink>
    </nav>
    <div class="mt-auto rounded-lg border border-white/[0.08] bg-white/[0.035] p-3">
      <div class="text-xs uppercase tracking-[0.18em] text-[var(--vc-subtle)]">runtime</div>
      <div class="mt-2 text-sm text-white">Go + SQLite</div>
      <div class="mt-1 text-xs text-[var(--vc-muted)]">Node stays in build time.</div>
    </div>
  </aside>
</template>

