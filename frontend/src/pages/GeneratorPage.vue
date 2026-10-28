<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Button from 'openvue/button'
import Textarea from 'openvue/textarea'
import Select from 'openvue/select'
import InputText from 'openvue/inputtext'
import { useToast } from 'openvue/usetoast'
import type { GameStyle, Genre, MapSize } from '@/api/types'
import { messageFromError } from '@/api/client'
import { formatNumber, formatTokens } from '@/lib/format'
import { genreLabel, shortGameTitle, styleLabel, stylePalette } from '@/lib/game'
import { useAuthStore } from '@/stores/auth'
import { useGameStore } from '@/stores/game'
import AppIcon from '@/components/ui/AppIcon.vue'

const router = useRouter()
const toast = useToast()
const auth = useAuthStore()
const game = useGameStore()

const prompt = ref('')
const genre = ref<Genre>('arcade')
const style = ref<GameStyle>('neon')
const customGenre = ref('')
const customStyle = ref('')
const title = ref('')
const description = ref('')
const mapSize = ref<MapSize>('medium')
const tempo = ref(6)
const advanced = ref(false)
const submitting = ref(false)
const error = ref('')

const maxPrompt = 2000
const currentPrice = computed(() => auth.prices.generate)
const previewTitle = computed(() => shortGameTitle(title.value, prompt.value))
const previewDescription = computed(
  () =>
    description.value.trim() ||
    prompt.value.trim() ||
    'Пара слов о том, во что играть, — её видно в каталоге',
)
const previewGenre = computed(() => genreLabel(genre.value, customGenre.value))
const previewStyle = computed(() => styleLabel(style.value, customStyle.value))
const previewCover = computed(() => {
  const palette = stylePalette(style.value)

  return { background: `linear-gradient(135deg, ${palette[0]}, ${palette[1]})` }
})
const activeGame = computed(() => {
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
  const games = auth.profile?.games ?? []

  return games.find((item) => busy.includes(item.status)) ?? null
})
const enoughTokens = computed(() => !auth.isAuthenticated || auth.balance >= currentPrice.value)
const customGenreReady = computed(
  () => genre.value !== 'custom' || customGenre.value.trim().length >= 2,
)
const customStyleReady = computed(
  () => style.value !== 'custom' || customStyle.value.trim().length >= 2,
)
const canSubmit = computed(
  () =>
    prompt.value.trim().length >= 3 &&
    prompt.value.length <= maxPrompt &&
    !submitting.value &&
    customGenreReady.value &&
    customStyleReady.value,
)
const examples = [
  {
    label: 'Неоновая змейка',
    prompt:
      'Неоновая змейка в киберпанк-городе: чем длиннее хвост, тем быстрее темп, на карте появляются стены.',
    genre: 'snake' as Genre,
    style: 'neon' as GameStyle,
    map: 'medium' as MapSize,
    tempo: 8,
  },
  {
    label: 'Пиксельный кликер',
    prompt:
      'Пиксельный кликер про космическую пекарню: печенье заменяют астероиды, есть апгрейды и редкие бонусы.',
    genre: 'clicker' as Genre,
    style: 'pixel' as GameStyle,
    map: 'small' as MapSize,
    tempo: 6,
  },
  {
    label: 'Платформер-побег',
    prompt:
      'Мультяшный платформер: лягушка убегает от дождя, прыгает по листьям и собирает светлячков.',
    genre: 'platformer' as Genre,
    style: 'cartoon' as GameStyle,
    map: 'large' as MapSize,
    tempo: 7,
  },
  {
    label: 'Головоломка',
    prompt:
      'Минималистичная головоломка с перемещением кристаллов: нужно собрать три одинаковых цвета в ряд за 30 ходов.',
    genre: 'puzzle' as Genre,
    style: 'minimal' as GameStyle,
    map: 'small' as MapSize,
    tempo: 4,
  },
]

const genreOptions = [
  { label: 'Аркада', value: 'arcade' },
  { label: 'Змейка', value: 'snake' },
  { label: 'Кликер', value: 'clicker' },
  { label: 'Платформер', value: 'platformer' },
  { label: 'Головоломка', value: 'puzzle' },
  { label: 'Своя механика', value: 'custom' },
]

