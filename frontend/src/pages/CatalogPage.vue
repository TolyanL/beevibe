<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import InputText from 'openvue/inputtext'
import type { CatalogSort, Genre } from '@/api/types'
import { genreLabel } from '@/lib/game'
import { useCatalogStore } from '@/stores/catalog'
import AppIcon from '@/components/ui/AppIcon.vue'
import GameCard from '@/components/catalog/GameCard.vue'
import EmptyState from '@/components/ui/EmptyState.vue'

const route = useRoute()
const router = useRouter()
const catalog = useCatalogStore()
const search = ref('')
const sentinel = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | null = null

const sortOptions: { value: CatalogSort; label: string; icon: 'clock' | 'star' | 'flame' }[] = [
  { value: 'fresh', label: 'Свежие', icon: 'clock' },
  { value: 'rating', label: 'По рейтингу', icon: 'star' },
  { value: 'hours', label: 'По времени', icon: 'flame' },
]

const genres: { value: Genre | ''; label: string }[] = [
  { value: '', label: 'Все жанры' },
  { value: 'arcade', label: 'Аркада' },
  { value: 'snake', label: 'Змейка' },
  { value: 'clicker', label: 'Кликер' },
  { value: 'platformer', label: 'Платформер' },
  { value: 'puzzle', label: 'Головоломка' },
  { value: 'custom', label: 'Своя механика' },
]

const filtered = computed(() => {
  const query = search.value.trim().toLowerCase()
  if (!query) return catalog.items

  return catalog.items.filter((card) => {
    const haystack = [card.prompt, genreLabel(card.genre), card.style].join(' ').toLowerCase()
    return haystack.includes(query)
  })
})

const hasFilters = computed(() => Boolean(search.value) || Boolean(catalog.genre) || catalog.sort !== 'fresh')

function syncQuery(): void {
  const query: Record<string, string> = {}
  if (catalog.sort !== 'fresh') query.sort = catalog.sort
  if (catalog.genre) query.genre = catalog.genre
  void router.replace({ query })
}

function chooseSort(sort: CatalogSort): void {
  catalog.sort = sort
  syncQuery()
  void catalog.load(true)
}

function chooseGenre(genre: Genre | ''): void {
  catalog.genre = genre
  syncQuery()
  void catalog.load(true)
}

function reset(): void {
  search.value = ''
  catalog.sort = 'fresh'
  catalog.genre = ''
  syncQuery()
  void catalog.load(true)
}

watch(
  () => catalog.hasMore,
  (hasMore) => {
    if (!observer) return
    if (hasMore && sentinel.value) observer.observe(sentinel.value)
    else if (sentinel.value) observer.unobserve(sentinel.value)
  },
)

onMounted(() => {
  const routeSort = route.query.sort
  const routeGenre = route.query.genre
  if (routeSort === 'rating' || routeSort === 'hours' || routeSort === 'fresh') catalog.sort = routeSort
  if (typeof routeGenre === 'string' && genres.some((item) => item.value === routeGenre)) {
    catalog.genre = routeGenre as Genre
  }

  void catalog.load(true)

  observer = new IntersectionObserver(
    (entries) => {
      if (entries.some((entry) => entry.isIntersecting) && catalog.hasMore && !catalog.loadingMore) {
        void catalog.loadMore()
      }
    },
    { rootMargin: '320px' },
  )

  if (sentinel.value) observer.observe(sentinel.value)
})

onBeforeUnmount(() => {
  observer?.disconnect()
  observer = null
})
</script>

