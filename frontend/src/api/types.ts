export type ThemeMode = 'light' | 'dark'

export type Genre = 'arcade' | 'snake' | 'clicker' | 'platformer' | 'puzzle' | 'custom'
export type MapSize = 'small' | 'medium' | 'large'
export type GameStyle = 'pixel' | 'cartoon' | 'neon' | 'minimal'
export type CatalogSort = 'rating' | 'hours' | 'fresh'

export type GameStatus =
  | 'queued'
  | 'analyzing'
  | 'planning'
  | 'coding'
  | 'assets'
  | 'reviewing'
  | 'assembling'
  | 'checking'
  | 'ready'
  | 'failed'
  | 'patching'

export interface User {
  id: string
  email: string
  balance: number
}

export interface UserEnvelope {
  user: User
}

export interface ProfileGame {
  id: string
  prompt: string
  genre: string
  status: GameStatus
  tokens_cost: number
  model_calls: number
  fixes: number
  created_at: string
}

export type TokenOpReason = 'signup_grant' | 'generate' | 'patch' | 'topup' | 'refund'

export interface ProfileTokenOp {
  game_id?: string | null
  delta: number
  reason: TokenOpReason
  created_at: string
}

export interface Profile {
  balance: number
  games: ProfileGame[]
  token_ops: ProfileTokenOp[]
  prices: {
    generate: number
    patch: number
  }
}

export interface RatingAggregate {
  average: number
  total: number
  counts: number[]
}

export interface PlaytimeAggregate {
  sessions: number
  seconds: number
  hours: number
}

export interface GameSpec {
  genre: string
  mechanic: string
  controls: string
  win_condition: string
  lose_condition: string
  palette: string[]
  map_size: string
  style: string
  audience: string
  assets: string[]
}

export interface GameConfig {
  tempo: number
  palette: string[]
  map_size: string
  difficulty: string
  controls: string
  style: string
}

export type VersionSource = 'generate' | 'review' | 'patch' | 'config' | 'repair' | 'rollback'

export interface GameVersion {
  id: string
  number: number
  source: VersionSource
  note?: string
  tokens_cost: number
  model_calls: number
  prompt_version?: string
  created_at: string
}

export interface GameCost {
  tokens: number
  calls: number
  fixes: number
}

export interface GameStage {
  stage: string
  status: 'running' | 'done'
  durationMs: number
  issues?: number
}

export interface Game {
  status: GameStatus
  stages: GameStage[]
  spec?: GameSpec
  config?: GameConfig
  versions: GameVersion[]
  cost: GameCost
  rating: RatingAggregate
  my_rating: number
}

export interface CatalogCard {
  id: string
  prompt: string
  genre: string
  map_size: string
  style: string
  created_at: string
  rating: RatingAggregate
  playtime: PlaytimeAggregate
  tokens_cost: number
  model_calls: number
}

export interface CatalogPage {
  items: CatalogCard[]
}

export interface CreateGameInput {
  prompt: string
  genre?: Genre
  tempo?: number
  mapSize?: MapSize
  style?: GameStyle
}

export interface CreateGameResponse {
  id: string
  status: GameStatus
}

export interface PatchGameResponse {
  versionId: string
  tokensSpent: number
}

export interface VersionResponse {
  versionId: string
}

export interface RateGameResponse {
  rating: number
  aggregate: RatingAggregate
}

export interface TopupResponse {
  balance: number
}

export interface PlaytimeResponse {
  totalSeconds: number
}

export interface ApiErrorEnvelope {
  error?: {
    code?: string
    message?: string
  }
}

export interface StageEventPayload {
  stage: string
  status: 'running' | 'done'
  durationMs: number
  issues?: number
}

export interface ArtifactEventPayload {
  kind: 'spec' | 'config'
  content: unknown
}

export interface DoneEventPayload {
  versionId: string
  cost: {
    tokens: number
    calls: number
  }
}

export interface ErrorEventPayload {
  code: string
  message: string
}
