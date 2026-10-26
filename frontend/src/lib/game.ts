import type { GameStatus } from '@/api/types'

export interface StatusMeta {
  label: string
  short: string
  tone: 'neutral' | 'live' | 'good' | 'bad'
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

export function stageLabel(stage: string): string {
  return PIPELINE.find((item) => item.id === stage)?.label ?? stage
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

  return labels[genre] ?? 'Игра'
}

export function styleLabel(style: string, custom?: string): string {
  if (style === 'custom' && custom && custom.trim()) return custom.trim()

  const labels: Record<string, string> = {
    pixel: 'Пиксель',
    cartoon: 'Мультяшный',
    neon: 'Неон',
    minimal: 'Минимализм',
  }

  return labels[style] ?? 'Стиль'
}

export function mapSizeLabel(size: string): string {
  const labels: Record<string, string> = {
    small: 'Компактная',
    medium: 'Средняя',
    large: 'Большая',
  }

  return labels[size] ?? size
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
