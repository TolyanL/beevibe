import { describe, expect, it } from 'vitest'

import { estimatedPercent, estimatedStage, PROGRESS_CEILING } from '@/lib/progress'

const NORMAL_SECONDS = 780
const FAST_SECONDS = 300
const PATCH_SECONDS = 180

describe('estimatedPercent', () => {
  it('starts at zero and never leaves the ceiling', () => {
    expect(estimatedPercent(0, NORMAL_SECONDS)).toBe(0)
    expect(estimatedPercent(NORMAL_SECONDS * 10, NORMAL_SECONDS)).toBe(PROGRESS_CEILING)
  })

  it('is close to the ceiling when a real run finishes', () => {
    expect(estimatedPercent(798, NORMAL_SECONDS)).toBeGreaterThanOrEqual(88)
    expect(estimatedPercent(312, FAST_SECONDS)).toBeGreaterThanOrEqual(88)
    expect(estimatedPercent(180, PATCH_SECONDS)).toBeGreaterThanOrEqual(88)
  })

  it('moves forward only', () => {
    const points = [0, 30, 90, 300, 600, 780, 1200]
    const values = points.map((elapsed) => estimatedPercent(elapsed, NORMAL_SECONDS))

    expect(values).toEqual([...values].sort((a, b) => a - b))
  })

  it('guards against a zero or negative estimate', () => {
    expect(estimatedPercent(5, 0)).toBe(estimatedPercent(5, 10))
    expect(estimatedPercent(5, -100)).toBeGreaterThan(0)
  })
})

describe('estimatedStage', () => {
  it('moves through the phases in order', () => {
    expect(estimatedStage(0, NORMAL_SECONDS)).toBe('Отправляем запрос модели')
    expect(estimatedStage(60, NORMAL_SECONDS)).toBe('Модель обдумывает ответ')
    expect(estimatedStage(300, NORMAL_SECONDS)).toBe('Собираем результат')
    expect(estimatedStage(780, NORMAL_SECONDS)).toContain('Осталось чуть-чуть')
  })

  it('scales with the estimate instead of the wall clock', () => {
    expect(estimatedStage(60, NORMAL_SECONDS)).not.toBe(estimatedStage(60, PATCH_SECONDS))
  })
})
