import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  deleteGame,
  getGame,
  listUserGames,
  patchGame,
  rateGame,
  recordPlaytime,
  updateGameConfig,
} from '../games'
import { getGamePublicCard } from '../catalog'

interface Captured {
  url: string
  init: RequestInit
}

let captured: Captured[] = []

function mockFetch(payload: unknown, status = 200): void {
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: RequestInfo | URL, init: RequestInit = {}) => {
      captured.push({ url: String(input), init })

      return new Response(JSON.stringify(payload), {
        status,
        headers: { 'Content-Type': 'application/json' },
      })
    }),
  )
}

beforeEach(() => {
  captured = []
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('game endpoints address a game by slug', () => {
  it('builds the rate path from the given reference as is', async () => {
    mockFetch({ rating: 4, aggregate: { average: 4, total: 1, counts: [0, 0, 0, 1, 0] } })

    await rateGame('prostoy-2d-platformer-da2f', 4)

    expect(captured[0]?.url).toBe('/api/games/prostoy-2d-platformer-da2f/rate')
    expect(captured[0]?.init.method).toBe('POST')
    expect(captured[0]?.init.body).toBe(JSON.stringify({ value: 4 }))
  })

  it('keeps the reference untouched for every route under /games/{id}', async () => {
    mockFetch({})

    await getGame('zmeyka-1a2b')
    await patchGame('zmeyka-1a2b', 'сделай сложнее')
    await updateGameConfig('zmeyka-1a2b', { tempo: 7 })
    await recordPlaytime('zmeyka-1a2b', 60)
    await deleteGame('zmeyka-1a2b')

    expect(captured.map((call) => `${call.init.method ?? 'GET'} ${call.url}`)).toEqual([
      'GET /api/games/zmeyka-1a2b',
      'POST /api/games/zmeyka-1a2b/patch',
      'PATCH /api/games/zmeyka-1a2b/config',
      'POST /api/games/zmeyka-1a2b/playtime',
      'DELETE /api/games/zmeyka-1a2b',
    ])
  })

  it('requests the public card by the same reference', async () => {
    mockFetch({ slug: 'zmeyka-1a2b' })

    await getGamePublicCard('zmeyka-1a2b')

    expect(captured[0]?.url).toBe('/api/games/zmeyka-1a2b/public')
  })

  it('passes list filters as query parameters', async () => {
    mockFetch({ items: [], total: 0 })

    await listUserGames({ status: 'ready', sort: 'tokens_cost', limit: 5, offset: 10 })

    expect(captured[0]?.url).toBe('/api/games?status=ready&sort=tokens_cost&limit=5&offset=10')
  })
})
