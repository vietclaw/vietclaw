<script setup lang="ts">
import type { MemoryRecord } from '~/types'

defineProps<{ memory: MemoryRecord }>()
</script>

<template>
  <article class="vc-bezel vc-noise group vc-transition">
    <div class="vc-bezel-inner relative overflow-hidden px-5 py-5">
      <div class="relative z-10 flex items-center gap-1.5">
        <StatusPill :label="memory.kind" tone="muted" size="sm" />
        <StatusPill :label="memory.confidence" tone="good" size="sm" />
      </div>
      <p class="relative z-10 mt-3.5 text-[13px] leading-relaxed text-[var(--fg-0)]">{{ memory.content }}</p>
      <div class="relative z-10 mt-4 flex items-center justify-between border-t border-[var(--border-0)] pt-3">
        <span class="text-[10px] font-medium text-[var(--fg-2)]">{{ memory.scope }}</span>
        <span class="text-[10px] text-[var(--fg-2)]">{{ memory.created_at }}</span>
      </div>
    </div>
  </article>
</template>
