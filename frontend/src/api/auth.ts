import { request } from './client'
import type { Profile, TopupResponse, UserEnvelope, UserInfo } from './types'

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

export function updateProfile(displayName: string): Promise<{ user: UserInfo }> {
  return request<{ user: UserInfo }>('/profile', {
    method: 'PATCH',
    body: { display_name: displayName },
  })
}

export function topup(amount: number): Promise<TopupResponse> {
  return request<TopupResponse>('/profile/topup', {
    method: 'POST',
    body: { amount },
    headers: { 'Idempotency-Key': crypto.randomUUID() },
  })
}
