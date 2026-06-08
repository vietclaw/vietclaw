<script setup lang="ts">
import type { ChannelStatus } from '~/types'
import { apiFetch } from '~/utils/api'

definePageMeta({ title: 'Channels' })

const { data: channels, refresh } = await useAsyncData('channels-page', () => apiFetch<ChannelStatus[]>('/api/channels'), { default: () => [] })
</script>

<template>
  <div class="space-y-5">
    <div class="grid gap-4 md:grid-cols-2">
      <ChannelStatusCard v-for="channel in channels" :key="channel.name" :channel="channel" />
    </div>
    <section class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)] p-5">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h2 class="text-sm font-medium text-white">Usage rules</h2>
          <div class="mt-3 space-y-2 text-sm text-[var(--vc-muted)]">
            <p>Guild/group: mention the bot or reply to the bot.</p>
            <p>DM/private: chat normally.</p>
            <p>No slash commands in this phase.</p>
          </div>
        </div>
        <button class="rounded-lg border border-white/[0.1] px-3 py-2 text-sm text-white" @click="() => refresh()">Refresh</button>
      </div>
    </section>
  </div>
</template>
