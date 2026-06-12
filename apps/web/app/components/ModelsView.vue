<script setup lang="ts">
import { apiFetch } from '~/utils/api'
import type { CatalogModelConfig } from '~/types/config'

const { config } = useSettings()
const { t } = useI18n()
const toast = useToast()

const inputClass = 'vc-input'
const monoClass = 'vc-input vc-input--mono'
const modelLists = ref<Record<string, string[]>>({})
const loadingModels = ref<Record<string, boolean>>({})

function newEntry(): CatalogModelConfig {
  return {
    id: `model-${Date.now()}`,
    provider: config.value?.providers[0]?.id || 'mock',
    model: config.value?.providers[0]?.default_model || 'mock-small',
    label: '',
    enabled: true,
  }
}

function addEntry() {
  if (!config.value) return
  if (!config.value.models) {
    config.value.models = { catalog: [], default_catalog_id: '' }
  }
  const entry = newEntry()
  config.value.models.catalog.push(entry)
  if (!config.value.models.default_catalog_id) {
    config.value.models.default_catalog_id = entry.id
  }
}

function removeEntry(id: string) {
  if (!config.value?.models) return
  config.value.models.catalog = config.value.models.catalog.filter(e => e.id !== id)
  if (config.value.models.default_catalog_id === id) {
    config.value.models.default_catalog_id = config.value.models.catalog[0]?.id || ''
  }
}

async function fetchProviderModels(providerId: string) {
  loadingModels.value[providerId] = true
  try {
    const res = await apiFetch<{ models: string[] }>(`/api/providers/${providerId}/models`)
    modelLists.value[providerId] = res.models || []
  } catch {
    modelLists.value[providerId] = []
    toast.add(t('models.fetchFailed'), 'error')
  } finally {
    loadingModels.value[providerId] = false
  }
}

async function importFromProvider(providerId: string) {
  if (!config.value) return
  if (!modelLists.value[providerId]?.length) {
    await fetchProviderModels(providerId)
  }
  const models = modelLists.value[providerId] || []
  if (!models.length) return
  if (!config.value.models) {
    config.value.models = { catalog: [], default_catalog_id: '' }
  }
  for (const model of models.slice(0, 8)) {
    const id = `${providerId}-${model}`.replace(/[^a-zA-Z0-9_-]/g, '-').slice(0, 48)
    if (config.value.models.catalog.some(e => e.id === id)) continue
    config.value.models.catalog.push({
      id,
      provider: providerId,
      model,
      label: model,
      enabled: true,
    })
  }
  if (!config.value.models.default_catalog_id && config.value.models.catalog[0]) {
    config.value.models.default_catalog_id = config.value.models.catalog[0].id
  }
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="vc-display text-2xl font-medium text-vc-text">{{ t('models.title') }}</h1>
      <p class="mt-1.5 text-sm text-vc-text-muted">{{ t('models.subtitle') }}</p>
    </div>

    <template v-if="config?.models">
      <section class="vc-card p-5 space-y-4">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-sm font-medium text-vc-text">{{ t('models.catalog') }}</h2>
          <button type="button" class="vc-btn vc-btn-outline text-xs" @click="addEntry">{{ t('common.add') }}</button>
        </div>

        <div v-if="config.models.catalog.length === 0" class="text-sm text-vc-text-muted">
          {{ t('models.empty') }}
        </div>

        <div v-for="entry in config.models.catalog" :key="entry.id" class="grid gap-3 border-t border-vc-border-subtle pt-4 sm:grid-cols-2">
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">{{ t('models.field.id') }}</label>
            <input v-model="entry.id" type="text" :class="monoClass" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">{{ t('models.field.label') }}</label>
            <input v-model="entry.label" type="text" :class="inputClass" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">{{ t('models.field.provider') }}</label>
            <select v-model="entry.provider" class="vc-input">
              <option v-for="p in config.providers" :key="p.id" :value="p.id">{{ p.id }}</option>
            </select>
          </div>
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">{{ t('models.field.model') }}</label>
            <div class="flex gap-2">
              <input v-model="entry.model" type="text" :class="monoClass" />
              <button type="button" class="vc-btn vc-btn-outline shrink-0 text-xs" @click="fetchProviderModels(entry.provider)">
                {{ loadingModels[entry.provider] ? '…' : t('common.models') }}
              </button>
            </div>
            <select
              v-if="modelLists[entry.provider]?.length"
              class="vc-input mt-2"
              @change="entry.model = ($event.target as HTMLSelectElement).value"
            >
              <option value="">{{ t('models.pickModel') }}</option>
              <option v-for="m in modelLists[entry.provider]" :key="m" :value="m">{{ m }}</option>
            </select>
          </div>
          <div class="flex items-center gap-3 sm:col-span-2">
            <VcToggle v-model="entry.enabled" :label="t('common.enable')" size="sm" />
            <button type="button" class="text-xs text-vc-error" @click="removeEntry(entry.id)">{{ t('models.remove') }}</button>
          </div>
        </div>
      </section>

      <section class="vc-card p-5 space-y-3">
        <h2 class="text-sm font-medium text-vc-text">{{ t('models.default') }}</h2>
        <select v-model="config.models.default_catalog_id" class="vc-input max-w-md">
          <option v-for="entry in config.models.catalog.filter(e => e.enabled)" :key="entry.id" :value="entry.id">
            {{ entry.label || entry.id }}
          </option>
        </select>
      </section>

      <section class="vc-card p-5 space-y-3">
        <h2 class="text-sm font-medium text-vc-text">{{ t('models.import') }}</h2>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="p in config.providers.filter(x => x.enabled)"
            :key="p.id"
            type="button"
            class="vc-btn vc-btn-outline text-xs"
            @click="importFromProvider(p.id)"
          >
            {{ t('models.importFrom', p.id) }}
          </button>
        </div>
      </section>
    </template>
  </div>
</template>
