<script setup lang="ts">
import type { OptionGroup } from '~/utils/configOptions'
import { OPTION_GROUPS } from '~/utils/configOptions'

const props = withDefaults(
  defineProps<{
    modelValue: string
    group: OptionGroup
    selectClass?: string
  }>(),
  { selectClass: 'vc-input' },
)

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const { option } = useI18n()

const values = computed(() => OPTION_GROUPS[props.group])

const current = computed({
  get: () => props.modelValue,
  set: (v: string) => emit('update:modelValue', v),
})

const options = computed(() => {
  const list = [...values.value]
  if (props.modelValue && !list.includes(props.modelValue)) {
    list.unshift(props.modelValue)
  }
  return list
})
</script>

<template>
  <select v-model="current" :class="selectClass">
    <option v-for="v in options" :key="v" :value="v">
      {{ option(group, v) }}
    </option>
  </select>
</template>
