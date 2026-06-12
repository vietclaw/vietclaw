<script setup lang="ts">
import { apiFetch } from '~/utils/api'
import type { ChannelStatus } from '~/types'
import type { ChannelEnvTest } from '~/types/config'
import { AlertCircle } from '@lucide/vue'

defineProps<{ embedded?: boolean }>()

const toast = useToast()
const { config } = useSettings()
const { t, channelStatus } = useI18n()
const runtime = ref<ChannelStatus[]>([])
const envTests = ref<Record<string, ChannelEnvTest>>({})

const fieldClass = 'vc-input vc-input--mono'

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
      toast.add(t('channels.envNotSet', res.token_env), 'error')
    } else {
      toast.add(t('status.envOk'), 'success')
    }
  } catch (err) {
    toast.add(err instanceof Error ? err.message : t('channels.testFailed'), 'error')
  }
}

onMounted(() => {
  void fetchRuntime()
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="vc-display text-2xl font-medium text-vc-text">{{ t('channels.title') }}</h1>
      <p class="mt-1.5 text-sm text-vc-text-muted">{{ t('channels.subtitle') }}</p>
    </div>

    <div v-if="config" class="space-y-3">
      <section class="vc-card p-5">
        <div class="flex items-center justify-between gap-3">
          <div>
            <h2 class="text-sm font-medium text-vc-text">{{ t('channels.discord') }}</h2>
            <p v-if="runtimeFor('discord')" class="text-xs text-vc-text-muted">
              {{ channelStatus(runtimeFor('discord')) }}
            </p>
          </div>
          <VcToggle v-model="config.channels.discord.enabled" :label="t('common.enable')" />
        </div>
        <div class="mt-3 space-y-3 border-t border-vc-border-subtle pt-3">
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">{{ t('channels.tokenEnv') }}</label>
            <input
              v-model="config.channels.discord.token_env"
              type="text"
              :class="fieldClass"
              placeholder="VIETCLAW_DISCORD_TOKEN"
            />
          </div>
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">{{ t('channels.allowedGuilds') }}</label>
            <input
              :value="listToText(config.channels.discord.allowed_guilds)"
              type="text"
              :class="fieldClass"
              @input="config.channels.discord.allowed_guilds = textToList(($event.target as HTMLInputElement).value)"
            />
          </div>
          <VcToggle v-model="config.channels.discord.respond_in_dm" :label="t('channels.respondInDm')" size="sm" />
          <button type="button" class="vc-btn vc-btn-outline text-xs" @click="testToken('discord')">
            {{ t('channels.testToken') }}
          </button>
          <p v-if="envTests.discord" class="text-xs" :class="envTests.discord.env_found ? 'text-vc-success' : 'text-vc-error'">
            {{ envTests.discord.env_found ? t('status.envOk') : t('status.envMissing') }}
          </p>
          <div v-if="runtimeFor('discord')?.error" class="flex gap-2 text-xs text-vc-error">
            <AlertCircle :size="14" class="shrink-0" />
            {{ runtimeFor('discord')?.error }}
          </div>
        </div>
      </section>

      <section class="vc-card p-5">
        <div class="flex items-center justify-between gap-3">
          <div>
            <h2 class="text-sm font-medium text-vc-text">{{ t('channels.telegram') }}</h2>
            <p v-if="runtimeFor('telegram')" class="text-xs text-vc-text-muted">
              {{ channelStatus(runtimeFor('telegram')) }}
            </p>
          </div>
          <VcToggle v-model="config.channels.telegram.enabled" :label="t('common.enable')" />
        </div>
        <div class="mt-3 space-y-3 border-t border-vc-border-subtle pt-3">
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">{{ t('channels.tokenEnv') }}</label>
            <input
              v-model="config.channels.telegram.token_env"
              type="text"
              :class="fieldClass"
              placeholder="VIETCLAW_TELEGRAM_TOKEN"
            />
          </div>
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">{{ t('channels.allowedChats') }}</label>
            <input
              :value="listToText(config.channels.telegram.allowed_chats)"
              type="text"
              :class="fieldClass"
              @input="config.channels.telegram.allowed_chats = textToList(($event.target as HTMLInputElement).value)"
            />
          </div>
          <VcToggle v-model="config.channels.telegram.respond_in_private" :label="t('channels.respondInPrivate')" size="sm" />
          <div class="grid gap-3 sm:grid-cols-2">
            <div>
              <label class="mb-1 block text-xs text-vc-text-muted">{{ t('channels.commandMode') }}</label>
              <VcSelect
                :model-value="config.channels.telegram.command_mode || 'slash'"
                group="telegram_command_mode"
                @update:model-value="config.channels.telegram.command_mode = $event"
              />
            </div>
            <div>
              <label class="mb-1 block text-xs text-vc-text-muted">{{ t('channels.commandPrefix') }}</label>
              <input v-model="config.channels.telegram.command_prefix" type="text" :class="fieldClass" placeholder="/" />
            </div>
          </div>
          <p class="text-xs text-vc-text-muted">{{ t('channels.modelsHint') }}</p>
          <button type="button" class="vc-btn vc-btn-outline text-xs" @click="testToken('telegram')">
            {{ t('channels.testToken') }}
          </button>
          <p v-if="envTests.telegram" class="text-xs" :class="envTests.telegram.env_found ? 'text-vc-success' : 'text-vc-error'">
            {{ envTests.telegram.env_found ? t('status.envOk') : t('status.envMissing') }}
          </p>
        </div>
      </section>

      <section class="vc-card p-5">
        <div class="flex items-center justify-between gap-3">
          <h2 class="text-sm font-medium text-vc-text">{{ t('channels.attachments') }}</h2>
          <VcToggle v-model="config.channels.attachments.enabled" :label="t('common.enable')" />
        </div>
        <div class="mt-3 grid grid-cols-2 gap-3 border-t border-vc-border-subtle pt-3">
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">{{ t('channels.maxFiles') }}</label>
            <input v-model.number="config.channels.attachments.max_files" type="number" min="0" class="vc-input vc-input--mono" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-vc-text-muted">{{ t('channels.maxBytes') }}</label>
            <input v-model.number="config.channels.attachments.max_bytes" type="number" min="0" class="vc-input vc-input--mono" />
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
