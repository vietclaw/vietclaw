<script setup lang="ts">
import { apiFetch } from '~/utils/api'

const props = defineProps<{ embedded?: boolean }>()

const toast = useToast()
const { config } = useSettings()
const modelLists = ref<Record<string, string[]>>({})
const loadingModels = ref<Record<string, boolean>>({})

async function fetchModels(providerId: string) {
  if (modelLists.value[providerId] || loadingModels.value[providerId]) return
  loadingModels.value[providerId] = true
  try {
    const res = await apiFetch<{ models: string[] }>(`/api/providers/${providerId}/models`)
    modelLists.value[providerId] = res.models || []
  } catch {
    toast.add('Không lấy được danh sách model', 'error')
  } finally {
    loadingModels.value[providerId] = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-vc-text">Providers</h1>
      <p v-if="config" class="mt-1 text-sm text-vc-text-muted">{{ config.providers.length }} provider trong config</p>
    </div>

    <div v-if="config" class="space-y-3">
      <div
        v-for="p in config.providers"
        :key="p.id"
        class="rounded-lg border border-vc-border bg-vc-surface p-4"
      >
        <div class="flex items-center justify-between gap-3">
          <div>
            <span class="text-sm font-medium text-vc-text">{{ p.id }}</span>
            <span class="ml-2 text-xs font-mono text-vc-text-muted">{{ p.type }}</span>
          </div>
          <VcToggle v-model="p.enabled" label="Bật" />
        </div>
        <div class="mt-4 grid gap-3 sm:grid-cols-2">
          <div>
            <span class="mb-1 block text-xs text-vc-text-muted">Model</span>
            <div class="flex gap-2">
              <select
                v-if="modelLists[p.id]?.length"
                v-model="p.default_model"
                class="flex-1 rounded-md border border-vc-border bg-vc-bg px-2 py-1.5 text-xs font-mono text-vc-text"
              >
                <option v-for="m in modelLists[p.id]" :key="m" :value="m">{{ m }}</option>
              </select>
              <input
                v-else
                v-model="p.default_model"
                type="text"
                class="flex-1 rounded-md border border-vc-border bg-vc-bg px-2 py-1.5 text-xs font-mono text-vc-text"
              />
              <button
                type="button"
                class="vc-btn vc-btn-ghost text-xs"
                :disabled="loadingModels[p.id]"
                @click="fetchModels(p.id)"
              >
                {{ loadingModels[p.id] ? '…' : 'models' }}
              </button>
            </div>
          </div>
          <div>
            <span class="mb-1 block text-xs text-vc-text-muted">API key env</span>
            <input
              v-model="p.api_key_env"
              type="text"
              class="w-full rounded-md border border-vc-border bg-vc-bg px-2 py-1.5 text-xs font-mono text-vc-text"
            />
          </div>
          <div class="sm:col-span-2">
            <span class="mb-1 block text-xs text-vc-text-muted">Base URL</span>
            <input
              v-model="p.base_url"
              type="text"
              class="w-full rounded-md border border-vc-border bg-vc-bg px-2 py-1.5 text-xs font-mono text-vc-text"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
