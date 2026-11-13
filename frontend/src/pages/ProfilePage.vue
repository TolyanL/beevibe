<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import Button from 'openvue/button'
import Dialog from 'openvue/dialog'
import InputText from 'openvue/inputtext'
import { useToast } from 'openvue/usetoast'
import { messageFromError } from '@/api/client'
import { listTokenOps, listUserGenerations, listUserGames } from '@/api/games'
import type { ProfileGame, ProfileGeneration, ProfileTokenOp } from '@/api/types'
import { formatNumber, formatTokens, plural } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'
import AppIcon from '@/components/ui/AppIcon.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import InlineError from '@/components/ui/InlineError.vue'
import PaginationBar from '@/components/ui/PaginationBar.vue'
import ProfileGameCard from '@/components/profile/ProfileGameCard.vue'
import GenerationHistoryItem from '@/components/profile/GenerationHistoryItem.vue'
import TokenOpItem from '@/components/profile/TokenOpItem.vue'

const GAMES_PAGE_SIZE = 6
const GENERATIONS_PAGE_SIZE = 6
const OPS_PAGE_SIZE = 8

const auth = useAuthStore()
const toast = useToast()
const loading = ref(true)
const profileError = ref('')
const toppingUp = ref(false)
const displayName = ref('')
const savingProfile = ref(false)
const topupOpen = ref(false)
const topupAmount = ref(250000)
const topupOptions = [
  { label: '250K', value: 250000 },
  { label: '500K', value: 500000 },
  { label: '1M', value: 1000000 },
]

const profile = computed(() => auth.profile)

const games = ref<ProfileGame[]>([])
const gamesTotal = ref(0)
const gamesPage = ref(1)
const gamesSort = ref<'created_at' | 'tokens_cost'>('created_at')
const gamesLoading = ref(false)
const gamesError = ref('')

const generations = ref<ProfileGeneration[]>([])
const generationsTotal = ref(0)
const generationsPage = ref(1)
const generationsLoading = ref(false)
const generationsError = ref('')

const tokenOps = ref<ProfileTokenOp[]>([])
const tokenOpsTotal = ref(0)
const tokenOpsPage = ref(1)
const tokenOpsLoading = ref(false)
const tokenOpsError = ref('')

const emailInitial = computed(() => auth.user?.email?.slice(0, 1).toUpperCase() ?? 'B')
const stats = computed(
  () =>
    auth.profile?.stats ?? {
      games_total: gamesTotal.value,
      games_ready: games.value.filter((item) => item.status === 'ready').length,
      tokens_spent: games.value.reduce((sum, item) => sum + (item.tokens_cost || 0), 0),
      fixes: games.value.reduce((sum, item) => sum + (item.fixes || 0), 0),
    },
)
const gamesLabel = computed(() => plural(gamesTotal.value, 'игра', 'игры', 'игр'))
const operationsLabel = computed(() =>
  plural(tokenOpsTotal.value, 'операция', 'операции', 'операций'),
)

function pageCount(total: number, size: number): number {
  return Math.max(1, Math.ceil(total / size))
}

function offsetFor(page: number, size: number): number {
  return (page - 1) * size
}

