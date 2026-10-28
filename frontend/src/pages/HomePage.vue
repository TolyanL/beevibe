<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getCatalog } from '@/api/catalog'
import type { CatalogCard } from '@/api/types'
import { formatNumber, formatRelative } from '@/lib/format'
import { shortGameTitle } from '@/lib/game'
import { useAuthStore } from '@/stores/auth'
import AppIcon from '@/components/ui/AppIcon.vue'
import GameCard from '@/components/catalog/GameCard.vue'

const auth = useAuthStore()
const fresh = ref<CatalogCard[]>([])
const top = ref<CatalogCard[]>([])
const loading = ref(true)

const lastGame = computed(() => {
  const games = auth.profile?.games ?? []

  return games.find((item) => item.status === 'ready') ?? games[0] ?? null
})

const activeGame = computed(() => {
  const games = auth.profile?.games ?? []
  const busy = [
    'queued',
    'analyzing',
    'planning',
    'coding',
    'assets',
    'reviewing',
    'assembling',
    'checking',
    'patching',
  ]

  return games.find((item) => busy.includes(item.status)) ?? null
})

const steps = [
  {
    icon: 'wand',
    title: 'Расскажите идею',
    text: 'Одно предложение: герой, механика, настроение.',
  },
  {
    icon: 'code',
    title: 'Агенты собирают',
    text: 'Семь этапов: разбор идеи, план, код, графика, ревью, сборка, проверка.',
  },
  {
    icon: 'gamepad',
    title: 'Играйте сразу',
    text: 'Игра открывается в браузере — один файл, ничего устанавливать не нужно.',
  },
  {
    icon: 'refresh',
    title: 'Правьте словами',
    text: 'Скажите, что поменять: новая версия соберётся сама, старые останутся.',
  },
] as const

