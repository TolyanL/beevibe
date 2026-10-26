export function formatTokens(value: number): string {
  if (!Number.isFinite(value)) return '0'
  if (Math.abs(value) >= 1_000_000) {
    return `${trim(value / 1_000_000)}M`
  }
  if (Math.abs(value) >= 1_000) {
    return `${trim(value / 1_000)}K`
  }
  return String(Math.round(value))
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat('ru-RU').format(Math.round(value))
}

export function formatPrice(value: number): string {
  if (value >= 1_000_000) return `${trim(value / 1_000_000)}M токенов`
  if (value >= 1_000) return `${trim(value / 1_000)}K токенов`

  return `${value} токенов`
}

function trim(value: number): string {
  return value.toFixed(value >= 10 ? 0 : 1).replace('.', ',').replace(/,0$/, '')
}

export function formatDate(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'

  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  }).format(date)
}

export function formatRelative(value: string): string {
  const date = new Date(value).getTime()
  if (Number.isNaN(date)) return '—'

  const diff = Date.now() - date
  const minutes = Math.round(diff / 60_000)

  if (minutes < 1) return 'только что'
  if (minutes < 60) return `${minutes} мин назад`

  const hours = Math.round(minutes / 60)
  if (hours < 24) return `${hours} ч назад`

  const days = Math.round(hours / 24)
  if (days < 7) return `${days} дн назад`

  return formatDate(value)
}

export function formatHours(hours: number): string {
  if (!hours) return '0 ч'
  if (hours < 1) return `${Math.round(hours * 60)} мин`

  return `${trim(hours)} ч`
}

export function plural(count: number, one: string, few: string, many: string): string {
  const mod10 = count % 10
  const mod100 = count % 100

  if (mod10 === 1 && mod100 !== 11) return one
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return few

  return many
}
