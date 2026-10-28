import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import {
  acceptReview,
  cancelGame,
  createGame,
  regenerateReview,
  deleteGame,
  fixGame,
  getGame,
  patchGame,
  rateGame,
  retryGame,
  rollbackGame,
  updateGameConfig,
  updateGameMeta,
  updateGameVisibility,
  generateGameMeta,
} from '@/api/games'
import { messageFromError } from '@/api/client'
import { openGameStream } from '@/api/sse'
import type {
  ActiveOperation,
  CreateGameInput,
  Game,
  GameConfig,
  GameMeta,
  GameSpec,
  GameStage,
  RatingAggregate,
  Visibility,
} from '@/api/types'

const PIPELINE_STAGES = 8

const PIPELINE_STAGE_NAMES = new Set([
  'analyzing',
  'planning',
  'coding',
  'assets',
  'reviewing',
  'assembling',
  'checking',
])

const ACTIVE_STATUSES = new Set([
  'queued',
  'analyzing',
  'planning',
  'coding',
  'assets',
  'reviewing',
  'assembling',
  'checking',
  'patching',
])

export const useGameStore = defineStore('game', () => {
  const game = ref<Game | null>(null)
  const stages = ref<GameStage[]>([])
  const spec = ref<GameSpec | null>(null)
  const config = ref<GameConfig | null>(null)
  const rating = ref<RatingAggregate>({ average: 0, total: 0, counts: [] })
  const myRating = ref(0)
  const operation = ref<ActiveOperation | null>(null)
  const loading = ref(false)
  const error = ref('')
  const streamError = ref('')
  const streaming = ref(false)
  const refreshKey = ref(0)
  const actionPending = ref('')

  let closeStream: (() => void) | null = null
  let currentId = ''
  let streamSeq = 0

  const status = computed(() => game.value?.status ?? 'queued')
  const isReady = computed(() => status.value === 'ready')
  const isFailed = computed(() => status.value === 'failed')
  const isActive = computed(() => ACTIVE_STATUSES.has(status.value))
  const pipelineStages = computed(() =>
    stages.value.filter((stage) => PIPELINE_STAGE_NAMES.has(stage.stage)),
  )

  const isGenerating = computed(
    () => ACTIVE_STATUSES.has(status.value) && status.value !== 'patching',
  )

  const stageProgress = computed(() => {
    if (isReady.value) return 100
    if (!pipelineStages.value.length) return 0

    const done = pipelineStages.value.filter((stage) => stage.status === 'done').length

    return Math.min(100, Math.round((done / PIPELINE_STAGES) * 100))
  })

  function hydrate(next: Game): void {
    game.value = next
    stages.value = next.stages ?? []
    spec.value = next.spec ?? null
    config.value = next.config ?? null
    rating.value = next.rating ?? { average: 0, total: 0, counts: [] }
    myRating.value = next.my_rating ?? 0
    operation.value = next.operation ?? null
    streamSeq = Math.max(streamSeq, next.cursor ?? 0)
  }

  async function load(id: string, options: { silent?: boolean } = {}): Promise<void> {
    currentId = id
    if (!options.silent) {
      loading.value = true
      error.value = ''
    }

    try {
      const next = await getGame(id)
      hydrate(next)
      if (ACTIVE_STATUSES.has(next.status)) startStream(id)
      else stopStream()
    } catch (err) {
      if (!options.silent) error.value = messageFromError(err)
    } finally {
      if (!options.silent) loading.value = false
    }
  }

  async function refresh(): Promise<void> {
    if (!currentId) return
    await load(currentId, { silent: true })
  }

  function startStream(id: string): void {
    if (streaming.value && currentId === id) return

    stopStream()
    streamError.value = ''
    streaming.value = true

    const baseline = streamSeq
    const isRunningOperation = (): boolean => status.value === 'patching'
    const isReplay = (seq: number): boolean => {
      if (seq && seq <= baseline) return true
      if (seq > streamSeq) streamSeq = seq

      return false
    }

    closeStream = openGameStream(id, {
      onOpen: () => {
        streamError.value = ''
      },
      onStage: (payload, seq) => {
        if (isReplay(seq)) return

        const index = stages.value.findIndex((stage) => stage.stage === payload.stage)

        if (index >= 0) stages.value[index] = payload
        else stages.value = [...stages.value, payload]

        if (payload.status === 'running') streamError.value = ''

        if (game.value && payload.status === 'running') {
          game.value.status = payload.stage as Game['status']
        }
      },
      onArtifact: (payload, seq) => {
        if (isReplay(seq)) return
        if (payload.kind === 'spec') spec.value = payload.content as GameSpec
        if (payload.kind === 'config') config.value = payload.content as GameConfig
        if (payload.kind === 'meta') {
          const meta = payload.content as GameMeta

          if (game.value) {
            game.value.title = meta.title
            game.value.description = meta.description
            game.value.cover = meta.cover
          }
        }
      },
      onDone: (_payload, seq) => {
        if (isReplay(seq) || !isRunningOperation()) return

        stopStream()
        refreshKey.value++
        void load(id, { silent: true })
      },
      onError: (payload, seq) => {
        if (isReplay(seq) || !isRunningOperation()) return

        streamError.value = friendlyStreamError(payload.code, payload.message)
        stopStream()
        void load(id, { silent: true })
      },
      onConnectionError: () => {
        streamError.value = 'Связь с сервером прервалась. Переподключаемся…'
      },
    })
  }

  function friendlyStreamError(code: string, message: string): string {
    switch (code) {
      case 'llm_unavailable':
        return 'Модель недоступна — правка отменена, токены вернулись на баланс.'
      case 'llm_budget_exceeded':
        return 'Исчерпан бюджет вызовов модели — токены вернулись на баланс.'
      case 'validation_failed':
        return 'Правка не прошла проверку — токены вернулись на баланс.'
      case 'timeout':
        return 'Модель не успела ответить — токены вернулись на баланс.'
      default:
        return message
    }
  }

  function stopStream(): void {
    closeStream?.()
    closeStream = null
    streaming.value = false
  }

  async function create(input: CreateGameInput): Promise<string> {
    actionPending.value = 'create'
    try {
      const response = await createGame(input)

      return response.slug
    } finally {
      actionPending.value = ''
    }
  }

  async function applyPatch(text: string): Promise<void> {
    if (!currentId) return
    actionPending.value = 'patch'
    streamError.value = ''
    try {
      await patchGame(currentId, text)
      await refresh()
    } finally {
      actionPending.value = ''
    }
  }

  async function applyConfig(update: Parameters<typeof updateGameConfig>[1]): Promise<void> {
    if (!currentId) return
    actionPending.value = 'config'
    try {
      await updateGameConfig(currentId, update)
      refreshKey.value++
      await refresh()
    } finally {
      actionPending.value = ''
    }
  }

  async function rollback(version: number): Promise<void> {
    if (!currentId) return
    actionPending.value = 'rollback'
    try {
      await rollbackGame(currentId, version)
      refreshKey.value++
      await refresh()
    } finally {
      actionPending.value = ''
    }
  }

  async function repair(runtimeError?: string): Promise<void> {
    if (!currentId) return
    actionPending.value = 'fix'
    streamError.value = ''
    try {
      await fixGame(currentId, runtimeError)
      await refresh()
    } finally {
      actionPending.value = ''
    }
  }

  async function acceptReviewDecision(): Promise<string> {
    if (!currentId) return ''

    actionPending.value = 'review'
    try {
      streamError.value = ''
      const response = await acceptReview(currentId)
      await refresh()

      return response.resume_from ?? ''
    } finally {
      actionPending.value = ''
    }
  }

  async function regenerateCode(): Promise<string> {
    if (!currentId) return ''

    actionPending.value = 'review'
    try {
      streamError.value = ''
      const response = await regenerateReview(currentId)
      await refresh()

      return response.resume_from ?? ''
    } finally {
      actionPending.value = ''
    }
  }

  async function retry(): Promise<string> {
    if (!currentId) return ''

    actionPending.value = 'retry'
    try {
      streamError.value = ''
      const response = await retryGame(currentId)
      await refresh()

      return response.resume_from ?? ''
    } finally {
      actionPending.value = ''
    }
  }

  async function cancel(): Promise<void> {
    if (!currentId) return
    actionPending.value = 'cancel'
    try {
      await cancelGame(currentId)
      stopStream()
      await refresh()
    } finally {
      actionPending.value = ''
    }
  }

  async function saveMeta(meta: Partial<GameMeta>): Promise<void> {
    if (!currentId) return
    actionPending.value = 'meta'
    try {
      await updateGameMeta(currentId, meta)
      await refresh()
    } finally {
      actionPending.value = ''
    }
  }

  async function generateMeta(): Promise<void> {
    if (!currentId) return
    actionPending.value = 'meta-generate'
    streamError.value = ''
    try {
      await generateGameMeta(currentId)
      await refresh()
    } finally {
      actionPending.value = ''
    }
  }

  async function remove(): Promise<void> {
    if (!currentId) return
    actionPending.value = 'delete'
    try {
      await deleteGame(currentId)
      stopStream()
    } finally {
      actionPending.value = ''
    }
  }

  async function setVisibility(visibility: Visibility): Promise<void> {
    if (!currentId) return
    actionPending.value = 'visibility'
    try {
      await updateGameVisibility(currentId, visibility)
      await refresh()
    } finally {
      actionPending.value = ''
    }
  }

  async function vote(value: number): Promise<void> {
    if (!currentId) return
    actionPending.value = 'rate'
    try {
      const response = await rateGame(currentId, value)
      rating.value = response.aggregate
      myRating.value = response.rating
    } finally {
      actionPending.value = ''
    }
  }

  function reset(): void {
    stopStream()
    currentId = ''
    game.value = null
    stages.value = []
    spec.value = null
    config.value = null
    rating.value = { average: 0, total: 0, counts: [] }
    myRating.value = 0
    operation.value = null
    error.value = ''
    streamError.value = ''
    refreshKey.value = 0
    actionPending.value = ''
  }

  return {
    game,
    stages,
    spec,
    config,
    rating,
    myRating,
    operation,
    loading,
    error,
    streamError,
    streaming,
    refreshKey,
    actionPending,
    status,
    isReady,
    isFailed,
    isActive,
    isGenerating,
    pipelineStages,
    stageProgress,
    load,
    refresh,
    create,
    applyPatch,
    applyConfig,
    cancel,
    retry,
    acceptReviewDecision,
    regenerateCode,
    saveMeta,
    generateMeta,
    remove,
    setVisibility,
    rollback,
    repair,
    vote,
    reset,
    stopStream,
  }
})
