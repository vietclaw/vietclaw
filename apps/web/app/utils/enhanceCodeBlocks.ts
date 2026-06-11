import hljs from 'highlight.js'
import { languageFromClassName, languageLabel, shouldHighlightLanguage } from '~/utils/codeLang'

export function enhanceCodeBlocks(
  root: Element,
  onCopy?: (text: string) => void,
  copyLabel = 'Copy',
) {
  root.querySelectorAll('pre').forEach((pre) => {
    if (pre.closest('.vc-code-block')) return

    const code = pre.querySelector('code')
    const codeEl = code as HTMLElement | null
    const lang = codeEl ? languageFromClassName(codeEl.className) : ''
    const raw = codeEl?.textContent || pre.textContent || ''

    if (codeEl && shouldHighlightLanguage(lang)) {
      hljs.highlightElement(codeEl)
    } else if (codeEl) {
      codeEl.classList.remove('hljs')
      codeEl.removeAttribute('data-highlighted')
    }

    const wrapper = document.createElement('div')
    wrapper.className = 'vc-code-block'

    const header = document.createElement('div')
    header.className = 'vc-code-block-header'

    const label = document.createElement('span')
    label.className = 'vc-code-block-lang'
    label.textContent = languageLabel(lang)

    const copyBtn = document.createElement('button')
    copyBtn.type = 'button'
    copyBtn.className = 'vc-code-block-copy'
    copyBtn.textContent = copyLabel
    copyBtn.addEventListener('click', (e) => {
      e.preventDefault()
      e.stopPropagation()
      if (onCopy) onCopy(raw)
      else void navigator.clipboard.writeText(raw)
    })

    header.append(label, copyBtn)

    const body = document.createElement('div')
    body.className = 'vc-code-block-body'

    const parent = pre.parentNode
    if (!parent) return

    parent.insertBefore(wrapper, pre)
    wrapper.appendChild(header)
    body.appendChild(pre)
    wrapper.appendChild(body)
  })
}
