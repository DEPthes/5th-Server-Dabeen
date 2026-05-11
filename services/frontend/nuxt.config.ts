export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',

  runtimeConfig: {
    public: {
      // nginx 리버스 프록시를 통해 /api 로 접근
      apiBase: process.env.NUXT_PUBLIC_API_BASE ?? '/api',
    },
  },

  app: {
    head: {
      title: 'TODO App - Docker Swarm 실습',
      meta: [{ charset: 'utf-8' }, { name: 'viewport', content: 'width=device-width, initial-scale=1' }],
    },
  },

  ssr: true,
})
