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
    description: 'Задача принята. Конвейер скоро начнёт работу.',
  },
  analyzing: {
    label: 'Анализ промпта',
    short: 'Анализ',
    tone: 'live',
    description: 'Агент разбирает идею, жанр и ограничения.',
  },
  planning: {
    label: 'Планирование',
    short: 'План',
    tone: 'live',
    description: 'Собираем механику, условия победы и структуру игры.',
  },
  coding: {
    label: 'Пишем код',
    short: 'Код',
    tone: 'live',
    description: 'Генерируем самодостаточную HTML5-игру.',
  },
  assets: {
    label: 'Ассеты',
    short: 'Ассеты',
    tone: 'live',
    description: 'Подбираем палитру и визуальные элементы.',
  },
  reviewing: {
    label: 'Ревью',
    short: 'Ревью',
    tone: 'live',
    description: 'Проверяем игру на ошибки и соответствие плану.',
  },
  needs_decision: {
    label: 'Нужно решение',
    short: 'Решение',
    tone: 'warn',
    description: 'Ревью нашло замечания. Оставьте код как есть или сгенерируйте его заново.',
  },
  assembling: {
    label: 'Сборка',
    short: 'Сборка',
    tone: 'live',
    description: 'Собираем артефакт в один playable-файл.',
  },
  checking: {
    label: 'Проверка',
    short: 'Проверка',
    tone: 'live',
    description: 'Финальные smoke-тесты перед публикацией.',
  },
  ready: {
    label: 'Готова',
    short: 'Готова',
    tone: 'good',
    description: 'Игра собрана и доступна для запуска.',
  },
  failed: {
    label: 'Ошибка',
    short: 'Ошибка',
    tone: 'bad',
    description: 'Конвейер не смог собрать игру. Можно попробовать починку.',
  },
  patching: {
    label: 'Патчим',
    short: 'Патч',
    tone: 'live',
    description: 'Вносим правки и собираем новую версию.',
  },
}

export function statusMeta(status: GameStatus | string): StatusMeta {
  return META[status as GameStatus] ?? META.queued
}

export const PIPELINE = [
  { id: 'queued', label: 'Очередь', short: 'Очередь' },
  { id: 'analyzing', label: 'Анализ', short: 'Анализ' },
  { id: 'planning', label: 'Планирование', short: 'План' },
  { id: 'coding', label: 'Код', short: 'Код' },
  { id: 'assets', label: 'Ассеты', short: 'Ассеты' },
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

// Каталог показывает короткое имя: название автора, иначе первые слова промпта.
export function shortGameTitle(
  title: string | undefined,
  prompt: string,
  words = shortTitleWords,
): string {
  const clean = (title ?? '').trim()
  const promptText = prompt.trim()

  // Сервер может подставить промпт вместо пустого названия — тогда его и сокращаем.
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
    patch: 'Патч',
    config: 'Настройки',
    repair: 'Починка',
    rollback: 'Откат',
  }

  return labels[source] ?? source
}
