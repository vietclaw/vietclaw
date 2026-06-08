<script setup lang="ts">
import type { ProviderConfig } from '~/types'
import { apiFetch } from '~/utils/api'

definePageMeta({ title: 'Providers' })

const { data: providers } = await useAsyncData('providers', () => apiFetch<ProviderConfig[]>('/api/providers'), { default: () => [] })
</script>

<template>
  <div class="space-y-5">
    <section class="vc-bezel vc-noise vc-ambient vc-stagger">
      <div class="vc-bezel-inner relative px-5 py-5">
        <div class="relative z-10 flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--bg-3)]">
            <svg class="h-4 w-4 text-[var(--fg-2)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <h2 class="text-[13px] font-bold text-[var(--fg-0)]">Provider router</h2>
            <p class="text-[11px] text-[var(--fg-2)]">Tokens live in env vars, not config. VietClaw starts with mock by default.</p>
          </div>
        </div>
      </div>
    </section>

    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3 vc-stagger" style="animation-delay: 100ms;">
      <ProviderCard v-for="provider in providers" :key="provider.id" :provider="provider" />
    </div>
  </div>
</template>
