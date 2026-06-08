<script setup lang="ts">
import { MessageSquare, Database, Server, DollarSign, Radio, FileText, History } from '@lucide/vue'

const tabs = [
  { id: 'chat', label: 'Chat', icon: MessageSquare },
  { id: 'sessions', label: 'Sessions', icon: History },
  { id: 'memory', label: 'Memory', icon: Database },
  { id: 'providers', label: 'Providers', icon: Server },
  { id: 'budget', label: 'Budget', icon: DollarSign },
  { id: 'channels', label: 'Channels', icon: Radio },
  { id: 'logs', label: 'Logs', icon: FileText },
] as const

const activeTab = ref<string>('chat')
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Tab Bar -->
    <div class="flex items-center gap-0.5 px-4 md:px-6 border-b border-zinc-800/60 bg-zinc-950/40 overflow-x-auto vc-scrollbar">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        class="flex items-center gap-1.5 px-3 py-2.5 text-[11px] font-medium transition-colors whitespace-nowrap border-b-2 -mb-px"
        :class="activeTab === tab.id
          ? 'text-zinc-100 border-zinc-100'
          : 'text-zinc-500 border-transparent hover:text-zinc-300'"
        @click="activeTab = tab.id"
      >
        <component :is="tab.icon" :size="13" />
        {{ tab.label }}
      </button>
    </div>

    <!-- Tab Content -->
    <div class="flex-1 overflow-hidden">
      <div v-if="activeTab === 'chat'" class="h-full">
        <ChatPanel />
      </div>
      <div v-else-if="activeTab === 'sessions'" class="h-full overflow-y-auto p-4 md:p-6 vc-scrollbar">
        <SessionsView />
      </div>
      <div v-else-if="activeTab === 'memory'" class="h-full overflow-y-auto p-4 md:p-6 vc-scrollbar">
        <MemoryView />
      </div>
      <div v-else-if="activeTab === 'providers'" class="h-full overflow-y-auto p-4 md:p-6 vc-scrollbar">
        <ProvidersView />
      </div>
      <div v-else-if="activeTab === 'budget'" class="h-full overflow-y-auto p-4 md:p-6 vc-scrollbar">
        <BudgetView />
      </div>
      <div v-else-if="activeTab === 'channels'" class="h-full overflow-y-auto p-4 md:p-6 vc-scrollbar">
        <ChannelsView />
      </div>
      <div v-else-if="activeTab === 'logs'" class="h-full overflow-y-auto p-4 md:p-6 vc-scrollbar">
        <LogsView />
      </div>
    </div>
  </div>
</template>
