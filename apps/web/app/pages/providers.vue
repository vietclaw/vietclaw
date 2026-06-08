<script setup lang="ts">
import type { ProviderConfig } from '~/types'
import { apiFetch } from '~/utils/api'

definePageMeta({ title: 'Providers' })

const { data: providers } = await useAsyncData('providers', () => apiFetch<ProviderConfig[]>('/api/providers'), { default: () => [] })
</script>

<template>
  <div class="space-y-5">
    <section class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)] p-5">
      <h2 class="text-sm font-medium text-white">Provider router</h2>
      <p class="mt-2 text-sm text-[var(--vc-muted)]">Tokens live in env vars, not config. VietClaw starts with mock by default.</p>
    </section>
    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <ProviderCard v-for="provider in providers" :key="provider.id" :provider="provider" />
    </div>
  </div>
</template>

