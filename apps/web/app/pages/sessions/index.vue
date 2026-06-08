<script setup lang="ts">
import type { Session } from '~/types'
import { apiFetch } from '~/utils/api'

definePageMeta({ title: 'Sessions' })

const { data: sessions } = await useAsyncData('sessions', () => apiFetch<Session[]>('/api/sessions'), { default: () => [] })
</script>

<template>
  <SessionList :sessions="sessions" />
</template>

