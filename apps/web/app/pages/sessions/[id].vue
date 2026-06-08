<script setup lang="ts">
import type { SessionDetail } from '~/types'
import { apiFetch } from '~/utils/api'

definePageMeta({ title: 'Session' })

const route = useRoute()
const id = computed(() => String(route.params.id))
const { data: detail } = await useAsyncData(`session-${id.value}`, () => apiFetch<SessionDetail>(`/api/sessions/${encodeURIComponent(id.value)}`), { default: () => null })
</script>

<template>
  <div class="space-y-4">
    <NuxtLink to="/sessions" class="text-sm text-[var(--vc-muted)] hover:text-white vc-focus">Back to sessions</NuxtLink>
    <section class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)]">
      <div v-for="message in detail?.messages || []" :key="message.id" class="border-b border-white/[0.06] px-5 py-4 last:border-b-0">
        <div class="text-xs uppercase tracking-[0.16em] text-[var(--vc-subtle)]">{{ message.role }}</div>
        <p class="mt-2 whitespace-pre-wrap text-sm leading-6 text-white">{{ message.content }}</p>
      </div>
      <div v-if="!detail || detail.messages.length === 0" class="p-8 text-sm text-[var(--vc-muted)]">No messages.</div>
    </section>
  </div>
</template>

