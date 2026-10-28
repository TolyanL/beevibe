<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import Rating from 'openvue/rating'
import { useToast } from 'openvue/usetoast'
import { getCatalog, getGamePublicCard } from '@/api/catalog'
import { messageFromError } from '@/api/client'
import { rateGame } from '@/api/games'
import type { CatalogCard, GamePublicCard } from '@/api/types'
import { formatDate, formatHours, plural } from '@/lib/format'
import {
  gameDescription as gameDescriptionOf,
  genreLabel,
  shortGameTitle,
  styleLabel,
} from '@/lib/game'
import { ratingTone } from '@/lib/rating'
import { usePlaytime } from '@/composables/usePlaytime'
import { useAuthStore } from '@/stores/auth'
import AppIcon from '@/components/ui/AppIcon.vue'
import GameCard from '@/components/catalog/GameCard.vue'
import GameFrame from '@/components/game/GameFrame.vue'
import RatingBadge from '@/components/ui/RatingBadge.vue'

const route = useRoute()
const toast = useToast()
const auth = useAuthStore()
const id = computed(() => String(route.params.id ?? ''))
const card = ref<GamePublicCard | null>(null)
const related = ref<CatalogCard[]>([])
const loading = ref(true)
const frameLoaded = ref(false)
const vote = ref(0)
const ratingKey = ref(0)
const votePending = ref(false)
const voteError = ref('')

const session = usePlaytime(
  () => id.value,
  () => frameLoaded.value,
)
const sessionSeconds = session.seconds
const more = computed(() => related.value.slice(0, 3))
const title = computed(() =>
  shortGameTitle(card.value?.title, card.value?.prompt ?? 'Игра BeeVibe'),
)
const description = computed(() =>
  gameDescriptionOf(card.value?.description, card.value?.prompt ?? ''),
)
const author = computed(() => card.value?.author ?? '')
const genre = computed(() => genreLabel(card.value?.genre ?? '', card.value?.custom_genre))
const styleName = computed(() => styleLabel(card.value?.style ?? '', card.value?.custom_style))
const playtime = computed(() => formatHours(card.value?.playtime?.hours ?? 0))
const canVote = computed(() => Boolean(card.value) && auth.isAuthenticated && !card.value?.is_owner)
const ratingInputColor = computed(() => (vote.value ? ratingTone(vote.value).hex : '#e6b800'))
const loginLink = computed(() => ({ name: 'login', query: { redirect: `/play/${id.value}` } }))

function sessionLabel(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
}

async function rate(value: number): Promise<void> {
  if (!card.value) return

  votePending.value = true
  voteError.value = ''

  try {
    const response = await rateGame(card.value.id, value)

    vote.value = response.rating
    card.value.my_rating = response.rating
    card.value.rating = response.aggregate
    toast.add({
      severity: 'success',
      summary: 'Оценка сохранена',
      detail: 'Спасибо — рейтинг обновлён.',
      life: 2200,
    })
  } catch (err) {
    vote.value = card.value.my_rating
    voteError.value = messageFromError(err)
  } finally {
    votePending.value = false
  }
}

function onRatingInput(value: number | null): void {
  if (!value) {
    // NOTE: OpenVue Rating clears the active value on a repeated click; remount restores it.
    ratingKey.value++
    return
  }

  vote.value = value
  void rate(value)
}

