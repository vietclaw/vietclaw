<script setup lang="ts">
import type { BudgetStatus, ChannelStatus, DaemonStatus } from '~/types'
import { apiFetch } from '~/utils/api'

definePageMeta({ title: 'Overview' })

const { data: status } = await useAsyncData('status', () => apiFetch<DaemonStatus>('/status'), { default: () => null })
const { data: budget } = await useAsyncData('budget', () => apiFetch<BudgetStatus>('/api/budget'), { default: () => null })
const { data: channels } = await useAsyncData('channels', () => apiFetch<ChannelStatus[]>('/api/channels'), { default: () => [] })

const quickLinks = [
  { label: 'Memory', to: '/memory', icon: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547' },
  { label: 'Providers', to: '/providers', icon: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4' },
  { label: 'Logs', to: '/logs', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' }
]
</script>

<template>
  <div class="space-y-6">
    <!-- Hero stat row -->
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4 vc-stagger">
      <div v-for="(stat, i) in [
        { label: 'status', value: status?.mode || 'eco', icon: 'M13 10V3L4 14h7v7l9-11h-7z', color: 'var(--accent)' },
        { label: 'uptime', value: status?.uptime || '-', icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z', color: 'var(--success)' },
        { label: 'tasks', value: String(status?.max_concurrent_tasks || 1), icon: 'M4 6h16M4 10h16M4 14h16M4 18h16', color: 'var(--warning)' },
        { label: 'version', value: status?.version || '-', icon: 'M11.48 3.499a.562.562 0 011.04 0l2.125 5.111a.563.563 0 00.475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 00-.182.557l1.285 5.385a.562.562 0 01-.84.61l-4.725-2.885a.563.563 0 00-.586 0L6.982 20.54a.562.562 0 01-.84-.61l1.285-5.386a.562.562 0 00-.182-.557l-4.204-3.602a.563.563 0 01.321-.988l5.518-.442a.563.563 0 00.475-.345L11.48 3.5z', color: 'var(--accent)' }
      ]" :key="stat.label" class="vc-bezel vc-noise vc-ambient">
        <div class="vc-bezel-inner relative overflow-hidden px-5 py-5">
          <div class="relative z-10 flex items-center gap-3.5">
            <div class="flex h-10 w-10 items-center justify-center rounded-xl" :style="{ background: `color-mix(in srgb, ${stat.color} 10%, transparent)` }">
              <svg class="h-4.5 w-4.5" :style="{ color: stat.color }" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
                <path stroke-linecap="round" stroke-linejoin="round" :d="stat.icon" />
              </svg>
            </div>
            <div>
              <div class="text-[10px] font-semibold uppercase tracking-wider text-[var(--fg-2)]">{{ stat.label }}</div>
              <div class="mt-0.5 text-[14px] font-bold text-[var(--fg-0)]">{{ stat.value }}</div>
            </div>
          </div>
          <!-- Ambient glow -->
          <div class="absolute -right-8 -top-8 h-24 w-24 rounded-full blur-2xl opacity-0 transition-opacity duration-700 group-hover:opacity-100" :style="{ background: `color-mix(in srgb, ${stat.color} 8%, transparent)` }" />
          <div v-if="stat.label === 'status'" class="relative z-10 mt-3 flex items-center gap-1.5">
            <div class="relative">
              <div class="h-1.5 w-1.5 rounded-full" :class="status?.db_ok ? 'bg-[var(--success)]' : 'bg-[var(--warning)]'" />
              <div class="absolute inset-0 h-1.5 w-1.5 animate-ping rounded-full opacity-40" :class="status?.db_ok ? 'bg-[var(--success)]' : 'bg-[var(--warning)]'" />
            </div>
            <span class="text-[11px] font-medium text-[var(--fg-2)]">{{ status?.db_ok ? 'db ok' : 'db check' }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Main grid -->
    <div class="grid gap-5 xl:grid-cols-[1fr_380px] vc-stagger" style="animation-delay: 200ms;">
      <!-- Left: Chat -->
      <ChatPanel compact />

      <!-- Right stack -->
      <div class="space-y-4">
        <BudgetCard :budget="budget" />
        <div class="space-y-3">
          <ChannelStatusCard v-for="channel in channels" :key="channel.name" :channel="channel" />
        </div>
      </div>
    </div>

    <!-- Quick links -->
    <div class="grid gap-3 sm:grid-cols-3 vc-stagger" style="animation-delay: 400ms;">
      <NuxtLink
        v-for="link in quickLinks"
        :key="link.to"
        :to="link.to"
        class="group vc-bezel vc-noise"
      >
        <div class="vc-bezel-inner relative overflow-hidden px-5 py-4">
          <div class="relative z-10 flex items-center gap-3.5">
            <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-[var(--bg-3)] vc-transition-fast group-hover:bg-[var(--accent)]/10">
              <svg class="h-4 w-4 text-[var(--fg-2)] vc-transition-fast group-hover:text-[var(--accent-light)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
                <path stroke-linecap="round" stroke-linejoin="round" :d="link.icon" />
              </svg>
            </div>
            <span class="text-[13px] font-semibold text-[var(--fg-1)] vc-transition-fast group-hover:text-[var(--fg-0)]">{{ link.label }}</span>
            <svg class="ml-auto h-3.5 w-3.5 text-[var(--fg-2)]/20 vc-transition-fast group-hover:translate-x-1 group-hover:text-[var(--accent-light)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </div>
          <!-- Hover glow -->
          <div class="absolute -right-10 -top-10 h-32 w-32 rounded-full bg-[var(--accent)]/[0.04] blur-3xl opacity-0 vc-transition-slow group-hover:opacity-100" />
        </div>
      </NuxtLink>
    </div>
  </div>
</template>
