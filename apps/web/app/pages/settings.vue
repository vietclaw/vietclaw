<script setup lang="ts">
import { apiFetch } from '~/utils/api'

type RuntimeConfig = {
  server: { host: string; port: number }
  runtime: { mode: string; max_concurrent_tasks: number }
  agent: {
    name: string
    language: string
    default_mode: string
    max_steps: number
    max_output_tokens: number
    skill_dirs: string[]
  }
  router: {
    default_provider: string
    default_model: string
    intent_mode: string
    cheap_first: boolean
    allow_escalation: boolean
  }
  budget: {
    daily_usd_limit: number
    require_approval_above_usd: number
  }
}

const { data, refresh } = await useAsyncData('settings', () => apiFetch<RuntimeConfig>('/api/settings'))
const draft = ref('')
const message = ref('')
const saving = ref(false)

watchEffect(() => {
  if (data.value && !draft.value) {
    draft.value = JSON.stringify(data.value, null, 2)
  }
})

async function saveSettings() {
  saving.value = true
  message.value = ''
  try {
    const parsed = JSON.parse(draft.value)
    await apiFetch('/api/settings', {
      method: 'PUT',
      body: JSON.stringify(parsed)
    })
    await refresh()
    draft.value = JSON.stringify(data.value, null, 2)
    message.value = 'saved and reloaded'
  } catch (error) {
    message.value = error instanceof Error ? error.message : 'save failed'
  } finally {
    saving.value = false
  }
}

async function reloadSettings() {
  saving.value = true
  message.value = ''
  try {
    await apiFetch('/api/settings/reload', { method: 'POST' })
    await refresh()
    draft.value = JSON.stringify(data.value, null, 2)
    message.value = 'reloaded from disk'
  } catch (error) {
    message.value = error instanceof Error ? error.message : 'reload failed'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="space-y-5">
    <section class="vc-surface p-5">
      <div class="relative z-10 flex flex-wrap items-center justify-between gap-3">
        <div>
          <p class="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">runtime</p>
          <h1 class="mt-2 text-2xl font-bold tracking-tight text-foreground">Settings</h1>
        </div>
        <div class="flex gap-2">
          <button class="rounded-md bg-background px-3 py-2 text-sm text-foreground vc-focus" :disabled="saving" @click="reloadSettings">
            Reload
          </button>
          <button class="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground vc-focus" :disabled="saving" @click="saveSettings">
            Save
          </button>
        </div>
      </div>
    </section>

    <section class="vc-surface p-5">
      <div class="relative z-10 space-y-3">
        <textarea
          v-model="draft"
          class="min-h-[520px] w-full rounded-lg border border-border bg-background p-4 font-mono text-xs leading-5 text-foreground outline-none focus:border-primary"
          spellcheck="false"
        />
        <p v-if="message" class="text-sm text-muted-foreground">{{ message }}</p>
      </div>
    </section>
  </div>
</template>
