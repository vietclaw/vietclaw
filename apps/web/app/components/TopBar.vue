<script setup lang="ts">
defineProps<{ title: string }>()

const nav = [
  { label: 'Chat', to: '/chat' },
  { label: 'Memory', to: '/memory' },
  { label: 'Logs', to: '/logs' }
]
</script>

<template>
  <header class="sticky top-0 z-20 border-b border-white/[0.07] bg-[#090a0d]/85 px-4 py-3 backdrop-blur sm:px-6 lg:px-8">
    <div class="flex items-center justify-between gap-4">
      <div>
        <div class="text-xs uppercase tracking-[0.18em] text-[var(--vc-subtle)]">VietClaw</div>
        <h1 class="mt-0.5 text-lg font-semibold tracking-tight text-white">{{ title }}</h1>
      </div>
      <nav class="flex gap-1 lg:hidden">
        <NuxtLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          class="rounded-lg px-3 py-2 text-sm text-[var(--vc-muted)] vc-focus"
          active-class="bg-white/[0.08] text-white"
        >
          {{ item.label }}
        </NuxtLink>
      </nav>
    </div>
  </header>
</template>

