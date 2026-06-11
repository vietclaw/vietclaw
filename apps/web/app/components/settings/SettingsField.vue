<script setup lang="ts">
defineProps<{
  label: string
  hint?: string
}>()
</script>

<template>
  <div class="space-y-1.5">
    <label class="text-xs font-medium text-vc-text-secondary">{{ label }}</label>
    <slot />
    <p v-if="hint" class="text-[11px] leading-relaxed text-vc-text-muted">{{ hint }}</p>
  </div>
</template>
