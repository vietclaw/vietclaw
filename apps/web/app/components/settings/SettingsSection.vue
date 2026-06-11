<script setup lang="ts">
const props = defineProps<{
  title: string
  description?: string
  defaultOpen?: boolean
}>()

const open = ref(props.defaultOpen ?? false)
</script>

<template>
  <section class="rounded-2xl border border-vc-border-subtle bg-vc-surface">
    <button
      type="button"
      class="flex w-full items-center justify-between gap-3 px-4 py-3.5 text-left transition-colors hover:bg-vc-surface-hover"
      @click="open = !open"
    >
      <div>
        <h3 class="text-sm font-semibold text-vc-text">{{ title }}</h3>
        <p v-if="description" class="mt-0.5 text-xs text-vc-text-muted">{{ description }}</p>
      </div>
      <span class="shrink-0 text-sm text-vc-text-muted">{{ open ? '−' : '+' }}</span>
    </button>
    <div v-show="open" class="space-y-4 border-t border-vc-border-subtle px-4 py-4">
      <slot />
    </div>
  </section>
</template>
