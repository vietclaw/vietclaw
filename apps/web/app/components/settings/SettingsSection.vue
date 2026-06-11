<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <section class="space-y-4 rounded-2xl border border-vc-border-subtle bg-vc-surface px-4 py-4">
    <div>
      <h3 class="text-sm font-semibold text-vc-text">{{ title }}</h3>
      <p v-if="description" class="mt-0.5 text-xs text-vc-text-muted">{{ description }}</p>
    </div>
    <slot />
  </section>
</template>
