<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <section class="vc-card space-y-4 px-5 py-5">
    <div>
      <h3 class="text-sm font-semibold tracking-tight text-vc-text">{{ title }}</h3>
      <p v-if="description" class="mt-0.5 text-xs text-vc-text-muted">{{ description }}</p>
    </div>
    <slot />
  </section>
</template>
