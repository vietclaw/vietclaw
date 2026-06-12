<script setup lang="ts">
import { Menu, Download } from '@lucide/vue'

defineEmits<{ toggleMobile: [] }>()

const { currentSession } = useChat()
const { status } = useDaemon()
const { t } = useI18n()

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
  <header class="flex h-13 shrink-0 items-center justify-between px-4 md:px-6">
    <div class="flex min-w-0 items-center gap-2">
      <button type="button" class="vc-btn-ghost rounded-full p-1.5 md:hidden" @click="$emit('toggleMobile')">
        <Menu :size="18" :stroke-width="1.5" />
      </button>
      <span class="truncate text-sm font-medium text-vc-text-secondary md:hidden">
        {{ currentSession()?.title || t('chat.title') }}
      </span>
    </div>

    <div class="flex items-center gap-2">
      <span v-if="status?.version" class="hidden rounded-full border border-vc-border-subtle bg-vc-surface px-2.5 py-0.5 font-mono text-[11px] text-vc-text-muted sm:inline">
        {{ status.version }}
      </span>
      <button
        type="button"
        class="vc-btn-ghost rounded-full p-2"
        :title="t('chat.exportJson')"
        @click="exportSession"
      >
        <Download :size="15" :stroke-width="1.5" />
      </button>
    </div>
  </header>
</template>
