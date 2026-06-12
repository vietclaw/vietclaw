<script setup lang="ts">
import { ArrowLeft, RefreshCw, Save } from '@lucide/vue'

const { currentSessionId, sessions } = useChat()
const { loading, saving, dirty, load, save, reload, discard } = useSettings()
const { t } = useI18n()

const chatHome = computed(() => {
  const id = currentSessionId.value || sessions.value[0]?.id
  return id ? `/p/${id}` : '/'
})

onMounted(() => {
  if (!useSettings().config.value) void load()
})
</script>

<template>
  <div class="flex h-full min-h-0 max-h-full flex-col overflow-hidden">
    <div class="flex min-w-0 shrink-0 items-center justify-between gap-4 px-4 py-3 md:px-6">
      <NuxtLink :to="chatHome" class="vc-link flex items-center gap-1.5 text-sm">
        <ArrowLeft :size="16" :stroke-width="1.5" />
        {{ t('nav.backToChat') }}
      </NuxtLink>
      <div class="flex items-center gap-2">
        <Transition name="dirty">
          <span
            v-if="dirty"
            class="rounded-full bg-vc-accent-soft px-2.5 py-0.5 text-xs font-medium text-vc-accent"
          >
            {{ t('common.unsaved') }}
          </span>
        </Transition>
        <button
          type="button"
          class="vc-btn vc-btn-ghost text-xs disabled:opacity-40"
          :disabled="saving"
          @click="discard"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          class="vc-btn vc-btn-ghost text-xs disabled:opacity-40"
          :disabled="saving"
          @click="reload"
        >
          <RefreshCw :size="14" :stroke-width="1.5" />
          {{ t('common.reload') }}
        </button>
        <button
          type="button"
          class="vc-btn vc-btn-primary text-xs disabled:opacity-40"
          :disabled="saving || !dirty"
          @click="save"
        >
          <Save :size="14" :stroke-width="1.5" />
          {{ t('common.save') }}
        </button>
      </div>
    </div>

    <div class="min-h-0 flex-1 overflow-y-auto vc-scrollbar">
      <div v-if="loading" class="mx-auto max-w-3xl space-y-3 p-6 md:p-8">
        <div v-for="i in 4" :key="i" class="h-24 animate-pulse rounded-2xl bg-vc-bg-subtle" :style="{ animationDelay: `${i * 0.1}s` }" />
      </div>
      <div v-else class="mx-auto max-w-3xl p-4 pb-16 md:p-8 md:pb-20">
        <slot />
      </div>
    </div>
  </div>
</template>

<style scoped>
.dirty-enter-active,
.dirty-leave-active {
  transition: opacity 0.25s var(--vc-ease), transform 0.25s var(--vc-ease);
}
.dirty-enter-from,
.dirty-leave-to {
  opacity: 0;
  transform: translateY(2px);
}
</style>
