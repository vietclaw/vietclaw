<script setup lang="ts">
import { Menu, Download } from '@lucide/vue'

defineEmits<{ toggleMobile: [] }>()

const { currentSession } = useChat()
const { status } = useDaemon()

function exportSession() {
  const s = currentSession()
  if (!s?.messages?.length) return
  const json = `data:text/json;charset=utf-8,${encodeURIComponent(JSON.stringify(s, null, 2))}`
  const link = document.createElement('a')
  link.href = json
  link.download = `${s.title.replace(/\s+/g, '_')}.json`
  document.body.appendChild(link)
  link.click()
  link.remove()
}
</script>

<template>
  <header class="flex h-12 shrink-0 items-center justify-between border-b border-vc-border-subtle px-4 md:px-6">
    <div class="flex min-w-0 items-center gap-2">
      <button type="button" class="vc-btn-ghost rounded-md p-1.5 md:hidden" @click="$emit('toggleMobile')">
        <Menu :size="18" :stroke-width="1.75" />
      </button>
      <span class="truncate text-sm text-vc-text-secondary md:hidden">
        {{ currentSession()?.title || 'Hội thoại' }}
      </span>
    </div>

    <div class="flex items-center gap-1">
      <button
        type="button"
        class="vc-btn-ghost rounded-md p-2"
        title="Xuất JSON"
        @click="exportSession"
      >
        <Download :size="15" :stroke-width="1.75" />
      </button>
      <span v-if="status?.version" class="hidden text-xs text-vc-text-muted sm:inline font-mono">
        {{ status.version }}
      </span>
    </div>
  </header>
</template>
