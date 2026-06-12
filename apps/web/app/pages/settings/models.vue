<script setup lang="ts">
const { t } = useI18n()
useHead({ title: () => `${t('nav.models')} · VietClaw` })
</script>

<template>
  <ModelsView />
</template>
