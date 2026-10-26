<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { getCatalog } from '@/api/catalog'
import type { CatalogCard } from '@/api/types'
import { formatDate, formatHours, plural } from '@/lib/format'
import { genreLabel, styleLabel } from '@/lib/game'
import { usePlaytime } from '@/composables/usePlaytime'
import AppIcon from '@/components/ui/AppIcon.vue'
import GameCard from '@/components/catalog/GameCard.vue'
import GameFrame from '@/components/game/GameFrame.vue'
import RatingBadge from '@/components/ui/RatingBadge.vue'

const route = useRoute()
const id = computed(() => String(route.params.id ?? ''))
const card = ref<CatalogCard | null>(null)
const items = ref<CatalogCard[]>([])
const loading = ref(true)
const frameLoaded = ref(false)

const session = usePlaytime(() => id.value, () => frameLoaded.value)
const sessionSeconds = session.seconds
const more = computed(() => items.value.filter((item) => item.id !== id.value).slice(0, 3))
const title = computed(() => card.value?.prompt ?? 'Игра BeeVibe')
const genre = computed(() => genreLabel(card.value?.genre ?? ''))
const styleName = computed(() => styleLabel(card.value?.style ?? ''))
const playtime = computed(() => formatHours(card.value?.playtime?.hours ?? 0))

function sessionLabel(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
}

onMounted(async () => {
  try {
    const response = await getCatalog({ limit: 100, sort: 'fresh' })
    items.value = response.items ?? []
    card.value = items.value.find((item) => item.id === id.value) ?? null
  } catch {
    card.value = null
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="pb-8 pt-6 sm:pt-8">
    <div class="app-container">
      <div class="mb-5 flex flex-wrap items-center justify-between gap-3">
        <RouterLink
          to="/catalog"
          class="inline-flex items-center gap-1.5 text-[11px] font-bold text-muted transition hover:text-ink"
        >
          <AppIcon name="back" :size="14" />
          В каталог
        </RouterLink>
        <div class="flex items-center gap-2">
          <span
            v-if="frameLoaded && sessionSeconds > 0"
            class="inline-flex items-center gap-1.5 rounded-full border border-line bg-surface/70 px-2.5 py-1 font-mono text-[9px] text-muted"
          >
            <AppIcon name="clock" :size="12" />
            {{ sessionLabel(sessionSeconds) }}
          </span>
          <span class="font-mono text-[9px] uppercase tracking-[0.13em] text-faint">public play</span>
        </div>
      </div>

      <header class="mb-5 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <span class="rounded-full border border-line bg-surface/70 px-2.5 py-1 font-mono text-[9px] font-bold uppercase tracking-[0.09em] text-muted">
              {{ genre }}
            </span>
            <span class="rounded-full border border-line bg-surface/70 px-2.5 py-1 font-mono text-[9px] font-bold uppercase tracking-[0.09em] text-muted">
              {{ styleName }}
            </span>
            <RatingBadge
              v-if="card"
              :average="card.rating?.average ?? 0"
              :total="card.rating?.total ?? 0"
            />
            <span v-if="card" class="font-mono text-[9px] uppercase tracking-[0.1em] text-faint">
              {{ playtime }} в игре
            </span>
          </div>

          <h1 class="mt-3 max-w-4xl font-display text-2xl font-bold leading-tight tracking-[-0.04em] sm:text-3xl lg:text-[2.2rem]">
            {{ title }}
          </h1>
        </div>

        <div class="flex shrink-0 items-center gap-2">
          <RouterLink
            to="/"
            class="inline-flex h-10 items-center gap-2 rounded-full bg-bee px-4 text-[11px] font-extrabold text-graphite transition hover:bg-bee-deep"
          >
            <AppIcon name="spark" :size="14" />
            Сделать свою
          </RouterLink>
        </div>
      </header>

      <div class="animate-rise">
        <GameFrame :game-id="id" status="ready" :version-token="1" @loaded="frameLoaded = true" />
      </div>

      <div class="mt-5 grid gap-5 lg:grid-cols-[1.4fr_.6fr]">
        <section class="panel p-5 sm:p-6">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="mono-label">Об игре</p>
              <h2 class="mt-1 font-display text-xl font-bold">Из промпта — в playable</h2>
            </div>
            <AppIcon name="gamepad" :size="20" class="text-muted" />
          </div>

          <p class="mt-4 text-[14px] leading-7 text-muted">
            {{ card?.prompt ?? 'Игра сгенерирована BeeVibe и доступна по публичной ссылке. Запускайте прямо в браузере, оценка и история — в каталоге.' }}
          </p>

          <div class="mt-5 grid grid-cols-2 gap-2 sm:grid-cols-4">
            <div class="rounded-xl border border-line bg-canvas/35 p-3">
              <p class="mono-label">Жанр</p>
              <p class="mt-1.5 text-xs font-bold">{{ genre }}</p>
            </div>
            <div class="rounded-xl border border-line bg-canvas/35 p-3">
              <p class="mono-label">Стиль</p>
              <p class="mt-1.5 text-xs font-bold">{{ styleName }}</p>
            </div>
            <div class="rounded-xl border border-line bg-canvas/35 p-3">
              <p class="mono-label">Время</p>
              <p class="mt-1.5 text-xs font-bold">{{ playtime }}</p>
            </div>
            <div class="rounded-xl border border-line bg-canvas/35 p-3">
              <p class="mono-label">Дата</p>
              <p class="mt-1.5 text-xs font-bold">{{ card ? formatDate(card.created_at) : '—' }}</p>
            </div>
          </div>
        </section>

        <aside class="panel p-5 sm:p-6">
          <p class="mono-label">Оцените игру</p>
          <div v-if="card" class="mt-4 flex items-end gap-3">
            <p class="font-display text-4xl font-bold leading-none">
              {{ card.rating.total ? card.rating.average.toFixed(1).replace('.', ',') : '—' }}
            </p>
            <div class="pb-1">
              <RatingBadge :average="card.rating.average" :total="card.rating.total" size="sm" />
              <p class="mt-1 font-mono text-[9px] uppercase tracking-[0.09em] text-faint">
                {{ card.rating.total }}
                {{ plural(card.rating.total, 'оценка', 'оценки', 'оценок') }}
              </p>
            </div>
          </div>
          <div v-else class="mt-4 rounded-xl border border-line bg-canvas/35 px-4 py-3 text-xs text-muted">
            Оценки доступны из каталога после загрузки карточки игры.
          </div>

          <p class="mt-4 text-[11px] leading-5 text-faint">
            Оценка хранится в аккаунте и меняет средний рейтинг в каталоге. Пять — восторг,
            один — совсем не зашло.
          </p>
        </aside>
      </div>

      <section v-if="!loading && more.length" class="mt-10">
        <div class="flex items-end justify-between gap-4">
          <div>
            <div class="eyebrow">Ещё играть</div>
            <h2 class="display-title mt-2 text-2xl sm:text-3xl">Другие игры</h2>
          </div>
          <RouterLink to="/catalog" class="group inline-flex items-center gap-2 text-xs font-bold text-muted transition hover:text-ink">
            Весь каталог
            <AppIcon name="arrow-right" :size="14" class="transition group-hover:translate-x-0.5" />
          </RouterLink>
        </div>
        <div class="mt-5 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <GameCard v-for="item in more" :key="item.id" :card="item" />
        </div>
      </section>
    </div>
  </div>
</template>
