const LANG_LABELS: Record<string, string> = {
  js: 'JavaScript',
  javascript: 'JavaScript',
  mjs: 'JavaScript',
  cjs: 'JavaScript',
  ts: 'TypeScript',
  typescript: 'TypeScript',
  tsx: 'TypeScript',
  jsx: 'JavaScript',
  py: 'Python',
  python: 'Python',
  go: 'Go',
  golang: 'Go',
  rs: 'Rust',
  rust: 'Rust',
  java: 'Java',
  kt: 'Kotlin',
  kotlin: 'Kotlin',
  cs: 'C#',
  csharp: 'C#',
  cpp: 'C++',
  c: 'C',
  h: 'C',
  rb: 'Ruby',
  ruby: 'Ruby',
  php: 'PHP',
  swift: 'Swift',
  sh: 'Shell',
  bash: 'Bash',
  zsh: 'Zsh',
  shell: 'Shell',
  ps1: 'PowerShell',
  powershell: 'PowerShell',
  sql: 'SQL',
  json: 'JSON',
  yaml: 'YAML',
  yml: 'YAML',
  toml: 'TOML',
  xml: 'XML',
  html: 'HTML',
  htm: 'HTML',
  css: 'CSS',
  scss: 'SCSS',
  md: 'Markdown',
  markdown: 'Markdown',
  vue: 'Vue',
  dockerfile: 'Dockerfile',
  text: 'Text',
  plaintext: 'Text',
  plain: 'Text',
  txt: 'Text',
}

export function languageFromClassName(className: string): string {
  for (const part of className.split(/\s+/)) {
    if (part.startsWith('language-')) {
      return part.slice('language-'.length)
    }
  }
  return ''
}

export function languageLabel(lang: string): string {
  if (!lang) return 'Text'
  const key = lang.toLowerCase()
  return LANG_LABELS[key] || lang
}

export function shouldHighlightLanguage(lang: string): boolean {
  if (!lang) return false
  const key = lang.toLowerCase()
  return key !== 'text' && key !== 'plaintext' && key !== 'plain' && key !== 'txt'
}
