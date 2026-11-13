import type { TokenOpReason, VersionSource } from '@/api/types'

type TokenIcon = 'spark' | 'wand' | 'code' | 'refresh' | 'coin' | 'plus' | 'bolt' | 'star'

export interface TokenOpMeta {
  label: string
  icon: TokenIcon
  tone: 'in' | 'out' | 'neutral'
}

export const TOKEN_OP_META: Record<TokenOpReason, TokenOpMeta> = {
  signup_grant: { label: 'Стартовый бонус', icon: 'spark', tone: 'in' },
  generate: { label: 'Сборка игры', icon: 'wand', tone: 'out' },
  patch: { label: 'Правка игры', icon: 'code', tone: 'out' },
  meta: { label: 'Название и описание', icon: 'star', tone: 'out' },
  topup: { label: 'Пополнение баланса', icon: 'plus', tone: 'in' },
  refund: { label: 'Возврат токенов', icon: 'refresh', tone: 'in' },
}

export function tokenOpMeta(reason: string): TokenOpMeta {
  return TOKEN_OP_META[reason as TokenOpReason] ?? { label: reason, icon: 'bolt', tone: 'neutral' }
}

type GenerationIcon = 'spark' | 'eye' | 'wand' | 'settings' | 'refresh' | 'back'

export interface GenerationSourceMeta {
  icon: GenerationIcon
  tone: string
}

export const GENERATION_SOURCE_META: Record<VersionSource, GenerationSourceMeta> = {
  generate: {
    icon: 'spark',
    tone: 'border-bee/40 bg-bee/10 text-bee-deep dark:text-bee',
  },
  review: {
    icon: 'eye',
    tone: 'border-emerald-500/25 bg-emerald-500/10 text-emerald-500',
  },
  patch: {
    icon: 'wand',
    tone: 'border-sky-500/25 bg-sky-500/10 text-sky-500',
  },
  config: {
    icon: 'settings',
    tone: 'border-line-strong bg-canvas/50 text-muted',
  },
  repair: {
    icon: 'refresh',
    tone: 'border-red-500/25 bg-red-500/10 text-red-500',
  },
  rollback: {
    icon: 'back',
    tone: 'border-violet-500/25 bg-violet-500/10 text-violet-500',
  },
}

export function generationSourceMeta(source: string): GenerationSourceMeta {
  return (
    GENERATION_SOURCE_META[source as VersionSource] ?? {
      icon: 'spark',
      tone: 'border-line-strong bg-canvas/50 text-muted',
    }
  )
}
