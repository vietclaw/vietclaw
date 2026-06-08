<script setup lang="ts">
import { apiFetch } from '~/utils/api'

definePageMeta({ title: 'Logs' })

const logs = ref<string[]>([])
const loading = ref(false)

async function refresh() {
  loading.value = true
  try {
    logs.value = await apiFetch<string[]>('/api/logs/recent')
  } finally {
    loading.value = false
  }
}

let timer: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  void refresh()
  timer = setInterval(refresh, 8000)
})
onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between vc-stagger">
      <div>
        <h2 class="text-[14px] font-bold text-[var(--fg-0)]">Daemon logs</h2>
        <p class="text-[11px] text-[var(--fg-2)]">Auto-refreshes every 8s</p>
      </div>
      <button class="vc-btn-magnetic border border-[var(--border-1)] bg-[var(--bg-2)]/80 text-[var(--fg-1)] hover:bg-[var(--bg-3)] hover:text-[var(--fg-0)]" @click="refresh">Refresh</button>
    </div>
    <LogViewer :logs="logs" :loading="loading" />
  </div>
</template>
