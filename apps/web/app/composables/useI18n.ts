import vi from '../../locales/vi.json'
import en from '../../locales/en.json'

type Locale = 'vi' | 'en'
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
    return template.replace(/%s/g, () => String(args.shift()))
  }

  function toolLabel(toolName: string): string {
    const normalized = normalizeToolName(toolName)
    const key = `tool.ui.${normalized}`
    const label = catalogs[lang.value]?.[key] ?? catalogs.vi[key]
    return label ?? toolName
  }

  function setLanguage(next: Locale) {
    lang.value = next
  }

  return { lang, t, toolLabel, setLanguage }
}
