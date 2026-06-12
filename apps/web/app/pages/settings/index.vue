<script setup lang="ts">
const { t } = useI18n()
useHead({ title: () => `${t('settings.overview.title')} · VietClaw` })
</script>

<template>
  <SettingsView />
</template>
