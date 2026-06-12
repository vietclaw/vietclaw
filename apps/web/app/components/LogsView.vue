<script setup lang="ts">
import { apiFetch } from '~/utils/api'
import { RefreshCw } from '@lucide/vue'

defineProps<{ embedded?: boolean }>()

const toast = useToast()
const { t } = useI18n()
const logs = ref('')
const loading = ref(true)

async function fetchLogs() {
  loading.value = true
  try {
    const data = await apiFetch<{ logs: string }>('/api/logs/recent')
    logs.value = data.logs || ''
  } catch (err) {
    toast.add(err instanceof Error ? err.message : t('logs.loadFailed'), 'error')
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
        <h1 class="vc-display text-2xl font-medium text-vc-text">{{ t('logs.title') }}</h1>
        <p class="mt-1.5 text-sm text-vc-text-muted">{{ t('logs.autoRefresh') }}</p>
      </div>
      <button type="button" class="vc-btn vc-btn-outline p-2" :aria-label="t('logs.reload')" @click="fetchLogs">
        <RefreshCw :size="15" :stroke-width="1.5" />
      </button>
    </div>

    <div v-if="loading && !logs" class="h-48 animate-pulse rounded-2xl bg-vc-bg-subtle" />

    <div v-else-if="!logs && !loading" class="vc-card flex flex-col items-center px-6 py-12 text-center">
      <p class="vc-display text-lg text-vc-text">{{ t('logs.empty.title') }}</p>
      <p class="mt-1.5 text-sm text-vc-text-muted">{{ t('logs.empty.desc') }}</p>
    </div>

    <div v-else class="vc-card overflow-hidden">
      <pre class="vc-scrollbar max-h-[65vh] overflow-auto whitespace-pre-wrap p-4 font-mono text-xs leading-relaxed text-vc-text-secondary">{{ logs }}</pre>
    </div>
  </div>
</template>
