export type SettingsNavItem = {
  to: string
  labelKey: string
  exact?: boolean
}

export const SETTINGS_NAV: SettingsNavItem[] = [
  { to: '/settings', labelKey: 'nav.settingsOverview', exact: true },
  { to: '/settings/providers', labelKey: 'nav.providers' },
  { to: '/settings/models', labelKey: 'nav.models' },
  { to: '/settings/budget', labelKey: 'nav.budget' },
  { to: '/settings/channels', labelKey: 'nav.channels' },
  { to: '/settings/memory', labelKey: 'nav.memory' },
  { to: '/settings/logs', labelKey: 'nav.logs' },
]

export function isSettingsNavActive(path: string, item: SettingsNavItem) {
  if (item.exact) return path === item.to
  return path === item.to || path.startsWith(item.to + '/')
}
