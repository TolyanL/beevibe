import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/pages/HomePage.vue'),
      meta: { title: 'Игры из одной фразы' },
    },
    {
      path: '/create',
      name: 'create',
      component: () => import('@/pages/GeneratorPage.vue'),
      meta: { title: 'Создать игру', requiresAuth: true },
    },
    {
      path: '/catalog',
      name: 'catalog',
      component: () => import('@/pages/CatalogPage.vue'),
      meta: { title: 'Каталог' },
    },
    {
      path: '/games/:id',
      name: 'game',
      component: () => import('@/pages/GameWorkspacePage.vue'),
      meta: { title: 'Игра', requiresAuth: true },
    },
    {
      path: '/play/:id',
      name: 'play',
      component: () => import('@/pages/PlayPage.vue'),
      meta: { title: 'Играть' },
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('@/pages/ProfilePage.vue'),
      meta: { title: 'Профиль', requiresAuth: true },
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/AuthPage.vue'),
      meta: { title: 'Вход', guestOnly: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/pages/AuthPage.vue'),
      meta: { title: 'Регистрация', guestOnly: true },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/pages/NotFoundPage.vue'),
      meta: { title: 'Не найдено' },
    },
  ],
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) return savedPosition
    if (to.hash) return { el: to.hash, behavior: 'smooth' }
    return { top: 0 }
  },
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.bootstrap()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return {
      name: 'login',
      query: { redirect: to.fullPath },
    }
  }

  if (to.meta.guestOnly && auth.isAuthenticated) {
    return { name: 'profile' }
  }

  return true
})

router.afterEach((to) => {
  const title = typeof to.meta.title === 'string' ? to.meta.title : ''
  document.title = title ? `${title} — BeeVibe` : 'BeeVibe — игры из одной фразы'
})

export default router
