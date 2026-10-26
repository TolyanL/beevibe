import { request } from './client'
import type {
  CreateGameInput,
  CreateGameResponse,
  Game,
  PatchGameResponse,
  PlaytimeResponse,
  RateGameResponse,
  VersionResponse,
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

export function patchGame(id: string, text: string): Promise<PatchGameResponse> {
  return request<PatchGameResponse>(`/games/${id}/patch`, {
    method: 'POST',
    body: { text },
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

export function fixGame(id: string, error?: string): Promise<VersionResponse> {
  return request<VersionResponse>(`/games/${id}/fix`, {
    method: 'POST',
    body: error ? { error } : {},
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
