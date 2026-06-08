<script setup lang="ts">
defineProps<{
  label: string
  tone?: 'good' | 'warn' | 'bad' | 'muted'
  size?: 'sm' | 'md'
}>()
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1 text-[11px] font-semibold"
    :class="[
      tone === 'good' && 'bg-[var(--success)]/10 text-[var(--success)]',
      tone === 'warn' && 'bg-[var(--warning)]/10 text-[var(--warning)]',
      tone === 'bad' && 'bg-[var(--danger)]/10 text-[var(--danger)]',
      (!tone || tone === 'muted') && 'bg-[var(--bg-3)]/80 text-[var(--fg-2)]',
      size === 'sm' && '!text-[10px] !px-2 !py-0.5'
    ]"
  >
    <span
      v-if="tone && tone !== 'muted'"
      class="h-1.5 w-1.5 rounded-full"
      :class="[
        tone === 'good' && 'bg-[var(--success)]',
        tone === 'warn' && 'bg-[var(--warning)]',
        tone === 'bad' && 'bg-[var(--danger)]'
      ]"
    />
    {{ label }}
  </span>
</template>