<template>
  <div class="pb-8 pt-10 sm:pt-14">
    <div class="app-container">
      <section class="relative overflow-hidden rounded-[1.6rem] border border-line bg-surface/70 px-5 py-8 sm:px-8 sm:py-10">
        <div class="absolute -right-20 -top-24 size-64 rounded-full bg-bee/15 blur-3xl" />
        <div class="relative flex flex-col gap-7 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <div class="eyebrow">Публичная витрина</div>
            <h1 class="display-title mt-4 max-w-2xl text-4xl sm:text-5xl lg:text-6xl">
              Игры, собранные <span class="text-bee-deep dark:text-bee">сообществом</span>
            </h1>
            <p class="mt-4 max-w-xl text-sm leading-6 text-muted">
              Запускайте в один клик, оценивайте по пятибалльной шкале и смотрите, сколько времени
              игроки провели внутри. Каждая игра — один HTML-файл.
            </p>
          </div>

          <div class="flex shrink-0 items-center gap-2.5">
            <RouterLink
              to="/"
              class="inline-flex h-11 items-center gap-2 rounded-full bg-bee px-5 text-xs font-extrabold text-graphite transition hover:bg-bee-deep"
            >
              <AppIcon name="plus" :size="15" :stroke="2.4" />
              Создать свою
            </RouterLink>
          </div>
        </div>
      </section>

      <section class="mt-7">
        <div class="panel p-3 sm:p-4">
          <div class="flex flex-col gap-3 lg:flex-row lg:items-center">
            <div class="relative flex-1">
              <AppIcon name="search" :size="17" class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-faint" />
              <InputText
                v-model="search"
                class="h-11 w-full pl-11"
                placeholder="Поиск по промпту: змейка, неон, кликер…"
                aria-label="Поиск по каталогу"
              />
            </div>

            <div class="flex flex-wrap items-center gap-1 rounded-full border border-line bg-canvas/60 p-1">
              <button
                v-for="option in sortOptions"
                :key="option.value"
                type="button"
                class="inline-flex items-center gap-1.5 rounded-full px-3.5 py-2 text-[11px] font-bold transition"
                :class="catalog.sort === option.value ? 'bg-ink text-canvas shadow-sm' : 'text-muted hover:text-ink'"
                @click="chooseSort(option.value)"
              >
                <AppIcon :name="option.icon" :size="13" />
                {{ option.label }}
              </button>
            </div>
          </div>

          <div class="mt-3 flex flex-wrap items-center gap-1.5 border-t border-line pt-3">
            <button
              v-for="item in genres"
              :key="item.value"
              type="button"
              class="rounded-full px-3 py-1.5 text-[11px] font-bold transition"
              :class="catalog.genre === item.value ? 'bg-bee text-graphite' : 'text-muted hover:bg-canvas hover:text-ink'"
              @click="chooseGenre(item.value)"
            >
              {{ item.label }}
            </button>
            <button
              v-if="hasFilters"
              type="button"
              class="ml-auto inline-flex items-center gap-1 rounded-full px-2.5 py-1.5 text-[11px] font-bold text-faint transition hover:text-ink"
              @click="reset"
            >
              <AppIcon name="x" :size="12" />
              Сбросить
            </button>
          </div>
        </div>
      </section>

      <div class="mt-6 flex items-center justify-between gap-4">
        <p class="text-sm text-muted">
          <span v-if="catalog.loading" class="inline-flex items-center gap-2">
            <span class="size-3 animate-spin rounded-full border-2 border-line border-t-bee-deep" />
            Загружаем игры…
          </span>
          <span v-else>
            <b class="font-bold text-ink">{{ filtered.length }}</b>
            {{ filtered.length === 1 ? 'игра' : filtered.length < 5 ? 'игры' : 'игр' }}
            <span v-if="search" class="text-faint">по запросу «{{ search }}»</span>
          </span>
        </p>
        <p class="hidden font-mono text-[9px] uppercase tracking-[0.13em] text-faint sm:block">
          sort: {{ catalog.sort }}
        </p>
      </div>

      <div v-if="catalog.loading && !catalog.items.length" class="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="index in 6" :key="index" class="overflow-hidden rounded-[1.35rem] border border-line bg-surface/60">
          <div class="aspect-[16/10] animate-pulse bg-line/50" />
          <div class="space-y-3 p-4">
            <div class="h-4 w-4/5 animate-pulse rounded bg-line/70" />
            <div class="h-3 w-2/5 animate-pulse rounded bg-line/50" />
          </div>
        </div>
      </div>

      <div v-else-if="catalog.error && !catalog.items.length" class="mt-4">
        <EmptyState icon="alert" title="Каталог не загрузился" :text="catalog.error">
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-full bg-bee px-5 py-2.5 text-xs font-extrabold text-graphite"
            @click="catalog.load(true)"
          >
            <AppIcon name="refresh" :size="14" />
            Повторить
          </button>
        </EmptyState>
      </div>

      <div v-else-if="filtered.length" class="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <GameCard v-for="card in filtered" :key="card.id" :card="card" />
      </div>

      <div v-else class="mt-4">
        <EmptyState
          icon="search"
          title="Ничего не нашлось"
          text="Попробуйте другой запрос, жанр или сбросьте фильтры. Новые игры появляются после генерации."
        >
          <button
            type="button"
            class="inline-flex items-center gap-2 rounded-full border border-line-strong px-5 py-2.5 text-xs font-bold transition hover:bg-surface"
            @click="reset"
          >
            <AppIcon name="refresh" :size="14" />
            Сбросить фильтры
          </button>
        </EmptyState>
      </div>

      <div ref="sentinel" class="h-4" />

      <div v-if="catalog.loadingMore" class="mt-5 flex items-center justify-center gap-2 text-xs text-muted">
        <span class="size-3 animate-spin rounded-full border-2 border-line border-t-bee-deep" />
        Ещё немного…
      </div>

      <div
        v-if="catalog.error && catalog.items.length"
        class="mt-5 flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-red-500/20 bg-red-500/10 px-4 py-3 text-xs text-red-500"
      >
        <span>{{ catalog.error }}</span>
        <button
          type="button"
          class="rounded-full border border-red-500/30 px-3 py-1.5 font-bold transition hover:bg-red-500/10"
          @click="catalog.load(false)"
        >
          Повторить загрузку
        </button>
      </div>
    </div>
  </div>
</template>
