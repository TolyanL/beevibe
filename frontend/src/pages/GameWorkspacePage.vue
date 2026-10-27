<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Button from 'openvue/button'
import InputText from 'openvue/inputtext'
import Select from 'openvue/select'
import Textarea from 'openvue/textarea'
import { useConfirm } from 'openvue/useconfirm'
import { useToast } from 'openvue/usetoast'
import { messageFromError } from '@/api/client'
import type { GameVersion } from '@/api/types'
import { formatDate, formatTokens, plural } from '@/lib/format'
import {
  gameDescription as gameDescriptionOf,
  genreLabel,
  mapSizeLabel,
  shortGameTitle,
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
import AiProgress from '@/components/ui/AiProgress.vue'
import GameFrame from '@/components/game/GameFrame.vue'
import StageTimeline from '@/components/game/StageTimeline.vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const confirm = useConfirm()
const auth = useAuthStore()
const game = useGameStore()

const id = computed(() => String(route.params.id ?? ''))
const promptText = computed(
  () =>
    auth.profile?.games.find((item) => item.id === id.value)?.prompt ??
    `Игра ${id.value.slice(0, 8)}`,
)
const profileGame = computed(() => auth.profile?.games.find((item) => item.id === id.value))
const gameTitle = computed(() => shortGameTitle(game.game?.title, promptText.value))
const gameDescription = computed(() => gameDescriptionOf(game.game?.description, promptText.value))
const versions = computed(() =>
  [...(game.game?.versions ?? [])].sort((a, b) => b.number - a.number),
)
const latestVersion = computed(() => versions.value[0])
const meta = computed(() => statusMeta(game.status))
const metaForm = reactive({ title: '', description: '' })
const metaSaving = ref(false)
const metaGenerating = ref(false)

const coverGradient = computed(() => {
  const palette = game.game?.cover?.palette ?? game.spec?.palette ?? []

  if (palette.length >= 2) {
    return { background: `linear-gradient(135deg, ${palette[0]}, ${palette[palette.length - 1]})` }
  }

  return { background: 'linear-gradient(135deg, var(--bv-yellow), var(--bv-ink))' }
})
const isPublic = computed(() => game.game?.visibility === 'public')
const operation = computed(() => game.operation)
const operationLabel = computed(() => {
  switch (game.operation?.kind) {
    case 'fix':
      return 'Чиним игру'
    case 'config':
      return 'Применяем настройки'
    case 'meta':
      return 'Придумываем название'
    default:
      return 'Применяем правки'
  }
})

const operationHint = computed(() => {
  switch (game.operation?.kind) {
    case 'meta':
      return 'Модель пишет название, описание и обложку'
    case 'fix':
      return 'Модель ищет и чинит ошибку'
    default:
      return 'Модель переписывает игру'
  }
})

const operationExpectedSeconds = computed(() => (game.operation?.kind === 'meta' ? 45 : 150))
const isMetaOperation = computed(() => game.operation?.kind === 'meta')
const reviewIssues = computed(() => game.game?.review_issues ?? [])
const needsDecision = computed(() => game.status === 'needs_decision')

function stageTitle(stage: string): string {
  switch (stage) {
    case 'assembling':
      return 'сборки'
    case 'coding':
      return 'кода'
    case 'assets':
      return 'ассетов'
    default:
      return 'конвейера'
  }
}

async function acceptReview(): Promise<void> {
  try {
    const resumeFrom = await game.acceptReviewDecision()
    toast.add({
      severity: 'success',
      summary: 'Код принят',
      detail: `Продолжаем с этапа «${stageLabel(resumeFrom || 'assembling')}» — перегенерации не будет.`,
      life: 3400,
    })
  } catch (err) {
    toast.add({
      severity: 'error',
      summary: 'Не удалось продолжить',
      detail: messageFromError(err),
      life: 3200,
    })
  }
}

function regenerateCode(): void {
  confirm.require({
    header: 'Сгенерировать код заново?',
    message:
      'Спека, план и ассеты сохранятся — заново пойдёт только этап кода. Это займёт время, но не полный проход конвейера.',
    acceptLabel: 'Перегенерировать код',
    rejectLabel: 'Оставить как есть',
    accept: async () => {
      try {
        const resumeFrom = await game.regenerateCode()
        toast.add({
          severity: 'success',
          summary: 'Перегенерация запущена',
          detail: `Продолжаем с этапа «${stageLabel(resumeFrom || 'coding')}».`,
          life: 3400,
        })
      } catch (err) {
        toast.add({
          severity: 'error',
          summary: 'Не удалось запустить',
          detail: messageFromError(err),
          life: 3200,
        })
      }
    },
  })
}

type EditTab = 'patch' | 'config' | 'fix'

const patchText = ref('')
const patchError = ref('')
const fixText = ref('')
const savingConfig = ref(false)
const copied = ref(false)
const editTab = ref<EditTab>('patch')
const session = usePlaytime(
  () => id.value,
  () => game.isReady,
)
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

const averageTone = computed(() => ratingTone(game.rating.average))

const ratingRows = computed(() =>
  [5, 4, 3, 2, 1].map((star) => {
    const count = game.rating.counts?.[star - 1] ?? 0
    const percent = game.rating.total ? (count / game.rating.total) * 100 : 0
    return { star, count, percent, tone: ratingTone(star) }
  }),
)
const maxRatingCount = computed(() => Math.max(1, ...ratingRows.value.map((row) => row.count)))
const canPatch = computed(
  () => patchText.value.trim().length >= 3 && !game.actionPending && !game.operation,
)
const showMaintenance = computed(() => game.isReady && !game.loading)
const editTabs = computed<
  { value: EditTab; label: string; icon: 'wand' | 'settings' | 'refresh' }[]
>(() => [
  { value: 'patch', label: 'Патч', icon: 'wand' },
  ...(game.config
    ? [{ value: 'config' as const, label: 'Настройки', icon: 'settings' as const }]
    : []),
  { value: 'fix', label: 'Починка', icon: 'refresh' },
])

watch(
  () => game.game?.title,
  (title) => {
    metaForm.title = title ?? ''
  },
  { immediate: true },
)

watch(
  () => game.game?.description,
  (description) => {
    metaForm.description = description ?? ''
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

function logLine(stage: {
  stage: string
  status: string
  durationMs: number
  operation?: string
}): string {
  const time = stage.durationMs ? ` ${(stage.durationMs / 1000).toFixed(1)}s` : ''
  const mark = stage.status === 'done' ? 'ok' : stage.status === 'error' ? '!!' : '..'

  return `[${mark}] ${stageLabel(stage.stage, stage.operation)}${time}`
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
    toast.add({
      severity: 'success',
      summary: 'Правка принята',
      detail: 'Собираем новую версию — прогресс в блоке сборки.',
      life: 3000,
    })
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
    toast.add({
      severity: 'error',
      summary: 'Не удалось сохранить',
      detail: messageFromError(err),
      life: 3200,
    })
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
        toast.add({
          severity: 'error',
          summary: 'Откат не удался',
          detail: messageFromError(err),
          life: 3200,
        })
      }
    },
  })
}

