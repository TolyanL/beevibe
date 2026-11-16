// NOTE: an LLM call has no real progress, so the bar eases towards a ceiling and waits there.
export const PROGRESS_CEILING = 95

const CURVE_STEEPNESS = 3.2
const MIN_EXPECTED_SECONDS = 10

const STAGE_BOUNDS = [
  { until: 0.05, label: 'Отправляем запрос модели' },
  { until: 0.25, label: 'Модель обдумывает ответ' },
  { until: 0.65, label: 'Собираем результат' },
] as const

const STAGE_LAST = 'Осталось чуть-чуть — ответ вот-вот придёт'

function expectedFloor(expectedSeconds: number): number {
  return Math.max(MIN_EXPECTED_SECONDS, expectedSeconds)
}

export function estimatedPercent(elapsedSeconds: number, expectedSeconds: number): number {
  const ratio = Math.max(0, elapsedSeconds) / expectedFloor(expectedSeconds)
  const eased = 1 - Math.exp(-CURVE_STEEPNESS * ratio)

  return Math.min(PROGRESS_CEILING, Math.round(eased * PROGRESS_CEILING))
}

export function estimatedStage(elapsedSeconds: number, expectedSeconds: number): string {
  const ratio = Math.max(0, elapsedSeconds) / expectedFloor(expectedSeconds)

  return STAGE_BOUNDS.find((bound) => ratio < bound.until)?.label ?? STAGE_LAST
}
