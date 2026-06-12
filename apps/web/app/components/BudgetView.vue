<script setup lang="ts">
import { apiFetch, formatMoney } from '~/utils/api'

defineProps<{ embedded?: boolean }>()

const { config } = useSettings()
const { t } = useI18n()
const todayCost = ref<number | null>(null)

onMounted(async () => {
  try {
    const res = await apiFetch<{ total_cost_usd: number }>('/api/budget')
    todayCost.value = res.total_cost_usd
  } catch {
    todayCost.value = null
  }
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="vc-display text-2xl font-medium text-vc-text">{{ t('budget.title') }}</h1>
      <p v-if="todayCost !== null" class="mt-1.5 text-sm text-vc-text-muted">
        {{ t('budget.todayCost') }}
        <span class="font-mono font-medium text-vc-text" style="font-variant-numeric: tabular-nums">{{ formatMoney(todayCost) }}</span>
      </p>
    </div>

    <div v-if="config" class="vc-card space-y-5 p-5">
      <div class="flex flex-wrap gap-4">
        <VcToggle v-model="config.router.cheap_first" :label="t('budget.cheapFirst')" size="sm" />
        <VcToggle v-model="config.router.allow_escalation" :label="t('budget.allowEscalation')" size="sm" />
      </div>
      <div class="grid gap-4 border-t border-vc-border-subtle pt-4 sm:grid-cols-2">
        <div>
          <label class="mb-1.5 block text-xs font-medium text-vc-text-secondary">{{ t('budget.dailyCap') }}</label>
          <input
            v-model.number="config.budget.daily_usd_limit"
            type="number"
            min="0"
            step="0.01"
            class="vc-input vc-input--mono"
          />
        </div>
        <div>
          <label class="mb-1.5 block text-xs font-medium text-vc-text-secondary">{{ t('budget.approvalAbove') }}</label>
          <input
            v-model.number="config.budget.require_approval_above_usd"
            type="number"
            min="0"
            step="0.01"
            class="vc-input vc-input--mono"
          />
        </div>
      </div>
    </div>
  </div>
</template>