function cancelRun(): void {
  confirm.require({
    header: 'Отменить сборку?',
    message:
      'Задача остановится и игра вернётся в готовое состояние. Токены вернутся, только если модель ещё не ответила.',
    acceptLabel: 'Отменить сборку',
    rejectLabel: 'Подождать',
    accept: async () => {
      try {
        await game.cancel()
        toast.add({ severity: 'info', summary: 'Задача отменена', life: 2400 })
      } catch (err) {
        toast.add({
          severity: 'error',
          summary: 'Не удалось отменить',
          detail: messageFromError(err),
          life: 3200,
        })
      }
    },
  })
}

async function saveMeta(): Promise<void> {
  metaSaving.value = true
  try {
    await game.saveMeta({ title: metaForm.title, description: metaForm.description })
    toast.add({ severity: 'success', summary: 'Название и описание сохранены', life: 2400 })
  } catch (err) {
    toast.add({
      severity: 'error',
      summary: 'Не удалось сохранить',
      detail: messageFromError(err),
      life: 3200,
    })
  } finally {
    metaSaving.value = false
  }
}

function generateMeta(): void {
  confirm.require({
    header: 'Сгенерировать название и описание?',
    message: `Модель придумает название, описание и обложку. Спишется ${formatTokens(auth.prices.meta)} токенов.`,
    acceptLabel: 'Сгенерировать',
    rejectLabel: 'Отмена',
    accept: async () => {
      metaGenerating.value = true
      try {
        await game.generateMeta()
        toast.add({
          severity: 'success',
          summary: 'Генерация меты запущена',
          detail: 'Результат появится в панели публикации.',
          life: 2800,
        })
      } catch (err) {
        toast.add({
          severity: 'error',
          summary: 'Не удалось запустить',
          detail: messageFromError(err),
          life: 3200,
        })
      } finally {
        metaGenerating.value = false
      }
    },
  })
}

function toggleVisibility(): void {
  const next = isPublic.value ? 'private' : 'public'

  confirm.require({
    header: next === 'public' ? 'Опубликовать игру?' : 'Скрыть игру из каталога?',
    message:
      next === 'public'
        ? 'Игра появится в каталоге и станет доступна по публичной ссылке.'
        : 'Игра пропадёт из каталога, публичная ссылка перестанет открываться.',
    acceptLabel: next === 'public' ? 'Опубликовать' : 'Скрыть',
    rejectLabel: 'Отмена',
    accept: async () => {
      try {
        await game.setVisibility(next)
        toast.add({
          severity: 'success',
          summary: next === 'public' ? 'Игра опубликована' : 'Игра скрыта',
          life: 2400,
        })
      } catch (err) {
        toast.add({
          severity: 'error',
          summary: 'Не удалось изменить видимость',
          detail: messageFromError(err),
          life: 3200,
        })
      }
    },
  })
}

