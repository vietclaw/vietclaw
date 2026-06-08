<script setup lang="ts">
import type { BudgetStatus } from '~/types'
import { formatMoney } from '~/utils/api'

defineProps<{ budget?: BudgetStatus | null }>()
</script>

<template>
  <section class="vc-bezel vc-noise vc-ambient">
    <div class="vc-bezel-inner">
      <div class="relative z-10 flex items-start justify-between border-b border-[var(--border-0)] px-5 py-4">
        <div class="flex items-center gap-3">
          <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-[var(--success)]/10">
            <svg class="h-4 w-4 text-[var(--success)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <h2 class="text-[13px] font-bold text-[var(--fg-0)]">Budget</h2>
            <p class="text-[11px] text-[var(--fg-2)]">Cheap first. Escalate only when useful.</p>
          </div>
        </div>
        <StatusPill :label="budget?.cheap_first ? 'cheap first' : 'manual'" tone="good" size="sm" />
      </div>
      <div class="relative z-10 grid grid-cols-3 divide-x divide-[var(--border-0)]">
        <div class="px-4 py-4 text-center">
          <div class="text-[10px] font-semibold uppercase tracking-wider text-[var(--fg-2)]">today</div>
          <div class="mt-1.5 text-lg font-bold tabular-nums text-[var(--fg-0)]">{{ formatMoney(budget?.total_cost_usd) }}</div>
        </div>
        <div class="px-4 py-4 text-center">
          <div class="text-[10px] font-semibold uppercase tracking-wider text-[var(--fg-2)]">daily cap</div>
          <div class="mt-1.5 text-lg font-bold tabular-nums text-[var(--fg-0)]">{{ formatMoney(budget?.daily_usd_limit) }}</div>
        </div>
        <div class="px-4 py-4 text-center">
          <div class="text-[10px] font-semibold uppercase tracking-wider text-[var(--fg-2)]">approval</div>
          <div class="mt-1.5 text-lg font-bold tabular-nums text-[var(--fg-0)]">{{ formatMoney(budget?.require_approval_above_usd) }}</div>
        </div>
      </div>
    </div>
  </section>
</template>
