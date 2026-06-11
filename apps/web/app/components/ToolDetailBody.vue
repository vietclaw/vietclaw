<script setup lang="ts">
import hljs from 'highlight.js'
import {
  parseToolInputDisplay,
  parseToolResultDisplay,
  type ToolDisplayView,
} from '~/utils/toolDisplay'
import { languageLabel, shouldHighlightLanguage } from '~/utils/codeLang'

const props = defineProps<{
  toolName: string
  raw?: string
  side: 'input' | 'result'
  inputRaw?: string
  label: string
}>()

const { t } = useI18n()
const toast = useToast()
const codeRef = ref<HTMLElement | null>(null)
const failureOutputRef = ref<HTMLElement | null>(null)

const view = computed<ToolDisplayView>(() => {
  if (props.side === 'input') {
    return parseToolInputDisplay(props.toolName, props.raw)
  }
  return parseToolResultDisplay(props.toolName, props.raw, props.inputRaw)
})

function highlightCode() {
  const el = codeRef.value
  if (!el) return
  if (view.value.mode !== 'code') return
  if (shouldHighlightLanguage(view.value.lang)) {
    hljs.highlightElement(el)
  } else {
    el.classList.remove('hljs')
    el.removeAttribute('data-highlighted')
  }
}

async function copyText(text: string) {
  await window.navigator.clipboard.writeText(text)
  toast.add(t('chat.copied'), 'success')
}

async function copyContent() {
  if (view.value.mode === 'code') {
    await copyText(view.value.content)
  } else if (view.value.mode === 'failure' && view.value.output) {
    await copyText(view.value.output)
  }
}

watch(view, async () => {
  if (view.value.mode === 'code' || (view.value.mode === 'failure' && view.value.output)) {
    await nextTick()
    highlightCode()
  }
}, { flush: 'post' })

onMounted(() => highlightCode())
</script>

<template>
  <div v-if="view.mode !== 'empty'" class="space-y-1.5">
    <p class="text-[11px] font-medium text-vc-text-muted">{{ label }}</p>

    <div v-if="view.mode === 'code'" class="vc-code-block">
      <div class="vc-code-block-header">
        <span class="vc-code-block-lang">
          {{ languageLabel(view.lang) }}
          <span v-if="view.path" class="vc-code-block-path"> · {{ view.path }}</span>
        </span>
        <button type="button" class="vc-code-block-copy" @click="copyContent">
          {{ t('chat.copy') }}
        </button>
      </div>
      <div class="vc-code-block-body">
        <pre class="vc-code-block-pre vc-scrollbar"><code
          ref="codeRef"
          :class="`language-${view.lang}`"
        >{{ view.content }}</code></pre>
      </div>
    </div>

    <div v-else-if="view.mode === 'command'" class="vc-tool-command">
      <span class="vc-tool-command-prompt">$</span>
      <span class="font-mono">{{ view.command }}</span>
    </div>

    <div v-else-if="view.mode === 'path'" class="vc-tool-path">
      <span class="font-mono text-xs text-vc-text-secondary">{{ view.path }}</span>
    </div>

    <div v-else-if="view.mode === 'failure'" class="vc-tool-failure">
      <p v-if="view.error" class="vc-tool-failure-error">{{ view.error }}</p>
      <p v-else-if="!view.output" class="vc-tool-failure-muted">{{ t('tool.ui.no_output') }}</p>
      <div v-if="view.output" class="vc-code-block">
        <div class="vc-code-block-header">
          <span class="vc-code-block-lang">{{ languageLabel('plaintext') }}</span>
          <button type="button" class="vc-code-block-copy" @click="copyContent">
            {{ t('chat.copy') }}
          </button>
        </div>
        <div class="vc-code-block-body">
          <pre class="vc-code-block-pre vc-scrollbar"><code
            ref="failureOutputRef"
            class="language-plaintext"
          >{{ view.output }}</code></pre>
        </div>
      </div>
    </div>

    <pre
      v-else-if="view.mode === 'raw'"
      class="vc-tool-raw vc-scrollbar"
    >{{ view.text }}</pre>
  </div>
</template>
