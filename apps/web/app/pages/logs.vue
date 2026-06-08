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
    <div class="flex items-center justify-between gap-4">
      <p class="text-sm text-[var(--vc-muted)]">Recent daemon logs. No websocket, just light refresh.</p>
      <button class="rounded-lg border border-white/[0.1] px-3 py-2 text-sm text-white" @click="refresh">Refresh</button>
    </div>
    <LogViewer :logs="logs" :loading="loading" />
  </div>
</template>

