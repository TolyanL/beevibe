<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import Button from 'openvue/button'
import Dialog from 'openvue/dialog'
import Rating from 'openvue/rating'
import Select from 'openvue/select'
import Textarea from 'openvue/textarea'
import { useConfirm } from 'openvue/useconfirm'
import { useToast } from 'openvue/usetoast'
import { messageFromError } from '@/api/client'
import type { GameVersion } from '@/api/types'
import { formatDate, formatTokens, plural } from '@/lib/format'
import {
  genreLabel,
  mapSizeLabel,
  stageLabel,
  statusMeta,
  styleLabel,
  versionSourceLabel,
} from '@/lib/game'
import { ratingTone } from '@/lib/rating'
import { usePlaytime } from '@/composables/usePlaytime'
import { useAuthStore } from '@/stores/auth'
import { useGameStore } from '@/stores/game'
import AppIcon from '@/components/ui/AppIcon.vue'
import StatusPill from '@/components/ui/StatusPill.vue'
import RatingBadge from '@/components/ui/RatingBadge.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import GameFrame from '@/components/game/GameFrame.vue'
import StageTimeline from '@/components/game/StageTimeline.vue'

const route = useRoute()
const toast = useToast()
const confirm = useConfirm()
const auth = useAuthStore()
const game = useGameStore()

const id = computed(() => String(route.params.id ?? ''))
const promptText = computed(
  () => auth.profile?.games.find((item) => item.id === id.value)?.prompt ?? `Игра ${id.value.slice(0, 8)}`,
)
const profileGame = computed(() => auth.profile?.games.find((item) => item.id === id.value))
const versions = computed(() =>
  [...(game.game?.versions ?? [])].sort((a, b) => b.number - a.number),
)
const latestVersion = computed(() => versions.value[0])
const meta = computed(() => statusMeta(game.status))

const patchText = ref('')
const patchError = ref('')
const patchOpen = ref(true)
const fixOpen = ref(false)
const fixText = ref('')
const savingConfig = ref(false)
const vote = ref(0)
const ratingKey = ref(0)
const voteError = ref('')
const copied = ref(false)
const session = usePlaytime(() => id.value, () => game.isReady)
const sessionSeconds = session.seconds

const configForm = reactive({
  tempo: 5,
  mapSize: 'medium',
  difficulty: 'normal',
  palette: ['#ffcc00', '#22d3ee', '#fb7185', '#84cc16'],
})

const mapOptions = [
  { label: 'Компактная', value: 'small' },
  { label: 'Средняя', value: 'medium' },
  { label: 'Большая', value: 'large' },
]

const difficultyOptions = [
  { label: 'Лёгкая', value: 'easy' },
  { label: 'Средняя', value: 'normal' },
  { label: 'Сложная', value: 'hard' },
]

const patchSuggestions = [
  'сделай управление плавнее',
  'добавь экран победы и рестарт',
  'увеличь контраст и читаемость',
  'добавь короткие звуковые эффекты',
  'исправь коллизии со стенами',
]

const ratingInputColor = computed(() => (vote.value ? ratingTone(vote.value).hex : '#e6b800'))
const averageTone = computed(() => ratingTone(game.rating.average))

const ratingRows = computed(() =>
  [5, 4, 3, 2, 1].map((star) => {
    const count = game.rating.counts?.[star - 1] ?? 0
    const percent = game.rating.total ? (count / game.rating.total) * 100 : 0
    return { star, count, percent, tone: ratingTone(star) }
  }),
)
const maxRatingCount = computed(() => Math.max(1, ...ratingRows.value.map((row) => row.count)))
const canPatch = computed(() => patchText.value.trim().length >= 3 && !game.actionPending)
const showMaintenance = computed(() => game.isReady && !game.loading)

watch(
  () => game.myRating,
  (value) => {
    vote.value = value
  },
  { immediate: true },
)

watch(
  () => game.config,
  (config) => {
    if (!config) return
    configForm.tempo = config.tempo ?? 5
    configForm.mapSize = config.map_size ?? 'medium'
    configForm.difficulty = config.difficulty ?? 'normal'
    configForm.palette = config.palette?.length ? [...config.palette] : configForm.palette
  },
  { immediate: true },
)

function sessionLabel(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
}

