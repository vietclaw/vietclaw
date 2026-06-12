<script setup lang="ts">
const { t } = useI18n()
useHead({ title: () => `${t('chat.title')} · VietClaw` })

const route = useRoute()
const {
  sessions,
  switchSession,
  createSession,
  hydrateSessionFromAPI,
  loadChildrenForParent,
  setExpandedRoot,
  isSpawnSessionId,
  parseParentId,
} = useChat()

const sessionId = computed(() => decodeURIComponent(String(route.params.id)))

watch(sessionId, async (id) => {
  if (sessions.value.some(s => s.id === id)) {
    switchSession(id)
    const current = sessions.value.find(s => s.id === id)
    if (current?.kind === 'spawn') {
      await hydrateSessionFromAPI(id)
      if (current.parentId) {
        await loadChildrenForParent(current.parentId)
      }
    } else {
      await loadChildrenForParent(id)
    }
    return
  }

  if (isSpawnSessionId(id)) {
    const parentId = parseParentId(id)
    setExpandedRoot(parentId)
    const hydrated = await hydrateSessionFromAPI(id)
    if (hydrated) {
      switchSession(id)
      await loadChildrenForParent(parentId)
      return
    }
  }

  const fallback = sessions.value.find(s => s.kind !== 'spawn')
  if (fallback) {
    await navigateTo(`/p/${encodeURIComponent(fallback.id)}`, { replace: true })
    return
  }

  const created = createSession()
  await navigateTo(`/p/${encodeURIComponent(created.id)}`, { replace: true })
}, { immediate: true })
</script>

<template>
  <div class="h-full min-h-0">
    <ChatPanel />
  </div>
</template>
