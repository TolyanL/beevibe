<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { formatTokens } from '@/lib/format'
import AppIcon from '@/components/ui/AppIcon.vue'
import BrandMark from './BrandMark.vue'
import ThemeToggle from './ThemeToggle.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const mobileOpen = ref(false)
const userOpen = ref(false)
const loggingOut = ref(false)
const userMenu = ref<HTMLElement | null>(null)

const nav = [
  { name: 'create', label: 'Создать', icon: 'wand' as const },
  { name: 'catalog', label: 'Каталог', icon: 'grid' as const },
]

watch(
  () => route.fullPath,
  () => {
    mobileOpen.value = false
    userOpen.value = false
  },
)

async function logout(): Promise<void> {
  loggingOut.value = true
  try {
    await auth.logout()
    await router.push({ name: 'home' })
  } finally {
    loggingOut.value = false
  }
}

function onPointerDown(event: PointerEvent): void {
  if (userOpen.value && userMenu.value && !userMenu.value.contains(event.target as Node)) {
    userOpen.value = false
  }
}

function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    userOpen.value = false
    mobileOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('pointerdown', onPointerDown)
  document.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onPointerDown)
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <header class="sticky top-0 z-40 border-b border-line/80 bg-canvas/80 backdrop-blur-xl">
    <div class="app-container flex h-[74px] items-center justify-between gap-4">
      <RouterLink to="/" class="shrink-0" aria-label="BeeVibe — на главную">
        <BrandMark />
      </RouterLink>

      <nav
        class="hidden items-center gap-1 rounded-full border border-line bg-surface/60 p-1 md:flex"
      >
        <RouterLink
          v-for="item in nav"
          :key="item.name"
          :to="{ name: item.name }"
          class="flex items-center gap-2 rounded-full px-4 py-2 text-[13px] font-bold transition"
          :class="
            route.name === item.name ? 'bg-ink text-canvas shadow-sm' : 'text-muted hover:text-ink'
          "
        >
          <AppIcon :name="item.icon" :size="15" />
          {{ item.label }}
        </RouterLink>
        <RouterLink
          v-if="auth.isAuthenticated"
          :to="{ name: 'profile' }"
          class="flex items-center gap-2 rounded-full px-4 py-2 text-[13px] font-bold transition"
          :class="
            route.name === 'profile' ? 'bg-ink text-canvas shadow-sm' : 'text-muted hover:text-ink'
          "
        >
          <AppIcon name="chart" :size="15" />
          Профиль
        </RouterLink>
      </nav>

      <div class="flex items-center gap-2">
        <RouterLink
          v-if="auth.isAuthenticated"
          :to="{ name: 'profile' }"
          class="hidden items-center gap-2 rounded-full border border-line bg-surface/75 py-1.5 pl-2.5 pr-3 text-xs font-bold text-ink transition hover:border-bee-deep sm:flex"
          title="Баланс токенов"
        >
          <span class="grid size-6 place-items-center rounded-full bg-bee text-graphite">
            <AppIcon name="bolt" :size="13" :stroke="2.2" />
          </span>
          <span class="font-mono tabular-nums">{{ formatTokens(auth.balance) }}</span>
        </RouterLink>

        <ThemeToggle />

        <div v-if="auth.isAuthenticated" ref="userMenu" class="relative hidden sm:block">
          <button
            type="button"
            class="flex h-10 max-w-[190px] items-center gap-2 rounded-full border border-line bg-surface/75 px-2 pr-3 text-left transition hover:border-bee-deep"
            :aria-expanded="userOpen"
            aria-haspopup="menu"
            @click="userOpen = !userOpen"
          >
            <span class="grid size-6 place-items-center rounded-full bg-ink text-canvas">
              <AppIcon name="user" :size="13" />
            </span>
            <span class="truncate text-[12px] font-bold">{{ auth.user?.email }}</span>
            <AppIcon name="chevron-down" :size="13" class="text-muted" />
          </button>

          <Transition name="menu">
            <div
              v-if="userOpen"
              class="absolute right-0 top-12 w-56 overflow-hidden rounded-2xl border border-line-strong bg-elevated p-1.5 shadow-lift"
              role="menu"
            >
              <RouterLink
                :to="{ name: 'profile' }"
                class="flex items-center gap-2.5 rounded-xl px-3 py-2.5 text-[13px] font-semibold hover:bg-hover-surface"
                role="menuitem"
              >
                <AppIcon name="chart" :size="16" class="text-muted" />
                Профиль
              </RouterLink>
              <RouterLink
                :to="{ name: 'create' }"
                class="flex items-center gap-2.5 rounded-xl px-3 py-2.5 text-[13px] font-semibold hover:bg-hover-surface"
                role="menuitem"
              >
                <AppIcon name="wand" :size="16" class="text-muted" />
                Новая игра
              </RouterLink>
              <div class="my-1 h-px bg-line" />
              <button
                type="button"
                class="flex w-full items-center gap-2.5 rounded-xl px-3 py-2.5 text-left text-[13px] font-semibold text-bad hover:bg-hover-surface disabled:opacity-50"
                role="menuitem"
                :disabled="loggingOut"
                @click="logout"
              >
                <AppIcon name="logout" :size="16" />
                {{ loggingOut ? 'Выходим…' : 'Выйти' }}
              </button>
            </div>
          </Transition>
        </div>

        <RouterLink
          v-else
          :to="{ name: 'login' }"
          class="hidden rounded-full bg-ink px-4 py-2.5 text-[12px] font-bold text-canvas transition hover:bg-hover-ink sm:inline-flex"
        >
          Войти
        </RouterLink>

        <button
          type="button"
          class="grid size-10 place-items-center rounded-full border border-line bg-surface/70 md:hidden"
          aria-label="Меню"
          :aria-expanded="mobileOpen"
          @click="mobileOpen = !mobileOpen"
        >
          <AppIcon :name="mobileOpen ? 'x' : 'menu'" :size="18" />
        </button>
      </div>
    </div>

    <Transition name="menu">
      <div v-if="mobileOpen" class="border-t border-line bg-canvas/95 px-4 pb-5 pt-3 md:hidden">
        <div class="grid gap-1.5">
          <RouterLink
            v-for="item in nav"
            :key="item.name"
            :to="{ name: item.name }"
            class="flex items-center gap-3 rounded-2xl px-4 py-3 text-sm font-bold"
            :class="
              route.name === item.name
                ? 'bg-bee text-graphite'
                : 'text-muted hover:bg-hover-surface'
            "
          >
            <AppIcon :name="item.icon" :size="18" />
            {{ item.label }}
          </RouterLink>
          <RouterLink
            v-if="auth.isAuthenticated"
            :to="{ name: 'profile' }"
            class="flex items-center gap-3 rounded-2xl px-4 py-3 text-sm font-bold"
            :class="
              route.name === 'profile'
                ? 'bg-bee text-graphite'
                : 'text-muted hover:bg-hover-surface'
            "
          >
            <AppIcon name="chart" :size="18" />
            Профиль
          </RouterLink>
        </div>

        <div
          v-if="auth.isAuthenticated"
          class="mt-4 flex items-center justify-between rounded-2xl border border-line bg-surface px-4 py-3"
        >
          <div class="min-w-0">
            <p class="truncate text-xs font-bold">{{ auth.user?.email }}</p>
            <p class="mt-0.5 font-mono text-[11px] text-muted">
              {{ formatTokens(auth.balance) }} токенов
            </p>
          </div>
          <button
            type="button"
            class="text-xs font-bold text-bad"
            :disabled="loggingOut"
            @click="logout"
          >
            Выйти
          </button>
        </div>
        <RouterLink
          v-else
          :to="{ name: 'login' }"
          class="mt-4 flex items-center justify-center rounded-full bg-bee px-4 py-3 text-sm font-extrabold text-graphite"
        >
          Войти в аккаунт
        </RouterLink>
      </div>
    </Transition>
  </header>
</template>

<style scoped>
.menu-enter-active,
.menu-leave-active {
  transition:
    opacity 0.16s ease,
    transform 0.16s ease;
}

.menu-enter-from,
.menu-leave-to {
  opacity: 0;
  transform: translateY(-5px);
}
</style>
