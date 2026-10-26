import { request } from './client'
import type { Profile, TopupResponse, UserEnvelope } from './types'

export interface Credentials {
  email: string
  password: string
}

export function login(credentials: Credentials): Promise<UserEnvelope> {
  return request<UserEnvelope>('/auth/login', {
    method: 'POST',
    body: credentials,
  })
}

export function register(credentials: Credentials): Promise<UserEnvelope> {
  return request<UserEnvelope>('/auth/register', {
    method: 'POST',
    body: credentials,
  })
}

export function logout(): Promise<void> {
  return request<void>('/auth/logout', { method: 'POST' })
}

export function getMe(): Promise<UserEnvelope> {
  return request<UserEnvelope>('/me')
}

export function getProfile(): Promise<Profile> {
  return request<Profile>('/profile')
}

export function topup(): Promise<TopupResponse> {
  return request<TopupResponse>('/profile/topup', { method: 'POST' })
}
