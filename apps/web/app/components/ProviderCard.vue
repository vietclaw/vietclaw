<script setup lang="ts">
import type { ProviderConfig } from '~/types'

defineProps<{ provider: ProviderConfig }>()
</script>

<template>
  <article class="vc-bezel vc-noise group vc-transition">
    <div class="vc-bezel-inner relative overflow-hidden px-5 py-5">
      <!-- Hover glow -->
      <div class="absolute inset-0 opacity-0 vc-transition-slow group-hover:opacity-100">
        <div class="absolute -left-10 -top-10 h-32 w-32 rounded-full bg-[var(--accent)]/[0.04] blur-3xl" />
      </div>

      <div class="relative z-10 flex items-start justify-between gap-3">
        <div class="flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--bg-3)] text-[var(--fg-2)] vc-transition-fast group-hover:bg-[var(--accent)]/10 group-hover:text-[var(--accent-light)]">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
          </div>
          <div>
            <h2 class="text-[13px] font-bold text-[var(--fg-0)]">{{ provider.id }}</h2>
            <p class="text-[11px] text-[var(--fg-2)]">{{ provider.type }}</p>
          </div>
        </div>
        <StatusPill :label="provider.enabled ? 'on' : 'off'" :tone="provider.enabled ? 'good' : 'muted'" size="sm" />
      </div>
      <div class="relative z-10 mt-5 space-y-3 border-t border-[var(--border-0)] pt-4">
        <div class="flex items-center justify-between">
          <span class="text-[11px] font-medium text-[var(--fg-2)]">model</span>
          <span class="text-[12px] font-semibold text-[var(--fg-0)]">{{ provider.default_model || 'not set' }}</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-[11px] font-medium text-[var(--fg-2)]">token</span>
          <span class="text-[12px] text-[var(--fg-2)]">{{ provider.api_key_env ? 'env var' : 'not needed' }}</span>
        </div>
      </div>
    </div>
  </article>
</template>
