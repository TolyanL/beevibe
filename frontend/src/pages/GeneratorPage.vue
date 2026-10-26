<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Button from 'openvue/button'
import Textarea from 'openvue/textarea'
import Select from 'openvue/select'
import { useToast } from 'openvue/usetoast'
import { getCatalog } from '@/api/catalog'
import type { CatalogCard, GameStyle, Genre, MapSize } from '@/api/types'
import { messageFromError } from '@/api/client'
import { formatTokens } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'
import { useGameStore } from '@/stores/game'
import AppIcon from '@/components/ui/AppIcon.vue'
import GameCard from '@/components/catalog/GameCard.vue'
import HeroPreview from '@/components/generator/HeroPreview.vue'

const router = useRouter()
const toast = useToast()
const auth = useAuthStore()
const game = useGameStore()

const prompt = ref('')
const genre = ref<Genre>('arcade')
const style = ref<GameStyle>('neon')
const mapSize = ref<MapSize>('medium')
const tempo = ref(6)
const advanced = ref(false)
const submitting = ref(false)
const error = ref('')
const latest = ref<CatalogCard[]>([])
const latestLoading = ref(true)

const maxPrompt = 2000
const currentPrice = computed(() => auth.prices.generate)
const enoughTokens = computed(() => !auth.isAuthenticated || auth.balance >= currentPrice.value)
const canSubmit = computed(
  () => prompt.value.trim().length >= 3 && prompt.value.length <= maxPrompt && !submitting.value,
)
const examples = [
  {
    label: 'Неоновая змейка',
    prompt: 'Неоновая змейка в киберпанк-городе: чем длиннее хвост, тем быстрее темп, на карте появляются стены.',
    genre: 'snake' as Genre,
    style: 'neon' as GameStyle,
    map: 'medium' as MapSize,
    tempo: 8,
  },
  {
    label: 'Пиксельный кликер',
    prompt: 'Пиксельный кликер про космическую пекарню: печенье заменяют астероиды, есть апгрейды и редкие бонусы.',
    genre: 'clicker' as Genre,
    style: 'pixel' as GameStyle,
    map: 'small' as MapSize,
    tempo: 6,
  },
  {
    label: 'Платформер-побег',
    prompt: 'Мультяшный платформер: лягушка убегает от дождя, прыгает по листьям и собирает светлячков.',
    genre: 'platformer' as Genre,
    style: 'cartoon' as GameStyle,
    map: 'large' as MapSize,
    tempo: 7,
  },
  {
    label: 'Головоломка',
    prompt: 'Минималистичная головоломка с перемещением кристаллов: нужно собрать три одинаковых цвета в ряд за 30 ходов.',
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

const features = [
  {
    icon: 'wand',
    title: 'Промпт вместо ТЗ',
    text: 'Опишите механику, настроение и цель. Конвейер сам соберёт спецификацию.',
  },
  {
    icon: 'code',
    title: 'Живой playable',
    text: 'Игра запускается сразу в широкой сцене — без установок и сборок.',
  },
  {
    icon: 'refresh',
    title: 'Правки без кода',
    text: 'Патч промптом, настройки, починка и откат к любой версии.',
  },
] as const

const styleOptions = [
  { label: 'Пиксель', value: 'pixel' },
  { label: 'Мультяшный', value: 'cartoon' },
  { label: 'Неон', value: 'neon' },
  { label: 'Минимализм', value: 'minimal' },
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

function scrollToComposer(): void {
  document.querySelector('#create')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

async function submit(): Promise<void> {
  error.value = ''

  if (!auth.isAuthenticated) {
    await router.push({ name: 'login', query: { redirect: '/#create' } })
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
      genre: genre.value,
      style: style.value,
      mapSize: mapSize.value,
      tempo: tempo.value,
    })

    toast.add({
      severity: 'success',
      summary: 'Конвейер запущен',
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

onMounted(async () => {
  if (auth.isAuthenticated && !auth.profile) void auth.refreshProfile()

  try {
    const response = await getCatalog({ sort: 'rating', limit: 4 })
    latest.value = response.items ?? []
  } catch {
    latest.value = []
  } finally {
    latestLoading.value = false
  }
})
</script>

<template>
  <div>
    <section class="relative overflow-hidden pb-4 pt-10 sm:pt-14 lg:pt-16">
      <div class="app-container">
        <div class="grid items-center gap-10 lg:grid-cols-[1.02fr_.98fr] lg:gap-14">
          <div class="animate-rise">
            <div class="eyebrow">GameGen · конвейер агентов</div>

            <h1 class="display-title mt-5 text-[clamp(2.8rem,6.8vw,5.6rem)]">
              Собери игру
              <span class="relative inline-block">
                <span class="relative z-10">из фразы.</span>
                <span class="absolute inset-x-0 bottom-1 z-0 h-[0.18em] -rotate-1 rounded-full bg-bee/70" />
              </span>
            </h1>

            <p class="mt-6 max-w-xl text-[15px] leading-7 text-muted sm:text-base">
              Опишите идею — BeeVibe проведёт её через восемь этапов и вернёт готовую
              HTML5-игру. Играйте сразу в браузере, правьте промптом и делитесь ссылкой.
            </p>

            <div class="mt-7 flex flex-wrap items-center gap-2.5">
              <Button class="bee-button" size="large" @click="scrollToComposer">
                <AppIcon name="spark" :size="17" />
                <span class="ml-2">Создать игру</span>
              </Button>
              <RouterLink
                to="/catalog"
                class="inline-flex h-[46px] items-center gap-2 rounded-full border border-line-strong bg-surface/50 px-5 text-sm font-bold transition hover:border-bee-deep hover:bg-bee/10"
              >
                <AppIcon name="grid" :size="16" />
                Смотреть каталог
              </RouterLink>
            </div>

            <div class="mt-9 grid max-w-lg grid-cols-3 gap-4 border-t border-line pt-5">
              <div>
                <p class="font-display text-xl font-bold">8</p>
                <p class="mt-1 font-mono text-[9px] uppercase tracking-[0.12em] text-muted">этапов</p>
              </div>
              <div>
                <p class="font-display text-xl font-bold">HTML5</p>
                <p class="mt-1 font-mono text-[9px] uppercase tracking-[0.12em] text-muted">один файл</p>
              </div>
              <div>
                <p class="font-display text-xl font-bold">0</p>
                <p class="mt-1 font-mono text-[9px] uppercase tracking-[0.12em] text-muted">установок</p>
              </div>
            </div>
          </div>

          <div class="animate-rise lg:translate-y-2" style="animation-delay: 90ms">
            <HeroPreview />
          </div>
        </div>
      </div>
    </section>

    <section id="create" class="scroll-mt-24 py-12 sm:py-16">
      <div class="app-container">
        <div class="mb-6 flex flex-wrap items-end justify-between gap-4">
          <div>
            <div class="eyebrow">Конструктор</div>
            <h2 class="display-title mt-3 text-3xl sm:text-4xl">Опишите, во что будем играть</h2>
          </div>
          <p class="max-w-xs text-right text-xs leading-5 text-muted sm:text-sm">
            Чем конкретнее механика, цель и атмосфера — тем точнее результат.
          </p>
        </div>

        <div class="panel overflow-hidden">
          <div class="grid lg:grid-cols-[1.35fr_.65fr]">
            <form class="p-5 sm:p-7 lg:p-8" @submit.prevent="submit">
              <div class="flex items-center justify-between gap-3">
                <label for="prompt" class="mono-label">Промпт игры</label>
                <span class="font-mono text-[10px] tabular-nums text-faint">{{ prompt.length }} / {{ maxPrompt }}</span>
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
                <span class="mr-1 text-[11px] font-semibold text-faint">Быстрый старт:</span>
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

              <div class="mt-6 grid gap-4 sm:grid-cols-2">
                <div>
                  <label class="mono-label mb-2 block">Жанр</label>
                  <Select v-model="genre" :options="genreOptions" option-label="label" option-value="value" />
                </div>
                <div>
                  <label class="mono-label mb-2 block">Стиль</label>
                  <Select v-model="style" :options="styleOptions" option-label="label" option-value="value" />
                </div>
              </div>

              <div class="mt-4 flex items-center justify-between rounded-2xl border border-line bg-canvas/40 px-4 py-3">
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
                <div v-if="advanced" class="mt-3 grid gap-4 rounded-2xl border border-line bg-canvas/30 p-4 sm:grid-cols-2">
                  <div>
                    <label class="mono-label mb-2 block">Размер карты</label>
                    <Select v-model="mapSize" :options="mapOptions" option-label="label" option-value="value" />
                  </div>
                  <div>
                    <div class="mb-2 flex items-center justify-between">
                      <label class="mono-label">Темп</label>
                      <span class="font-mono text-[10px] font-bold text-bee-deep dark:text-bee">{{ tempo }} / 10</span>
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

              <div class="mt-6 flex flex-col justify-between gap-4 border-t border-line pt-5 sm:flex-row sm:items-center">
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

                <Button type="submit" class="bee-button w-full sm:w-auto" size="large" :loading="submitting" :disabled="!canSubmit">
                  <AppIcon name="spark" :size="17" />
                  <span class="ml-2">{{ submitting ? 'Запускаем…' : 'Запустить конвейер' }}</span>
                  <AppIcon name="arrow-right" :size="16" class="ml-2" />
                </Button>
              </div>

              <p v-if="!auth.isAuthenticated" class="mt-3 text-center text-[11px] text-faint sm:text-right">
                <RouterLink to="/login" class="font-bold text-ink underline decoration-bee decoration-2 underline-offset-2">
                  Войдите
                </RouterLink>
                или зарегистрируйтесь — при старте начислим токены.
              </p>
            </form>

            <aside class="relative overflow-hidden border-t border-line bg-[#11110e] p-5 text-white sm:p-7 lg:border-l lg:border-t-0 lg:p-8">
              <div class="dot-grid absolute inset-0 opacity-15" />
              <div class="relative">
                <p class="flex items-center gap-2 font-mono text-[9px] font-bold uppercase tracking-[0.16em] text-white/40">
                  <span class="size-1.5 animate-pulse-soft rounded-full bg-bee" />
                  внутри конвейера
                </p>
                <h3 class="mt-5 font-display text-2xl font-bold leading-tight">От идеи до playable за один проход</h3>

                <div class="mt-7 space-y-0">
                  <div v-for="(step, index) in ['Анализ промпта', 'План и механика', 'Генерация HTML', 'Ревью и сборка', 'Playable-релиз']" :key="step" class="flex gap-3">
                    <div class="flex flex-col items-center">
                      <span class="grid size-6 shrink-0 place-items-center rounded-full border border-white/15 bg-white/5 font-mono text-[9px] text-white/55">
                        {{ String(index + 1).padStart(2, '0') }}
                      </span>
                      <span v-if="index < 4" class="my-1 h-7 w-px bg-white/10" />
                    </div>
                    <p class="pt-0.5 text-[12px] font-semibold text-white/65">{{ step }}</p>
                  </div>
                </div>

                <div class="mt-8 rounded-2xl border border-white/10 bg-white/5 p-4">
                  <div class="flex items-center justify-between">
                    <span class="text-[11px] font-bold text-white/70">Что получите</span>
                    <AppIcon name="shield" :size="15" class="text-bee" />
                  </div>
                  <ul class="mt-3 space-y-2 text-[11px] leading-4 text-white/45">
                    <li class="flex gap-2"><AppIcon name="check" :size="12" class="mt-0.5 text-bee" /> готовую игру без сборки</li>
                    <li class="flex gap-2"><AppIcon name="check" :size="12" class="mt-0.5 text-bee" /> правки промптом и настройками</li>
                    <li class="flex gap-2"><AppIcon name="check" :size="12" class="mt-0.5 text-bee" /> версии, откат и публичную ссылку</li>
                  </ul>
                </div>
              </div>
            </aside>
          </div>
        </div>
      </div>
    </section>

    <section class="py-10 sm:py-14">
      <div class="app-container">
        <div class="flex flex-wrap items-end justify-between gap-4">
          <div>
            <div class="eyebrow">Уже играют</div>
            <h2 class="display-title mt-3 text-3xl sm:text-4xl">Свежие игры из каталога</h2>
          </div>
          <RouterLink to="/catalog" class="group inline-flex items-center gap-2 text-sm font-bold text-muted transition hover:text-ink">
            Все игры
            <AppIcon name="arrow-right" :size="16" class="transition group-hover:translate-x-0.5" />
          </RouterLink>
        </div>

        <div v-if="latestLoading" class="mt-7 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div v-for="index in 4" :key="index" class="h-64 animate-pulse rounded-[1.35rem] bg-surface/70" />
        </div>
        <div v-else-if="latest.length" class="mt-7 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <GameCard v-for="card in latest" :key="card.id" :card="card" />
        </div>
        <div v-else class="mt-7 rounded-[1.35rem] border border-dashed border-line-strong px-6 py-12 text-center">
          <p class="text-sm font-bold">Каталог пока пуст</p>
          <p class="mt-1.5 text-xs text-muted">Сгенерируйте первую игру — она появится здесь.</p>
        </div>
      </div>
    </section>

    <section class="py-12 sm:py-16">
      <div class="app-container">
        <div class="grid gap-px overflow-hidden rounded-[1.35rem] border border-line bg-line sm:grid-cols-3">
          <div v-for="(item, index) in features" :key="item.title" class="bg-surface/80 p-6 sm:p-7">
            <span class="grid size-11 place-items-center rounded-2xl bg-bee/15 text-bee-deep dark:text-bee">
              <AppIcon :name="item.icon" :size="21" />
            </span>
            <p class="mt-5 font-display text-lg font-bold">{{ item.title }}</p>
            <p class="mt-2 text-[13px] leading-6 text-muted">{{ item.text }}</p>
            <p class="mt-5 font-mono text-[9px] uppercase tracking-[0.14em] text-faint">0{{ index + 1 }} / 03</p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.expand-enter-active,
.expand-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  transform: translateY(-5px);
}
</style>
