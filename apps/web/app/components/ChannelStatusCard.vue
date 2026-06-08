<script setup lang="ts">
import type { ChannelStatus } from '~/types'

defineProps<{ channel: ChannelStatus }>()
</script>

<template>
  <section class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)] p-5">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h2 class="text-sm font-medium capitalize text-white">{{ channel.name }}</h2>
        <p class="mt-1 text-sm text-[var(--vc-muted)]">
          {{ channel.name === 'discord' ? 'Guilds need @bot or reply.' : 'Groups need @botusername or reply.' }}
        </p>
      </div>
      <StatusPill
        :label="channel.enabled ? (channel.running ? 'running' : 'enabled') : 'disabled'"
        :tone="channel.enabled ? (channel.running ? 'good' : 'warn') : 'muted'"
      />
    </div>
    <p v-if="channel.error" class="mt-4 rounded-lg border border-amber-300/15 bg-amber-300/10 px-3 py-2 text-sm text-[var(--vc-warn)]">
      {{ channel.error }}
    </p>
  </section>
</template>

