<script setup lang="ts">
const { toasts, remove } = useToast()

function toastClass(type: string) {
  if (type === 'error' || type === 'warning') return 'text-vc-error'
  return 'text-vc-text'
}
</script>

<template>
  <div class="fixed bottom-4 right-4 z-50 flex max-w-sm flex-col gap-2">
    <TransitionGroup name="toast">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="rounded-lg border border-vc-border bg-vc-surface px-4 py-3 text-sm shadow-sm"
        :class="toastClass(t.type)"
        @click="remove(t.id)"
      >
        {{ t.msg }}
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-enter-active {
  transition: all 0.2s ease-out;
}
.toast-leave-active {
  transition: all 0.15s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateY(6px);
}
.toast-leave-to {
  opacity: 0;
}
</style>
