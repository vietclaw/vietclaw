<script setup lang="ts">
withDefaults(
  defineProps<{
    label?: string
    disabled?: boolean
    size?: 'sm' | 'md'
  }>(),
  { disabled: false, size: 'md' },
)

const model = defineModel<boolean>({ default: false })
</script>

<template>
  <label
    class="vc-toggle"
    :class="[`vc-toggle--${size}`, { 'vc-toggle--disabled': disabled }]"
  >
    <input
      v-model="model"
      type="checkbox"
      class="vc-toggle-input"
      :disabled="disabled"
    />
    <span class="vc-toggle-track" aria-hidden="true">
      <span class="vc-toggle-thumb" />
    </span>
    <span v-if="label" class="vc-toggle-label">{{ label }}</span>
  </label>
</template>
