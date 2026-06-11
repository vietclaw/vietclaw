<script setup lang="ts">
const route = useRoute()
await navigateTo(`/p/${route.params.id}`, { replace: true })
</script>

<template>
  <div />
</template>