onMounted(async () => {
  try {
    card.value = await getGamePublicCard(id.value)
    vote.value = card.value.my_rating ?? 0
  } catch {
    card.value = null
  } finally {
    loading.value = false
  }

  try {
    const response = await getCatalog({ limit: 3, sort: 'fresh' })

    related.value = (response.items ?? []).filter((item) => item.id !== id.value)
  } catch {
    related.value = []
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
          <span class="font-mono text-[9px] uppercase tracking-[0.13em] text-faint"
            >public play</span
          >
        </div>
      </div>

      <header class="mb-5 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <span
              v-if="genre"
              class="rounded-full border border-line bg-surface/70 px-2.5 py-1 font-mono text-[9px] font-bold uppercase tracking-[0.09em] text-muted"
            >
              {{ genre }}
            </span>
            <span
              v-if="styleName"
              class="rounded-full border border-line bg-surface/70 px-2.5 py-1 font-mono text-[9px] font-bold uppercase tracking-[0.09em] text-muted"
            >
              {{ styleName }}
            </span>
            <RatingBadge
              v-if="card"
              :average="card.rating?.average ?? 0"
              :total="card.rating?.total ?? 0"
            />
            <span v-if="author" class="font-mono text-[9px] uppercase tracking-[0.1em] text-faint">
              {{ author }}
            </span>
            <span v-if="card" class="font-mono text-[9px] uppercase tracking-[0.1em] text-faint">
              {{ playtime }} в игре
            </span>
          </div>

          <h1
            class="mt-3 max-w-4xl font-display text-2xl font-bold leading-tight tracking-[-0.04em] sm:text-3xl lg:text-[2.2rem]"
          >
            {{ title }}
          </h1>
        </div>

        <div class="flex shrink-0 items-center gap-2">
          <RouterLink
            :to="{ name: 'create' }"
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

      <div class="mt-5 grid gap-5 lg:grid-cols-[minmax(0,1.4fr)_minmax(0,.6fr)]">
        <section class="panel p-5 sm:p-6">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="mono-label">Об игре</p>
              <h2 class="mt-1 font-display text-xl font-bold">Из промпта — в playable</h2>
            </div>
            <AppIcon name="gamepad" :size="20" class="text-muted" />
          </div>

          <p class="mt-4 text-[14px] leading-7 text-muted">
            {{
              description ||
              'Игра сгенерирована BeeVibe и доступна по публичной ссылке. Запускайте прямо в браузере, оценка и история — в каталоге.'
            }}
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
          <div
            v-else
            class="mt-4 rounded-xl border border-line bg-canvas/35 px-4 py-3 text-xs text-muted"
          >
            Оценки доступны из каталога после загрузки карточки игры.
          </div>

          <div v-if="card && !card.is_owner" class="mt-5 border-t border-line pt-4">
            <template v-if="canVote">
              <p class="mono-label mb-2">Ваша оценка</p>
              <div class="flex items-center gap-3">
                <div class="rating-choice" :style="{ '--rating-color': ratingInputColor }">
                  <Rating
                    :key="ratingKey"
                    :model-value="vote"
                    :disabled="votePending"
                    @update:model-value="onRatingInput"
                  />
                </div>
                <span v-if="card.my_rating" class="font-mono text-[10px] text-muted"
                  >{{ card.my_rating }} / 5</span
                >
              </div>
              <p v-if="voteError" class="mt-2 text-[10px] text-red-500">{{ voteError }}</p>
              <p class="mt-2 text-[10px] leading-4 text-faint">
                Оценку можно изменить в любой момент.
              </p>
            </template>

            <template v-else-if="!auth.isAuthenticated">
              <p class="text-[11px] leading-5 text-faint">
                <RouterLink :to="loginLink" class="font-bold text-bee-deep hover-link dark:text-bee"
                  >Войдите</RouterLink
                >, чтобы поставить оценку — она попадёт в рейтинг каталога.
              </p>
            </template>
          </div>

          <p
            v-else-if="card?.is_owner"
            class="mt-4 border-t border-line pt-3 text-[11px] leading-5 text-faint"
          >
            Это ваша игра: оценивать её могут только другие игроки, поэтому здесь только статистика.
          </p>

          <p v-if="!card" class="mt-4 text-[11px] leading-5 text-faint">
            Оценка хранится в аккаунте и меняет средний рейтинг в каталоге. Пять — восторг, один —
            совсем не зашло.
          </p>
        </aside>
      </div>

      <section v-if="!loading && more.length" class="mt-10">
        <div class="flex items-end justify-between gap-4">
          <div>
            <div class="eyebrow">Ещё играть</div>
            <h2 class="display-title mt-2 text-2xl sm:text-3xl">Другие игры</h2>
          </div>
          <RouterLink
            to="/catalog"
            class="group inline-flex items-center gap-2 text-xs font-bold text-muted transition hover:text-ink"
          >
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

<style scoped>
.rating-choice :deep(.p-rating-option-active),
.rating-choice :deep(.p-rating-option-active .p-rating-icon) {
  color: var(--rating-color);
}
</style>
