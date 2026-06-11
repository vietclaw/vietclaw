<script setup lang="ts">
import { ArrowLeft, Plus, Trash2, X } from '@lucide/vue'
import { sessionPath } from '~/composables/useChat'
import { SETTINGS_NAV, isSettingsNavActive } from '~/utils/settingsNav'

defineProps<{ open: boolean }>()
defineEmits<{ close: [] }>()

const { sessions, createSession, deleteSession } = useChat()
const route = useRoute()
const { online } = useDaemon()

const isSettings = computed(() => route.path.startsWith('/settings'))

const activeSessionId = computed(() => {
  if (route.path.startsWith('/p/')) {
    return String(route.params.id)
  }
  return ''
})

const chatHome = computed(() => {
  const id = activeSessionId.value || sessions.value[0]?.id
  return id ? sessionPath(id) : '/'
})

const gridRows = computed(() =>
  isSettings.value
    ? 'grid-rows-[auto_minmax(0,1fr)_auto]'
    : 'grid-rows-[auto_minmax(0,1fr)_auto_auto]',
)

async function startNewSession() {
  const session = createSession()
  await navigateTo(sessionPath(session.id))
}
</script>

<template>
  <aside
    class="fixed inset-y-0 left-0 z-30 grid w-60 shrink-0 border-r border-vc-border bg-vc-surface transition-transform duration-200 ease-out md:static md:z-auto md:translate-x-0 -translate-x-full"
    :class="[gridRows, open ? 'translate-x-0' : '']"
  >
    <div class="flex items-center justify-between px-4 pt-5 pb-3">
      <span class="text-[15px] font-semibold tracking-tight text-vc-text">VietClaw</span>
      <button type="button" class="vc-btn-ghost rounded-md p-1 md:hidden" @click="$emit('close')">
        <X :size="18" :stroke-width="1.75" />
      </button>
    </div>

    <div class="min-h-0 overflow-y-auto px-2 vc-scrollbar">
      <template v-if="isSettings">
        <div class="pb-2">
          <NuxtLink
            :to="chatHome"
            class="vc-btn vc-btn-ghost w-full justify-start gap-2 px-2 py-2 text-sm"
          >
            <ArrowLeft :size="16" :stroke-width="1.75" />
            Quay lại chat
          </NuxtLink>
        </div>
        <p class="px-3 pb-2 text-xs font-semibold text-vc-text-muted">Cài đặt</p>
        <nav class="space-y-0.5 py-1">
          <NuxtLink
            v-for="item in SETTINGS_NAV"
            :key="item.to"
            :to="item.to"
            class="block rounded-md px-3 py-2 text-sm transition-colors"
            :class="isSettingsNavActive(route.path, item)
              ? 'bg-vc-bg-subtle font-medium text-vc-text'
              : 'text-vc-text-secondary hover:bg-vc-bg-subtle hover:text-vc-text'"
          >
            {{ item.label }}
          </NuxtLink>
        </nav>
      </template>

      <template v-else>
        <div class="pb-2">
          <button type="button" class="vc-btn vc-btn-ghost w-full justify-start gap-2 px-2 py-2" @click="startNewSession">
            <Plus :size="16" :stroke-width="1.75" />
            Hội thoại mới
          </button>
        </div>
        <div class="space-y-0.5 py-1">
          <div
            v-for="session in sessions"
            :key="session.id"
            class="group flex items-center gap-1 rounded-md"
            :class="session.id === activeSessionId ? 'bg-vc-bg-subtle' : ''"
          >
            <NuxtLink
              :to="sessionPath(session.id)"
              class="min-w-0 flex-1 truncate px-3 py-2 text-left text-sm transition-colors"
              :class="session.id === activeSessionId
                ? 'font-medium text-vc-text'
                : 'text-vc-text-secondary hover:text-vc-text'"
            >
              {{ session.title }}
            </NuxtLink>
            <button
              v-if="sessions.length > 1"
              type="button"
              class="vc-btn-ghost rounded-md p-1.5 opacity-0 group-hover:opacity-100"
              @click.stop="deleteSession(session.id)"
            >
              <Trash2 :size="13" :stroke-width="1.75" />
            </button>
          </div>
        </div>
      </template>
    </div>

    <div v-if="!isSettings" class="border-t border-vc-border-subtle px-4 py-3">
      <NuxtLink
        to="/settings"
        class="block text-sm text-vc-text-secondary transition-colors hover:text-vc-text"
      >
        Cài đặt
      </NuxtLink>
    </div>

    <div class="border-t border-vc-border-subtle px-4 py-3">
      <p class="text-xs text-vc-text-muted">
        {{ online ? 'Đã kết nối daemon' : 'Chưa kết nối' }}
      </p>
    </div>
  </aside>
</template>
