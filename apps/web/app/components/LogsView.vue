<script setup lang="ts">
import { apiFetch } from '~/utils/api'
import { FileText, RefreshCw } from '@lucide/vue'

const toast = useToast()
const logs = ref('')
const loading = ref(true)

async function fetchLogs() {
  loading.value = true
  try {
    const data = await apiFetch<{ logs: string }>('/api/logs/recent')
    logs.value = data.logs || ''
  } catch (err) {
    toast.add(err instanceof Error ? err.message : 'Failed to load logs', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(fetchLogs)

let interval: ReturnType<typeof setInterval>
onMounted(() => {
  interval = setInterval(fetchLogs, 8000)
})
onUnmounted(() => clearInterval(interval))
</script>

<template>
  <div class="max-w-3xl mx-auto space-y-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <FileText :size="16" class="text-zinc-400" />
        <h2 class="text-sm font-semibold text-zinc-200">Logs</h2>
        <span class="text-[10px] text-zinc-500 font-mono">auto-refresh 8s</span>
      </div>
      <button
        class="p-1.5 rounded hover:bg-zinc-900 text-zinc-500 hover:text-zinc-300 transition-colors"
        @click="fetchLogs"
      >
        <RefreshCw :size="14" />
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading && !logs" class="h-64 rounded-lg bg-zinc-900/40 animate-pulse" />

    <!-- Empty -->
    <div v-else-if="!logs && !loading" class="text-center py-16">
      <FileText :size="32" class="mx-auto text-zinc-700 mb-3" />
      <p class="text-xs text-zinc-500">No logs available.</p>
    </div>

    <!-- Logs -->
    <div v-else class="rounded-lg border border-zinc-900 bg-zinc-950/60 overflow-hidden">
      <pre class="p-4 font-mono text-[11px] leading-5 text-zinc-400 max-h-[60vh] overflow-auto vc-scrollbar whitespace-pre-wrap">{{ logs }}</pre>
    </div>
  </div>
</template>
