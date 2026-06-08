<script setup lang="ts">
defineProps<{ logs: string[]; loading?: boolean }>()
</script>

<template>
  <section class="vc-bezel vc-noise overflow-hidden">
    <div class="vc-bezel-inner">
      <div v-if="loading" class="relative z-10 flex items-center gap-2 px-5 py-4 text-[12px] font-medium text-[var(--fg-2)]">
        <span class="h-1.5 w-1.5 animate-pulse rounded-full bg-[var(--accent)]" />
        Loading logs...
      </div>
      <div v-else-if="logs.length === 0" class="relative z-10 px-5 py-14 text-center">
        <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-[var(--bg-3)]">
          <svg class="h-5 w-5 text-[var(--fg-2)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
        </div>
        <p class="mt-3 text-[13px] font-medium text-[var(--fg-2)]">No logs yet.</p>
      </div>
      <pre v-else class="relative z-10 vc-scrollbar max-h-[62vh] overflow-auto p-5 font-mono text-[11px] leading-5 text-[var(--fg-1)]">{{ logs.join('\n') }}</pre>
    </div>
  </section>
</template>
