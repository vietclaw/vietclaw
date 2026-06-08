<script setup lang="ts">
import type { Session } from '~/types'
import { apiFetch } from '~/utils/api'

definePageMeta({ title: 'Sessions' })

const { data: sessions } = await useAsyncData('sessions', () => apiFetch<Session[]>('/api/sessions'), { default: () => [] })
</script>

<template>
  <div class="space-y-4 vc-stagger">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-[14px] font-bold text-[var(--fg-0)]">All sessions</h2>
        <p class="text-[11px] text-[var(--fg-2)]">{{ sessions.length }} session{{ sessions.length !== 1 ? 's' : '' }}</p>
      </div>
    </div>
    <SessionList :sessions="sessions" />
  </div>
</template>
