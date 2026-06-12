<script setup lang="ts">
const { t } = useI18n()
useHead({ title: () => `${t('nav.providers')} · VietClaw` })
</script>

<template>
  <ProvidersView embedded />
</template>
