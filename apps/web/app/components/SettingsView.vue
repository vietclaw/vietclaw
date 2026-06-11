<script setup lang="ts">
import SettingsField from '~/components/settings/SettingsField.vue'
import SettingsSection from '~/components/settings/SettingsSection.vue'

const { config } = useSettings()

const inputClass = 'w-full rounded-md border border-vc-border bg-vc-surface px-3 py-2 text-sm text-vc-text focus:border-vc-accent focus:outline-none'
const monoClass = `${inputClass} font-mono text-xs`
const selectClass = `${inputClass} cursor-pointer`
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-lg font-semibold tracking-tight text-vc-text">Tổng quan</h1>
      <p class="mt-1 text-sm text-vc-text-muted">
        Agent, routing, tools và runtime. Providers, budget, kênh, memory và logs có trang riêng trong menu.
      </p>
    </div>

    <template v-if="config">
      <SettingsSection title="Agent" description="Hành vi và giới hạn của agent">
        <div class="grid gap-4 sm:grid-cols-2">
          <SettingsField label="Tên">
            <input v-model="config.agent.name" type="text" :class="inputClass" />
          </SettingsField>
          <SettingsField label="Ngôn ngữ">
            <select v-model="config.agent.language" :class="selectClass">
              <option value="vi">Tiếng Việt</option>
              <option value="en">English</option>
            </select>
          </SettingsField>
          <SettingsField label="Experience">
            <select v-model="config.agent.experience" :class="selectClass">
              <option value="prompt">prompt</option>
              <option value="pro">pro</option>
            </select>
          </SettingsField>
          <SettingsField label="Style">
            <input v-model="config.agent.style" type="text" :class="monoClass" />
          </SettingsField>
          <SettingsField label="Max steps" hint="0 = không giới hạn">
            <input v-model.number="config.agent.max_steps" type="number" min="0" :class="monoClass" />
          </SettingsField>
          <SettingsField label="Max output tokens" hint="0 = không giới hạn">
            <input v-model.number="config.agent.max_output_tokens" type="number" min="0" :class="monoClass" />
          </SettingsField>
          <SettingsField label="Max context chars">
            <input v-model.number="config.agent.max_context_chars" type="number" min="0" :class="monoClass" />
          </SettingsField>
          <SettingsField label="Max history messages">
            <input v-model.number="config.agent.max_history_messages" type="number" min="0" :class="monoClass" />
          </SettingsField>
        </div>
        <div class="flex flex-wrap gap-4 pt-2">
          <VcToggle v-model="config.agent.reflexion.enabled" label="Reflexion" size="sm" />
          <VcToggle v-model="config.agent.memory_tools.enabled" label="Memory tools" size="sm" />
          <VcToggle v-model="config.agent.heartbeat.enabled" label="Heartbeat" size="sm" />
        </div>
        <div v-if="config.agent.heartbeat.enabled" class="grid gap-4 sm:grid-cols-2 border-t border-vc-border-subtle pt-4">
          <SettingsField label="Interval (giây)">
            <input v-model.number="config.agent.heartbeat.interval_seconds" type="number" min="60" :class="monoClass" />
          </SettingsField>
          <SettingsField label="Session ID">
            <input v-model="config.agent.heartbeat.session_id" type="text" :class="monoClass" />
          </SettingsField>
          <SettingsField label="Prompt" class="sm:col-span-2">
            <textarea v-model="config.agent.heartbeat.prompt" rows="3" :class="inputClass" />
          </SettingsField>
        </div>
      </SettingsSection>

      <SettingsSection title="Router" description="Provider và model mặc định">
        <div class="grid gap-4 sm:grid-cols-2">
          <SettingsField label="Default provider">
            <select v-model="config.router.default_provider" :class="selectClass">
              <option v-for="p in config.providers" :key="p.id" :value="p.id">{{ p.id }}</option>
            </select>
          </SettingsField>
          <SettingsField label="Default model">
            <input v-model="config.router.default_model" type="text" :class="monoClass" />
          </SettingsField>
          <SettingsField label="Intent mode">
            <input v-model="config.router.intent_mode" type="text" :class="monoClass" />
          </SettingsField>
          <SettingsField label="Agent routing">
            <input v-model="config.router.agent_routing" type="text" :class="monoClass" />
          </SettingsField>
        </div>
        <p class="mt-3 text-xs text-vc-text-muted">
          Cheap first và escalation nằm trong trang Budget.
        </p>
      </SettingsSection>

      <SettingsSection title="Tools">
        <div class="space-y-4">
          <div class="flex flex-wrap gap-4">
            <VcToggle v-model="config.tools.shell.enabled" label="Shell" size="sm" />
            <VcToggle v-model="config.tools.files.enabled" label="Files" size="sm" />
            <VcToggle v-model="config.tools.files.workspace_only" label="Files workspace only" size="sm" />
          </div>
          <div v-if="config.tools.shell.enabled" class="grid gap-4 sm:grid-cols-2 border-t border-vc-border-subtle pt-4">
            <SettingsField label="Sandbox">
              <select v-model="config.tools.shell.sandbox" :class="selectClass">
                <option value="none">none</option>
                <option value="docker">docker</option>
              </select>
            </SettingsField>
            <SettingsField label="Workspace mode">
              <select v-model="config.tools.shell.workspace_mode" :class="selectClass">
                <option value="ro">ro</option>
                <option value="rw">rw</option>
              </select>
            </SettingsField>
            <SettingsField label="Docker image">
              <input v-model="config.tools.shell.docker_image" type="text" :class="monoClass" />
            </SettingsField>
            <SettingsField label="Timeout (s)">
              <input v-model.number="config.tools.shell.timeout_seconds" type="number" min="0" :class="monoClass" />
            </SettingsField>
          </div>
        </div>
      </SettingsSection>

      <SettingsSection title="Framework">
        <div class="flex flex-wrap gap-4">
          <VcToggle v-model="config.framework.enabled" label="Enabled" size="sm" />
          <VcToggle v-model="config.framework.delegate_enabled" label="Delegate" size="sm" />
          <VcToggle v-model="config.framework.hooks_enabled" label="Hooks" size="sm" />
        </div>
      </SettingsSection>

      <SettingsSection title="Runtime & Server" description="Thay port/host cần restart daemon">
        <div class="grid gap-4 sm:grid-cols-2">
          <SettingsField label="Mode">
            <input v-model="config.runtime.mode" type="text" :class="monoClass" />
          </SettingsField>
          <SettingsField label="Max concurrent tasks">
            <input v-model.number="config.runtime.max_concurrent_tasks" type="number" min="1" :class="monoClass" />
          </SettingsField>
          <SettingsField label="Host">
            <input v-model="config.server.host" type="text" :class="monoClass" />
          </SettingsField>
          <SettingsField label="Port">
            <input v-model.number="config.server.port" type="number" min="1" max="65535" :class="monoClass" />
          </SettingsField>
        </div>
      </SettingsSection>
    </template>

    <p v-else class="text-sm text-vc-text-muted">Không tải được cấu hình.</p>
  </div>
</template>
