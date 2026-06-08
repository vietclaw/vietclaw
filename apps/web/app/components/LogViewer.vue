<script setup lang="ts">
defineProps<{ logs: string[]; loading?: boolean }>()
</script>

<template>
  <section class="rounded-xl border border-white/[0.08] bg-[#07080b] p-4">
    <div v-if="loading" class="text-sm text-[var(--vc-muted)]">Loading logs...</div>
    <pre v-else class="vc-scrollbar max-h-[70vh] overflow-auto whitespace-pre-wrap text-xs leading-6 text-[#c9d3e4]">{{ logs.join('\n') || 'No logs yet.' }}</pre>
  </section>
</template>

