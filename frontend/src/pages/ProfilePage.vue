<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import Button from 'openvue/button'
import Dialog from 'openvue/dialog'
import InputText from 'openvue/inputtext'
import { useToast } from 'openvue/usetoast'
import { messageFromError } from '@/api/client'
import { listTokenOps, listUserGames } from '@/api/games'
import type { ProfileGame, ProfileTokenOp } from '@/api/types'
import { formatDate, formatNumber, formatTokens } from '@/lib/format'
import { genreLabel } from '@/lib/game'
import { useAuthStore } from '@/stores/auth'
import AppIcon from '@/components/ui/AppIcon.vue'
import StatusPill from '@/components/ui/StatusPill.vue'
import EmptyState from '@/components/ui/EmptyState.vue'

const PAGE_SIZE = 20

const auth = useAuthStore()
const toast = useToast()
const loading = ref(true)
const error = ref('')
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
const gamesLoading = ref(false)

const tokenOps = ref<ProfileTokenOp[]>([])
const tokenOpsTotal = ref(0)
const tokenOpsLoading = ref(false)

const totalSpent = computed(() =>
  games.value.reduce((sum, item) => sum + (item.tokens_cost || 0), 0),
)
const totalFixes = computed(() => games.value.reduce((sum, item) => sum + (item.fixes || 0), 0))
const readyGames = computed(() => games.value.filter((item) => item.status === 'ready').length)
const moreGames = computed(() => games.value.length < gamesTotal.value)
const moreOps = computed(() => tokenOps.value.length < tokenOpsTotal.value)
const emailInitial = computed(() => auth.user?.email?.slice(0, 1).toUpperCase() ?? 'B')

