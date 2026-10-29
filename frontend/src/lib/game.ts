import type { GameStatus } from '@/api/types'

export interface StatusMeta {
  label: string
  short: string
  tone: 'neutral' | 'live' | 'good' | 'bad' | 'warn'
  description: string
}

const META: Record<GameStatus, StatusMeta> = {
  queued: {
    label: 'В очереди',
    short: 'Очередь',
    tone: 'neutral',
    description: 'Игра ждёт свободного агента — сборка начнётся сама.',
  },
  analyzing: {
    label: 'Разбор идеи',
    short: 'Разбор',
    tone: 'live',
    description: 'Агент читает идею: жанр, настроение, ограничения.',
  },
  planning: {
    label: 'План и правила',
    short: 'План',
    tone: 'live',
    description: 'Придумываем правила: что делает игрок, как победить, как проиграть.',
  },
  coding: {
    label: 'Пишем код',
    short: 'Код',
    tone: 'live',
    description: 'Пишем игру: один файл, который работает в браузере.',
  },
  assets: {
    label: 'Графика',
    short: 'Графика',
    tone: 'live',
    description: 'Подбираем палитру, шрифты и графику под выбранный стиль.',
  },
  reviewing: {
    label: 'Ревью',
    short: 'Ревью',
    tone: 'live',
    description: 'Смотрим свежим взглядом: запускается ли игра и совпадает ли с планом.',
  },
  needs_decision: {
    label: 'Нужно решение',
    short: 'Решение',
    tone: 'warn',
    description: 'Проверка нашла замечания. Оставьте код как есть или попросите переписать его.',
  },
  assembling: {
    label: 'Сборка',
    short: 'Сборка',
    tone: 'live',
    description: 'Склеиваем всё в один файл, который открывается в браузере.',
  },
  checking: {
    label: 'Проверка',
    short: 'Проверка',
    tone: 'live',
    description: 'Последняя проверка перед публикацией.',
  },
  ready: {
    label: 'Готова',
    short: 'Готова',
    tone: 'good',
    description: 'Игра собрана: открывайте и играйте.',
  },
  failed: {
    label: 'Ошибка',
    short: 'Ошибка',
    tone: 'bad',
    description:
      'Собрать не получилось. Запустите починку — агент посмотрит на ошибку и попробует снова.',
  },
  patching: {
    label: 'Правим',
    short: 'Правка',
    tone: 'live',
    description: 'Вносим правки и собираем новую версию игры.',
  },
}

export function statusMeta(status: GameStatus | string): StatusMeta {
  return META[status as GameStatus] ?? META.queued
}

export const PIPELINE = [
  { id: 'queued', label: 'Очередь', short: 'Очередь' },
  { id: 'analyzing', label: 'Разбор идеи', short: 'Разбор' },
  { id: 'planning', label: 'План и правила', short: 'План' },
  { id: 'coding', label: 'Код', short: 'Код' },
  { id: 'assets', label: 'Графика', short: 'Графика' },
  { id: 'reviewing', label: 'Ревью', short: 'Ревью' },
  { id: 'assembling', label: 'Сборка', short: 'Сборка' },
  { id: 'checking', label: 'Проверка', short: 'Проверка' },
] as const

const OPERATION_STAGE_LABELS: Record<string, string> = {
  patching: 'Правка игры',
  meta: 'Название и описание',
}

export function stageLabel(stage: string, operation?: string): string {
  if (stage === 'patching') {
    if (operation === 'fix') return 'Починка игры'
    if (operation === 'meta') return 'Название и описание'
  }

  return PIPELINE.find((item) => item.id === stage)?.label ?? OPERATION_STAGE_LABELS[stage] ?? stage
}

export function genreLabel(genre: string, custom?: string): string {
  if (genre === 'custom' && custom && custom.trim()) return custom.trim()

  const labels: Record<string, string> = {
    arcade: 'Аркада',
    snake: 'Змейка',
    clicker: 'Кликер',
    platformer: 'Платформер',
    puzzle: 'Головоломка',
    custom: 'Своя механика',
  }

  return labels[genre] ?? ''
}

const STYLE_PALETTES: Record<string, [string, string]> = {
  pixel: ['#84cc16', '#0d9488'],
  cartoon: ['#fb7185', '#f59e0b'],
  neon: ['#ffcc00', '#7c3aed'],
  minimal: ['#f3f0e6', '#98917f'],
  custom: ['#22d3ee', '#2563eb'],
}

export function stylePalette(style: string): [string, string] {
  return STYLE_PALETTES[style] ?? ['#ffcc00', '#7c3aed']
}

export function styleLabel(style: string, custom?: string): string {
  if (style === 'custom' && custom && custom.trim()) return custom.trim()

  const labels: Record<string, string> = {
    pixel: 'Пиксель',
    cartoon: 'Мультяшный',
    neon: 'Неон',
    minimal: 'Минимализм',
  }

  return labels[style] ?? ''
}

const shortTitleWords = 5

export function shortGameTitle(
  title: string | undefined,
  prompt: string,
  words = shortTitleWords,
): string {
  const clean = (title ?? '').trim()
  const promptText = prompt.trim()

  if (clean && clean !== promptText) return clean

  const parts = promptText.split(/\s+/).filter(Boolean)

  if (parts.length === 0) return 'Игра'

  const truncated = parts.length > words
  const head = parts
    .slice(0, words)
    .join(' ')
    .replace(/[,:;.\-—]+$/, '')

  return truncated ? `${head}…` : head
}

export function gameDescription(description: string | undefined, prompt: string): string {
  const clean = (description ?? '').trim()

  return clean || prompt.trim()
}

export function mapSizeLabel(size: string): string {
  const labels: Record<string, string> = {
    small: 'Компактная',
    medium: 'Средняя',
    large: 'Большая',
  }

  return labels[size] ?? ''
}

export function versionSourceLabel(source: string): string {
  const labels: Record<string, string> = {
    generate: 'Генерация',
    review: 'Ревью',
    patch: 'Правка',
    config: 'Настройки',
    repair: 'Починка',
    rollback: 'Откат',
  }

  return labels[source] ?? source
}