onMounted(async () => {
  if (auth.isAuthenticated && !auth.profile) void auth.refreshProfile()

  try {
    const [freshPage, topPage] = await Promise.all([
      getCatalog({ sort: 'fresh', limit: 6 }),
      getCatalog({ sort: 'rating', limit: 3 }),
    ])

    fresh.value = freshPage.items ?? []
    top.value = (topPage.items ?? []).filter((card) => (card.rating?.total ?? 0) > 0)
  } catch {
    fresh.value = []
    top.value = []
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="pb-10 pt-8 sm:pt-10">
    <div class="app-container">
      <section
        class="relative overflow-hidden rounded-[1.7rem] border border-line bg-surface/70 p-6 sm:p-9"
      >
        <div class="absolute -right-24 -top-28 size-72 rounded-full bg-bee/15 blur-3xl" />
        <div class="absolute -bottom-32 -left-20 size-72 rounded-full bg-bee/[0.07] blur-3xl" />

        <div
          class="relative grid items-center gap-8 lg:grid-cols-[minmax(0,1.15fr)_minmax(0,.85fr)]"
        >
          <div>
            <div class="eyebrow">BeeVibe · лаборатория игр</div>
            <h1
              class="display-title mt-3 text-[2.6rem] leading-[0.95] sm:text-[3.4rem] lg:text-[3.9rem]"
            >
              Игры из<br />
              <span class="text-bee-deep dark:text-bee">одной фразы</span>
            </h1>
            <p class="mt-4 max-w-xl text-sm leading-6 text-muted sm:text-[15px]">
              Опишите идею одной фразой — агенты придумают правила, напишут код и нарисуют графику.
              Через пару минут можно играть, а правки просить словами.
            </p>

            <div class="mt-6 flex flex-wrap items-center gap-2.5">
              <RouterLink
                :to="{ name: 'create' }"
                class="inline-flex items-center gap-2 rounded-full bg-bee px-6 py-3 text-[13px] font-extrabold text-graphite transition hover:bg-bee-deep"
              >
                <AppIcon name="spark" :size="16" />
                Создать игру
              </RouterLink>
              <RouterLink
                :to="{ name: 'catalog' }"
                class="inline-flex items-center gap-2 rounded-full border border-line-strong px-5 py-3 text-[13px] font-bold transition hover:border-bee-deep hover:text-ink"
              >
                <AppIcon name="grid" :size="15" />
                Смотреть каталог
              </RouterLink>
              <span
                v-if="auth.isAuthenticated"
                class="inline-flex items-center gap-1.5 rounded-full border border-line bg-canvas/50 px-3 py-2 font-mono text-[10px] text-muted"
              >
                <AppIcon name="bolt" :size="12" />
                {{ formatNumber(auth.balance) }} токенов
              </span>
            </div>

            <div
              v-if="activeGame"
              class="mt-4 flex items-center gap-2 rounded-2xl border border-bee-deep/25 bg-bee/[0.06] px-4 py-3 text-xs"
            >
              <span class="size-2 shrink-0 animate-pulse-soft rounded-full bg-bee-deep" />
              <span class="shrink-0 font-bold">Идёт сборка:</span>
              <RouterLink
                :to="{ name: 'game', params: { id: activeGame.slug } }"
                class="min-w-0 truncate text-bee-deep hover-link dark:text-bee"
              >
                {{ shortGameTitle(activeGame.title, activeGame.prompt) }}
              </RouterLink>
            </div>

            <div
              v-else-if="lastGame"
              class="mt-4 flex items-center gap-2 rounded-2xl border border-line bg-canvas/40 px-4 py-3 text-xs"
            >
              <span class="shrink-0 text-faint">Последняя игра:</span>
              <RouterLink
                :to="{ name: 'game', params: { id: lastGame.slug } }"
                class="min-w-0 truncate font-bold hover-link"
              >
                {{ shortGameTitle(lastGame.title, lastGame.prompt) }}
              </RouterLink>
            </div>
          </div>

          <div class="relative hidden lg:block">
            <div class="rounded-[1.4rem] border border-line bg-canvas/60 p-4 shadow-lift">
              <div class="flex items-center justify-between">
                <span
                  class="inline-flex items-center gap-1.5 font-mono text-[9px] uppercase tracking-[0.12em] text-faint"
                >
                  <AppIcon name="bolt" :size="12" />
                  сборка
                </span>
                <span
                  class="rounded-full bg-bee/15 px-2 py-0.5 font-mono text-[9px] font-bold uppercase tracking-[0.1em] text-bee-deep dark:text-bee"
                >
                  живая
                </span>
              </div>
              <div class="mt-3 space-y-2">
                <div
                  v-for="(step, index) in steps"
                  :key="step.title"
                  class="flex items-center gap-3 rounded-xl border border-line bg-surface/70 px-3 py-2.5"
                >
                  <span
                    class="grid size-7 shrink-0 place-items-center rounded-lg bg-bee/15 font-mono text-[10px] font-bold text-bee-deep dark:text-bee"
                  >
                    0{{ index + 1 }}
                  </span>
                  <div class="min-w-0">
                    <p class="text-[12px] font-bold leading-tight">{{ step.title }}</p>
                    <p class="mt-0.5 truncate text-[10px] text-faint">{{ step.text }}</p>
                  </div>
                  <AppIcon :name="step.icon" :size="15" class="ml-auto shrink-0 text-faint" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="mt-8">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <div class="eyebrow">Как это работает</div>
            <h2 class="mt-2 font-display text-xl font-bold sm:text-2xl">
              Как из фразы получается игра
            </h2>
          </div>
          <RouterLink
            :to="{ name: 'create' }"
            class="group inline-flex items-center gap-2 text-xs font-bold text-muted transition hover:text-ink"
          >
            Попробовать
            <AppIcon name="arrow-right" :size="14" class="transition group-hover:translate-x-0.5" />
          </RouterLink>
        </div>
        <div class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div v-for="(step, index) in steps" :key="step.title" class="panel p-4">
            <div class="flex items-center justify-between">
              <span class="mono-label">шаг {{ index + 1 }}</span>
              <AppIcon :name="step.icon" :size="16" class="text-faint" />
            </div>
            <p class="mt-2 text-sm font-bold">{{ step.title }}</p>
            <p class="mt-1 text-[11px] leading-5 text-muted">{{ step.text }}</p>
          </div>
        </div>
      </section>

      <section class="mt-9">
        <div class="flex items-end justify-between gap-4">
          <div>
            <div class="eyebrow">Свежее</div>
            <h2 class="display-title mt-2 text-2xl sm:text-3xl">Только что собрали</h2>
          </div>
          <RouterLink
            :to="{ name: 'catalog' }"
            class="group inline-flex items-center gap-2 text-xs font-bold text-muted transition hover:text-ink"
          >
            Весь каталог
            <AppIcon name="arrow-right" :size="14" class="transition group-hover:translate-x-0.5" />
          </RouterLink>
        </div>

        <div v-if="loading" class="mt-5 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <div
            v-for="index in 6"
            :key="index"
            class="overflow-hidden rounded-[1.35rem] border border-line bg-surface/60"
          >
            <div class="aspect-[16/10] animate-pulse bg-line/50" />
            <div class="space-y-3 p-4">
              <div class="h-4 w-4/5 animate-pulse rounded bg-line/70" />
              <div class="h-3 w-2/5 animate-pulse rounded bg-line/50" />
            </div>
          </div>
        </div>

        <div v-else-if="fresh.length" class="mt-5 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <GameCard v-for="card in fresh" :key="card.id" :card="card" />
        </div>

        <div v-else class="panel mt-5 flex flex-col items-center gap-3 px-6 py-10 text-center">
          <span
            class="grid size-11 place-items-center rounded-2xl bg-bee/15 text-bee-deep dark:text-bee"
          >
            <AppIcon name="spark" :size="20" />
          </span>
          <p class="text-sm font-bold">Каталог пока пуст</p>
          <p class="max-w-sm text-xs leading-5 text-muted">
            Пока никто не опубликовал игру. Соберите первую — она появится здесь, как только вы
            откроете её для всех.
          </p>
          <RouterLink
            :to="{ name: 'create' }"
            class="mt-1 inline-flex items-center gap-2 rounded-full bg-bee px-5 py-2.5 text-xs font-extrabold text-graphite"
          >
            <AppIcon name="wand" :size="14" />
            Создать игру
          </RouterLink>
        </div>
      </section>

      <section v-if="top.length" class="mt-9">
        <div class="flex items-end justify-between gap-4">
          <div>
            <div class="eyebrow">Рейтинг</div>
            <h2 class="display-title mt-2 text-2xl sm:text-3xl">Выбор игроков</h2>
          </div>
          <RouterLink
            :to="{ name: 'catalog', query: { sort: 'rating' } }"
            class="group inline-flex items-center gap-2 text-xs font-bold text-muted transition hover:text-ink"
          >
            По рейтингу
            <AppIcon name="arrow-right" :size="14" class="transition group-hover:translate-x-0.5" />
          </RouterLink>
        </div>
        <div class="mt-5 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <GameCard v-for="card in top" :key="card.id" :card="card" />
        </div>
      </section>

      <section
        v-if="fresh.length"
        class="panel mt-9 flex flex-col items-start gap-4 p-6 sm:flex-row sm:items-center sm:justify-between"
      >
        <div>
          <p class="mono-label">Своя игра</p>
          <h2 class="mt-1 font-display text-xl font-bold">Расскажите, во что хотите сыграть</h2>
          <p class="mt-1 text-xs text-muted">
            Первая игра — за наш счёт: при регистрации даём {{ formatNumber(1_000_000) }} токенов.
          </p>
        </div>
        <RouterLink
          :to="{ name: 'create' }"
          class="inline-flex shrink-0 items-center gap-2 rounded-full bg-bee px-5 py-2.5 text-xs font-extrabold text-graphite transition hover:bg-bee-deep"
        >
          <AppIcon name="wand" :size="14" />
          Создать игру
        </RouterLink>
      </section>

      <p
        v-if="fresh.length"
        class="mt-6 text-center font-mono text-[9px] uppercase tracking-[0.13em] text-faint"
      >
        последняя сборка · {{ formatRelative(fresh[0]!.created_at) }}
      </p>
    </div>
  </div>
</template>