function scrollTo(id: string): void {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

async function loadProfile(): Promise<void> {
  loading.value = true
  profileError.value = ''
  const result = await auth.refreshProfile()
  if (!result) profileError.value = 'Не удалось загрузить профиль. Попробуйте ещё раз.'
  loading.value = false
}

async function loadGames(): Promise<void> {
  gamesLoading.value = true
  gamesError.value = ''

  try {
    const page = await listUserGames({
      sort: gamesSort.value,
      limit: GAMES_PAGE_SIZE,
      offset: offsetFor(gamesPage.value, GAMES_PAGE_SIZE),
    })

    games.value = page.items ?? []
    gamesTotal.value = page.total ?? 0
  } catch {
    gamesError.value = 'Не удалось загрузить игры. Попробуйте ещё раз.'
  } finally {
    gamesLoading.value = false
  }
}

async function loadGenerations(): Promise<void> {
  generationsLoading.value = true
  generationsError.value = ''

  try {
    const page = await listUserGenerations({
      limit: GENERATIONS_PAGE_SIZE,
      offset: offsetFor(generationsPage.value, GENERATIONS_PAGE_SIZE),
    })

    generations.value = page.items ?? []
    generationsTotal.value = page.total ?? 0
  } catch {
    generationsError.value = 'Не удалось загрузить историю генераций.'
  } finally {
    generationsLoading.value = false
  }
}

async function loadTokenOps(): Promise<void> {
  tokenOpsLoading.value = true
  tokenOpsError.value = ''

  try {
    const page = await listTokenOps({
      limit: OPS_PAGE_SIZE,
      offset: offsetFor(tokenOpsPage.value, OPS_PAGE_SIZE),
    })

    tokenOps.value = page.items ?? []
    tokenOpsTotal.value = page.total ?? 0
  } catch {
    tokenOpsError.value = 'Не удалось загрузить операции.'
  } finally {
    tokenOpsLoading.value = false
  }
}

function setGamesSort(sort: 'created_at' | 'tokens_cost'): void {
  if (gamesSort.value === sort) return
  gamesSort.value = sort
  gamesPage.value = 1
  void loadGames()
}

function changeGamesPage(page: number): void {
  gamesPage.value = Math.min(Math.max(1, page), pageCount(gamesTotal.value, GAMES_PAGE_SIZE))
  void loadGames().then(() => scrollTo('games'))
}

function changeGenerationsPage(page: number): void {
  generationsPage.value = Math.min(
    Math.max(1, page),
    pageCount(generationsTotal.value, GENERATIONS_PAGE_SIZE),
  )
  void loadGenerations().then(() => scrollTo('generations'))
}

function changeTokenOpsPage(page: number): void {
  tokenOpsPage.value = Math.min(Math.max(1, page), pageCount(tokenOpsTotal.value, OPS_PAGE_SIZE))
  void loadTokenOps().then(() => scrollTo('operations'))
}

async function saveDisplayName(): Promise<void> {
  savingProfile.value = true
  try {
    await auth.updateDisplayName(displayName.value)
    toast.add({ severity: 'success', summary: 'Имя сохранено', life: 2200 })
  } catch (err) {
    toast.add({
      severity: 'error',
      summary: 'Не удалось сохранить',
      detail: messageFromError(err),
      life: 3200,
    })
  } finally {
    savingProfile.value = false
  }
}

async function topup(): Promise<void> {
  toppingUp.value = true
  try {
    await auth.topup(topupAmount.value)
    topupOpen.value = false
    toast.add({
      severity: 'success',
      summary: 'Баланс пополнен',
      detail: 'Токены уже доступны для новых игр.',
      life: 2800,
    })
    tokenOpsPage.value = 1
    void loadTokenOps()
  } catch (err) {
    toast.add({
      severity: 'error',
      summary: 'Не получилось',
      detail: messageFromError(err),
      life: 3200,
    })
  } finally {
    toppingUp.value = false
  }
}

watch(
  () => auth.user?.display_name,
  (value) => {
    displayName.value = value ?? ''
  },
  { immediate: true },
)

onMounted(() => {
  if (auth.profile) {
    loading.value = false
    void auth.refreshProfile()
  } else {
    void loadProfile()
  }

  void loadGames()
  void loadGenerations()
  void loadTokenOps()
})
</script>

<template>
  <div class="pb-8 pt-9 sm:pt-12">
    <div class="app-container">
      <section
        class="relative overflow-hidden rounded-[1.6rem] border border-line bg-surface/70 p-5 sm:p-8"
      >
        <div class="absolute -right-24 -top-24 size-64 rounded-full bg-bee/15 blur-3xl" />
        <div class="relative flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
          <div class="flex items-center gap-4">
            <span
              class="grid size-14 shrink-0 place-items-center rounded-2xl bg-ink font-display text-xl font-bold text-canvas"
            >
              {{ emailInitial }}
            </span>
            <div class="min-w-0">
              <div class="eyebrow">Личный кабинет</div>
              <h1 class="mt-2 truncate font-display text-2xl font-bold sm:text-3xl">
                {{ auth.user?.display_name || auth.user?.email }}
              </h1>
              <p class="mt-1 font-mono text-[10px] uppercase tracking-[0.1em] text-faint">
                {{ auth.user?.email }} · id {{ auth.user?.id?.slice(0, 8) }}
              </p>
            </div>
          </div>

          <div class="flex flex-wrap items-center gap-2.5">
            <div
              class="flex items-center gap-2 rounded-full border border-line bg-canvas/50 py-1.5 pl-2 pr-4"
            >
              <span class="grid size-7 place-items-center rounded-full bg-bee text-graphite">
                <AppIcon name="bolt" :size="14" :stroke="2.3" />
              </span>
              <div>
                <p class="font-mono text-sm font-bold leading-none tabular-nums">
                  {{ formatNumber(auth.balance) }}
                </p>
                <p class="mt-1 text-[9px] uppercase tracking-[0.1em] text-faint">токенов</p>
              </div>
            </div>
            <Button class="bee-button" :loading="toppingUp" @click="topupOpen = true">
              <AppIcon name="plus" :size="15" :stroke="2.4" />
              <span class="ml-1.5">Пополнить</span>
            </Button>
            <RouterLink
              :to="{ name: 'create' }"
              class="inline-flex items-center gap-1.5 rounded-full border border-line-strong px-4 py-2.5 text-[11px] font-extrabold text-ink transition hover:border-bee-deep hover:bg-hover-surface"
            >
              <AppIcon name="spark" :size="13" :stroke="2.2" />
              Создать игру
            </RouterLink>
          </div>
        </div>

        <div
          class="relative mt-7 grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-line bg-line sm:grid-cols-4"
        >
          <div class="bg-canvas/45 p-4">
            <p class="mono-label">Всего игр</p>
            <p class="mt-2 font-display text-2xl font-bold">
              {{ formatNumber(stats.games_total) }}
            </p>
          </div>
          <div class="bg-canvas/45 p-4">
            <p class="mono-label">Готовых</p>
            <p class="mt-2 font-display text-2xl font-bold">
              {{ formatNumber(stats.games_ready) }}
            </p>
          </div>
          <div class="bg-canvas/45 p-4">
            <p class="mono-label">Потрачено</p>
            <p class="mt-2 font-display text-2xl font-bold">
              {{ formatTokens(stats.tokens_spent) }}
            </p>
          </div>
          <div class="bg-canvas/45 p-4">
            <p class="mono-label">Починок</p>
            <p class="mt-2 font-display text-2xl font-bold">{{ formatNumber(stats.fixes) }}</p>
          </div>
        </div>
      </section>

      <section class="panel mt-5 p-5 sm:p-6">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <p class="mono-label">Профиль</p>
            <h2 class="mt-1 font-display text-lg font-bold">Публичное имя автора</h2>
            <p class="mt-1 text-[11px] text-muted">
              Его видят в каталоге и на странице игры вместо email.
            </p>
          </div>
          <div class="flex w-full gap-2 sm:w-auto">
            <InputText
              v-model="displayName"
              maxlength="40"
              class="w-full sm:w-64"
              placeholder="Например: Мастер кристаллов"
            />
            <Button class="bee-button" :loading="savingProfile" @click="saveDisplayName"
              >Сохранить</Button
            >
          </div>
        </div>
      </section>

      <div
        v-if="loading && !profile"
        class="mt-6 grid gap-5 lg:grid-cols-[minmax(0,1.5fr)_minmax(0,.65fr)]"
      >
        <div class="h-96 animate-pulse rounded-[1.35rem] bg-surface/70" />
        <div class="h-96 animate-pulse rounded-[1.35rem] bg-surface/70" />
      </div>

      <div v-else-if="profileError" class="mt-6">
        <EmptyState icon="alert" title="Профиль не загрузился" :text="profileError">
          <Button class="bee-button" @click="loadProfile">Повторить</Button>
        </EmptyState>
      </div>

      <div
        v-else
        class="mt-6 grid items-start gap-5 lg:grid-cols-[minmax(0,1.5fr)_minmax(0,.65fr)]"
      >
        <div class="min-w-0 space-y-5">
          <section id="games" class="panel scroll-mt-24 overflow-hidden">
            <div
              class="flex flex-wrap items-end justify-between gap-3 border-b border-line px-4 py-4 sm:px-6"
            >
              <div>
                <p class="mono-label">Библиотека</p>
                <div class="mt-1 flex flex-wrap items-baseline gap-2">
                  <h2 class="font-display text-xl font-bold">Мои игры</h2>
                  <span class="font-mono text-[10px] text-faint">
                    {{ gamesLoading && !games.length ? '…' : formatNumber(gamesTotal) }}
                    {{ gamesLabel }}
                  </span>
                </div>
                <p class="mt-1 text-[11px] text-muted">
                  Отсюда можно открыть игру, посмотреть сборку или сразу играть.
                </p>
              </div>

              <div class="flex flex-wrap items-center gap-2">
                <div
                  class="flex items-center rounded-full border border-line bg-canvas/40 p-0.5"
                  aria-label="Сортировка игр"
                >
                  <button
                    type="button"
                    class="rounded-full px-3 py-1.5 text-[10px] font-bold transition"
                    :class="
                      gamesSort === 'created_at'
                        ? 'bg-bee text-graphite'
                        : 'text-muted hover:text-ink'
                    "
                    :aria-pressed="gamesSort === 'created_at'"
                    @click="setGamesSort('created_at')"
                  >
                    Новые
                  </button>
                  <button
                    type="button"
                    class="rounded-full px-3 py-1.5 text-[10px] font-bold transition"
                    :class="
                      gamesSort === 'tokens_cost'
                        ? 'bg-bee text-graphite'
                        : 'text-muted hover:text-ink'
                    "
                    :aria-pressed="gamesSort === 'tokens_cost'"
                    @click="setGamesSort('tokens_cost')"
                  >
                    Дорогие
                  </button>
                </div>

                <RouterLink
                  :to="{ name: 'create' }"
                  class="inline-flex items-center gap-1.5 rounded-full bg-bee px-3.5 py-2 text-[11px] font-extrabold text-graphite transition hover:bg-bee-deep"
                >
                  <AppIcon name="plus" :size="13" :stroke="2.4" />
                  Новая
                </RouterLink>
              </div>
            </div>

            <div v-if="gamesError && games.length" class="border-b border-line px-4 py-3 sm:px-6">
              <p class="flex items-center gap-2 text-[11px] text-muted">
                <AppIcon name="alert" :size="13" class="text-red-500" />
                {{ gamesError }}
                <button
                  type="button"
                  class="font-bold text-hover-accent hover:underline"
                  @click="loadGames"
                >
                  Повторить
                </button>
              </p>
            </div>

            <div v-if="gamesLoading && !games.length" class="grid gap-3 p-4 sm:grid-cols-2 sm:p-5">
              <div
                v-for="index in 2"
                :key="index"
                class="overflow-hidden rounded-[1.2rem] border border-line"
              >
                <div class="aspect-[16/9] animate-pulse bg-line/40" />
                <div class="space-y-2 p-4">
                  <div class="h-3 w-2/3 animate-pulse rounded bg-line/70" />
                  <div class="h-2.5 w-full animate-pulse rounded bg-line/50" />
                  <div class="h-2.5 w-1/2 animate-pulse rounded bg-line/50" />
                </div>
              </div>
            </div>

            <InlineError
              v-else-if="gamesError && !games.length"
              :text="gamesError"
              @retry="loadGames"
            />

            <div v-else-if="games.length" class="grid gap-3 p-4 sm:grid-cols-2 sm:p-5">
              <ProfileGameCard v-for="item in games" :key="item.id" :game="item" />
            </div>

            <div v-else class="p-5">
              <EmptyState
                icon="wand"
                title="Пока ни одной игры"
                text="Соберите первую игру — она появится здесь."
              >
                <RouterLink
                  :to="{ name: 'create' }"
                  class="inline-flex items-center gap-2 rounded-full bg-bee px-5 py-2.5 text-xs font-extrabold text-graphite"
                >
                  <AppIcon name="spark" :size="14" />
                  Создать игру
                </RouterLink>
              </EmptyState>
            </div>

            <PaginationBar
              :page="gamesPage"
              :page-size="GAMES_PAGE_SIZE"
              :total="gamesTotal"
              :loading="gamesLoading"
              label="игр"
              @update:page="changeGamesPage"
            />
          </section>

          <section id="generations" class="panel scroll-mt-24 overflow-hidden">
            <div
              class="flex flex-wrap items-end justify-between gap-3 border-b border-line px-4 py-4 sm:px-6"
            >
              <div>
                <p class="mono-label">История</p>
                <div class="mt-1 flex flex-wrap items-baseline gap-2">
                  <h2 class="font-display text-xl font-bold">Генерации</h2>
                  <span class="font-mono text-[10px] text-faint">
                    {{
                      generationsLoading && !generations.length
                        ? '…'
                        : formatNumber(generationsTotal)
                    }}
                  </span>
                </div>
                <p class="mt-1 text-[11px] text-muted">
                  Каждая версия: первая сборка, правки, починки и изменения настроек.
                </p>
              </div>
              <span
                class="hidden size-9 place-items-center rounded-xl bg-bee/15 text-bee-deep dark:text-bee sm:grid"
              >
                <AppIcon name="layers" :size="18" />
              </span>
            </div>

            <div
              v-if="generationsError && generations.length"
              class="border-b border-line px-4 py-3 sm:px-6"
            >
              <p class="flex items-center gap-2 text-[11px] text-muted">
                <AppIcon name="alert" :size="13" class="text-red-500" />
                {{ generationsError }}
                <button
                  type="button"
                  class="font-bold text-hover-accent hover:underline"
                  @click="loadGenerations"
                >
                  Повторить
                </button>
              </p>
            </div>

            <div v-if="generationsLoading && !generations.length" class="divide-y divide-line">
              <div
                v-for="index in 4"
                :key="index"
                class="flex items-center gap-3 px-4 py-4 sm:px-5"
              >
                <span class="size-10 shrink-0 animate-pulse rounded-xl bg-line/50" />
                <div class="flex-1 space-y-2">
                  <div class="h-2.5 w-1/4 animate-pulse rounded bg-line/60" />
                  <div class="h-3 w-2/5 animate-pulse rounded bg-line/70" />
                  <div class="h-2 w-1/3 animate-pulse rounded bg-line/50" />
                </div>
              </div>
            </div>

            <InlineError
              v-else-if="generationsError && !generations.length"
              :text="generationsError"
              @retry="loadGenerations"
            />

            <div v-else-if="generations.length" class="divide-y divide-line">
              <GenerationHistoryItem
                v-for="item in generations"
                :key="`${item.game_id}-${item.version_number}`"
                :generation="item"
              />
            </div>

            <div v-else class="flex flex-col items-center gap-2 px-5 py-10 text-center">
              <span
                class="grid size-12 place-items-center rounded-2xl border border-line bg-canvas/40 text-bee-deep dark:text-bee"
              >
                <AppIcon name="layers" :size="22" />
              </span>
              <p class="text-sm font-bold">Генераций пока нет</p>
              <p class="max-w-xs text-[11px] leading-5 text-muted">
                Здесь появится история версий: сборка, правки и починки игр.
              </p>
            </div>

            <PaginationBar
              :page="generationsPage"
              :page-size="GENERATIONS_PAGE_SIZE"
              :total="generationsTotal"
              :loading="generationsLoading"
              label="генераций"
              @update:page="changeGenerationsPage"
            />
          </section>
        </div>

        <div class="min-w-0 space-y-5">
          <section id="operations" class="panel scroll-mt-24 overflow-hidden">
            <div class="flex items-start justify-between gap-3 border-b border-line px-5 py-4">
              <div>
                <p class="mono-label">Журнал</p>
                <div class="mt-1 flex flex-wrap items-baseline gap-2">
                  <h2 class="font-display text-lg font-bold">Операции</h2>
                  <span class="font-mono text-[10px] text-faint">
                    {{ tokenOpsLoading && !tokenOps.length ? '…' : formatNumber(tokenOpsTotal) }}
                    {{ operationsLabel }}
                  </span>
                </div>
              </div>
              <AppIcon name="chart" :size="18" class="mt-1 text-muted" />
            </div>

            <div v-if="tokenOpsError && tokenOps.length" class="border-b border-line px-5 py-3">
              <p class="flex items-center gap-2 text-[11px] text-muted">
                <AppIcon name="alert" :size="13" class="text-red-500" />
                {{ tokenOpsError }}
                <button
                  type="button"
                  class="font-bold text-hover-accent hover:underline"
                  @click="loadTokenOps"
                >
                  Повторить
                </button>
              </p>
            </div>

            <div v-if="tokenOpsLoading && !tokenOps.length" class="divide-y divide-line">
              <div v-for="index in 4" :key="index" class="flex items-center gap-3 px-5 py-3.5">
                <span class="size-8 shrink-0 animate-pulse rounded-full bg-line/50" />
                <div class="flex-1 space-y-2">
                  <div class="h-3 w-1/3 animate-pulse rounded bg-line/70" />
                  <div class="h-2 w-1/2 animate-pulse rounded bg-line/50" />
                </div>
              </div>
            </div>

            <InlineError
              v-else-if="tokenOpsError && !tokenOps.length"
              :text="tokenOpsError"
              @retry="loadTokenOps"
            />

            <div v-else-if="tokenOps.length" class="divide-y divide-line">
              <TokenOpItem
                v-for="(op, index) in tokenOps"
                :key="`${op.created_at}-${index}`"
                :op="op"
              />
            </div>

            <div v-else class="px-5 py-8 text-center text-xs text-faint">Операций пока нет.</div>

            <PaginationBar
              :page="tokenOpsPage"
              :page-size="OPS_PAGE_SIZE"
              :total="tokenOpsTotal"
              :loading="tokenOpsLoading"
              label="операций"
              @update:page="changeTokenOpsPage"
            />
          </section>

          <section class="panel p-5">
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="mono-label">Тарифы</p>
                <h2 class="mt-1 font-display text-lg font-bold">Стоимость действий</h2>
              </div>
              <span
                class="grid size-9 place-items-center rounded-xl bg-bee/15 text-bee-deep dark:text-bee"
              >
                <AppIcon name="coin" :size="18" />
              </span>
            </div>

            <div class="mt-4 space-y-2.5">
              <div
                class="flex items-center justify-between rounded-xl border border-line bg-canvas/35 px-3.5 py-3"
              >
                <div>
                  <p class="text-xs font-bold">Новая игра</p>
                  <p class="mt-0.5 text-[10px] text-faint">полная сборка с нуля</p>
                </div>
                <span class="font-mono text-xs font-bold">{{
                  formatTokens(auth.prices.generate)
                }}</span>
              </div>
              <div
                class="flex items-center justify-between rounded-xl border border-line bg-canvas/35 px-3.5 py-3"
              >
                <div>
                  <p class="text-xs font-bold">Быстрая игра</p>
                  <p class="mt-0.5 text-[10px] text-faint">два вызова модели</p>
                </div>
                <span class="font-mono text-xs font-bold">{{
                  formatTokens(auth.prices.generate_fast)
                }}</span>
              </div>
              <div
                class="flex items-center justify-between rounded-xl border border-line bg-canvas/35 px-3.5 py-3"
              >
                <div>
                  <p class="text-xs font-bold">Правка</p>
                  <p class="mt-0.5 text-[10px] text-faint">изменение готовой игры</p>
                </div>
                <span class="font-mono text-xs font-bold">{{
                  formatTokens(auth.prices.patch)
                }}</span>
              </div>
              <div
                class="flex items-center justify-between rounded-xl border border-line bg-canvas/35 px-3.5 py-3"
              >
                <div>
                  <p class="text-xs font-bold">Название и обложка</p>
                  <p class="mt-0.5 text-[10px] text-faint">метаданные для каталога</p>
                </div>
                <span class="font-mono text-xs font-bold">{{
                  formatTokens(auth.prices.meta)
                }}</span>
              </div>
            </div>

            <Button
              class="p-button-secondary mt-4 w-full"
              :loading="toppingUp"
              @click="topupOpen = true"
            >
              <AppIcon name="plus" :size="14" />
              <span class="ml-1.5">Пополнить баланс</span>
            </Button>
          </section>

          <section class="panel p-5">
            <div class="flex items-start gap-3">
              <span
                class="grid size-9 shrink-0 place-items-center rounded-xl bg-bee/15 text-bee-deep dark:text-bee"
              >
                <AppIcon name="bolt" :size="18" />
              </span>
              <div>
                <p class="mono-label">Совет</p>
                <h2 class="mt-1 font-display text-lg font-bold">Больше деталей — точнее игра</h2>
              </div>
            </div>
            <p class="mt-3 text-xs leading-5 text-muted">
              Добавьте в описание героя, цель, управление и стиль — чем конкретнее идея, тем меньше
              правок понадобится после сборки.
            </p>
            <div class="mt-4 flex flex-wrap gap-1.5">
              <span
                v-for="tag in ['герой', 'цель', 'управление', 'стиль']"
                :key="tag"
                class="rounded-full border border-line bg-canvas/45 px-2.5 py-1 font-mono text-[9px] uppercase tracking-[0.06em] text-faint"
                >{{ tag }}</span
              >
            </div>
            <RouterLink
              to="/#create"
              class="mt-4 inline-flex items-center gap-1.5 text-[11px] font-extrabold text-bee-deep transition hover:text-ink dark:text-bee"
            >
              <AppIcon name="wand" :size="13" />
              Собрать новую игру
              <AppIcon name="arrow-right" :size="13" />
            </RouterLink>
          </section>
        </div>
      </div>
    </div>

    <Dialog
      v-model:visible="topupOpen"
      modal
      header="Пополнить баланс"
      :style="{ width: 'min(92vw, 420px)' }"
      :draggable="false"
    >
      <p class="text-xs leading-5 text-muted">
        Это демо-токены: реальных денег не нужно. На пополнение есть суточный лимит.
      </p>
      <div class="mt-4 flex flex-wrap gap-2">
        <button
          v-for="option in topupOptions"
          :key="option.value"
          type="button"
          class="rounded-full border px-3.5 py-2 text-[11px] font-bold transition"
          :class="
            topupAmount === option.value
              ? 'border-transparent bg-bee text-graphite'
              : 'border-line-strong text-muted hover:border-bee-deep hover:text-ink'
          "
          @click="topupAmount = option.value"
        >
          {{ option.label }}
        </button>
      </div>
      <Button class="bee-button mt-5 w-full" :loading="toppingUp" @click="topup">
        <AppIcon name="plus" :size="14" />
        <span class="ml-1.5">Пополнить на {{ formatTokens(topupAmount) }}</span>
      </Button>
    </Dialog>
  </div>
</template>
