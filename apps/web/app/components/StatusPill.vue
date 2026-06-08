<script setup lang="ts">
defineProps<{
  label: string
  tone?: 'good' | 'warn' | 'bad' | 'muted'
}>()
</script>

<template>
  <span
    class="inline-flex items-center rounded-full border px-2.5 py-1 text-xs"
    :class="{
      'border-emerald-300/20 bg-emerald-300/10 text-[var(--vc-good)]': tone === 'good',
      'border-amber-300/20 bg-amber-300/10 text-[var(--vc-warn)]': tone === 'warn',
      'border-red-300/20 bg-red-300/10 text-[var(--vc-bad)]': tone === 'bad',
      'border-white/10 bg-white/[0.04] text-[var(--vc-muted)]': !tone || tone === 'muted'
    }"
  >
    {{ label }}
  </span>
</template>

