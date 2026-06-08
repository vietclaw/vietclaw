<script setup lang="ts">
import type { ChannelStatus } from '~/types'
import { apiFetch } from '~/utils/api'
import { Radio, AlertCircle } from '@lucide/vue'

const toast = useToast()
const channels = ref<ChannelStatus[]>([])
const loading = ref(true)

const channelIcons: Record<string, string> = {
  discord: 'M20.317 4.37a19.791 19.791 0 00-4.885-1.515.074.074 0 00-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 00-5.487 0 12.64 12.64 0 00-.617-1.25.077.077 0 00-.079-.037A19.736 19.736 0 003.677 4.37a.07.07 0 00-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 00.031.057 19.9 19.9 0 005.993 3.03.078.078 0 00.084-.028c.462-.63.874-1.295 1.226-1.994a.076.076 0 00-.041-.106 13.107 13.107 0 01-1.872-.892.077.077 0 01-.008-.128 10.2 10.2 0 00.372-.292.074.074 0 01.077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 01.078.01c.12.098.246.198.373.292a.077.077 0 01-.006.127 12.299 12.299 0 01-1.873.892.077.077 0 00-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 00.084.028 19.839 19.839 0 006.002-3.03.077.077 0 00.032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 00-.031-.03z',
  telegram: 'M11.944 0A12 12 0 000 12a12 12 0 0012 12 12 12 0 0012-12A12 12 0 0012 0a12 12 0 00-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 01.171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.479.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z'
}

async function fetchChannels() {
  loading.value = true
  try {
    channels.value = await apiFetch<ChannelStatus[]>('/api/channels')
  } catch (err) {
    toast.add(err instanceof Error ? err.message : 'Failed to load channels', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(fetchChannels)
</script>

<template>
  <div class="max-w-3xl mx-auto space-y-4">
    <div class="flex items-center gap-2">
      <Radio :size="16" class="text-zinc-400" />
      <h2 class="text-sm font-semibold text-zinc-200">Channels</h2>
      <span class="text-[10px] text-zinc-500 font-mono">{{ channels.length }} active</span>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <div v-for="i in 2" :key="i" class="h-20 rounded-lg bg-zinc-900/40 animate-pulse" />
    </div>

    <!-- Empty -->
    <div v-else-if="channels.length === 0" class="text-center py-16">
      <Radio :size="32" class="mx-auto text-zinc-700 mb-3" />
      <p class="text-xs text-zinc-500">No channels configured.</p>
    </div>

    <!-- List -->
    <div v-else class="space-y-2">
      <div
        v-for="ch in channels"
        :key="ch.name"
        class="rounded-lg border border-zinc-900 bg-zinc-950/40 p-4 hover:border-zinc-800 transition-colors"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 rounded-lg bg-zinc-900 border border-zinc-800 flex items-center justify-center">
              <svg v-if="channelIcons[ch.name]" class="w-4 h-4 text-zinc-400" viewBox="0 0 24 24" fill="currentColor">
                <path :d="channelIcons[ch.name]" />
              </svg>
              <Radio v-else :size="16" class="text-zinc-500" />
            </div>
            <div>
              <h3 class="text-xs font-semibold text-zinc-200 capitalize">{{ ch.name }}</h3>
              <p class="text-[10px] text-zinc-500">
                {{ ch.name === 'discord' ? 'Needs @bot or reply in guilds' : 'Needs @botusername or reply in groups' }}
              </p>
            </div>
          </div>
          <div class="flex items-center gap-1.5 px-2 py-1 rounded text-[10px] font-mono"
            :class="ch.enabled
              ? (ch.running ? 'bg-emerald-950/30 text-emerald-400 border border-emerald-900/30' : 'bg-amber-950/30 text-amber-400 border border-amber-900/30')
              : 'bg-zinc-900 text-zinc-500 border border-zinc-800'"
          >
            <span class="w-1.5 h-1.5 rounded-full"
              :class="ch.enabled ? (ch.running ? 'bg-emerald-500' : 'bg-amber-500') : 'bg-zinc-600'"
            />
            {{ ch.enabled ? (ch.running ? 'running' : 'enabled') : 'disabled' }}
          </div>
        </div>
        <div v-if="ch.error" class="mt-3 flex items-start gap-2 p-2.5 rounded bg-rose-950/20 border border-rose-900/20">
          <AlertCircle :size="12" class="text-rose-400 mt-0.5 shrink-0" />
          <span class="text-[11px] text-rose-400">{{ ch.error }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
