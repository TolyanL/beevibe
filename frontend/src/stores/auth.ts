import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import {
  getMe,
  getProfile,
  login as loginRequest,
  logout as logoutRequest,
  register as registerRequest,
  topup as topupRequest,
  updateProfile as updateProfileRequest,
} from '@/api/auth'
import { isApiError, messageFromError } from '@/api/client'
import type { Profile, User } from '@/api/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const profile = ref<Profile | null>(null)
  const loading = ref(false)
  const initialized = ref(false)
  const error = ref('')

  const isAuthenticated = computed(() => user.value !== null)
  const balance = computed(() => user.value?.balance ?? 0)
  const prices = computed(
    () => profile.value?.prices ?? { generate: 320_000, patch: 40_000, meta: 10_000 },
  )

  function setUser(next: User | null): void {
    user.value = next
    if (next && profile.value) profile.value.balance = next.balance
  }

  function syncProfile(next: Profile): void {
    profile.value = next
    if (user.value) user.value.balance = next.balance
  }

  async function bootstrap(): Promise<void> {
    if (initialized.value) return

    try {
      const response = await getMe()
      setUser(response.user)
      initialized.value = true
    } catch (err) {
      if (isApiError(err) && err.status === 401) {
        setUser(null)
        initialized.value = true
        return
      }

      error.value = messageFromError(err)
      setUser(null)
    }
  }

  async function login(email: string, password: string): Promise<void> {
    loading.value = true
    error.value = ''

    try {
      const response = await loginRequest({ email, password })
      setUser(response.user)
      initialized.value = true
      await refreshProfile()
    } catch (err) {
      error.value = messageFromError(err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function register(email: string, password: string): Promise<void> {
    loading.value = true
    error.value = ''

    try {
      const response = await registerRequest({ email, password })
      setUser(response.user)
      initialized.value = true
      await refreshProfile()
    } catch (err) {
      error.value = messageFromError(err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function logout(): Promise<void> {
    loading.value = true
    try {
      await logoutRequest()
    } catch (err) {
      if (!isApiError(err) || err.status !== 401) throw err
    } finally {
      setUser(null)
      profile.value = null
      loading.value = false
    }
  }

  async function refreshMe(): Promise<void> {
    try {
      const response = await getMe()
      setUser(response.user)
    } catch (err) {
      if (isApiError(err) && err.status === 401) {
        setUser(null)
        return
      }
      error.value = messageFromError(err)
    }
  }

  async function refreshProfile(): Promise<Profile | null> {
    if (!user.value) return null

    try {
      const next = await getProfile()
      syncProfile(next)
      return next
    } catch (err) {
      if (isApiError(err) && err.status === 401) {
        setUser(null)
        return null
      }
      error.value = messageFromError(err)
      return null
    }
  }

  async function updateDisplayName(name: string): Promise<void> {
    const response = await updateProfileRequest(name)

    if (user.value) user.value.display_name = response.user.display_name
  }

  async function topup(amount: number): Promise<void> {
    loading.value = true
    try {
      const response = await topupRequest(amount)
      if (user.value) user.value.balance = response.balance
      if (profile.value) profile.value.balance = response.balance
      await refreshProfile()
    } finally {
      loading.value = false
    }
  }

  function clearError(): void {
    error.value = ''
  }

  return {
    user,
    profile,
    loading,
    initialized,
    error,
    isAuthenticated,
    balance,
    prices,
    bootstrap,
    login,
    register,
    logout,
    refreshMe,
    refreshProfile,
    topup,
    updateDisplayName,
    clearError,
  }
})