function removeGame(): void {
  confirm.require({
    header: 'Удалить игру?',
    message: 'Игра исчезнет из профиля и каталога. Файлы и версии останутся в архиве.',
    acceptLabel: 'Удалить',
    rejectLabel: 'Отмена',
    accept: async () => {
      try {
        await game.remove()
        toast.add({ severity: 'success', summary: 'Игра удалена', life: 2400 })
        await router.push({ name: 'profile' })
      } catch (err) {
        toast.add({
          severity: 'error',
          summary: 'Не удалось удалить',
          detail: messageFromError(err),
          life: 3200,
        })
      }
    },
  })
}

async function retryGame(): Promise<void> {
  try {
    const resumeFrom = await game.retry()
    const label = resumeFrom ? stageLabel(resumeFrom) : ''

    toast.add({
      severity: 'success',
      summary: 'Сборка перезапущена',
      detail: label
        ? `Продолжаем с этапа «${label}» — пройденные шаги сохранены.`
        : 'Конвейер начнёт с первого этапа.',
      life: 3200,
    })
  } catch (err) {
    toast.add({
      severity: 'error',
      summary: 'Не удалось перезапустить',
      detail: messageFromError(err),
      life: 3200,
    })
  }
}

function openFix(): void {
  fixText.value = ''
  openEdit('fix')
}

