<script setup lang="ts">
import { apiFetch } from '~/utils/api'
import type { ChannelStatus } from '~/types'
import type { ChannelEnvTest } from '~/types/config'
import { AlertCircle } from '@lucide/vue'

defineProps<{ embedded?: boolean }>()

const toast = useToast()
const { config } = useSettings()
const runtime = ref<ChannelStatus[]>([])
const envTests = ref<Record<string, ChannelEnvTest>>({})

const fieldClass = 'w-full rounded-md border border-vc-border bg-vc-bg px-2 py-1.5 text-xs font-mono text-vc-text focus:border-vc-accent focus:outline-none'

function runtimeFor(name: string): ChannelStatus | undefined {
  return runtime.value.find(c => c.name === name)
}

function listToText(items: string[] | undefined): string {
  return (items || []).join(', ')
}

function textToList(text: string): string[] {
  return text.split(',').map(s => s.trim()).filter(Boolean)
}

async function fetchRuntime() {
  try {
    runtime.value = await apiFetch<ChannelStatus[]>('/api/channels')
  } catch {
    runtime.value = []
  }
}

async function testToken(channel: 'discord' | 'telegram') {
  try {
    const res = await apiFetch<ChannelEnvTest>(`/api/channels/${channel}/test`, { method: 'POST' })
    envTests.value[channel] = res
    if (!res.env_found) {
      toast.add(`Biến môi trường ${res.token_env} chưa được set`, 'error')
    } else {
      toast.add('Token env OK', 'success')
    }
  } catch (err) {
    toast.add(err instanceof Error ? err.message : 'Test thất bại', 'error')
  }
}

onMounted(() => {
  void fetchRuntime()
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-vc-text">Kênh</h1>
      <p class="mt-1 text-sm text-vc-text-muted">Discord, Telegram và file đính kèm</p>
    </div>

    <div v-if="config" class="space-y-3">
      <section class="rounded-lg border border-vc-border bg-vc-surface p-4">
        <div class="flex items-center justify-between gap-3">
          <div>
            <h2 class="text-sm font-medium text-vc-text">Discord</h2>
            <p v-if="runtimeFor('discord')" class="text-xs text-vc-text-muted">
              {{ runtimeFor('discord')?.running ? 'running' : runtimeFor('discord')?.enabled ? 'enabled' : 'off' }}
            </p>
          </div>
          <VcToggle v-model="config.channels.discord.enabled" label="Bật" />
        </div>
        <div class="mt-3 space-y-3 border-t border-vc-border-subtle pt-3">
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">Token env</label>
            <input
              v-model="config.channels.discord.token_env"
              type="text"
              :class="fieldClass"
              placeholder="VIETCLAW_DISCORD_TOKEN"
            />
          </div>
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">Allowed guilds (id, phẩy)</label>
            <input
              :value="listToText(config.channels.discord.allowed_guilds)"
              type="text"
              :class="fieldClass"
              @input="config.channels.discord.allowed_guilds = textToList(($event.target as HTMLInputElement).value)"
            />
          </div>
          <VcToggle v-model="config.channels.discord.respond_in_dm" label="Respond in DM" size="sm" />
          <button type="button" class="vc-btn vc-btn-ghost text-xs" @click="testToken('discord')">
            Kiểm tra token
          </button>
          <p v-if="envTests.discord" class="text-xs" :class="envTests.discord.env_found ? 'text-vc-success' : 'text-vc-error'">
            {{ envTests.discord.env_found ? 'env OK' : 'env missing' }}
          </p>
          <div v-if="runtimeFor('discord')?.error" class="flex gap-2 text-xs text-vc-error">
            <AlertCircle :size="14" class="shrink-0" />
            {{ runtimeFor('discord')?.error }}
          </div>
        </div>
      </section>

      <section class="rounded-lg border border-vc-border bg-vc-surface p-4">
        <div class="flex items-center justify-between gap-3">
          <div>
            <h2 class="text-sm font-medium text-vc-text">Telegram</h2>
            <p v-if="runtimeFor('telegram')" class="text-xs text-vc-text-muted">
              {{ runtimeFor('telegram')?.running ? 'running' : runtimeFor('telegram')?.enabled ? 'enabled' : 'off' }}
            </p>
          </div>
          <VcToggle v-model="config.channels.telegram.enabled" label="Bật" />
        </div>
        <div class="mt-3 space-y-3 border-t border-vc-border-subtle pt-3">
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">Token env</label>
            <input
              v-model="config.channels.telegram.token_env"
              type="text"
              :class="fieldClass"
              placeholder="VIETCLAW_TELEGRAM_TOKEN"
            />
          </div>
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">Allowed chats (id, phẩy)</label>
            <input
              :value="listToText(config.channels.telegram.allowed_chats)"
              type="text"
              :class="fieldClass"
              @input="config.channels.telegram.allowed_chats = textToList(($event.target as HTMLInputElement).value)"
            />
          </div>
          <VcToggle v-model="config.channels.telegram.respond_in_private" label="Respond in private" size="sm" />
          <button type="button" class="vc-btn vc-btn-ghost text-xs" @click="testToken('telegram')">
            Kiểm tra token
          </button>
          <p v-if="envTests.telegram" class="text-xs" :class="envTests.telegram.env_found ? 'text-vc-success' : 'text-vc-error'">
            {{ envTests.telegram.env_found ? 'env OK' : 'env missing' }}
          </p>
        </div>
      </section>

      <section class="rounded-lg border border-vc-border bg-vc-surface p-4">
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-sm font-medium text-vc-text">File đính kèm</h2>
          <VcToggle v-model="config.channels.attachments.enabled" label="Bật" />
        </div>
        <div class="mt-3 grid grid-cols-2 gap-3 border-t border-vc-border-subtle pt-3">
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">max files</label>
            <input v-model.number="config.channels.attachments.max_files" type="number" min="0" class="w-full rounded-md border border-vc-border bg-vc-bg px-2 py-1.5 text-xs font-mono" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">max bytes</label>
            <input v-model.number="config.channels.attachments.max_bytes" type="number" min="0" class="w-full rounded-md border border-vc-border bg-vc-bg px-2 py-1.5 text-xs font-mono" />
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
