<script setup lang="ts">
import type { ProviderConfig } from '~/types'

defineProps<{ provider: ProviderConfig }>()
</script>

<template>
  <article class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)] p-5">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h2 class="text-base font-semibold text-white">{{ provider.id }}</h2>
        <p class="mt-1 text-sm text-[var(--vc-muted)]">{{ provider.type }}</p>
      </div>
      <StatusPill :label="provider.enabled ? 'enabled' : 'disabled'" :tone="provider.enabled ? 'good' : 'muted'" />
    </div>
    <div class="mt-5 space-y-2 text-sm">
      <div class="flex justify-between gap-4">
        <span class="text-[var(--vc-subtle)]">model</span>
        <span class="text-white">{{ provider.default_model || 'not set' }}</span>
      </div>
      <div class="flex justify-between gap-4">
        <span class="text-[var(--vc-subtle)]">token</span>
        <span class="text-[var(--vc-muted)]">{{ provider.api_key_env ? 'env var' : 'not needed' }}</span>
      </div>
    </div>
  </article>
</template>

