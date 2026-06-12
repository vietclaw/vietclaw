<script setup lang="ts">
import { ArrowLeft, Menu, Download } from '@lucide/vue'
import type { SpawnStatus } from '~/composables/useChat'

defineEmits<{ toggleMobile: [] }>()

const { currentSession, sessions, sessionPath, parseParentId } = useChat()
const { status } = useDaemon()
const { t } = useI18n()

const session = computed(() => currentSession())
const isReadOnly = computed(() => session.value?.readOnly === true)

const parentSession = computed(() => {
  const parentId = session.value?.parentId ?? parseParentId(session.value?.id ?? '')
  return sessions.value.find(s => s.id === parentId && s.kind !== 'spawn')
})

const parentSessionPath = computed(() => {
  const id = parentSession.value?.id
  return id ? sessionPath(id) : '/'
})

function spawnStatusLabel(status?: SpawnStatus): string {
  if (!status) return ''
  const key = `chat.spawn.${status}`
  const label = t(key)
  return label === key ? status : label
}

function spawnStatusClass(status?: SpawnStatus): string {
  if (status === 'done') return 'text-vc-success'
  if (status === 'failed') return 'text-vc-error'
  return 'text-vc-accent'
}

function exportSession() {
  const s = currentSession()
  if (!s?.messages?.length || s.readOnly) return
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
  <header class="flex h-13 shrink-0 items-center justify-between gap-3 border-b border-vc-border-subtle px-4 md:px-6">
    <div class="flex min-w-0 flex-1 items-center gap-2">
      <button type="button" class="vc-btn-ghost rounded-full p-1.5 md:hidden" @click="$emit('toggleMobile')">
        <Menu :size="18" :stroke-width="1.5" />
      </button>

      <template v-if="isReadOnly">
        <NuxtLink
          :to="parentSessionPath"
          class="vc-link flex shrink-0 items-center gap-1 text-xs"
        >
          <ArrowLeft :size="14" :stroke-width="1.75" />
          <span class="hidden max-w-[8rem] truncate sm:inline">{{ parentSession?.title || t('nav.backToParent') }}</span>
        </NuxtLink>
        <span class="text-vc-text-muted">/</span>
        <span class="truncate text-sm font-medium text-vc-text">
          {{ session?.agentId || session?.title }}
        </span>
        <span
          v-if="session?.spawnStatus"
          class="shrink-0 text-xs"
          :class="spawnStatusClass(session.spawnStatus)"
        >
          {{ spawnStatusLabel(session.spawnStatus) }}
        </span>
      </template>

      <span v-else class="truncate text-sm font-medium text-vc-text-secondary">
        {{ session?.title || t('chat.title') }}
      </span>
    </div>

    <div class="flex shrink-0 items-center gap-2">
      <span v-if="status?.version" class="hidden rounded-full border border-vc-border-subtle bg-vc-surface px-2.5 py-0.5 font-mono text-[11px] text-vc-text-muted sm:inline">
        {{ status.version }}
      </span>
      <button
        v-if="!isReadOnly"
        type="button"
        class="vc-btn-ghost rounded-full p-2"
        :title="t('chat.exportJson')"
        @click="exportSession"
      >
        <Download :size="15" :stroke-width="1.75" />
      </button>
    </div>
  </header>
</template>
