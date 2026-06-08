<script setup lang="ts">
import type { BudgetStatus, ChannelStatus, DaemonStatus } from '~/types'
import { apiFetch } from '~/utils/api'

definePageMeta({ title: 'Overview' })

const { data: status } = await useAsyncData('status', () => apiFetch<DaemonStatus>('/status'), { default: () => null })
const { data: budget } = await useAsyncData('budget', () => apiFetch<BudgetStatus>('/api/budget'), { default: () => null })
const { data: channels } = await useAsyncData('channels', () => apiFetch<ChannelStatus[]>('/api/channels'), { default: () => [] })
</script>

<template>
  <div class="space-y-6">
    <section class="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
      <div class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)] p-6">
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div>
            <p class="text-xs uppercase tracking-[0.2em] text-[var(--vc-subtle)]">daemon</p>
            <h2 class="mt-2 text-3xl font-semibold tracking-tight text-white">Daemon online</h2>
            <p class="mt-3 max-w-2xl text-sm leading-6 text-[var(--vc-muted)]">
              Lightweight personal agent runtime. Model routing, memory, channels, and tools stay behind one Go process.
            </p>
          </div>
          <StatusPill :label="status?.db_ok ? 'db ok' : 'db check'" :tone="status?.db_ok ? 'good' : 'warn'" />
        </div>
        <div class="mt-8 grid gap-3 sm:grid-cols-3">
          <div class="rounded-lg bg-white/[0.04] p-4">
            <div class="text-xs text-[var(--vc-subtle)]">mode</div>
            <div class="mt-1 text-lg font-semibold">{{ status?.mode || 'eco' }}</div>
          </div>
          <div class="rounded-lg bg-white/[0.04] p-4">
            <div class="text-xs text-[var(--vc-subtle)]">uptime</div>
            <div class="mt-1 text-lg font-semibold">{{ status?.uptime || '-' }}</div>
          </div>
          <div class="rounded-lg bg-white/[0.04] p-4">
            <div class="text-xs text-[var(--vc-subtle)]">tasks</div>
            <div class="mt-1 text-lg font-semibold">{{ status?.max_concurrent_tasks || 1 }}</div>
          </div>
        </div>
      </div>
      <BudgetCard :budget="budget" />
    </section>

    <section class="grid gap-4 xl:grid-cols-[1fr_380px]">
      <ChatPanel compact />
      <div class="space-y-4">
        <ChannelStatusCard v-for="channel in channels" :key="channel.name" :channel="channel" />
        <div class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)] p-5">
          <h2 class="text-sm font-medium text-white">Quick links</h2>
          <div class="mt-4 grid gap-2">
            <NuxtLink to="/memory" class="rounded-lg bg-white/[0.04] px-3 py-2 text-sm text-[var(--vc-muted)] hover:text-white vc-focus">Memory</NuxtLink>
            <NuxtLink to="/providers" class="rounded-lg bg-white/[0.04] px-3 py-2 text-sm text-[var(--vc-muted)] hover:text-white vc-focus">Providers</NuxtLink>
            <NuxtLink to="/logs" class="rounded-lg bg-white/[0.04] px-3 py-2 text-sm text-[var(--vc-muted)] hover:text-white vc-focus">Logs</NuxtLink>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

