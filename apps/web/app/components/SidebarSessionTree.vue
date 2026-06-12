<script setup lang="ts">
import { Trash2 } from '@lucide/vue'
import { sessionPath } from '~/composables/useChat'
import type { SpawnStatus } from '~/composables/useChat'

const {
  rootSessions,
  childrenOf,
  expandedRootId,
  activeRootId,
  deleteSession,
  loadChildrenForParent,
} = useChat()
const route = useRoute()
const { t } = useI18n()

const activeSessionId = computed(() => {
  if (route.path.startsWith('/p/')) {
    return decodeURIComponent(String(route.params.id))
  }
  return ''
})

watch(activeRootId, (rootId) => {
  if (rootId) void loadChildrenForParent(rootId)
}, { immediate: true })

function spawnStatusLabel(status?: SpawnStatus): string {
  if (!status) return ''
  const key = `chat.spawn.${status}`
  const label = t(key)
  return label === key ? status : label
}

function spawnStatusClass(status?: SpawnStatus): string {
  if (status === 'done') return 'text-vc-success'
  if (status === 'failed') return 'text-vc-error'
  return 'text-vc-accent'
}
</script>

<template>
  <div class="space-y-0.5 py-1">
    <template v-for="root in rootSessions" :key="root.id">
      <div
        class="group relative flex items-center gap-1 rounded-lg transition-colors duration-200"
        :class="activeSessionId === root.id ? 'bg-vc-bg-subtle' : 'hover:bg-vc-bg-subtle/60'"
      >
        <span
          v-if="activeSessionId === root.id"
          class="absolute left-0 top-1/2 h-4 w-0.5 -translate-y-1/2 rounded-full bg-vc-accent"
          aria-hidden="true"
        />
        <NuxtLink
          :to="sessionPath(root.id)"
          class="min-w-0 flex-1 truncate px-3 py-2 text-left text-sm transition-colors duration-200"
          :class="activeSessionId === root.id
            ? 'font-medium text-vc-text'
            : 'text-vc-text-secondary group-hover:text-vc-text'"
        >
          {{ root.title }}
        </NuxtLink>
        <button
          v-if="rootSessions.length > 1"
          type="button"
          class="vc-btn-ghost mr-1 rounded-full p-1.5 opacity-0 transition-opacity duration-200 group-hover:opacity-100 focus-visible:opacity-100"
          :aria-label="`Xóa ${root.title}`"
          @click.stop="deleteSession(root.id)"
        >
          <Trash2 :size="13" :stroke-width="1.5" />
        </button>
      </div>

      <div
        v-if="expandedRootId === root.id && childrenOf(root.id).length"
        class="ml-3 space-y-0.5 border-l border-vc-border-subtle py-0.5 pl-2"
      >
        <NuxtLink
          v-for="child in childrenOf(root.id)"
          :key="child.id"
          :to="sessionPath(child.id)"
          class="block rounded-md px-2 py-1.5 text-xs transition-colors duration-200"
          :class="activeSessionId === child.id
            ? 'bg-vc-bg-subtle font-medium text-vc-text'
            : 'text-vc-text-secondary hover:bg-vc-bg-subtle/60 hover:text-vc-text'"
        >
          <div class="flex items-center gap-1.5">
            <span class="min-w-0 flex-1 truncate font-medium">{{ child.agentId || child.title }}</span>
            <span class="shrink-0" :class="spawnStatusClass(child.spawnStatus)">
              {{ spawnStatusLabel(child.spawnStatus) }}
            </span>
          </div>
          <p v-if="child.taskPreview" class="mt-0.5 truncate text-[10px] leading-snug text-vc-text-muted">
            {{ child.taskPreview }}
          </p>
        </NuxtLink>
      </div>
    </template>
  </div>
</template>
