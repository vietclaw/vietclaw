<script setup lang="ts">
const { t } = useI18n()
useHead({ title: () => `${t('nav.logs')} · VietClaw` })
</script>

<template>
  <LogsView embedded />
</template>
