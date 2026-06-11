<script setup lang="ts">
const { sessions, currentSessionId, createSession } = useChat()

const targetId = currentSessionId.value || sessions.value[0]?.id || createSession().id
await navigateTo(`/p/${targetId}`, { replace: true })
</script>

<template>
  <div />
</template>
