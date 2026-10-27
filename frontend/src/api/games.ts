import { request } from './client'
import type {
  CreateGameInput,
  CreateGameResponse,
  Game,
  GameMeta,
  ListUserGamesPage,
  GameJobResponse,
  PlaytimeResponse,
  RateGameResponse,
  TokenOpsPage,
  VersionResponse,
  Visibility,
} from './types'

export function createGame(input: CreateGameInput): Promise<CreateGameResponse> {
  return request<CreateGameResponse>('/games', {
    method: 'POST',
    body: input,
    headers: {
      'Idempotency-Key': crypto.randomUUID(),
    },
  })
}

export function getGame(id: string, signal?: AbortSignal): Promise<Game> {
  return request<Game>(`/games/${id}`, { signal })
}

export function patchGame(id: string, text: string): Promise<GameJobResponse> {
  return request<GameJobResponse>(`/games/${id}/patch`, {
    method: 'POST',
    body: { text },
    headers: { 'Idempotency-Key': crypto.randomUUID() },
  })
}

export interface ConfigUpdate {
  tempo?: number
  palette?: string[]
  mapSize?: 'small' | 'medium' | 'large'
  difficulty?: string
}

export function updateGameConfig(id: string, update: ConfigUpdate): Promise<VersionResponse> {
  return request<VersionResponse>(`/games/${id}/config`, {
    method: 'PATCH',
    body: update,
  })
}

export function rollbackGame(id: string, version: number): Promise<VersionResponse> {
  return request<VersionResponse>(`/games/${id}/rollback`, {
    method: 'POST',
    body: { version },
  })
}

export function fixGame(id: string, error?: string): Promise<GameJobResponse> {
  return request<GameJobResponse>(`/games/${id}/fix`, {
    method: 'POST',
    body: error ? { error } : {},
    headers: { 'Idempotency-Key': crypto.randomUUID() },
  })
}

export function retryGame(id: string): Promise<{ status: string; resume_from?: string }> {
  return request<{ status: string; resume_from?: string }>(`/games/${id}/retry`, { method: 'POST' })
}

export function cancelGame(id: string): Promise<{ status: string }> {
  return request<{ status: string }>(`/games/${id}/cancel`, { method: 'POST' })
}

export function deleteGame(id: string): Promise<{ deleted: boolean }> {
  return request<{ deleted: boolean }>(`/games/${id}`, { method: 'DELETE' })
}

export function updateGameVisibility(id: string, visibility: Visibility): Promise<{ visibility: Visibility }> {
  return request<{ visibility: Visibility }>(`/games/${id}/visibility`, {
    method: 'PATCH',
    body: { visibility },
  })
}

export function updateGameMeta(id: string, meta: Partial<GameMeta>): Promise<GameMeta> {
  return request<GameMeta>(`/games/${id}/meta`, { method: 'PATCH', body: meta })
}

export function generateGameMeta(id: string): Promise<{ status: string }> {
  return request<{ status: string }>(`/games/${id}/meta/generate`, {
    method: 'POST',
    headers: { 'Idempotency-Key': crypto.randomUUID() },
  })
}

export function rateGame(id: string, value: number): Promise<RateGameResponse> {
  return request<RateGameResponse>(`/games/${id}/rate`, {
    method: 'POST',
    body: { value },
  })
}

export function recordPlaytime(id: string, seconds: number): Promise<PlaytimeResponse> {
  return request<PlaytimeResponse>(`/games/${id}/playtime`, {
    method: 'POST',
    body: { seconds },
  })
}

export interface ListUserGamesQuery {
  status?: string
  sort?: 'created_at' | 'tokens_cost'
  limit?: number
  offset?: number
}

export function listUserGames(query: ListUserGamesQuery = {}): Promise<ListUserGamesPage> {
  const params = new URLSearchParams()

  if (query.status) params.set('status', query.status)
  if (query.sort) params.set('sort', query.sort)
  if (query.limit) params.set('limit', String(query.limit))
  if (query.offset) params.set('offset', String(query.offset))

  const suffix = params.size ? `?${params.toString()}` : ''

  return request<ListUserGamesPage>(`/games${suffix}`)
}

export function listTokenOps(query: { limit?: number; offset?: number } = {}): Promise<TokenOpsPage> {
  const params = new URLSearchParams()

  if (query.limit) params.set('limit', String(query.limit))
  if (query.offset) params.set('offset', String(query.offset))

  const suffix = params.size ? `?${params.toString()}` : ''

  return request<TokenOpsPage>(`/me/token-ops${suffix}`)
}
