<script setup lang="ts">
import type { BudgetStatus } from '~/types'
import { apiFetch } from '~/utils/api'

definePageMeta({ title: 'Budget' })

const { data: budget, refresh } = await useAsyncData('budget', () => apiFetch<BudgetStatus>('/api/budget'), { default: () => null })
</script>

<template>
  <div class="space-y-5">
    <BudgetCard :budget="budget" />
    <section class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)] p-5">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h2 class="text-sm font-medium text-white">Policy</h2>
          <p class="mt-2 max-w-2xl text-sm leading-6 text-[var(--vc-muted)]">
            VietClaw dùng model rẻ trước, nâng model khi cần. Request vượt ngưỡng approval sẽ bị giữ lại.
          </p>
        </div>
        <button class="rounded-lg border border-white/[0.1] px-3 py-2 text-sm text-white" @click="() => refresh()">Refresh</button>
      </div>
    </section>
  </div>
</template>
