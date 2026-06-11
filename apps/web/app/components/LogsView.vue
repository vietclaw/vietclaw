<script setup lang="ts">
import { apiFetch } from '~/utils/api'
import { RefreshCw } from '@lucide/vue'

defineProps<{ embedded?: boolean }>()

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

let interval: ReturnType<typeof setInterval>
onMounted(() => {
  void fetchLogs()
  interval = setInterval(fetchLogs, 8000)
})
onUnmounted(() => clearInterval(interval))
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="text-lg font-semibold tracking-tight text-vc-text">Logs</h1>
        <p class="mt-1 text-sm text-vc-text-muted">Tự động cập nhật mỗi 8 giây</p>
      </div>
      <button type="button" class="vc-btn vc-btn-ghost p-2" @click="fetchLogs">
        <RefreshCw :size="15" :stroke-width="1.75" />
      </button>
    </div>

    <div v-if="loading && !logs" class="h-48 rounded-lg bg-vc-bg-subtle animate-pulse" />

    <p v-else-if="!logs && !loading" class="text-sm text-vc-text-muted">Chưa có log.</p>

    <div v-else class="rounded-lg border border-vc-border bg-vc-surface overflow-hidden">
      <pre class="max-h-[65vh] overflow-auto p-4 font-mono text-xs leading-relaxed text-vc-text-secondary vc-scrollbar whitespace-pre-wrap">{{ logs }}</pre>
    </div>
  </div>
</template>
