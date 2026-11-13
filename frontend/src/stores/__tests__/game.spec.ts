import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

import { getGame } from '@/api/games'
import { openGameStream } from '@/api/sse'
import { useGameStore } from '@/stores/game'
import type { Game, GameStage } from '@/api/types'

vi.mock('@/api/games', () => ({ getGame: vi.fn() }))
vi.mock('@/api/sse', () => ({ openGameStream: vi.fn(() => () => {}) }))

function game(id: string, cursor: number): Game {
  return {
    id,
    prompt: 'змейка',
    params: { speed: 'fast' },
    status: 'analyzing',
    visibility: 'private',
    title: '',
    description: '',
    slug: id,
    stages: [],
    versions: [],
    cost: { tokens: 0, calls: 0, fixes: 0 },
    rating: { average: 0, total: 0, counts: [] },
    my_rating: 0,
    cursor,
  }
}

function lastStreamHandlers() {
  const calls = vi.mocked(openGameStream).mock.calls
  const handlers = calls.at(-1)?.[1]
  if (!handlers?.onStage) throw new Error('stream was not opened')

  return handlers
}

describe('game store stream baseline', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.mocked(openGameStream).mockClear()
    vi.mocked(getGame).mockReset()
  })

  it('applies events of a game opened after another one', async () => {
    vi.mocked(getGame)
      .mockResolvedValueOnce(game('first', 15))
      .mockResolvedValueOnce(game('second', 0))

    const store = useGameStore()

    await store.load('first')
    await store.load('second')

    lastStreamHandlers().onStage({ stage: 'analyzing', status: 'done' } as GameStage, 1)

    expect(store.stages.map((stage) => stage.stage)).toEqual(['analyzing'])
    expect(store.stageProgress).toBe(33)
  })

  it('ignores replayed history of the game it reconnects to', async () => {
    vi.mocked(getGame).mockResolvedValue(game('same', 7))

    const store = useGameStore()

    await store.load('same')
    lastStreamHandlers().onStage({ stage: 'coding', status: 'done' } as GameStage, 3)

    expect(store.stages).toEqual([])
  })
})
