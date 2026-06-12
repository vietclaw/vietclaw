<script setup lang="ts">
import SettingsField from '~/components/settings/SettingsField.vue'
import SettingsSection from '~/components/settings/SettingsSection.vue'

const { config } = useSettings()
const { t } = useI18n()

const inputClass = 'vc-input'
const monoClass = 'vc-input vc-input--mono'
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="vc-display text-2xl font-medium text-vc-text">{{ t('settings.overview.title') }}</h1>
      <p class="mt-1.5 max-w-lg text-sm leading-relaxed text-vc-text-muted">
        {{ t('settings.overview.desc') }}
      </p>
    </div>

    <template v-if="config">
      <SettingsSection :title="t('settings.section.agent')" :description="t('settings.section.agent.desc')">
        <div class="grid gap-4 sm:grid-cols-2">
          <SettingsField :label="t('settings.field.name')">
            <input v-model="config.agent.name" type="text" :class="inputClass" />
          </SettingsField>
          <SettingsField :label="t('settings.field.language')">
            <VcSelect v-model="config.agent.language" group="language" />
          </SettingsField>
          <SettingsField :label="t('settings.field.experience')">
            <VcSelect v-model="config.agent.experience" group="experience" />
          </SettingsField>
          <SettingsField :label="t('settings.field.style')">
            <VcSelect v-model="config.agent.style" group="style" />
          </SettingsField>
          <SettingsField :label="t('settings.field.maxSteps')" :hint="t('settings.hint.unlimited')">
            <input v-model.number="config.agent.max_steps" type="number" min="0" :class="monoClass" />
          </SettingsField>
          <SettingsField :label="t('settings.field.maxOutputTokens')" :hint="t('settings.hint.unlimited')">
            <input v-model.number="config.agent.max_output_tokens" type="number" min="0" :class="monoClass" />
          </SettingsField>
          <SettingsField :label="t('settings.field.maxContextChars')">
            <input v-model.number="config.agent.max_context_chars" type="number" min="0" :class="monoClass" />
          </SettingsField>
          <SettingsField :label="t('settings.field.maxHistoryMessages')">
            <input v-model.number="config.agent.max_history_messages" type="number" min="0" :class="monoClass" />
          </SettingsField>
        </div>
        <div class="flex flex-wrap gap-4 pt-2">
          <VcToggle v-model="config.agent.reflexion.enabled" :label="t('settings.field.reflexion')" size="sm" />
          <VcToggle v-model="config.agent.memory_tools.enabled" :label="t('settings.field.memoryTools')" size="sm" />
          <VcToggle v-model="config.agent.heartbeat.enabled" :label="t('settings.field.heartbeat')" size="sm" />
        </div>
        <div v-if="config.agent.heartbeat.enabled" class="grid gap-4 border-t border-vc-border-subtle pt-4 sm:grid-cols-2">
          <SettingsField :label="t('settings.field.heartbeatInterval')">
            <input v-model.number="config.agent.heartbeat.interval_seconds" type="number" min="60" :class="monoClass" />
          </SettingsField>
          <SettingsField :label="t('settings.field.sessionId')">
            <input v-model="config.agent.heartbeat.session_id" type="text" :class="monoClass" />
          </SettingsField>
          <SettingsField :label="t('settings.field.prompt')" class="sm:col-span-2">
            <textarea v-model="config.agent.heartbeat.prompt" rows="3" :class="inputClass" />
          </SettingsField>
        </div>
      </SettingsSection>

      <SettingsSection :title="t('settings.section.router')" :description="t('settings.section.router.desc')">
        <div class="grid gap-4 sm:grid-cols-2">
          <SettingsField :label="t('settings.field.defaultProvider')">
            <select v-model="config.router.default_provider" class="vc-input">
              <option v-for="p in config.providers" :key="p.id" :value="p.id">{{ p.id }}</option>
            </select>
          </SettingsField>
          <SettingsField :label="t('settings.field.defaultModel')">
            <input v-model="config.router.default_model" type="text" :class="monoClass" />
          </SettingsField>
          <SettingsField :label="t('settings.field.intentMode')">
            <VcSelect v-model="config.router.intent_mode" group="intent_mode" />
          </SettingsField>
          <SettingsField :label="t('settings.field.agentRouting')">
            <VcSelect v-model="config.router.agent_routing" group="agent_routing" />
          </SettingsField>
        </div>
        <p class="mt-3 text-xs text-vc-text-muted">
          {{ t('settings.section.router.hint') }}
        </p>
      </SettingsSection>

      <SettingsSection :title="t('settings.section.tools')">
        <div class="space-y-4">
          <div class="flex flex-wrap gap-4">
            <VcToggle v-model="config.tools.shell.enabled" :label="t('settings.field.shell')" size="sm" />
            <VcToggle v-model="config.tools.files.enabled" :label="t('settings.field.files')" size="sm" />
            <VcToggle v-model="config.tools.files.workspace_only" :label="t('settings.field.filesWorkspaceOnly')" size="sm" />
          </div>
          <div v-if="config.tools.shell.enabled" class="grid gap-4 border-t border-vc-border-subtle pt-4 sm:grid-cols-2">
            <SettingsField :label="t('settings.field.sandbox')">
              <VcSelect
                :model-value="config.tools.shell.sandbox ?? 'none'"
                group="sandbox"
                @update:model-value="config.tools.shell.sandbox = $event"
              />
            </SettingsField>
            <SettingsField :label="t('settings.field.workspaceMode')">
              <VcSelect
                :model-value="config.tools.shell.workspace_mode ?? 'ro'"
                group="workspace_mode"
                @update:model-value="config.tools.shell.workspace_mode = $event"
              />
            </SettingsField>
            <SettingsField :label="t('settings.field.dockerImage')">
              <input v-model="config.tools.shell.docker_image" type="text" :class="monoClass" />
            </SettingsField>
            <SettingsField :label="t('settings.field.timeout')">
              <input v-model.number="config.tools.shell.timeout_seconds" type="number" min="0" :class="monoClass" />
            </SettingsField>
          </div>
        </div>
      </SettingsSection>

      <SettingsSection :title="t('settings.section.framework')">
        <div class="flex flex-wrap gap-4">
          <VcToggle v-model="config.framework.enabled" :label="t('settings.field.frameworkEnabled')" size="sm" />
          <VcToggle v-model="config.framework.delegate_enabled" :label="t('settings.field.delegate')" size="sm" />
          <VcToggle v-model="config.framework.hooks_enabled" :label="t('settings.field.hooks')" size="sm" />
          <VcToggle v-model="config.framework.allow_auto_create" :label="t('settings.field.allowAutoCreate')" size="sm" />
        </div>
        <div class="mt-4 grid gap-4 sm:grid-cols-2">
          <SettingsField :label="t('settings.field.maxTotalAgents')">
            <input v-model.number="config.framework.max_total_agents" type="number" min="1" :class="monoClass" />
          </SettingsField>
          <SettingsField :label="t('settings.field.maxConcurrentSpawns')">
            <input v-model.number="config.framework.max_concurrent_spawns" type="number" min="1" :class="monoClass" />
          </SettingsField>
        </div>
      </SettingsSection>

      <SettingsSection :title="t('settings.section.runtime')" :description="t('settings.section.runtime.desc')">
        <div class="grid gap-4 sm:grid-cols-2">
          <SettingsField :label="t('settings.field.runtimeMode')">
            <VcSelect v-model="config.runtime.mode" group="runtime_mode" />
          </SettingsField>
          <SettingsField :label="t('settings.field.maxConcurrentTasks')">
            <input v-model.number="config.runtime.max_concurrent_tasks" type="number" min="1" :class="monoClass" />
          </SettingsField>
          <SettingsField :label="t('settings.field.host')">
            <input v-model="config.server.host" type="text" :class="monoClass" />
          </SettingsField>
          <SettingsField :label="t('settings.field.port')">
            <input v-model.number="config.server.port" type="number" min="1" max="65535" :class="monoClass" />
          </SettingsField>
        </div>
      </SettingsSection>
    </template>

    <p v-else class="text-sm text-vc-text-muted">{{ t('settings.loadFailed') }}</p>
  </div>
</template>
