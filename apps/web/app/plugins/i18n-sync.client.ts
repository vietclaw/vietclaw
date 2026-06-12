export default defineNuxtPlugin(() => {
  const { lang, setLanguage } = useI18n()
  const { config } = useSettings()

  watch(
    () => config.value?.agent.language,
    (next) => {
      if (next === 'vi' || next === 'en') setLanguage(next)
    },
    { immediate: true },
  )

  watch(
    lang,
    (next) => {
      useHead({ htmlAttrs: { lang: next } })
    },
    { immediate: true },
  )
})
