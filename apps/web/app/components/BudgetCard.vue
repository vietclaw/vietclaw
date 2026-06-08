<script setup lang="ts">
import type { BudgetStatus } from '~/types'
import { formatMoney } from '~/utils/api'

defineProps<{ budget?: BudgetStatus | null }>()
</script>

<template>
  <section class="rounded-xl border border-white/[0.08] bg-[var(--vc-panel)] p-5">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h2 class="text-sm font-medium text-white">Budget</h2>
        <p class="mt-1 text-sm text-[var(--vc-muted)]">Cheap first. Escalate only when useful.</p>
      </div>
      <StatusPill :label="budget?.cheap_first ? 'cheap first' : 'manual'" tone="good" />
    </div>
    <div class="mt-6 grid grid-cols-3 gap-3">
      <div class="rounded-lg bg-white/[0.04] p-3">
        <div class="text-xs text-[var(--vc-subtle)]">today</div>
        <div class="mt-1 text-lg font-semibold">{{ formatMoney(budget?.total_cost_usd) }}</div>
      </div>
      <div class="rounded-lg bg-white/[0.04] p-3">
        <div class="text-xs text-[var(--vc-subtle)]">daily cap</div>
        <div class="mt-1 text-lg font-semibold">{{ formatMoney(budget?.daily_usd_limit) }}</div>
      </div>
      <div class="rounded-lg bg-white/[0.04] p-3">
        <div class="text-xs text-[var(--vc-subtle)]">approval</div>
        <div class="mt-1 text-lg font-semibold">{{ formatMoney(budget?.require_approval_above_usd) }}</div>
      </div>
    </div>
  </section>
</template>

