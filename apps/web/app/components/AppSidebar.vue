<script setup lang="ts">
import { ArrowLeft, Plus, X } from '@lucide/vue'
import { sessionPath } from '~/composables/useChat'
import { SETTINGS_NAV, isSettingsNavActive } from '~/utils/settingsNav'

defineProps<{ open: boolean }>()
defineEmits<{ close: [] }>()

const { sessions, createSession } = useChat()
const route = useRoute()
const { online } = useDaemon()
const { t } = useI18n()

const isSettings = computed(() => route.path.startsWith('/settings'))

const activeSessionId = computed(() => {
  if (route.path.startsWith('/p/')) {
    return decodeURIComponent(String(route.params.id))
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
    class="fixed inset-y-0 left-0 z-30 grid w-64 shrink-0 border-r border-vc-border-subtle bg-vc-surface transition-transform duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] md:static md:z-auto md:translate-x-0 -translate-x-full"
    :class="[gridRows, open ? 'translate-x-0' : '']"
  >
    <div class="flex items-center justify-between px-5 pt-5 pb-4">
      <NuxtLink :to="chatHome" class="group flex items-center gap-2.5">
        <span
          class="flex h-7 w-7 items-center justify-center rounded-lg bg-vc-accent shadow-sm transition-transform duration-300 ease-[cubic-bezier(0.32,0.72,0,1)] group-hover:-rotate-6"
          aria-hidden="true"
        >
          <svg viewBox="0 0 32 32" class="h-4 w-4" fill="none" stroke="#fdfcf9" stroke-width="3" stroke-linecap="round">
            <path d="M9 8c2.5 4.5 2.5 11.5 0 16M16 7c3 5 3 13 0 18M23 8c2.5 4.5 2.5 11.5 0 16" />
          </svg>
        </span>
        <span class="text-[15px] font-semibold tracking-tight text-vc-text">VietClaw</span>
      </NuxtLink>
      <button type="button" class="vc-btn-ghost rounded-full p-1 md:hidden" @click="$emit('close')">
        <X :size="18" :stroke-width="1.5" />
      </button>
    </div>

    <div class="min-h-0 overflow-y-auto px-3 vc-scrollbar">
      <template v-if="isSettings">
        <div class="pb-3">
          <NuxtLink
            :to="chatHome"
            class="vc-btn vc-btn-outline w-full justify-start gap-2 px-3 py-2 text-sm"
          >
            <ArrowLeft :size="15" :stroke-width="1.5" />
            {{ t('nav.backToChat') }}
          </NuxtLink>
        </div>
        <p class="vc-eyebrow px-3 pb-2 pt-1">{{ t('nav.settings') }}</p>
        <nav class="space-y-0.5 py-1">
          <NuxtLink
            v-for="item in SETTINGS_NAV"
            :key="item.to"
            :to="item.to"
            class="relative block rounded-lg px-3 py-2 text-sm transition-colors duration-200"
            :class="isSettingsNavActive(route.path, item)
              ? 'bg-vc-bg-subtle font-medium text-vc-text'
              : 'text-vc-text-secondary hover:bg-vc-bg-subtle/60 hover:text-vc-text'"
          >
            <span
              v-if="isSettingsNavActive(route.path, item)"
              class="absolute left-0 top-1/2 h-4 w-0.5 -translate-y-1/2 rounded-full bg-vc-accent"
              aria-hidden="true"
            />
            {{ item.labelKey ? t(item.labelKey) : '' }}
          </NuxtLink>
        </nav>
      </template>

      <template v-else>
        <div class="pb-3">
          <button
            type="button"
            class="vc-btn vc-btn-outline w-full justify-start gap-2 px-3 py-2"
            @click="startNewSession"
          >
            <span class="flex h-5 w-5 items-center justify-center rounded-full bg-vc-accent-soft text-vc-accent">
              <Plus :size="13" :stroke-width="2" />
            </span>
            {{ t('nav.newChat') }}
          </button>
        </div>
        <p class="vc-eyebrow px-3 pb-2 pt-1">{{ t('nav.conversations') }}</p>
        <SidebarSessionTree />
      </template>
    </div>

    <div v-if="!isSettings" class="border-t border-vc-border-subtle px-5 py-3">
      <NuxtLink
        to="/settings"
        class="block text-sm text-vc-text-secondary transition-colors duration-200 hover:text-vc-text"
      >
        {{ t('nav.settings') }}
      </NuxtLink>
    </div>

    <div class="flex items-center gap-2 border-t border-vc-border-subtle px-5 py-3.5">
      <span class="vc-status-dot" :class="online ? 'vc-status-dot--on' : 'vc-status-dot--off'" aria-hidden="true" />
      <p class="text-xs text-vc-text-muted">
        {{ online ? t('status.connected') : t('status.disconnected') }}
      </p>
    </div>
  </aside>
</template>
