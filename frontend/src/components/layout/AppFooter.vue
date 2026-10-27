<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import BrandMark from './BrandMark.vue'

const auth = useAuthStore()
const router = useRouter()
const loggingOut = ref(false)

async function logout(): Promise<void> {
  loggingOut.value = true
  try {
    await auth.logout()
    await router.push({ name: 'create' })
  } finally {
    loggingOut.value = false
  }
}
</script>

<template>
  <footer class="mt-16 border-t border-line/70 pb-8 pt-10">
    <div class="app-container">
      <div class="flex flex-col gap-8 md:flex-row md:items-start md:justify-between">
        <div class="max-w-sm">
          <BrandMark />
          <p class="mt-4 text-sm leading-6 text-muted">
            Промпт превращается в self-contained HTML5-игру: конвейер агентов, playable-превью,
            правки и каталог.
          </p>
        </div>

        <div class="grid grid-cols-2 gap-10 text-sm">
          <div>
            <p class="mono-label">Сервис</p>
            <div class="mt-3 grid gap-2.5 text-muted">
              <RouterLink :to="{ name: 'create' }" class="transition hover:text-ink">Создать игру</RouterLink>
              <RouterLink to="/catalog" class="transition hover:text-ink">Каталог</RouterLink>
              <RouterLink to="/profile" class="transition hover:text-ink">Профиль</RouterLink>
            </div>
          </div>
          <div>
            <p class="mono-label">Аккаунт</p>
            <div class="mt-3 grid gap-2.5 text-muted">
              <template v-if="auth.isAuthenticated">
                <RouterLink to="/profile" class="transition hover:text-ink">Мои игры</RouterLink>
                <button
                  type="button"
                  class="text-left transition hover:text-ink disabled:opacity-50"
                  :disabled="loggingOut"
                  @click="logout"
                >
                  {{ loggingOut ? 'Выходим…' : 'Выйти' }}
                </button>
              </template>
              <template v-else>
                <RouterLink to="/login" class="transition hover:text-ink">Войти</RouterLink>
                <RouterLink to="/register" class="transition hover:text-ink">Регистрация</RouterLink>
              </template>
            </div>
          </div>
        </div>
      </div>

      <div class="mt-10 flex flex-col gap-2 border-t border-line/60 pt-5 text-[11px] text-faint sm:flex-row sm:items-center sm:justify-between">
        <p>BeeVibe · игры, которые собираются из идеи</p>
        <p class="font-mono uppercase tracking-[0.13em]">Vue · OpenVue · Tailwind</p>
      </div>
    </div>
  </footer>
</template>
