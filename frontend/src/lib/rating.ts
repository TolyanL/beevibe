export interface RatingTone {
  label: string
  text: string
  background: string
  border: string
  solid: string
  hex: string
}

export function ratingTone(score: number): RatingTone {
  if (!score || score <= 0) {
    return {
      label: 'Без оценок',
      text: 'text-faint',
      background: 'bg-faint/10',
      border: 'border-faint/20',
      solid: 'bg-faint',
      hex: '#98917f',
    }
  }

  if (score >= 4.5) {
    return {
      label: 'Восторг',
      text: 'text-emerald-500',
      background: 'bg-emerald-500/10',
      border: 'border-emerald-500/25',
      solid: 'bg-emerald-500',
      hex: '#22c55e',
    }
  }

  if (score >= 3.5) {
    return {
      label: 'Хорошо',
      text: 'text-lime-500',
      background: 'bg-lime-500/10',
      border: 'border-lime-500/25',
      solid: 'bg-lime-500',
      hex: '#84cc16',
    }
  }

  if (score >= 2.5) {
    return {
      label: 'Нормально',
      text: 'text-yellow-500',
      background: 'bg-yellow-500/10',
      border: 'border-yellow-500/25',
      solid: 'bg-yellow-500',
      hex: '#eab308',
    }
  }

  if (score >= 1.5) {
    return {
      label: 'Слабо',
      text: 'text-orange-500',
      background: 'bg-orange-500/10',
      border: 'border-orange-500/25',
      solid: 'bg-orange-500',
      hex: '#f97316',
    }
  }

  return {
    label: 'Плохо',
    text: 'text-red-500',
    background: 'bg-red-500/10',
    border: 'border-red-500/25',
    solid: 'bg-red-500',
    hex: '#ef4444',
  }
}

export function ratingPercent(score: number): number {
  return Math.max(0, Math.min(100, (score / 5) * 100))
}
