<script setup lang="ts">
const { t } = useI18n()
useHead({ title: () => `${t('nav.memory')} · VietClaw` })
</script>

<template>
  <MemoryView embedded />
</template>
