<script setup lang="ts">
import type { ChannelStatus } from '~/types'
import { apiFetch } from '~/utils/api'

definePageMeta({ title: 'Channels' })

const { data: channels, refresh } = await useAsyncData('channels-page', () => apiFetch<ChannelStatus[]>('/api/channels'), { default: () => [] })
</script>

<template>
  <div class="space-y-5">
    <div class="grid gap-4 md:grid-cols-2 vc-stagger">
      <ChannelStatusCard v-for="channel in channels" :key="channel.name" :channel="channel" />
    </div>

    <section class="vc-bezel vc-noise vc-ambient vc-stagger" style="animation-delay: 100ms;">
      <div class="vc-bezel-inner relative px-5 py-5">
        <div class="relative z-10 flex items-start justify-between gap-4">
          <div class="flex items-start gap-3">
            <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--bg-3)]">
              <svg class="h-4 w-4 text-[var(--fg-2)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
                <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <h2 class="text-[13px] font-bold text-[var(--fg-0)]">Usage rules</h2>
              <div class="mt-2 space-y-1.5 text-[12px] text-[var(--fg-2)]">
                <p>Guild/group: mention the bot or reply to the bot.</p>
                <p>DM/private: chat normally.</p>
                <p>No slash commands in this phase.</p>
              </div>
            </div>
          </div>
          <button class="vc-btn-magnetic border border-[var(--border-1)] bg-[var(--bg-2)]/80 text-[var(--fg-1)] hover:bg-[var(--bg-3)] hover:text-[var(--fg-0)]" @click="() => refresh()">Refresh</button>
        </div>
      </div>
    </section>
  </div>
</template>
