<script setup lang="ts">
import { apiFetch } from '~/utils/api'

defineProps<{ embedded?: boolean }>()

const toast = useToast()
const { config } = useSettings()
const { t } = useI18n()
const modelLists = ref<Record<string, string[]>>({})
const loadingModels = ref<Record<string, boolean>>({})

async function fetchModels(providerId: string) {
  if (modelLists.value[providerId] || loadingModels.value[providerId]) return
  loadingModels.value[providerId] = true
  try {
    const res = await apiFetch<{ models: string[] }>(`/api/providers/${providerId}/models`)
    modelLists.value[providerId] = res.models || []
  } catch {
    toast.add(t('providers.fetchModelsFailed'), 'error')
  } finally {
    loadingModels.value[providerId] = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="vc-display text-2xl font-medium text-vc-text">{{ t('providers.title') }}</h1>
      <p v-if="config" class="mt-1.5 text-sm text-vc-text-muted">{{ t('providers.count', config.providers.length) }}</p>
    </div>

    <div v-if="config" class="space-y-3">
      <div
        v-for="p in config.providers"
        :key="p.id"
        class="vc-card p-5 transition-shadow duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] hover:shadow-[var(--vc-shadow-md)]"
      >
        <div class="flex items-center justify-between gap-3">
          <div class="flex items-center gap-2">
            <span class="vc-status-dot" :class="p.enabled ? 'vc-status-dot--on' : 'vc-status-dot--off'" aria-hidden="true" />
            <span class="text-sm font-semibold text-vc-text">{{ p.id }}</span>
            <span class="rounded-md bg-vc-bg-subtle px-1.5 py-0.5 font-mono text-[11px] text-vc-text-muted">{{ p.type }}</span>
          </div>
          <VcToggle v-model="p.enabled" :label="t('common.enable')" />
        </div>
        <div class="mt-4 grid gap-3 border-t border-vc-border-subtle pt-4 sm:grid-cols-2">
          <div>
            <span class="mb-1.5 block text-xs font-medium text-vc-text-secondary">{{ t('providers.model') }}</span>
            <div class="flex gap-2">
              <select
                v-if="modelLists[p.id]?.length"
                v-model="p.default_model"
                class="vc-input vc-input--mono flex-1"
              >
                <option v-for="m in modelLists[p.id]" :key="m" :value="m">{{ m }}</option>
              </select>
              <input
                v-else
                v-model="p.default_model"
                type="text"
                class="vc-input vc-input--mono flex-1"
              />
              <button
                type="button"
                class="vc-btn vc-btn-ghost shrink-0 text-xs"
                :disabled="loadingModels[p.id]"
                @click="fetchModels(p.id)"
              >
                {{ loadingModels[p.id] ? '…' : t('common.models') }}
              </button>
            </div>
          </div>
          <div>
            <span class="mb-1.5 block text-xs font-medium text-vc-text-secondary">{{ t('providers.apiKeyEnv') }}</span>
            <input
              v-model="p.api_key_env"
              type="text"
              class="vc-input vc-input--mono"
            />
          </div>
          <div class="sm:col-span-2">
            <span class="mb-1.5 block text-xs font-medium text-vc-text-secondary">{{ t('providers.baseUrl') }}</span>
            <input
              v-model="p.base_url"
              type="text"
              class="vc-input vc-input--mono"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