function logLine(stage: { stage: string; status: string; durationMs: number }): string {
  const time = stage.durationMs ? ` ${(stage.durationMs / 1000).toFixed(1)}s` : ''
  return `[${stage.status === 'done' ? 'ok' : '..'}] ${stageLabel(stage.stage)}${time}`
}

async function load(): Promise<void> {
  await game.load(id.value)
  if (!auth.profile) await auth.refreshProfile()
}

async function copyId(): Promise<void> {
  try {
    await navigator.clipboard.writeText(id.value)
    copied.value = true
    window.setTimeout(() => (copied.value = false), 1600)
  } catch {
    toast.add({ severity: 'error', summary: 'Не удалось скопировать', life: 2000 })
  }
}

async function sendPatch(): Promise<void> {
  patchError.value = ''
  if (patchText.value.trim().length < 3) {
    patchError.value = 'Опишите правку чуть подробнее.'
    return
  }

  try {
    await game.applyPatch(patchText.value.trim())
    patchText.value = ''
    toast.add({ severity: 'success', summary: 'Патч применён', detail: 'Новая версия собрана.', life: 2600 })
  } catch (err) {
    patchError.value = messageFromError(err)
  }
}

async function saveConfig(): Promise<void> {
  savingConfig.value = true
  try {
    await game.applyConfig({
      tempo: configForm.tempo,
      mapSize: configForm.mapSize as 'small' | 'medium' | 'large',
      difficulty: configForm.difficulty,
      palette: configForm.palette.filter(Boolean),
    })
    toast.add({ severity: 'success', summary: 'Настройки сохранены', life: 2400 })
  } catch (err) {
    toast.add({ severity: 'error', summary: 'Не удалось сохранить', detail: messageFromError(err), life: 3200 })
  } finally {
    savingConfig.value = false
  }
}

function rollback(version: GameVersion): void {
  confirm.require({
    header: 'Откатить версию?',
    message: `Игра вернётся к версии #${version.number}. Текущий playable будет заменён.`,
    acceptLabel: 'Откатить',
    rejectLabel: 'Отмена',
    accept: async () => {
      try {
        await game.rollback(version.number)
        toast.add({ severity: 'success', summary: `Версия #${version.number} активна`, life: 2400 })
      } catch (err) {
        toast.add({ severity: 'error', summary: 'Откат не удался', detail: messageFromError(err), life: 3200 })
      }
    },
  })
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

async function rate(value: number): Promise<void> {
  voteError.value = ''
  try {
    await game.vote(value)
    toast.add({ severity: 'success', summary: 'Оценка сохранена', detail: 'Спасибо — это помогает каталогу.', life: 2200 })
  } catch (err) {
    vote.value = game.myRating
    voteError.value = messageFromError(err)
  }
}

function openFix(): void {
  fixText.value = ''
  fixOpen.value = true
}

async function repair(): Promise<void> {
  try {
    await game.repair(fixText.value.trim() || undefined)
    fixOpen.value = false
    toast.add({ severity: 'success', summary: 'Починка запущена', detail: 'Следим за конвейером справа.', life: 2800 })
  } catch (err) {
    toast.add({ severity: 'error', summary: 'Починка не запустилась', detail: messageFromError(err), life: 3200 })
  }
}

function useSuggestion(text: string): void {
  patchText.value = text
  patchOpen.value = true
  document.querySelector('#patch')?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

let poll: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  void load()
  poll = setInterval(() => {
    if (game.isActive && !game.loading) void game.refresh()
  }, 15_000)
})

onBeforeUnmount(() => {
  if (poll) clearInterval(poll)
  poll = null
  game.reset()
})
</script>

