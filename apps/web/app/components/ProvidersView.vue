<script setup lang="ts">
import type { ProviderConfig } from '~/types'
import { apiFetch } from '~/utils/api'
import { Server, Code, ToggleLeft, ToggleRight } from '@lucide/vue'

const toast = useToast()
const providers = ref<ProviderConfig[]>([])
const loading = ref(true)

async function fetchProviders() {
  loading.value = true
  try {
    providers.value = await apiFetch<ProviderConfig[]>('/api/providers')
  } catch (err) {
    toast.add(err instanceof Error ? err.message : 'Failed to load providers', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(fetchProviders)
</script>

<template>
  <div class="max-w-3xl mx-auto space-y-4">
    <div class="flex items-center gap-2">
      <Server :size="16" class="text-zinc-400" />
      <h2 class="text-sm font-semibold text-zinc-200">Providers</h2>
      <span class="text-[10px] text-zinc-500 font-mono">{{ providers.length }} configured</span>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-3">
      <div v-for="i in 3" :key="i" class="h-24 rounded-lg bg-zinc-900/40 animate-pulse" />
    </div>

    <!-- Empty -->
    <div v-else-if="providers.length === 0" class="text-center py-16">
      <Server :size="32" class="mx-auto text-zinc-700 mb-3" />
      <p class="text-xs text-zinc-500">No providers configured.</p>
    </div>

    <!-- List -->
    <div v-else class="space-y-2">
      <div
        v-for="p in providers"
        :key="p.id"
        class="rounded-lg border border-zinc-900 bg-zinc-950/40 p-4 hover:border-zinc-800 transition-colors"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 rounded-lg bg-zinc-900 border border-zinc-800 flex items-center justify-center">
              <Code :size="16" class="text-zinc-400" />
            </div>
            <div>
              <h3 class="text-xs font-semibold text-zinc-200">{{ p.id }}</h3>
              <p class="text-[10px] text-zinc-500 font-mono">{{ p.type }}</p>
            </div>
          </div>
          <div class="flex items-center gap-1.5 px-2 py-1 rounded text-[10px] font-mono"
            :class="p.enabled ? 'bg-emerald-950/30 text-emerald-400 border border-emerald-900/30' : 'bg-zinc-900 text-zinc-500 border border-zinc-800'"
          >
            <span class="w-1.5 h-1.5 rounded-full" :class="p.enabled ? 'bg-emerald-500' : 'bg-zinc-600'" />
            {{ p.enabled ? 'active' : 'off' }}
          </div>
        </div>
        <div class="mt-3 pt-3 border-t border-zinc-900/60 grid grid-cols-2 gap-3">
          <div>
            <span class="text-[10px] text-zinc-600 block mb-0.5">model</span>
            <span class="text-[11px] text-zinc-300 font-mono">{{ p.default_model || 'not set' }}</span>
          </div>
          <div>
            <span class="text-[10px] text-zinc-600 block mb-0.5">api key</span>
            <span class="text-[11px] text-zinc-400 font-mono">{{ p.api_key_env ? `env:${p.api_key_env}` : 'not needed' }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
