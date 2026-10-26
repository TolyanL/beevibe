import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { createGame, fixGame, getGame, patchGame, rateGame, rollbackGame, updateGameConfig } from '@/api/games'
import { messageFromError } from '@/api/client'
import { openGameStream } from '@/api/sse'
import type {
  CreateGameInput,
  Game,
  GameConfig,
  GameSpec,
  GameStage,
  RatingAggregate,
} from '@/api/types'

const PIPELINE_STAGES = 8

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
  const loading = ref(false)
  const error = ref('')
  const streamError = ref('')
  const streaming = ref(false)
  const refreshKey = ref(0)
  const actionPending = ref('')

  let closeStream: (() => void) | null = null
  let currentId = ''

  const status = computed(() => game.value?.status ?? 'queued')
  const isReady = computed(() => status.value === 'ready')
  const isFailed = computed(() => status.value === 'failed')
  const isActive = computed(() => ACTIVE_STATUSES.has(status.value))
  const stageProgress = computed(() => {
    if (isReady.value) return 100
    if (!stages.value.length) return 0

    const done = stages.value.filter((stage) => stage.status === 'done').length

    return Math.min(100, Math.round((done / PIPELINE_STAGES) * 100))
  })

  function hydrate(next: Game): void {
    game.value = next
    stages.value = next.stages ?? []
    spec.value = next.spec ?? null
    config.value = next.config ?? null
    rating.value = next.rating ?? { average: 0, total: 0, counts: [] }
    myRating.value = next.my_rating ?? 0
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

    closeStream = openGameStream(id, {
      onOpen: () => {
        streamError.value = ''
      },
      onStage: (payload) => {
        const index = stages.value.findIndex((stage) => stage.stage === payload.stage)

        if (index >= 0) stages.value[index] = payload
        else stages.value = [...stages.value, payload]

        if (game.value && payload.status === 'running') {
          game.value.status = payload.stage as Game['status']
        }
      },
      onArtifact: (payload) => {
        if (payload.kind === 'spec') spec.value = payload.content as GameSpec
        if (payload.kind === 'config') config.value = payload.content as GameConfig
      },
      onDone: () => {
        stopStream()
        refreshKey.value++
        void load(id, { silent: true })
      },
      onError: (payload) => {
        streamError.value = payload.message
        if (game.value) game.value.status = 'failed'
        stopStream()
      },
      onConnectionError: () => {
        streamError.value = 'Связь с конвейером прервана. Переподключаемся…'
      },
    })
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
      return response.id
    } finally {
      actionPending.value = ''
    }
  }

  async function applyPatch(text: string): Promise<void> {
    if (!currentId) return
    actionPending.value = 'patch'
    try {
      await patchGame(currentId, text)
      refreshKey.value++
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
    try {
      await fixGame(currentId, runtimeError)
      refreshKey.value++
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
    stageProgress,
    load,
    refresh,
    create,
    applyPatch,
    applyConfig,
    rollback,
    repair,
    vote,
    reset,
    stopStream,
  }
})
