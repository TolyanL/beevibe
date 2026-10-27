import type { ApiErrorEnvelope } from './types'

const API_BASE =
  (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') ?? '/api'

export function apiUrl(path: string): string {
  return `${API_BASE}${path}`
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

type RequestOptions = Omit<RequestInit, 'body'> & {
  body?: unknown
  rawBody?: BodyInit
}

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { body, rawBody, headers, ...rest } = options
  const requestHeaders = new Headers(headers)

  requestHeaders.set('Accept', 'application/json')

  if (body !== undefined) {
    requestHeaders.set('Content-Type', 'application/json')
  }

  let response: Response

  try {
    response = await fetch(`${API_BASE}${path}`, {
      ...rest,
      credentials: 'include',
      headers: requestHeaders,
      body: body !== undefined ? JSON.stringify(body) : rawBody,
    })
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw error
    }

    throw new ApiError(0, 'network_error', 'Сервер недоступен. Проверьте соединение.')
  }

  if (response.status === 204) {
    return undefined as T
  }

  const contentType = response.headers.get('content-type') ?? ''
  const payload = contentType.includes('application/json')
    ? ((await response.json().catch(() => null)) as ApiErrorEnvelope | T | null)
    : null

  if (!response.ok) {
    const envelope = payload as ApiErrorEnvelope | null
    const code = envelope?.error?.code ?? `http_${response.status}`
    const message = envelope?.error?.message ?? defaultMessage(response.status)

    throw new ApiError(response.status, code, message)
  }

  return payload as T
}

function defaultMessage(status: number): string {
  if (status === 401) return 'Нужно войти в аккаунт.'
  if (status === 403) return 'Недостаточно прав.'
  if (status === 404) return 'Ничего не найдено.'
  if (status === 409) return 'Действие сейчас недоступно.'
  if (status === 429) return 'Слишком много запросов. Попробуйте позже.'
  if (status >= 500) return 'Сервер не смог выполнить запрос.'

  return 'Не получилось выполнить запрос.'
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}

export function messageFromError(error: unknown): string {
  if (isApiError(error)) return error.message
  if (error instanceof Error) return error.message

  return 'Что-то пошло не так.'
}