const reasonLabels: Record<string, string> = {
  signup_grant: 'Стартовый бонус',
  generate: 'Генерация игры',
  patch: 'Патч',
  topup: 'Пополнение',
  refund: 'Возврат',
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

async function loadGames(reset = true): Promise<void> {
  if (reset) gamesLoading.value = true

  try {
    const page = await listUserGames({
      sort: 'created_at',
      limit: PAGE_SIZE,
      offset: reset ? 0 : games.value.length,
    })

    games.value = reset ? (page.items ?? []) : [...games.value, ...(page.items ?? [])]
    gamesTotal.value = page.total ?? 0
  } catch {
    if (reset) error.value = 'Не удалось загрузить игры. Попробуйте ещё раз.'
  } finally {
    if (reset) gamesLoading.value = false
  }
}

async function loadTokenOps(reset = true): Promise<void> {
  if (reset) tokenOpsLoading.value = true

  try {
    const page = await listTokenOps({
      limit: PAGE_SIZE,
      offset: reset ? 0 : tokenOps.value.length,
    })

    tokenOps.value = reset ? (page.items ?? []) : [...tokenOps.value, ...(page.items ?? [])]
    tokenOpsTotal.value = page.total ?? 0
  } catch {
    if (reset) error.value = 'Не удалось загрузить операции. Попробуйте ещё раз.'
  } finally {
    if (reset) tokenOpsLoading.value = false
  }
}

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  const result = await auth.refreshProfile()
  if (!result) error.value = 'Не удалось загрузить профиль. Попробуйте ещё раз.'
  loading.value = false
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
    void load()
  }

  void loadGames(true)
  void loadTokenOps(true)
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
          </div>
        </div>

        <div
          class="relative mt-7 grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-line bg-line sm:grid-cols-4"
        >
          <div class="bg-canvas/45 p-4">
            <p class="mono-label">Всего игр</p>
            <p class="mt-2 font-display text-2xl font-bold">
              {{ formatNumber(gamesTotal || games.length) }}
            </p>
          </div>
          <div class="bg-canvas/45 p-4">
            <p class="mono-label">Готовы</p>
            <p class="mt-2 font-display text-2xl font-bold">{{ formatNumber(readyGames) }}</p>
          </div>
          <div class="bg-canvas/45 p-4">
            <p class="mono-label">Потрачено</p>
            <p class="mt-2 font-display text-2xl font-bold">{{ formatTokens(totalSpent) }}</p>
          </div>
          <div class="bg-canvas/45 p-4">
            <p class="mono-label">Починок</p>
            <p class="mt-2 font-display text-2xl font-bold">{{ formatNumber(totalFixes) }}</p>
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
        class="mt-6 grid gap-5 lg:grid-cols-[minmax(0,1.4fr)_minmax(0,.6fr)]"
      >
        <div class="h-80 animate-pulse rounded-[1.35rem] bg-surface/70" />
        <div class="h-80 animate-pulse rounded-[1.35rem] bg-surface/70" />
      </div>

      <div v-else-if="error" class="mt-6">
        <EmptyState icon="alert" title="Профиль не загрузился" :text="error">
          <Button class="bee-button" @click="load">Повторить</Button>
        </EmptyState>
      </div>

      <div v-else class="mt-6 grid items-start gap-5 lg:grid-cols-[minmax(0,1.4fr)_minmax(0,.6fr)]">
        <div class="min-w-0 space-y-5">
          <section class="panel overflow-hidden">
            <div
              class="flex items-center justify-between gap-3 border-b border-line px-5 py-4 sm:px-6"
            >
              <div>
                <p class="mono-label">Библиотека</p>
                <h2 class="mt-1 font-display text-xl font-bold">Мои игры</h2>
              </div>
              <RouterLink
                :to="{ name: 'create' }"
                class="inline-flex items-center gap-1.5 rounded-full bg-bee px-3.5 py-2 text-[11px] font-extrabold text-graphite transition hover:bg-bee-deep"
              >
                <AppIcon name="plus" :size="13" :stroke="2.4" />
                Новая
              </RouterLink>
            </div>

            <div v-if="gamesLoading && !games.length" class="divide-y divide-line">
              <div
                v-for="index in 4"
                :key="index"
                class="flex items-center gap-3 px-4 py-4 sm:px-6"
              >
                <span class="size-10 shrink-0 animate-pulse rounded-xl bg-line/50" />
                <div class="flex-1 space-y-2">
                  <div class="h-3 w-2/5 animate-pulse rounded bg-line/70" />
                  <div class="h-2.5 w-1/4 animate-pulse rounded bg-line/50" />
                </div>
              </div>
            </div>

            <div v-else-if="games.length" class="divide-y divide-line">
              <RouterLink
                v-for="item in games"
                :key="item.id"
                :to="{ name: 'game', params: { id: item.slug || item.id } }"
                class="group hover-row flex items-center gap-3 px-4 py-4 sm:px-6"
              >
                <span
                  class="grid size-10 shrink-0 place-items-center rounded-xl border border-line bg-canvas/60 text-bee-deep transition group-hover:border-bee-deep/40 group-hover:bg-bee/10 dark:text-bee"
                >
                  <AppIcon name="gamepad" :size="18" />
                </span>

                <div class="min-w-0 flex-1">
                  <p class="truncate text-[13px] font-bold">{{ item.prompt }}</p>
                  <p
                    class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 font-mono text-[9px] uppercase tracking-[0.08em] text-faint"
                  >
                    <span>{{ genreLabel(item.genre) }}</span>
                    <span>·</span>
                    <span>{{ formatDate(item.created_at) }}</span>
                    <span>·</span>
                    <span>{{ formatTokens(item.tokens_cost) }} токенов</span>
                  </p>
                </div>

                <StatusPill :status="item.status" compact class="hidden sm:inline-flex" />
                <AppIcon
                  name="arrow-right"
                  :size="15"
                  class="text-faint transition group-hover:translate-x-0.5 group-hover:text-ink"
                />
              </RouterLink>
            </div>

            <button
              v-if="moreGames"
              type="button"
              class="flex w-full items-center justify-center gap-2 border-t border-line px-5 py-3.5 text-[11px] font-bold text-muted transition hover:bg-hover-surface hover:text-ink"
              :disabled="gamesLoading"
              @click="loadGames(false)"
            >
              <span
                v-if="gamesLoading"
                class="size-3 animate-spin rounded-full border-2 border-line border-t-bee-deep"
              />
              {{
                gamesLoading ? 'Загружаем…' : `Показать ещё · осталось ${gamesTotal - games.length}`
              }}
            </button>

            <div v-else-if="!gamesLoading" class="p-5">
              <EmptyState
                icon="wand"
                title="Пока ни одной игры"
                text="Опишите идею на главной — первая игра появится здесь."
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
              Добавьте в промпт героя, цель, управление и стиль. Чем конкретнее идея, тем меньше
              правок понадобится после первой сборки.
            </p>
            <div class="mt-4 flex flex-wrap gap-1.5">
              <span
                class="rounded-full border border-line bg-canvas/45 px-2.5 py-1 font-mono text-[9px] uppercase tracking-[0.06em] text-faint"
                >герой</span
              >
              <span
                class="rounded-full border border-line bg-canvas/45 px-2.5 py-1 font-mono text-[9px] uppercase tracking-[0.06em] text-faint"
                >цель</span
              >
              <span
                class="rounded-full border border-line bg-canvas/45 px-2.5 py-1 font-mono text-[9px] uppercase tracking-[0.06em] text-faint"
                >управление</span
              >
              <span
                class="rounded-full border border-line bg-canvas/45 px-2.5 py-1 font-mono text-[9px] uppercase tracking-[0.06em] text-faint"
                >стиль</span
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

        <div class="min-w-0 space-y-5">
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
                  <p class="mt-0.5 text-[10px] text-faint">полный проход конвейера</p>
                </div>
                <span class="font-mono text-xs font-bold">{{
                  formatTokens(profile?.prices.generate ?? 0)
                }}</span>
              </div>
              <div
                class="flex items-center justify-between rounded-xl border border-line bg-canvas/35 px-3.5 py-3"
              >
                <div>
                  <p class="text-xs font-bold">Патч</p>
                  <p class="mt-0.5 text-[10px] text-faint">правка существующей игры</p>
                </div>
                <span class="font-mono text-xs font-bold">{{
                  formatTokens(profile?.prices.patch ?? 0)
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

          <section class="panel overflow-hidden">
            <div class="flex items-center justify-between gap-3 border-b border-line px-5 py-4">
              <div>
                <p class="mono-label">Журнал</p>
                <h2 class="mt-1 font-display text-lg font-bold">Операции</h2>
              </div>
              <AppIcon name="chart" :size="18" class="text-muted" />
            </div>

            <div v-if="tokenOpsLoading && !tokenOps.length" class="divide-y divide-line">
              <div v-for="index in 4" :key="index" class="flex items-center gap-3 px-5 py-3.5">
                <span class="size-8 shrink-0 animate-pulse rounded-full bg-line/50" />
                <div class="flex-1 space-y-2">
                  <div class="h-3 w-1/3 animate-pulse rounded bg-line/70" />
                  <div class="h-2 w-1/4 animate-pulse rounded bg-line/50" />
                </div>
              </div>
            </div>

            <template v-else-if="tokenOps.length">
              <div
                v-for="(op, index) in tokenOps"
                :key="`${op.created_at}-${index}`"
                class="flex items-center gap-3 px-5 py-3.5"
              >
                <span
                  class="grid size-8 shrink-0 place-items-center rounded-full border"
                  :class="
                    op.delta >= 0
                      ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-500'
                      : 'border-line bg-canvas/50 text-faint'
                  "
                >
                  <AppIcon :name="op.delta >= 0 ? 'plus' : 'arrow-right'" :size="13" />
                </span>
                <div class="min-w-0 flex-1">
                  <p class="text-xs font-bold">{{ reasonLabels[op.reason] ?? op.reason }}</p>
                  <p class="mt-0.5 text-[10px] text-faint">{{ formatDate(op.created_at) }}</p>
                </div>
                <span
                  class="font-mono text-[11px] font-bold tabular-nums"
                  :class="op.delta >= 0 ? 'text-emerald-500' : 'text-ink'"
                >
                  {{ op.delta >= 0 ? '+' : '' }}{{ formatTokens(op.delta) }}
                </span>
              </div>

              <button
                v-if="moreOps"
                type="button"
                class="w-full border-t border-line px-5 py-3 text-[11px] font-bold text-muted transition hover:bg-hover-surface hover:text-ink"
                :disabled="tokenOpsLoading"
                @click="loadTokenOps(false)"
              >
                {{
                  tokenOpsLoading
                    ? 'Загружаем…'
                    : `Показать ещё · осталось ${tokenOpsTotal - tokenOps.length}`
                }}
              </button>
            </template>

            <div v-else class="px-5 py-8 text-center text-xs text-faint">Операций пока нет.</div>
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
        Демо-пополнение без реальной оплаты: сумма ограничена сервером, действует суточный лимит.
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
