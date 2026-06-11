<script setup lang="ts">
import { X, Database, Server, DollarSign, Radio, FileText, History, Sliders } from '@lucide/vue'

const props = defineProps<{ open: boolean }>()
defineEmits<{ close: [] }>()

const tabs = [
  { id: 'sessions', label: 'Phiên', icon: History },
  { id: 'memory', label: 'Memory', icon: Database },
  { id: 'providers', label: 'Providers', icon: Server },
  { id: 'budget', label: 'Budget', icon: DollarSign },
  { id: 'channels', label: 'Kênh', icon: Radio },
  { id: 'logs', label: 'Logs', icon: FileText },
  { id: 'system', label: 'Hệ thống', icon: Sliders },
] as const

const activeTab = ref<string>('memory')
const { status, framework, online, refresh } = useDaemon()

watch(() => props.open, (v) => {
  if (v) refresh()
})
</script>

<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="open" class="fixed inset-0 z-50 flex justify-end">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="$emit('close')" />
        <aside class="relative flex h-full w-full max-w-2xl flex-col border-l border-zinc-800 bg-zinc-950 shadow-2xl">
          <header class="flex items-center justify-between border-b border-zinc-800 px-5 py-4">
            <div>
              <h2 class="text-sm font-semibold text-zinc-100">Công cụ nâng cao</h2>
              <p class="text-[11px] text-zinc-500">Không bắt buộc — chat là đủ cho hầu hết việc.</p>
            </div>
            <button class="rounded p-1.5 text-zinc-500 hover:bg-zinc-900 hover:text-zinc-300" @click="$emit('close')">
              <X :size="18" />
            </button>
          </header>

          <div class="flex gap-1 overflow-x-auto border-b border-zinc-800/80 px-4 py-2 vc-scrollbar">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              class="flex items-center gap-1.5 rounded-md px-3 py-2 text-[11px] font-medium whitespace-nowrap transition-colors"
              :class="activeTab === tab.id ? 'bg-zinc-900 text-zinc-100' : 'text-zinc-500 hover:text-zinc-300'"
              @click="activeTab = tab.id"
            >
              <component :is="tab.icon" :size="13" />
              {{ tab.label }}
            </button>
          </div>

          <div class="flex-1 overflow-y-auto p-4 md:p-6 vc-scrollbar">
            <div v-if="activeTab === 'sessions'" class="h-full min-h-[320px]">
              <SessionsView />
            </div>
            <div v-else-if="activeTab === 'memory'">
              <MemoryView />
            </div>
            <div v-else-if="activeTab === 'providers'">
              <ProvidersView />
            </div>
            <div v-else-if="activeTab === 'budget'">
              <BudgetView />
            </div>
            <div v-else-if="activeTab === 'channels'">
              <ChannelsView />
            </div>
            <div v-else-if="activeTab === 'logs'">
              <LogsView />
            </div>
            <div v-else-if="activeTab === 'system'" class="space-y-4 max-w-lg">
              <div class="rounded-lg border border-zinc-800 bg-zinc-900/30 p-4">
                <div class="flex items-center gap-2 text-xs">
                  <span
                    class="h-2 w-2 rounded-full"
                    :class="online ? 'bg-emerald-500' : 'bg-rose-500'"
                  />
                  <span class="text-zinc-300">{{ online ? 'Daemon đang chạy' : 'Không kết nối daemon' }}</span>
                </div>
                <dl class="mt-3 space-y-2 text-[11px] font-mono text-zinc-400">
                  <div class="flex justify-between gap-4"><dt>version</dt><dd class="text-zinc-200">{{ status?.version || '—' }}</dd></div>
                  <div class="flex justify-between gap-4"><dt>uptime</dt><dd class="text-zinc-200">{{ status?.uptime || '—' }}</dd></div>
                  <div class="flex justify-between gap-4"><dt>mode</dt><dd class="text-zinc-200">{{ status?.mode || '—' }}</dd></div>
                </dl>
              </div>
              <div class="rounded-lg border border-zinc-800 bg-zinc-900/30 p-4 text-[11px] text-zinc-400">
                <p class="mb-2 font-medium text-zinc-300">Agent framework</p>
                <ul class="space-y-1 font-mono">
                  <li>delegate: {{ framework?.delegate_enabled ? 'on' : 'off' }}</li>
                  <li>hooks: {{ framework?.hooks_enabled ? 'on' : 'off' }} ({{ framework?.hooks_registered ?? 0 }})</li>
                  <li>agents: {{ framework?.agents?.length ?? 0 }}</li>
                </ul>
              </div>
              <p class="text-[11px] leading-relaxed text-zinc-500">
                Cấu hình server: chỉnh <code class="text-zinc-400">config.json</code> trong data dir hoặc dùng CLI
                <code class="text-zinc-400">vietclaw doctor</code>. UI chat không cần cấu hình thêm.
              </p>
            </div>
          </div>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.drawer-enter-active,
.drawer-leave-active {
  transition: opacity 0.2s ease;
}
.drawer-enter-active aside,
.drawer-leave-active aside {
  transition: transform 0.25s ease;
}
.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
}
.drawer-enter-from aside,
.drawer-leave-to aside {
  transform: translateX(100%);
}
</style>
