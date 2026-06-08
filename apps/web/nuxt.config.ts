import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  ssr: false,
  srcDir: 'app/',
  css: ['~/assets/css/main.css'],
  modules: ['@nuxtjs/color-mode'],
  colorMode: {
    preference: 'dark',
    fallback: 'dark',
    classSuffix: ''
  },
  vite: {
    plugins: [tailwindcss()]
  },
  nitro: {
    preset: 'static'
  },
  app: {
    head: {
      title: 'VietClaw',
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Lightweight personal agent runtime' }
      ]
    }
  },
  compatibilityDate: '2026-06-08'
})

