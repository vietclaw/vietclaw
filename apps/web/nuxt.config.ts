import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  ssr: false,
  srcDir: 'app/',
  css: ['~/assets/css/main.css'],
  modules: ['@nuxtjs/color-mode'],
  colorMode: {
    preference: 'light',
    fallback: 'light',
    classSuffix: ''
  },
  vite: {
    plugins: [tailwindcss()]
  },
  nitro: {
    preset: 'static'
  },
  app: {
    buildAssetsDir: 'nuxt',
    head: {
      title: 'VietClaw',
      htmlAttrs: { lang: 'vi' },
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Lightweight personal agent runtime' },
        { name: 'theme-color', content: '#f6f3ee' }
      ],
      link: [
        {
          rel: 'icon',
          type: 'image/svg+xml',
          href: "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'%3E%3Crect width='32' height='32' rx='8' fill='%23b0482e'/%3E%3Cpath d='M9 8c2.5 4.5 2.5 11.5 0 16M16 7c3 5 3 13 0 18M23 8c2.5 4.5 2.5 11.5 0 16' stroke='%23fdfcf9' stroke-width='2.4' stroke-linecap='round' fill='none'/%3E%3C/svg%3E"
        },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=Be+Vietnam+Pro:wght@400;500;600;700&family=Fraunces:opsz,wght@9..144,500;9..144,600&family=IBM+Plex+Mono:wght@400;500&display=swap'
        }
      ]
    }
  },
  compatibilityDate: '2026-06-08'
})
