<script setup lang="ts">
const { t } = useI18n()
useHead({ title: () => `${t('nav.budget')} · VietClaw` })
</script>

<template>
  <BudgetView embedded />
</template>