const styleOptions = [
  { label: 'Пиксель', value: 'pixel' },
  { label: 'Мультяшный', value: 'cartoon' },
  { label: 'Неон', value: 'neon' },
  { label: 'Минимализм', value: 'minimal' },
  { label: 'Свой стиль', value: 'custom' },
]

const mapOptions = [
  { label: 'Компактная', value: 'small' },
  { label: 'Средняя', value: 'medium' },
  { label: 'Большая', value: 'large' },
]

function applyExample(example: (typeof examples)[number]): void {
  prompt.value = example.prompt
  genre.value = example.genre
  style.value = example.style
  mapSize.value = example.map
  tempo.value = example.tempo
  advanced.value = true
  error.value = ''
}

async function submit(): Promise<void> {
  error.value = ''

  if (!auth.isAuthenticated) {
    await router.push({ name: 'login', query: { redirect: '/create' } })
    return
  }

  if (prompt.value.trim().length < 3) {
    error.value = 'Опишите игру хотя бы в двух словах.'
    return
  }

  if (!enoughTokens.value) {
    error.value = 'Не хватает токенов на генерацию. Пополните баланс в профиле.'
    return
  }

  submitting.value = true

  try {
    const id = await game.create({
      prompt: prompt.value.trim(),
      title: title.value.trim() || undefined,
      description: description.value.trim() || undefined,
      genre: genre.value,
      style: style.value,
      customGenre: genre.value === 'custom' ? customGenre.value.trim() : undefined,
      customStyle: style.value === 'custom' ? customStyle.value.trim() : undefined,
      mapSize: mapSize.value,
      tempo: tempo.value,
    })

    toast.add({
      severity: 'success',
      summary: 'Сборка запущена',
      detail: 'Игра появится в рабочем пространстве через несколько минут.',
      life: 3500,
    })

    await router.push({ name: 'game', params: { id } })
  } catch (err) {
    error.value = messageFromError(err)
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  if (!auth.profile) void auth.refreshProfile()
})
</script>

