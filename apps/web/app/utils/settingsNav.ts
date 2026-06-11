export type SettingsNavItem = {
  to: string
  label: string
  exact?: boolean
}

export const SETTINGS_NAV: SettingsNavItem[] = [
  { to: '/settings', label: 'Tổng quan', exact: true },
  { to: '/settings/providers', label: 'Providers' },
  { to: '/settings/budget', label: 'Budget' },
  { to: '/settings/channels', label: 'Kênh' },
  { to: '/settings/memory', label: 'Memory' },
  { to: '/settings/logs', label: 'Logs' },
]

export function isSettingsNavActive(path: string, item: SettingsNavItem) {
  if (item.exact) return path === item.to
  return path === item.to || path.startsWith(item.to + '/')
}
