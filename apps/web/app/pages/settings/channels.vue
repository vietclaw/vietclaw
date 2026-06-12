<script setup lang="ts">
const { t } = useI18n()
useHead({ title: () => `${t('nav.channels')} · VietClaw` })
</script>

<template>
  <ChannelsView embedded />
</template>