function openEdit(tab: EditTab): void {
  editTab.value = tab
  document.querySelector('#edits')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

async function repair(): Promise<void> {
  try {
    await game.repair(fixText.value.trim() || undefined)
    fixText.value = ''
    toast.add({
      severity: 'success',
      summary: 'Починка принята',
      detail: 'Собираем новую версию — прогресс в блоке сборки.',
      life: 3000,
    })
  } catch (err) {
    toast.add({
      severity: 'error',
      summary: 'Починка не запустилась',
      detail: messageFromError(err),
      life: 3200,
    })
  }
}

function useSuggestion(text: string): void {
  patchText.value = text
  openEdit('patch')
}

let poll: ReturnType<typeof setInterval> | null = null
let lastVersionNumber = 0

function pollDelay(): number {
  return game.operation ? 2_500 : 15_000
}

function schedulePoll(): void {
  if (poll) clearInterval(poll)

  poll = setInterval(() => {
    if (game.loading || (!game.isActive && !game.operation)) return
    void game.refresh()
  }, pollDelay())
}

watch(
  () => Boolean(game.operation),
  () => schedulePoll(),
)

watch(
  () => latestVersion.value?.number ?? 0,
  (number) => {
    if (!number) return
    if (lastVersionNumber && number > lastVersionNumber) {
      toast.add({ severity: 'success', summary: `Версия #${number} готова`, life: 3000 })
    }
    lastVersionNumber = number
  },
)

watch(
  () => game.streamError,
  (message) => {
    if (!message) return
    toast.add({ severity: 'error', summary: 'Правка не применилась', detail: message, life: 4000 })
  },
)

onMounted(() => {
  void load()
  lastVersionNumber = 0
  schedulePoll()
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
          <span class="font-mono text-[9px] uppercase tracking-[0.13em] text-faint"
            >owner workspace</span
          >
          <span class="size-0.5 rounded-full bg-faint" />
          <button
            type="button"
            class="font-mono text-[9px] text-faint transition hover:text-ink"
            @click="copyId"
          >
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
          :text="
            game.error || 'Возможно, игра удалена или у вас нет доступа к рабочему пространству.'
          "
        >
          <div class="flex flex-wrap justify-center gap-2">
            <Button class="bee-button" @click="load">Повторить</Button>
            <RouterLink
              to="/profile"
              class="rounded-full border border-line-strong px-5 py-2.5 text-xs font-bold"
            >
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
              <span
                v-if="operation"
                class="inline-flex items-center gap-1.5 rounded-full border border-bee-deep/30 bg-bee/10 px-2.5 py-1 text-[10px] font-bold text-bee-deep dark:text-bee"
              >
                <span class="size-2 animate-pulse-soft rounded-full bg-bee" />
                {{ operationLabel }}…
              </span>
              <span class="font-mono text-[9px] uppercase tracking-[0.12em] text-faint">
                {{ genreLabel(profileGame?.genre ?? game.spec?.genre ?? '') }}
                <span v-if="profileGame?.created_at">
                  · {{ formatDate(profileGame.created_at) }}</span
                >
              </span>
            </div>
            <h1
              class="mt-3 max-w-4xl font-display text-2xl font-bold leading-tight tracking-[-0.04em] sm:text-3xl lg:text-[2.2rem]"
            >
              {{ gameTitle }}
            </h1>
            <p
              v-if="gameDescription && gameDescription !== gameTitle"
              class="mt-2 max-w-3xl text-[13px] leading-6 text-muted"
            >
              {{ gameDescription }}
            </p>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <div
              v-if="game.isReady"
              class="flex items-center gap-1.5 rounded-full border border-line bg-surface/70 px-3 py-2 font-mono text-[10px] text-muted"
            >
              <AppIcon name="clock" :size="13" />
              сессия {{ sessionLabel(sessionSeconds) }}
            </div>
            <RouterLink
              v-if="game.isReady"
              :to="{ name: 'play', params: { id: game.game?.slug || id } }"
              target="_blank"
              class="inline-flex items-center gap-1.5 rounded-full border border-line-strong px-3.5 py-2 text-[11px] font-bold transition hover:border-bee-deep hover:bg-bee/10"
            >
              <AppIcon name="external" :size="13" />
              Публичная страница
            </RouterLink>
            <button
              v-if="game.isReady"
              type="button"
              class="inline-flex items-center gap-1.5 rounded-full border px-3.5 py-2 text-[11px] font-bold transition"
              :class="
                isPublic
                  ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-500 hover:bg-emerald-500/15'
                  : 'border-line-strong text-muted hover:border-bee-deep hover:text-ink'
              "
              :disabled="Boolean(game.actionPending)"
              @click="toggleVisibility"
            >
              <AppIcon :name="isPublic ? 'external' : 'shield'" :size="13" />
              {{ isPublic ? 'В каталоге' : 'Скрыта' }}
            </button>
            <Button
              v-if="game.isReady"
              class="p-button-secondary"
              :loading="game.actionPending === 'fix'"
              @click="openFix"
            >
              <AppIcon name="refresh" :size="14" />
              <span class="ml-1.5">Починить</span>
            </Button>
          </div>
        </header>

        <div class="grid items-start gap-5 xl:grid-cols-[minmax(0,1fr)_350px]">
          <div class="min-w-0 space-y-5">
            <div v-if="promptText" class="relative animate-rise">
              <GameFrame
                :game-id="id"
                :slug="game.game?.slug"
                :status="game.status"
                :version-token="game.refreshKey"
                @loaded="session.flush()"
              />

              <div
                v-if="operation && !isMetaOperation"
                class="pointer-events-none absolute inset-x-0 bottom-0 flex justify-center p-3"
              >
                <span
                  class="inline-flex items-center gap-2 rounded-full border border-bee-deep/30 bg-canvas/95 px-3.5 py-2 text-[11px] font-bold text-ink shadow-lift backdrop-blur"
                >
                  <span class="size-2.5 animate-pulse-soft rounded-full bg-bee-deep" />
                  {{ operationLabel }} — собираем новую версию
                </span>
              </div>
            </div>

            <div
              v-if="game.streamError"
              class="flex items-start gap-2.5 rounded-2xl border border-orange-500/20 bg-orange-500/10 px-4 py-3 text-xs text-orange-500"
            >
              <AppIcon name="alert" :size="16" class="mt-0.5" />
              {{ game.streamError }}
            </div>

            <div v-if="needsDecision" class="panel border-orange-500/25 p-5 sm:p-6">
              <div class="flex items-start gap-3">
                <span
                  class="grid size-10 shrink-0 place-items-center rounded-xl bg-orange-500/10 text-orange-500"
                >
                  <AppIcon name="alert" :size="20" />
                </span>
                <div class="flex-1">
                  <h2 class="font-display text-lg font-bold">Ревью нашло замечания</h2>
                  <p class="mt-1.5 text-xs leading-5 text-muted">
                    Конвейер остановился, чтобы не тратить время на перегенерацию без вашего
                    согласия. Оставьте код как есть — соберём версию из него, — или перегенерируйте
                    только код.
                  </p>

                  <ul v-if="reviewIssues.length" class="mt-3 space-y-1.5">
                    <li
                      v-for="(issue, index) in reviewIssues"
                      :key="`${issue.code}-${index}`"
                      class="flex gap-2 rounded-xl border border-line bg-canvas/40 px-3 py-2 text-[11px] leading-5 text-muted"
                    >
                      <span class="font-mono text-[9px] uppercase tracking-[0.08em] text-faint">{{
                        issue.code
                      }}</span>
                      <span class="flex-1">{{ issue.message }}</span>
                      <span v-if="issue.line" class="font-mono text-[9px] text-faint"
                        >:{{ issue.line }}</span
                      >
                    </li>
                  </ul>

                  <div class="mt-4 flex flex-wrap gap-2">
                    <Button
                      class="bee-button"
                      :loading="game.actionPending === 'review'"
                      @click="acceptReview"
                    >
                      <AppIcon name="check" :size="14" />
                      <span class="ml-1.5">Оставить как есть</span>
                    </Button>
                    <Button
                      class="p-button-secondary"
                      :disabled="Boolean(game.actionPending)"
                      @click="regenerateCode"
                    >
                      <AppIcon name="refresh" :size="14" />
                      <span class="ml-1.5">Сгенерировать код заново</span>
                    </Button>
                  </div>

                  <p class="mt-3 text-[10px] leading-4 text-faint">
                    «Оставить как есть» продолжает с этапа {{ stageTitle('assembling') }} — быстро и
                    без новых вызовов модели. Перегенерация начинается с этапа кода.
                  </p>
                </div>
              </div>
            </div>

            <div v-if="game.isFailed" class="panel border-red-500/20 p-5 sm:p-6">
              <div class="flex items-start gap-3">
                <span
                  class="grid size-10 shrink-0 place-items-center rounded-xl bg-red-500/10 text-red-500"
                >
                  <AppIcon name="alert" :size="20" />
                </span>
                <div class="flex-1">
                  <h2 class="font-display text-lg font-bold">Конвейер остановился</h2>
                  <p class="mt-1.5 text-xs leading-5 text-muted">
                    Игра не собралась. Запустите починку — агент получит контекст ошибки и попробует
                    собрать новую версию.
                  </p>
                  <div class="mt-4 flex flex-wrap gap-2">
                    <Button
                      class="bee-button"
                      :loading="game.actionPending === 'fix'"
                      @click="openFix"
                    >
                      <AppIcon name="refresh" :size="14" />
                      <span class="ml-1.5">Починить игру</span>
                    </Button>
                    <Button
                      class="p-button-secondary"
                      :loading="game.actionPending === 'retry'"
                      @click="retryGame"
                    >
                      <AppIcon name="refresh" :size="14" />
                      <span class="ml-1.5">Перезапустить сборку</span>
                    </Button>
                    <RouterLink
                      :to="{ name: 'create' }"
                      class="rounded-full border border-line-strong px-4 py-2.5 text-xs font-bold"
                    >
                      Начать заново
                    </RouterLink>
                  </div>
                </div>
              </div>
            </div>

            <AiProgress
              v-if="game.isGenerating && !operation"
              :started-at="profileGame?.created_at"
              label="Нейросеть собирает игру"
              hint="этапы конвейера ниже"
              :expected-seconds="240"
            />

            <div v-if="game.isActive" class="panel overflow-hidden">
              <div class="flex items-center justify-between gap-3 border-b border-line px-5 py-4">
                <div>
                  <p class="mono-label">Live build</p>
                  <h2 class="mt-1 font-display text-lg font-bold">
                    {{ game.isGenerating ? 'Сборка в реальном времени' : 'Правка в работе' }}
                  </h2>
                </div>
                <div class="flex items-center gap-2">
                  <span
                    class="flex items-center gap-1.5 font-mono text-[9px] uppercase tracking-[0.1em] text-bee-deep dark:text-bee"
                  >
                    <span class="size-1.5 animate-pulse-soft rounded-full bg-bee" />
                    {{ game.streaming ? 'stream' : 'poll' }}
                  </span>
                  <button
                    type="button"
                    class="rounded-full border border-line-strong px-2.5 py-1 text-[10px] font-bold text-muted transition hover:border-bad/40 hover:text-bad disabled:opacity-50"
                    :disabled="Boolean(game.actionPending)"
                    @click="cancelRun"
                  >
                    Отменить
                  </button>
                </div>
              </div>

              <div class="grid gap-0 md:grid-cols-[minmax(0,.9fr)_minmax(0,1.1fr)]">
                <div class="border-b border-line p-5 md:border-b-0 md:border-r">
                  <p class="mono-label">{{ game.isGenerating ? 'Этапы' : 'Операция' }}</p>
                  <p class="mt-2 text-sm leading-6 text-muted">{{ meta.description }}</p>

                  <div v-if="game.isGenerating" class="mt-4 grid grid-cols-3 gap-2">
                    <div class="rounded-xl border border-line bg-canvas/45 p-3">
                      <p class="font-display text-lg font-bold">{{ game.stageProgress }}%</p>
                      <p class="mt-0.5 font-mono text-[8px] uppercase tracking-[0.1em] text-faint">
                        прогресс
                      </p>
                    </div>
                    <div class="rounded-xl border border-line bg-canvas/45 p-3">
                      <p class="font-display text-lg font-bold">
                        {{ game.pipelineStages.filter((item) => item.status === 'done').length }}/8
                      </p>
                      <p class="mt-0.5 font-mono text-[8px] uppercase tracking-[0.1em] text-faint">
                        этапов
                      </p>
                    </div>
                    <div class="rounded-xl border border-line bg-canvas/45 p-3">
                      <p class="font-display text-lg font-bold">
                        {{ game.stages.filter((item) => item.issues).length }}
                      </p>
                      <p class="mt-0.5 font-mono text-[8px] uppercase tracking-[0.1em] text-faint">
                        issues
                      </p>
                    </div>
                  </div>

                  <div v-else class="mt-4 grid grid-cols-3 gap-2">
                    <div class="rounded-xl border border-line bg-canvas/45 p-3">
                      <p class="font-display text-lg font-bold">{{ operationLabel }}</p>
                      <p class="mt-0.5 font-mono text-[8px] uppercase tracking-[0.1em] text-faint">
                        операция
                      </p>
                    </div>
                    <div class="rounded-xl border border-line bg-canvas/45 p-3">
                      <p class="font-display text-lg font-bold">{{ versions.length }}</p>
                      <p class="mt-0.5 font-mono text-[8px] uppercase tracking-[0.1em] text-faint">
                        версий
                      </p>
                    </div>
                    <div class="rounded-xl border border-line bg-canvas/45 p-3">
                      <p class="font-display text-lg font-bold">
                        {{ game.actionPending ? '✓' : '…' }}
                      </p>
                      <p class="mt-0.5 font-mono text-[8px] uppercase tracking-[0.1em] text-faint">
                        {{ game.streaming ? 'подписка' : 'опрос' }}
                      </p>
                    </div>
                  </div>
                </div>

                <div class="p-4">
                  <AiProgress
                    v-if="operation && !isMetaOperation"
                    class="mb-3"
                    compact
                    :started-at="operation.started_at"
                    :label="operationLabel"
                    :hint="operationHint"
                    :expected-seconds="operationExpectedSeconds"
                  />

                  <div class="terminal h-[210px] p-4">
                    <p><span class="tok-comment">// beevibe pipeline</span></p>
                    <p
                      v-for="stage in game.stages"
                      :key="`${stage.stage}-${stage.operation ?? ''}`"
                    >
                      <span
                        :class="
                          stage.status === 'done'
                            ? 'tok-string'
                            : stage.status === 'error'
                              ? 'tok-key'
                              : 'tok-key'
                        "
                      >
                        {{ logLine(stage) }}
                      </span>
                    </p>
                    <p v-if="!game.stages.length" class="tok-comment">ожидаем первый этап…</p>
                    <p class="mt-1 flex items-center gap-1 text-bee">
                      <span>$</span><span class="animate-pulse-soft">_</span>
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="game.isGenerating && !game.spec" class="panel p-5">
              <div class="flex items-center gap-3">
                <span
                  class="size-4 animate-spin rounded-full border-2 border-line border-t-bee-deep"
                />
                <p class="text-xs text-muted">Спецификация появится после этапа планирования.</p>
              </div>
            </div>

            <div v-if="showMaintenance" id="edits" class="panel overflow-hidden">
              <div
                class="flex flex-wrap items-start justify-between gap-3 border-b border-line px-5 py-4"
              >
                <div>
                  <p class="mono-label">Правки</p>
                  <h2 class="mt-1 font-display text-xl font-bold">Изменить игру</h2>
                </div>
                <span class="font-mono text-[9px] uppercase tracking-[0.1em] text-faint">
                  патч · {{ formatTokens(auth.prices.patch) }} токенов
                </span>
              </div>

              <div class="flex flex-wrap items-center gap-1 border-b border-line px-4 py-2.5">
                <button
                  v-for="tab in editTabs"
                  :key="tab.value"
                  type="button"
                  class="inline-flex items-center gap-1.5 rounded-full px-3.5 py-2 text-[11px] font-bold transition"
                  :class="
                    editTab === tab.value
                      ? 'bg-ink text-canvas shadow-sm'
                      : 'text-muted hover:bg-hover-surface hover:text-ink'
                  "
                  @click="editTab = tab.value"
                >
                  <AppIcon :name="tab.icon" :size="13" />
                  {{ tab.label }}
                </button>
              </div>

              <div v-show="editTab === 'patch'" class="p-4 sm:p-5">
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
                    Патч создаёт новую версию. Старые можно вернуть в списке версий ниже.
                  </p>
                  <Button
                    class="bee-button"
                    :loading="Boolean(operation) || game.actionPending === 'patch'"
                    :disabled="!canPatch"
                    @click="sendPatch"
                  >
                    <AppIcon name="wand" :size="15" />
                    <span class="ml-1.5">Применить патч</span>
                  </Button>
                </div>
              </div>

              <div v-show="editTab === 'config'" class="p-4 sm:p-5">
                <div v-if="game.config" class="grid gap-4 sm:grid-cols-2">
                  <div class="sm:col-span-2">
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
                    <Select
                      v-model="configForm.mapSize"
                      :options="mapOptions"
                      option-label="label"
                      option-value="value"
                    />
                  </div>

                  <div>
                    <label class="mono-label mb-2 block">Сложность</label>
                    <Select
                      v-model="configForm.difficulty"
                      :options="difficultyOptions"
                      option-label="label"
                      option-value="value"
                    />
                  </div>

                  <div class="sm:col-span-2">
                    <label class="mono-label mb-2 block">Палитра</label>
                    <div class="flex gap-2">
                      <label
                        v-for="(_color, index) in configForm.palette"
                        :key="index"
                        class="relative size-8 cursor-pointer overflow-hidden rounded-full border border-line-strong"
                        :style="{ backgroundColor: configForm.palette[index] }"
                        :title="configForm.palette[index]"
                      >
                        <input
                          v-model="configForm.palette[index]"
                          type="color"
                          class="absolute inset-0 cursor-pointer opacity-0"
                        />
                      </label>
                    </div>
                  </div>
                </div>

                <div
                  class="mt-4 flex flex-wrap items-center justify-between gap-3 border-t border-line pt-4"
                >
                  <p class="text-[10px] leading-4 text-faint">
                    Тюнинг применяется к игре без вызова LLM и создаёт новую версию.
                  </p>
                  <Button
                    class="p-button-secondary"
                    :loading="savingConfig"
                    :disabled="!game.config"
                    @click="saveConfig"
                  >
                    Сохранить настройки
                  </Button>
                </div>
              </div>

              <div v-show="editTab === 'fix'" class="p-4 sm:p-5">
                <p class="text-xs leading-5 text-muted">
                  Опишите, что сломалось. Если оставить поле пустым, агент попробует восстановить
                  игру по коду и последним ошибкам.
                </p>
                <Textarea
                  v-model="fixText"
                  rows="4"
                  maxlength="2000"
                  class="mt-3 text-[13px]"
                  placeholder="Например: после победы не перезапускается уровень, кнопка рестарта не реагирует…"
                />
                <div class="mt-4 flex flex-wrap items-center justify-between gap-3">
                  <p class="text-[10px] leading-4 text-faint">Починка тоже создаёт новую версию.</p>
                  <Button
                    class="bee-button"
                    :loading="Boolean(operation) || game.actionPending === 'fix'"
                    :disabled="Boolean(operation)"
                    @click="repair"
                  >
                    <AppIcon name="refresh" :size="14" />
                    <span class="ml-1.5">Запустить починку</span>
                  </Button>
                </div>
              </div>
            </div>

            <div v-if="showMaintenance" class="panel overflow-hidden">
              <div
                class="flex flex-wrap items-start justify-between gap-3 border-b border-line px-5 py-4"
              >
                <div>
                  <p class="mono-label">Публикация</p>
                  <h2 class="mt-1 font-display text-xl font-bold">Название и обложка</h2>
                </div>
                <span class="font-mono text-[9px] uppercase tracking-[0.1em] text-faint">
                  мета · {{ formatTokens(auth.prices.meta) }} токенов
                </span>
              </div>

              <div v-if="isMetaOperation" class="px-4 pt-3 sm:px-5">
                <AiProgress
                  compact
                  :started-at="operation?.started_at"
                  label="Нейросеть придумывает название"
                  :hint="operationHint"
                  :expected-seconds="operationExpectedSeconds"
                />
              </div>

              <div class="grid gap-4 p-4 sm:grid-cols-[140px_minmax(0,1fr)] sm:p-5">
                <div class="overflow-hidden rounded-2xl border border-line">
                  <div class="aspect-square" :style="coverGradient" />
                  <p
                    class="px-2.5 py-2 font-mono text-[9px] uppercase tracking-[0.08em] text-faint"
                  >
                    {{ game.game?.cover?.style ?? game.spec?.style ?? 'градиент' }}
                  </p>
                </div>

                <div class="space-y-3">
                  <div>
                    <label class="mono-label mb-1.5 block">Название</label>
                    <InputText
                      v-model="metaForm.title"
                      maxlength="80"
                      class="w-full"
                      placeholder="Например: Кристаллы"
                    />
                  </div>
                  <div>
                    <label class="mono-label mb-1.5 block">Описание</label>
                    <Textarea
                      v-model="metaForm.description"
                      rows="3"
                      maxlength="500"
                      class="text-[13px]"
                    />
                  </div>
                </div>
              </div>

              <div
                class="flex flex-wrap items-center justify-between gap-3 border-t border-line px-4 py-3 sm:px-5"
              >
                <p class="text-[10px] leading-4 text-faint">
                  Ручная правка бесплатна. В каталоге и на публичной странице показывается название,
                  пока его нет — промпт.
                </p>
                <div class="flex flex-wrap gap-2">
                  <Button
                    class="p-button-secondary"
                    :loading="metaGenerating || isMetaOperation"
                    :disabled="Boolean(operation)"
                    @click="generateMeta"
                  >
                    <AppIcon name="spark" :size="14" />
                    <span class="ml-1.5">Сгенерировать</span>
                  </Button>
                  <Button class="bee-button" :loading="metaSaving" @click="saveMeta"
                    >Сохранить</Button
                  >
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
                <div
                  v-for="version in versions"
                  :key="version.id"
                  class="flex items-center gap-3 px-5 py-3.5"
                >
                  <span
                    class="grid size-8 shrink-0 place-items-center rounded-lg border font-mono text-[10px] font-bold"
                    :class="
                      version.number === latestVersion?.number
                        ? 'border-bee-deep/30 bg-bee/10 text-bee-deep dark:text-bee'
                        : 'border-line bg-canvas/40 text-faint'
                    "
                  >
                    #{{ version.number }}
                  </span>
                  <div class="min-w-0 flex-1">
                    <p class="text-xs font-bold">
                      {{ versionSourceLabel(version.source) }}
                      <span v-if="version.note" class="font-normal text-muted"
                        >— {{ version.note }}</span
                      >
                    </p>
                    <p class="mt-0.5 font-mono text-[9px] text-faint">
                      {{ formatDate(version.created_at) }} ·
                      {{ formatTokens(version.tokens_cost) }} токенов
                    </p>
                  </div>
                  <span
                    v-if="version.number === latestVersion?.number"
                    class="rounded-full bg-emerald-500/10 px-2 py-0.5 font-mono text-[8px] font-bold uppercase tracking-[0.08em] text-emerald-500"
                  >
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

            <div v-if="game.spec" class="panel p-5 sm:p-6">
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="mono-label">Game spec</p>
                  <h2 class="mt-1 font-display text-xl font-bold">Что за игра</h2>
                </div>
                <span
                  class="rounded-full border border-line bg-canvas/50 px-3 py-1 font-mono text-[9px] uppercase tracking-[0.1em] text-muted"
                >
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
          </div>

          <aside class="space-y-5 xl:sticky xl:top-[90px]">
            <section v-if="operation" class="panel p-5">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="mono-label">Сейчас</p>
                  <h2 class="mt-1 font-display text-lg font-bold">{{ operationLabel }}</h2>
                </div>
                <span
                  class="grid size-9 shrink-0 place-items-center rounded-xl bg-bee/15 text-bee-deep dark:text-bee"
                >
                  <AppIcon name="spark" :size="17" />
                </span>
              </div>

              <AiProgress
                class="mt-4"
                compact
                :started-at="operation.started_at"
                :label="operationLabel"
                :hint="operationHint"
                :expected-seconds="operationExpectedSeconds"
              />

              <p class="mt-3 text-[11px] leading-5 text-muted">
                Пока модель работает, игра остаётся доступной. Результат появится здесь же —
                обновлять страницу не нужно.
              </p>

              <button
                type="button"
                class="mt-3 w-full rounded-full border border-line-strong px-3 py-2 text-[10px] font-bold text-muted transition hover:border-bad/40 hover:text-bad"
                :disabled="Boolean(game.actionPending)"
                @click="cancelRun"
              >
                Отменить
              </button>
            </section>

            <section v-else class="panel p-5">
              <StageTimeline :stages="game.stages" :status="game.status" />
              <div class="mt-5 grid grid-cols-2 gap-2 border-t border-line pt-4">
                <div class="rounded-xl border border-line bg-canvas/35 p-3">
                  <p class="mono-label">Стоимость</p>
                  <p class="mt-1.5 font-mono text-sm font-bold">
                    {{ formatTokens(game.game.cost.tokens) }}
                  </p>
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

              <div
                class="mt-4 flex items-end gap-3 rounded-2xl border border-line bg-canvas/35 p-4"
              >
                <p class="font-display text-4xl font-bold leading-none">
                  {{ game.rating.total ? game.rating.average.toFixed(1).replace('.', ',') : '—' }}
                </p>
                <div class="pb-0.5">
                  <div class="flex gap-0.5">
                    <span
                      v-for="star in 5"
                      :key="star"
                      class="size-2 rounded-full"
                      :class="
                        star <= Math.round(game.rating.average)
                          ? averageTone.solid
                          : 'bg-line-strong'
                      "
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
                  <span class="w-5 text-right font-mono text-[9px] tabular-nums text-faint">{{
                    row.count
                  }}</span>
                </div>
              </div>

              <p class="mt-4 border-t border-line pt-3 text-[10px] leading-4 text-faint">
                Оценки ставят игроки на публичной странице. Автор не может оценивать свою игру,
                поэтому здесь только статистика.
              </p>
            </section>

            <section class="panel p-5">
              <div class="flex items-center gap-2.5">
                <span
                  class="grid size-9 place-items-center rounded-xl bg-bee/15 text-bee-deep dark:text-bee"
                >
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

              <button
                type="button"
                class="mt-3 flex w-full items-center justify-center gap-1.5 rounded-full border border-bad/30 px-3 py-2 text-[10px] font-bold text-bad transition hover:bg-bad/10 disabled:opacity-50"
                :disabled="Boolean(game.actionPending)"
                @click="removeGame"
              >
                <AppIcon name="x" :size="12" />
                Удалить игру
              </button>
            </section>
          </aside>
        </div>
      </template>
    </div>
  </div>
</template>
