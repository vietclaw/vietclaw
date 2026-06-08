<script setup lang="ts">
import type { BudgetStatus } from '~/types'
import { apiFetch, formatMoney } from '~/utils/api'
import { DollarSign } from '@lucide/vue'

const toast = useToast()
const budget = ref<BudgetStatus | null>(null)
const loading = ref(true)

async function fetchBudget() {
  loading.value = true
  try {
    budget.value = await apiFetch<BudgetStatus>('/api/budget')
  } catch (err) {
    toast.add(err instanceof Error ? err.message : 'Failed to load budget', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(fetchBudget)
</script>

<template>
  <div class="max-w-3xl mx-auto space-y-4">
    <div class="flex items-center gap-2">
      <DollarSign :size="16" class="text-zinc-400" />
      <h2 class="text-sm font-semibold text-zinc-200">Budget</h2>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="h-32 rounded-lg bg-zinc-900/40 animate-pulse" />

    <!-- Content -->
    <div v-else-if="budget" class="rounded-lg border border-zinc-900 bg-zinc-950/40 overflow-hidden">
      <div class="px-5 py-4 border-b border-zinc-900/60 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="text-xs text-zinc-400">Policy</span>
          <span class="px-1.5 py-0.5 rounded text-[9px] font-mono bg-emerald-950/30 text-emerald-400 border border-emerald-900/30">
            {{ budget.cheap_first ? 'cheap first' : 'manual' }}
          </span>
        </div>
        <span class="text-[10px] text-zinc-600">
          escalation {{ budget.allow_escalation ? 'enabled' : 'disabled' }}
        </span>
      </div>
      <div class="grid grid-cols-3 divide-x divide-zinc-900/60">
        <div class="px-4 py-5 text-center">
          <div class="text-[10px] font-medium text-zinc-500 uppercase tracking-wider mb-1.5">today</div>
          <div class="text-lg font-bold tabular-nums text-zinc-100 font-mono">{{ formatMoney(budget.total_cost_usd) }}</div>
        </div>
        <div class="px-4 py-5 text-center">
          <div class="text-[10px] font-medium text-zinc-500 uppercase tracking-wider mb-1.5">daily cap</div>
          <div class="text-lg font-bold tabular-nums text-zinc-100 font-mono">{{ formatMoney(budget.daily_usd_limit) }}</div>
        </div>
        <div class="px-4 py-5 text-center">
          <div class="text-[10px] font-medium text-zinc-500 uppercase tracking-wider mb-1.5">approval</div>
          <div class="text-lg font-bold tabular-nums text-zinc-100 font-mono">{{ formatMoney(budget.require_approval_above_usd) }}</div>
        </div>
      </div>
    </div>

    <!-- Empty -->
    <div v-else class="text-center py-16">
      <DollarSign :size="32" class="mx-auto text-zinc-700 mb-3" />
      <p class="text-xs text-zinc-500">No budget data available.</p>
    </div>
  </div>
</template>
