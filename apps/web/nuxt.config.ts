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
        { name: 'theme-color', content: '#fafaf8' }
      ],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500&family=Plus+Jakarta+Sans:wght@400;500;600;700&display=swap'
        }
      ]
    }
  },
  compatibilityDate: '2026-06-08'
})
