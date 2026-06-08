<script setup lang="ts">
import type { MemoryRecord } from '~/types'

defineProps<{ memory: MemoryRecord }>()
</script>

<template>
  <article class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)] p-5">
    <div class="flex flex-wrap items-center gap-2">
      <StatusPill :label="memory.kind" tone="muted" />
      <StatusPill :label="memory.confidence" tone="good" />
    </div>
    <p class="mt-4 text-sm leading-6 text-white">{{ memory.content }}</p>
    <div class="mt-4 text-xs text-[var(--vc-subtle)]">{{ memory.scope }}</div>
  </article>
</template>