<template>
  <div>
    <section class="pb-2 pt-8 sm:pt-10">
      <div class="app-container">
        <RouterLink
          :to="{ name: 'home' }"
          class="inline-flex items-center gap-1.5 text-[11px] font-bold text-muted transition hover:text-ink"
        >
          <AppIcon name="back" :size="14" />
          На главную
        </RouterLink>

        <div class="mt-4 flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
          <div class="animate-rise">
            <div class="eyebrow">Сборка игры</div>
            <h1 class="display-title mt-3 text-[clamp(2.2rem,4.6vw,3.4rem)]">
              Собери игру из фразы
            </h1>
            <p class="mt-3 max-w-2xl text-sm leading-6 text-muted">
              Опишите идею — агенты придумают правила, напишут код и нарисуют графику. Получится
              страница с игрой: открыть в браузере, править словами, показать друзьям.
            </p>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <span
              class="inline-flex items-center gap-2 rounded-full border border-line bg-surface/70 px-3.5 py-2 text-[11px] font-bold"
            >
              <AppIcon name="coin" :size="13" class="text-bee-deep dark:text-bee" />
              Генерация · {{ formatTokens(currentPrice) }}
            </span>
            <span
              v-if="auth.isAuthenticated"
              class="inline-flex items-center gap-2 rounded-full border border-line bg-surface/70 px-3.5 py-2 font-mono text-[11px]"
            >
              <AppIcon name="bolt" :size="13" />
              {{ formatNumber(auth.balance) }}
            </span>
          </div>
        </div>

        <div
          v-if="activeGame"
          class="mt-4 flex flex-wrap items-center gap-2 rounded-2xl border border-bee-deep/25 bg-bee/[0.06] px-4 py-3 text-xs"
        >
          <span class="size-2 animate-pulse-soft rounded-full bg-bee-deep" />
          <span class="font-bold">У вас уже идёт сборка:</span>
          <RouterLink
            :to="{ name: 'game', params: { id: activeGame.slug } }"
            class="truncate text-bee-deep hover-link dark:text-bee"
          >
            {{ activeGame.title?.trim() || activeGame.prompt }}
          </RouterLink>
          <span class="text-faint">· новая игра встанет в очередь после неё</span>
        </div>
      </div>
    </section>

    <section class="py-7 sm:py-9">
      <div class="app-container">
        <div class="panel overflow-hidden">
          <div class="grid lg:grid-cols-[minmax(0,1.35fr)_minmax(0,.65fr)]">
            <form class="p-5 sm:p-7 lg:p-8" @submit.prevent="submit">
              <div class="grid gap-4">
                <div>
                  <label for="game-title" class="mono-label mb-2 block"
                    >Название (по желанию)</label
                  >
                  <InputText
                    id="game-title"
                    v-model="title"
                    maxlength="80"
                    class="w-full"
                    placeholder="Например: Кристаллы"
                  />
                </div>
                <div>
                  <div class="mb-2 flex items-center justify-between">
                    <label for="game-description" class="mono-label">Описание (по желанию)</label>
                    <span class="font-mono text-[10px] tabular-nums text-faint"
                      >{{ description.length }} / 500</span
                    >
                  </div>
                  <Textarea
                    id="game-description"
                    v-model="description"
                    rows="2"
                    auto-resize
                    maxlength="500"
                    class="w-full text-[13px] leading-6"
                    placeholder="Например: аркада про змейку, которая растёт от съеденных кристаллов"
                  />
                </div>
              </div>

              <p class="mt-2 text-[10px] leading-4 text-faint">
                Можно оставить пустыми — название и описание придумаем после сборки вместе с
                обложкой.
              </p>

              <div class="mt-6 flex items-center justify-between gap-3">
                <label for="prompt" class="mono-label">Идея игры</label>
                <span class="font-mono text-[10px] tabular-nums text-faint"
                  >{{ prompt.length }} / {{ maxPrompt }}</span
                >
              </div>

              <Textarea
                id="prompt"
                v-model="prompt"
                rows="5"
                auto-resize
                maxlength="2000"
                class="mt-2.5 text-[15px] leading-6"
                placeholder="Например: неоновая змейка в киберпанк-городе, темп растёт с длиной хвоста, есть бонусы и таблица рекордов…"
                @update:model-value="error = ''"
              />

              <div class="mt-3 flex flex-wrap items-center gap-2">
                <span class="mr-1 text-[11px] font-semibold text-faint">Начните с примера:</span>
                <button
                  v-for="example in examples"
                  :key="example.label"
                  type="button"
                  class="rounded-full border border-line bg-canvas/60 px-3 py-1.5 text-[11px] font-semibold text-muted transition hover:border-bee-deep hover:text-ink"
                  @click="applyExample(example)"
                >
                  {{ example.label }}
                </button>
              </div>

              <div class="mt-4 grid gap-4 sm:grid-cols-2">
                <div>
                  <label class="mono-label mb-2 block">Жанр</label>
                  <Select
                    v-model="genre"
                    :options="genreOptions"
                    option-label="label"
                    option-value="value"
                  />
                  <InputText
                    v-if="genre === 'custom'"
                    v-model="customGenre"
                    maxlength="32"
                    class="mt-2 w-full"
                    placeholder="Например: ритм-игра, tower defense"
                  />
                </div>
                <div>
                  <label class="mono-label mb-2 block">Стиль</label>
                  <Select
                    v-model="style"
                    :options="styleOptions"
                    option-label="label"
                    option-value="value"
                  />
                  <Textarea
                    v-if="style === 'custom'"
                    v-model="customStyle"
                    rows="2"
                    auto-resize
                    maxlength="160"
                    class="mt-2 w-full text-[13px] leading-6"
                    placeholder="Например: аниме-нуар с бумажным коллажем, приглушённые цвета и мягкие тени"
                  />
                  <p
                    v-if="style === 'custom'"
                    class="mt-1 font-mono text-[9px] uppercase tracking-[0.1em] text-faint"
                  >
                    {{ customStyle.length }} / 160
                  </p>
                </div>
              </div>

              <div
                class="mt-4 flex items-center justify-between rounded-2xl border border-line bg-canvas/40 px-4 py-3"
              >
                <div>
                  <p class="text-sm font-bold">Расширенные настройки</p>
                  <p class="mt-0.5 text-[11px] text-muted">Размер карты и темп</p>
                </div>
                <button
                  type="button"
                  role="switch"
                  :aria-checked="advanced"
                  class="relative h-7 w-12 rounded-full border transition"
                  :class="advanced ? 'border-bee-deep bg-bee' : 'border-line-strong bg-line'"
                  @click="advanced = !advanced"
                >
                  <span
                    class="absolute top-0.5 size-5 rounded-full bg-white shadow-sm transition-all"
                    :class="advanced ? 'left-[26px]' : 'left-0.5'"
                  />
                </button>
              </div>

              <Transition name="expand">
                <div
                  v-if="advanced"
                  class="mt-3 grid gap-4 rounded-2xl border border-line bg-canvas/30 p-4 sm:grid-cols-2"
                >
                  <div>
                    <label class="mono-label mb-2 block">Размер карты</label>
                    <Select
                      v-model="mapSize"
                      :options="mapOptions"
                      option-label="label"
                      option-value="value"
                    />
                  </div>
                  <div>
                    <div class="mb-2 flex items-center justify-between">
                      <label class="mono-label">Темп</label>
                      <span class="font-mono text-[10px] font-bold text-bee-deep dark:text-bee"
                        >{{ tempo }} / 10</span
                      >
                    </div>
                    <input
                      v-model.number="tempo"
                      type="range"
                      min="1"
                      max="10"
                      step="1"
                      class="mt-3 w-full accent-[var(--bv-yellow)]"
                      aria-label="Темп игры"
                    />
                  </div>
                </div>
              </Transition>

              <div
                v-if="error"
                class="mt-4 flex items-start gap-2.5 rounded-2xl border border-red-500/20 bg-red-500/10 px-4 py-3 text-sm text-red-500"
              >
                <AppIcon name="alert" :size="17" class="mt-0.5" />
                <div class="flex-1">
                  {{ error }}
                  <RouterLink
                    v-if="error.includes('токен')"
                    to="/profile"
                    class="ml-1 font-bold underline underline-offset-2"
                  >
                    В профиль
                  </RouterLink>
                </div>
              </div>

              <div
                class="mt-6 flex flex-col justify-between gap-4 border-t border-line pt-5 sm:flex-row sm:items-center"
              >
                <div class="flex items-center gap-2 text-xs text-muted">
                  <AppIcon name="coin" :size="16" class="text-bee-deep" />
                  <span>
                    Стоимость:
                    <b class="font-mono text-ink">{{ formatTokens(currentPrice) }}</b>
                    токенов
                  </span>
                  <span v-if="auth.isAuthenticated" class="text-faint">·</span>
                  <span v-if="auth.isAuthenticated" class="text-faint">
                    баланс {{ formatTokens(auth.balance) }}
                  </span>
                </div>

                <Button
                  type="submit"
                  class="bee-button w-full sm:w-auto"
                  size="large"
                  :loading="submitting"
                  :disabled="!canSubmit"
                >
                  <AppIcon name="spark" :size="17" />
                  <span class="ml-2">{{ submitting ? 'Запускаем…' : 'Собрать игру' }}</span>
                  <AppIcon name="arrow-right" :size="16" class="ml-2" />
                </Button>
              </div>

              <p
                v-if="!auth.isAuthenticated"
                class="mt-3 text-center text-[11px] text-faint sm:text-right"
              >
                <RouterLink
                  to="/login"
                  class="font-bold text-ink underline decoration-bee decoration-2 underline-offset-2"
                >
                  Войдите
                </RouterLink>
                или зарегистрируйтесь — при старте начислим токены.
              </p>
            </form>

            <aside
              class="relative overflow-hidden border-t border-line bg-[#11110e] p-5 text-white sm:p-7 lg:border-l lg:border-t-0 lg:p-8"
            >
              <div class="dot-grid absolute inset-0 opacity-15" />
              <div class="relative">
                <p
                  class="flex items-center gap-2 font-mono text-[9px] font-bold uppercase tracking-[0.16em] text-white/40"
                >
                  <span class="size-1.5 animate-pulse-soft rounded-full bg-bee" />
                  внутри сборки
                </p>
                <h3 class="mt-5 font-display text-2xl font-bold leading-tight">
                  От идеи до игры — один проход
                </h3>

                <div class="mt-7 space-y-0">
                  <div
                    v-for="(step, index) in [
                      'Разбор идеи',
                      'План и правила',
                      'Код игры',
                      'Ревью и сборка',
                      'Готовая игра',
                    ]"
                    :key="step"
                    class="flex gap-3"
                  >
                    <div class="flex flex-col items-center">
                      <span
                        class="grid size-6 shrink-0 place-items-center rounded-full border border-white/15 bg-white/5 font-mono text-[9px] text-white/55"
                      >
                        {{ String(index + 1).padStart(2, '0') }}
                      </span>
                      <span v-if="index < 4" class="my-1 h-7 w-px bg-white/10" />
                    </div>
                    <p class="pt-0.5 text-[12px] font-semibold text-white/65">{{ step }}</p>
                  </div>
                </div>

                <div class="mt-7 rounded-2xl border border-white/10 bg-white/5 p-3">
                  <div class="flex items-center justify-between">
                    <span class="text-[11px] font-bold text-white/70">Карточка в каталоге</span>
                    <AppIcon name="grid" :size="14" class="text-bee" />
                  </div>

                  <div
                    class="mt-2.5 overflow-hidden rounded-xl border border-white/10 bg-[#0b0b09]"
                  >
                    <div class="relative flex aspect-[16/7] items-end p-2.5" :style="previewCover">
                      <div class="dot-grid absolute inset-0 opacity-30" />
                      <span
                        class="relative rounded-full border border-white/15 bg-black/35 px-2 py-0.5 font-mono text-[8px] font-bold uppercase tracking-[0.12em] text-white/80 backdrop-blur"
                      >
                        {{ previewGenre || 'жанр' }}
                      </span>
                    </div>
                    <div class="p-3">
                      <p
                        class="line-clamp-1 text-[13px] font-bold leading-snug"
                        :title="previewTitle"
                      >
                        {{ previewTitle }}
                      </p>
                      <p class="mt-1 line-clamp-2 text-[10px] leading-4 text-white/45">
                        {{ previewDescription }}
                      </p>
                      <p
                        class="mt-2 flex flex-wrap items-center gap-x-2 gap-y-1 font-mono text-[8px] uppercase tracking-[0.1em] text-white/35"
                      >
                        <span>{{ previewStyle || 'стиль' }}</span>
                        <span class="size-0.5 rounded-full bg-white/30" />
                        <span>черновик</span>
                      </p>
                    </div>
                  </div>
                </div>

                <div class="mt-8 rounded-2xl border border-white/10 bg-white/5 p-4">
                  <div class="flex items-center justify-between">
                    <span class="text-[11px] font-bold text-white/70">Что получите</span>
                    <AppIcon name="shield" :size="15" class="text-bee" />
                  </div>
                  <ul class="mt-3 space-y-2 text-[11px] leading-4 text-white/45">
                    <li class="flex gap-2">
                      <AppIcon name="check" :size="12" class="mt-0.5 text-bee" /> игру, которая
                      открывается в браузере
                    </li>
                    <li class="flex gap-2">
                      <AppIcon name="check" :size="12" class="mt-0.5 text-bee" /> правки словами и
                      готовыми настройками
                    </li>
                    <li class="flex gap-2">
                      <AppIcon name="check" :size="12" class="mt-0.5 text-bee" /> историю версий,
                      откат и публичную ссылку
                    </li>
                  </ul>
                </div>
              </div>
            </aside>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.expand-enter-active,
.expand-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  transform: translateY(-5px);
}
</style>
