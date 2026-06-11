<script setup lang="ts">
definePageMeta({ title: 'Chat' })

const route = useRoute()
const { sessions, switchSession, createSession } = useChat()

const sessionId = computed(() => String(route.params.id))

watch(sessionId, async (id) => {
  if (sessions.value.some(s => s.id === id)) {
    switchSession(id)
    return
  }

  const fallback = sessions.value[0]
  if (fallback) {
    await navigateTo(`/p/${fallback.id}`, { replace: true })
    return
  }

  const created = createSession()
  await navigateTo(`/p/${created.id}`, { replace: true })
}, { immediate: true })
</script>

<template>
  <div class="h-full min-h-0">
    <ChatPanel />
  </div>
</template>
