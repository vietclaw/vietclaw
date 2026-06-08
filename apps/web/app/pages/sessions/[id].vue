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
    <NuxtLink to="/sessions" class="inline-flex items-center gap-1.5 text-[13px] font-medium text-[var(--fg-2)] vc-transition-fast hover:text-[var(--fg-0)] vc-focus">
      <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" /></svg>
      Sessions
    </NuxtLink>

    <section class="vc-bezel vc-noise overflow-hidden vc-stagger">
      <div class="vc-bezel-inner">
        <div v-for="message in detail?.messages || []" :key="message.id" class="border-b border-[var(--border-0)] px-5 py-5 last:border-b-0">
          <div class="flex items-center gap-3">
            <div
              class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg text-[10px] font-bold"
              :class="message.role === 'user' ? 'bg-gradient-to-br from-[var(--accent)] to-[var(--accent-dim)] text-white' : 'bg-[var(--bg-3)] text-[var(--fg-2)]'"
            >
              {{ message.role === 'user' ? 'U' : 'V' }}
            </div>
            <span class="text-[11px] font-semibold uppercase tracking-wider text-[var(--fg-2)]">{{ message.role }}</span>
          </div>
          <p class="mt-3 whitespace-pre-wrap pl-[38px] text-[13px] leading-relaxed text-[var(--fg-0)]">{{ message.content }}</p>
        </div>

        <!-- Empty state -->
        <div v-if="!detail || detail.messages.length === 0" class="px-5 py-14 text-center">
          <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-[var(--bg-3)]">
            <svg class="h-5 w-5 text-[var(--fg-2)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <p class="mt-3 text-[13px] font-medium text-[var(--fg-2)]">No messages.</p>
        </div>
      </div>
    </section>
  </div>
</template>
