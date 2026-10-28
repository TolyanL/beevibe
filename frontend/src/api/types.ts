export type ThemeMode = 'light' | 'dark'

export type Genre = 'arcade' | 'snake' | 'clicker' | 'platformer' | 'puzzle' | 'custom'
export type MapSize = 'small' | 'medium' | 'large'
export type GameStyle = 'pixel' | 'cartoon' | 'neon' | 'minimal' | 'custom'
export type CatalogSort = 'rating' | 'hours' | 'fresh'

export type Visibility = 'public' | 'private'

export interface GameCover {
  palette: string[]
  style: string
  seed: string
}

export interface GameMeta {
  title: string
  description: string
  cover?: GameCover
}

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
  | 'needs_decision'

export interface User {
  id: string
  email: string
  display_name: string
  balance: number
}

export interface UserEnvelope {
  user: User
}

export interface ProfileGame {
  id: string
  prompt: string
  title: string
  slug: string
  genre: string
  status: GameStatus
  visibility: Visibility
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
    meta: number
  }
}

export type UserInfo = User

export interface ListUserGamesPage {
  items: ProfileGame[]
  total: number
}

export interface TokenOpsPage {
  items: ProfileTokenOp[]
  total: number
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

export interface GamePublicCard {
  id: string
  prompt: string
  genre: string
  style: string
  map_size: string
  created_at: string
  rating: RatingAggregate
  playtime: PlaytimeAggregate
  my_rating: number
  is_owner: boolean
  custom_genre?: string
  custom_style?: string
  title: string
  description: string
  slug: string
  author: string
  cover?: GameCover
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
  status: 'running' | 'done' | 'error'
  durationMs: number
  issues?: number
  operation?: string
}

export interface ReviewIssue {
  code: string
  message: string
  line?: number
}

export interface ActiveOperation {
  kind: 'patch' | 'fix' | 'config' | 'meta'
  started_at: string
}

export interface Game {
  id: string
  prompt: string
  params?: {
    genre?: string
    style?: string
    map_size?: string
    tempo?: number
  }
  status: GameStatus
  visibility: Visibility
  title: string
  description: string
  slug: string
  cover?: GameCover
  stages: GameStage[]
  spec?: GameSpec
  config?: GameConfig
  versions: GameVersion[]
  cost: GameCost
  rating: RatingAggregate
  my_rating: number
  operation?: ActiveOperation | null
  review_issues?: ReviewIssue[]
  cursor: number
}

export interface CatalogCard {
  id: string
  prompt: string
  title: string
  description: string
  cover?: GameCover
  author: string
  slug: string
  custom_genre?: string
  custom_style?: string
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
  total: number
}

export interface CreateGameInput {
  prompt: string
  title?: string
  customGenre?: string
  customStyle?: string
  genre?: Genre
  tempo?: number
  mapSize?: MapSize
  style?: GameStyle
}

export interface CreateGameResponse {
  slug: string
  id: string
  status: GameStatus
}

export interface GameJobResponse {
  status: 'patching'
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
  status: 'running' | 'done' | 'error'
  durationMs: number
  issues?: number
  operation?: string
}

export interface ArtifactEventPayload {
  kind: 'spec' | 'config' | 'version' | 'meta'
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
