import vi from '../../locales/vi.json'
import en from '../../locales/en.json'
import type { OptionGroup } from '~/utils/configOptions'

export type Locale = 'vi' | 'en'
type Catalog = Record<string, string>

const catalogs: Record<Locale, Catalog> = { vi, en }

function normalizeToolName(name: string): string {
  return name.trim().replace(/\./g, '_')
}

export function useI18n() {
  const lang = useState<Locale>('appLanguage', () => 'vi')

  function t(key: string, ...args: unknown[]): string {
    const template = catalogs[lang.value]?.[key] ?? catalogs.vi[key] ?? key
    if (args.length === 0) return template
    let i = 0
    return template.replace(/%s/g, () => String(args[i++]))
  }

  function option(group: OptionGroup, value: string): string {
    const key = `option.${group}.${value}`
    const label = catalogs[lang.value]?.[key] ?? catalogs.vi[key]
    return label ?? value
  }

  function toolLabel(toolName: string, failed = false): string {
    const base = toolName.split(':')[0] ?? toolName
    const normalized = normalizeToolName(base)
    const key = failed ? `tool.ui.${normalized}_failed` : `tool.ui.${normalized}`
    const label = catalogs[lang.value]?.[key] ?? catalogs.vi[key]
    if (label) return label
    if (failed) {
      const fallback = catalogs[lang.value]?.[`tool.ui.${normalized}`] ?? catalogs.vi[`tool.ui.${normalized}`]
      if (fallback) return fallback
    }
    return toolName
  }

  function channelStatus(channel: { running?: boolean; enabled?: boolean } | undefined): string {
    if (!channel) return t('status.off')
    if (channel.running) return t('status.running')
    if (channel.enabled) return t('status.enabled')
    return t('status.off')
  }

  function setLanguage(next: Locale) {
    lang.value = next
  }

  return { lang, t, option, toolLabel, channelStatus, setLanguage }
}
