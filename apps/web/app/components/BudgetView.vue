<script setup lang="ts">
import { apiFetch, formatMoney } from '~/utils/api'

defineProps<{ embedded?: boolean }>()

const { config } = useSettings()
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
      <h1 class="text-lg font-semibold tracking-tight text-vc-text">Budget</h1>
      <p v-if="todayCost !== null" class="mt-1 text-sm text-vc-text-muted">
        Chi phí hôm nay: {{ formatMoney(todayCost) }}
      </p>
    </div>

    <div v-if="config" class="rounded-lg border border-vc-border bg-vc-surface p-5 space-y-5">
      <div class="flex flex-wrap gap-4">
        <VcToggle v-model="config.router.cheap_first" label="Cheap first" size="sm" />
        <VcToggle v-model="config.router.allow_escalation" label="Allow escalation" size="sm" />
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <label class="mb-1.5 block text-xs text-vc-text-muted">Daily cap (USD)</label>
          <input
            v-model.number="config.budget.daily_usd_limit"
            type="number"
            min="0"
            step="0.01"
            class="w-full rounded-md border border-vc-border bg-vc-bg px-3 py-2 text-sm font-mono text-vc-text"
          />
        </div>
        <div>
          <label class="mb-1.5 block text-xs text-vc-text-muted">Approval above (USD)</label>
          <input
            v-model.number="config.budget.require_approval_above_usd"
            type="number"
            min="0"
            step="0.01"
            class="w-full rounded-md border border-vc-border bg-vc-bg px-3 py-2 text-sm font-mono text-vc-text"
          />
        </div>
      </div>
    </div>
  </div>
</template>