<template>
  <div class="pb-8 pt-6 sm:pt-8">
    <div class="app-container">
      <div class="mb-5 flex flex-wrap items-center justify-between gap-3">
        <RouterLink
          to="/profile"
          class="inline-flex items-center gap-1.5 text-[11px] font-bold text-muted transition hover:text-ink"
        >
          <AppIcon name="back" :size="14" />
          Мои игры
        </RouterLink>
        <div class="flex items-center gap-2">
          <span class="font-mono text-[9px] uppercase tracking-[0.13em] text-faint">owner workspace</span>
          <span class="size-0.5 rounded-full bg-faint" />
          <button type="button" class="font-mono text-[9px] text-faint transition hover:text-ink" @click="copyId">
            {{ copied ? 'id скопирован' : `id · ${id.slice(0, 8)}` }}
          </button>
        </div>
      </div>

      <div v-if="game.loading && !game.game" class="space-y-5">
        <div class="h-16 w-2/3 animate-pulse rounded-2xl bg-surface/70" />
        <div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_350px]">
          <div class="aspect-video animate-pulse rounded-[1.35rem] bg-surface/70" />
          <div class="h-96 animate-pulse rounded-[1.35rem] bg-surface/70" />
        </div>
      </div>

      <div v-else-if="game.error && !game.game" class="pt-6">
        <EmptyState
          icon="alert"
          title="Игра недоступна"
          :text="game.error || 'Возможно, игра удалена или у вас нет доступа к рабочему пространству.'"
        >
          <div class="flex flex-wrap justify-center gap-2">
            <Button class="bee-button" @click="load">Повторить</Button>
            <RouterLink to="/profile" class="rounded-full border border-line-strong px-5 py-2.5 text-xs font-bold">
              В профиль
            </RouterLink>
          </div>
        </EmptyState>
      </div>

      <template v-else-if="game.game">
        <header class="mb-5 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <StatusPill :status="game.status" />
              <span class="font-mono text-[9px] uppercase tracking-[0.12em] text-faint">
                {{ genreLabel(profileGame?.genre ?? game.spec?.genre ?? '') }}
                <span v-if="profileGame?.created_at"> · {{ formatDate(profileGame.created_at) }}</span>
              </span>
            </div>
            <h1 class="mt-3 max-w-4xl font-display text-2xl font-bold leading-tight tracking-[-0.04em] sm:text-3xl lg:text-[2.2rem]">
              {{ promptText }}
            </h1>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <div v-if="game.isReady" class="flex items-center gap-1.5 rounded-full border border-line bg-surface/70 px-3 py-2 font-mono text-[10px] text-muted">
              <AppIcon name="clock" :size="13" />
              сессия {{ sessionLabel(sessionSeconds) }}
            </div>
            <RouterLink
              v-if="game.isReady"
              :to="{ name: 'play', params: { id } }"
              target="_blank"
              class="inline-flex items-center gap-1.5 rounded-full border border-line-strong px-3.5 py-2 text-[11px] font-bold transition hover:border-bee-deep hover:bg-bee/10"
            >
              <AppIcon name="external" :size="13" />
              Публичная страница
            </RouterLink>
            <Button v-if="game.isReady" class="p-button-secondary" :loading="game.actionPending === 'fix'" @click="openFix">
              <AppIcon name="refresh" :size="14" />
              <span class="ml-1.5">Починить</span>
            </Button>
          </div>
        </header>

        <div class="grid items-start gap-5 xl:grid-cols-[minmax(0,1fr)_350px]">
          <div class="min-w-0 space-y-5">
            <div v-if="promptText" class="animate-rise">
              <GameFrame :game-id="id" :status="game.status" :version-token="game.refreshKey" @loaded="session.flush()" />
            </div>

            <div
              v-if="game.streamError"
              class="flex items-start gap-2.5 rounded-2xl border border-orange-500/20 bg-orange-500/10 px-4 py-3 text-xs text-orange-500"
            >
              <AppIcon name="alert" :size="16" class="mt-0.5" />
              {{ game.streamError }}
            </div>

            <div v-if="game.isFailed" class="panel border-red-500/20 p-5 sm:p-6">
              <div class="flex items-start gap-3">
                <span class="grid size-10 shrink-0 place-items-center rounded-xl bg-red-500/10 text-red-500">
                  <AppIcon name="alert" :size="20" />
                </span>
                <div class="flex-1">
                  <h2 class="font-display text-lg font-bold">Конвейер остановился</h2>
                  <p class="mt-1.5 text-xs leading-5 text-muted">
                    Игра не собралась. Запустите починку — агент получит контекст ошибки и
                    попробует собрать новую версию.
                  </p>
                  <div class="mt-4 flex flex-wrap gap-2">
                    <Button class="bee-button" :loading="game.actionPending === 'fix'" @click="openFix">
                      <AppIcon name="refresh" :size="14" />
                      <span class="ml-1.5">Починить игру</span>
                    </Button>
                    <RouterLink to="/" class="rounded-full border border-line-strong px-4 py-2.5 text-xs font-bold">
                      Начать заново
                    </RouterLink>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="game.isActive" class="panel overflow-hidden">
              <div class="flex items-center justify-between gap-3 border-b border-line px-5 py-4">
                <div>
                  <p class="mono-label">Live build</p>
                  <h2 class="mt-1 font-display text-lg font-bold">Сборка в реальном времени</h2>
                </div>
                <span class="flex items-center gap-1.5 font-mono text-[9px] uppercase tracking-[0.1em] text-bee-deep dark:text-bee">
                  <span class="size-1.5 animate-pulse-soft rounded-full bg-bee" />
                  {{ game.streaming ? 'stream' : 'poll' }}
                </span>
              </div>

              <div class="grid gap-0 md:grid-cols-[.9fr_1.1fr]">
                <div class="border-b border-line p-5 md:border-b-0 md:border-r">
                  <p class="mono-label">Этапы</p>
                  <p class="mt-2 text-sm leading-6 text-muted">{{ meta.description }}</p>
                  <div class="mt-4 grid grid-cols-3 gap-2">
                    <div class="rounded-xl border border-line bg-canvas/45 p-3">
                      <p class="font-display text-lg font-bold">{{ game.stageProgress }}%</p>
                      <p class="mt-0.5 font-mono text-[8px] uppercase tracking-[0.1em] text-faint">прогресс</p>
                    </div>
                    <div class="rounded-xl border border-line bg-canvas/45 p-3">
                      <p class="font-display text-lg font-bold">{{ game.stages.filter((item) => item.status === 'done').length }}/8</p>
                      <p class="mt-0.5 font-mono text-[8px] uppercase tracking-[0.1em] text-faint">этапов</p>
                    </div>
                    <div class="rounded-xl border border-line bg-canvas/45 p-3">
                      <p class="font-display text-lg font-bold">{{ game.stages.filter((item) => item.issues).length }}</p>
                      <p class="mt-0.5 font-mono text-[8px] uppercase tracking-[0.1em] text-faint">issues</p>
                    </div>
                  </div>
                </div>

                <div class="p-4">
                  <div class="terminal h-[210px] p-4">
                    <p><span class="tok-comment">// beevibe pipeline</span></p>
                    <p v-for="stage in game.stages" :key="stage.stage">
                      <span :class="stage.status === 'done' ? 'tok-string' : 'tok-key'">{{ logLine(stage) }}</span>
                    </p>
                    <p v-if="!game.stages.length" class="tok-comment">ожидаем первый этап…</p>
                    <p class="mt-1 flex items-center gap-1 text-bee">
                      <span>$</span><span class="animate-pulse-soft">_</span>
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="game.isActive && !game.spec" class="panel p-5">
              <div class="flex items-center gap-3">
                <span class="size-4 animate-spin rounded-full border-2 border-line border-t-bee-deep" />
                <p class="text-xs text-muted">Спецификация появится после этапа планирования.</p>
              </div>
            </div>

            <div v-if="game.spec" class="panel p-5 sm:p-6">
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="mono-label">Game spec</p>
                  <h2 class="mt-1 font-display text-xl font-bold">Что за игра</h2>
                </div>
                <span class="rounded-full border border-line bg-canvas/50 px-3 py-1 font-mono text-[9px] uppercase tracking-[0.1em] text-muted">
                  {{ genreLabel(game.spec.genre) }} · {{ styleLabel(game.spec.style) }}
                </span>
              </div>

              <div class="mt-5 grid gap-3 sm:grid-cols-2">
                <div class="rounded-2xl border border-line bg-canvas/35 p-4">
                  <p class="mono-label">Механика</p>
                  <p class="mt-2 text-[13px] leading-6">{{ game.spec.mechanic }}</p>
                </div>
                <div class="rounded-2xl border border-line bg-canvas/35 p-4">
                  <p class="mono-label">Управление</p>
                  <p class="mt-2 text-[13px] leading-6">{{ game.spec.controls }}</p>
                </div>
                <div class="rounded-2xl border border-line bg-canvas/35 p-4">
                  <p class="mono-label">Победа</p>
                  <p class="mt-2 text-[13px] leading-6">{{ game.spec.win_condition }}</p>
                </div>
                <div class="rounded-2xl border border-line bg-canvas/35 p-4">
                  <p class="mono-label">Проигрыш</p>
                  <p class="mt-2 text-[13px] leading-6">{{ game.spec.lose_condition }}</p>
                </div>
              </div>

              <div class="mt-4 flex flex-wrap items-center gap-2 border-t border-line pt-4">
                <span class="mono-label mr-1">Палитра</span>
                <span
                  v-for="color in game.spec.palette"
                  :key="color"
                  class="size-7 rounded-full border border-black/10 shadow-inner"
                  :style="{ backgroundColor: color }"
                  :title="color"
                />
                <span class="ml-auto font-mono text-[9px] uppercase tracking-[0.1em] text-faint">
                  {{ mapSizeLabel(game.spec.map_size) }} карта
                </span>
              </div>
            </div>

            <div v-if="showMaintenance" id="patch" class="panel p-5 sm:p-6">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p class="mono-label">Патч промптом</p>
                  <h2 class="mt-1 font-display text-xl font-bold">Что изменить?</h2>
                </div>
                <span class="font-mono text-[9px] uppercase tracking-[0.1em] text-faint">
                  {{ formatTokens(auth.prices.patch) }} токенов
                </span>
              </div>

              <button
                type="button"
                class="mt-4 flex w-full items-center justify-between gap-3 rounded-xl border border-line bg-canvas/35 px-4 py-3 text-left text-xs font-semibold text-muted transition hover:border-bee-deep/40 md:hidden"
                @click="patchOpen = !patchOpen"
              >
                {{ patchOpen ? 'Свернуть редактор' : 'Открыть редактор правок' }}
                <AppIcon name="chevron-down" :size="14" :class="patchOpen ? 'rotate-180' : ''" />
              </button>

              <div v-show="patchOpen" class="mt-4">
                <Textarea
                  v-model="patchText"
                  rows="3"
                  auto-resize
                  maxlength="2000"
                  class="text-[13px] leading-6"
                  placeholder="Например: добавь двойной прыжок, ускорь анимацию и сделай фон темнее…"
                  @update:model-value="patchError = ''"
                />

                <div class="mt-2.5 flex flex-wrap gap-1.5">
                  <button
                    v-for="suggestion in patchSuggestions"
                    :key="suggestion"
                    type="button"
                    class="rounded-full border border-line bg-canvas/50 px-2.5 py-1 text-[10px] font-semibold text-muted transition hover:border-bee-deep hover:text-ink"
                    @click="useSuggestion(suggestion)"
                  >
                    {{ suggestion }}
                  </button>
                </div>

                <div v-if="patchError" class="mt-3 text-xs text-red-500">{{ patchError }}</div>

                <div class="mt-4 flex flex-wrap items-center justify-between gap-3">
                  <p class="text-[10px] leading-4 text-faint">
                    Патч создаёт новую версию. Старые можно вернуть в любой момент.
                  </p>
                  <Button class="bee-button" :loading="game.actionPending === 'patch'" :disabled="!canPatch" @click="sendPatch">
                    <AppIcon name="wand" :size="15" />
                    <span class="ml-1.5">Применить патч</span>
                  </Button>
                </div>
              </div>
            </div>

            <div v-if="showMaintenance" class="panel overflow-hidden">
              <div class="flex items-center justify-between gap-3 border-b border-line px-5 py-4">
                <div>
                  <p class="mono-label">История</p>
                  <h2 class="mt-1 font-display text-lg font-bold">Версии игры</h2>
                </div>
                <span class="font-mono text-[10px] text-faint">{{ versions.length }}</span>
              </div>

              <div v-if="versions.length" class="divide-y divide-line">
                <div v-for="version in versions" :key="version.id" class="flex items-center gap-3 px-5 py-3.5">
                  <span
                    class="grid size-8 shrink-0 place-items-center rounded-lg border font-mono text-[10px] font-bold"
                    :class="version.number === latestVersion?.number ? 'border-bee-deep/30 bg-bee/10 text-bee-deep dark:text-bee' : 'border-line bg-canvas/40 text-faint'"
                  >
                    #{{ version.number }}
                  </span>
                  <div class="min-w-0 flex-1">
                    <p class="text-xs font-bold">
                      {{ versionSourceLabel(version.source) }}
                      <span v-if="version.note" class="font-normal text-muted">— {{ version.note }}</span>
                    </p>
                    <p class="mt-0.5 font-mono text-[9px] text-faint">
                      {{ formatDate(version.created_at) }} · {{ formatTokens(version.tokens_cost) }} токенов
                    </p>
                  </div>
                  <span v-if="version.number === latestVersion?.number" class="rounded-full bg-emerald-500/10 px-2 py-0.5 font-mono text-[8px] font-bold uppercase tracking-[0.08em] text-emerald-500">
                    active
                  </span>
                  <button
                    v-else
                    type="button"
                    class="rounded-full border border-line px-2.5 py-1 text-[9px] font-bold text-muted transition hover:border-bee-deep hover:text-ink"
                    :disabled="Boolean(game.actionPending)"
                    @click="rollback(version)"
                  >
                    Вернуть
                  </button>
                </div>
              </div>
              <div v-else class="px-5 py-8 text-center text-xs text-faint">Версий пока нет.</div>
            </div>
          </div>

          <aside class="space-y-5 xl:sticky xl:top-[90px]">
            <section class="panel p-5">
              <StageTimeline :stages="game.stages" :status="game.status" />
              <div class="mt-5 grid grid-cols-2 gap-2 border-t border-line pt-4">
                <div class="rounded-xl border border-line bg-canvas/35 p-3">
                  <p class="mono-label">Стоимость</p>
                  <p class="mt-1.5 font-mono text-sm font-bold">{{ formatTokens(game.game.cost.tokens) }}</p>
                </div>
                <div class="rounded-xl border border-line bg-canvas/35 p-3">
                  <p class="mono-label">Вызовы</p>
                  <p class="mt-1.5 font-mono text-sm font-bold">{{ game.game.cost.calls }}</p>
                </div>
              </div>
            </section>

            <section v-if="game.isReady" class="panel p-5">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="mono-label">Оценка</p>
                  <h2 class="mt-1 font-display text-lg font-bold">Как игра?</h2>
                </div>
                <RatingBadge :average="game.rating.average" :total="game.rating.total" />
              </div>

              <div class="mt-4 flex items-end gap-3 rounded-2xl border border-line bg-canvas/35 p-4">
                <p class="font-display text-4xl font-bold leading-none">
                  {{ game.rating.total ? game.rating.average.toFixed(1).replace('.', ',') : '—' }}
                </p>
                <div class="pb-0.5">
                  <div class="flex gap-0.5">
                    <span
                      v-for="star in 5"
                      :key="star"
                      class="size-2 rounded-full"
                      :class="star <= Math.round(game.rating.average) ? averageTone.solid : 'bg-line-strong'"
                    />
                  </div>
                  <p class="mt-1.5 font-mono text-[9px] uppercase tracking-[0.09em] text-faint">
                    {{ game.rating.total }}
                    {{ plural(game.rating.total, 'оценка', 'оценки', 'оценок') }}
                  </p>
                </div>
              </div>

              <div class="mt-4 space-y-2">
                <div v-for="row in ratingRows" :key="row.star" class="flex items-center gap-2.5">
                  <span class="w-3 text-right font-mono text-[9px] text-faint">{{ row.star }}</span>
                  <div class="h-1.5 flex-1 overflow-hidden rounded-full bg-line">
                    <div
                      class="h-full rounded-full transition-all duration-500"
                      :class="row.tone.solid"
                      :style="{ width: `${(row.count / maxRatingCount) * 100}%` }"
                    />
                  </div>
                  <span class="w-5 text-right font-mono text-[9px] tabular-nums text-faint">{{ row.count }}</span>
                </div>
              </div>

              <div class="mt-5 border-t border-line pt-4">
                <p class="mono-label mb-2">Ваша оценка</p>
                <div class="flex items-center gap-3">
                  <div class="rating-choice" :style="{ '--rating-color': ratingInputColor }">
                    <Rating
                      :key="ratingKey"
                      :model-value="vote"
                      @update:model-value="onRatingInput"
                    />
                  </div>
                  <span v-if="game.myRating" class="font-mono text-[10px] text-muted">{{ game.myRating }} / 5</span>
                </div>
                <p v-if="voteError" class="mt-2 text-[10px] text-red-500">{{ voteError }}</p>
                <p class="mt-2 text-[10px] leading-4 text-faint">Можно изменить оценку в любой момент.</p>
              </div>
            </section>

            <section v-if="showMaintenance && game.config" class="panel p-5">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="mono-label">Тюнинг</p>
                  <h2 class="mt-1 font-display text-lg font-bold">Настройки</h2>
                </div>
                <AppIcon name="settings" :size="18" class="text-muted" />
              </div>

              <div class="mt-4 space-y-4">
                <div>
                  <div class="mb-2 flex items-center justify-between">
                    <label class="mono-label">Темп</label>
                    <span class="font-mono text-[10px] font-bold">{{ configForm.tempo }}/10</span>
                  </div>
                  <input
                    v-model.number="configForm.tempo"
                    type="range"
                    min="1"
                    max="10"
                    step="1"
                    class="w-full accent-[var(--bv-yellow)]"
                    aria-label="Темп игры"
                  />
                </div>

                <div>
                  <label class="mono-label mb-2 block">Размер карты</label>
                  <Select v-model="configForm.mapSize" :options="mapOptions" option-label="label" option-value="value" />
                </div>

                <div>
                  <label class="mono-label mb-2 block">Сложность</label>
                  <Select v-model="configForm.difficulty" :options="difficultyOptions" option-label="label" option-value="value" />
                </div>

                <div>
                  <label class="mono-label mb-2 block">Палитра</label>
                  <div class="flex gap-2">
                    <label
                      v-for="(_color, index) in configForm.palette"
                      :key="index"
                      class="relative size-8 cursor-pointer overflow-hidden rounded-full border border-line-strong"
                      :style="{ backgroundColor: configForm.palette[index] }"
                      :title="configForm.palette[index]"
                    >
                      <input v-model="configForm.palette[index]" type="color" class="absolute inset-0 cursor-pointer opacity-0" />
                    </label>
                  </div>
                </div>
              </div>

              <Button class="p-button-secondary mt-5 w-full" :loading="savingConfig" @click="saveConfig">
                Сохранить настройки
              </Button>
            </section>

            <section class="panel p-5">
              <div class="flex items-center gap-2.5">
                <span class="grid size-9 place-items-center rounded-xl bg-bee/15 text-bee-deep dark:text-bee">
                  <AppIcon name="shield" :size="17" />
                </span>
                <div>
                  <p class="text-xs font-bold">Рабочее пространство</p>
                  <p class="mt-0.5 text-[10px] text-faint">доступно только владельцу</p>
                </div>
              </div>
              <div class="mt-4 space-y-2 border-t border-line pt-3 text-[10px] text-faint">
                <p class="flex items-center justify-between gap-2">
                  <span>Версия</span>
                  <span class="font-mono text-muted">#{{ latestVersion?.number ?? '—' }}</span>
                </p>
                <p class="flex items-center justify-between gap-2">
                  <span>Починок</span>
                  <span class="font-mono text-muted">{{ game.game.cost.fixes }}</span>
                </p>
                <p class="flex items-center justify-between gap-2">
                  <span>Промпт</span>
                  <span class="font-mono text-muted">{{ promptText.length }} симв.</span>
                </p>
              </div>
            </section>
          </aside>
        </div>
      </template>
    </div>

    <Dialog
      v-model:visible="fixOpen"
      modal
      header="Починить игру"
      :style="{ width: 'min(92vw, 480px)' }"
      :draggable="false"
    >
      <p class="text-xs leading-5 text-muted">
        Опишите, что сломалось. Если оставить поле пустым, агент попробует восстановить игру по
        коду и последним ошибкам.
      </p>
      <Textarea
        v-model="fixText"
        rows="4"
        maxlength="2000"
        class="mt-4 text-[13px]"
        placeholder="Например: после победы не перезапускается уровень, кнопка рестарта не реагирует…"
      />
      <template #footer>
        <div class="flex justify-end gap-2">
          <Button class="p-button-secondary" @click="fixOpen = false">Отмена</Button>
          <Button class="bee-button" :loading="game.actionPending === 'fix'" @click="repair">
            <AppIcon name="refresh" :size="14" />
            <span class="ml-1.5">Запустить починку</span>
          </Button>
        </div>
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.rating-choice :deep(.p-rating-option-active),
.rating-choice :deep(.p-rating-option-active .p-rating-icon) {
  color: var(--rating-color);
}
</style>
