<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Button from 'openvue/button'
import InputText from 'openvue/inputtext'
import { useToast } from 'openvue/usetoast'
import { messageFromError } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import AppIcon from '@/components/ui/AppIcon.vue'
import BrandMark from '@/components/layout/BrandMark.vue'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const auth = useAuthStore()

const mode = ref<'login' | 'register'>(route.name === 'register' ? 'register' : 'login')
const email = ref('')
const password = ref('')
const confirm = ref('')
const showPassword = ref(false)
const formError = ref('')
const submitting = ref(false)

const isRegister = computed(() => mode.value === 'register')
const redirect = computed(() => {
  const value = typeof route.query.redirect === 'string' ? route.query.redirect : '/'

  return value.startsWith('/') && !value.startsWith('//') ? value : '/'
})
const title = computed(() => (isRegister.value ? 'Создайте аккаунт' : 'С возвращением'))
const subtitle = computed(() =>
  isRegister.value
    ? 'Один аккаунт — генератор, каталог, оценки и история игр.'
    : 'Войдите, чтобы продолжить сборку игр и управлять балансом.',
)

watch(
  () => route.name,
  (name) => {
    mode.value = name === 'register' ? 'register' : 'login'
    formError.value = ''
    auth.clearError()
  },
)

onMounted(() => auth.clearError())

function switchMode(next: 'login' | 'register'): void {
  if (mode.value === next) return
  formError.value = ''
  auth.clearError()
  void router.replace({ name: next, query: route.query })
}

function validate(): string {
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value.trim())) {
    return 'Введите корректный email.'
  }
  if (password.value.length < 8) {
    return 'Пароль должен быть не короче 8 символов.'
  }
  if (isRegister.value && password.value !== confirm.value) {
    return 'Пароли не совпадают.'
  }

  return ''
}

async function submit(): Promise<void> {
  formError.value = validate()
  if (formError.value) return

  submitting.value = true

  try {
    if (isRegister.value) {
      await auth.register(email.value.trim(), password.value)
      toast.add({
        severity: 'success',
        summary: 'Аккаунт создан',
        detail: 'Стартовые токены уже на балансе.',
        life: 3200,
      })
    } else {
      await auth.login(email.value.trim(), password.value)
      toast.add({ severity: 'success', summary: 'Вы вошли', life: 2000 })
    }

    await router.push(redirect.value || '/')
  } catch (err) {
    formError.value = messageFromError(err)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="pb-10 pt-8 sm:pt-12">
    <div class="app-container">
      <div class="grid overflow-hidden rounded-[1.6rem] border border-line bg-surface/70 shadow-panel lg:grid-cols-[.92fr_1.08fr]">
        <aside class="relative hidden overflow-hidden bg-[#10100d] p-8 text-white lg:flex lg:flex-col lg:justify-between">
          <div class="dot-grid absolute inset-0 opacity-20" />
          <div class="absolute -left-20 top-1/3 size-72 rounded-full bg-bee/10 blur-3xl" />

          <div class="relative">
            <BrandMark />
          </div>

          <div class="relative my-12">
            <p class="font-mono text-[9px] font-bold uppercase tracking-[0.18em] text-bee">game generation, simplified</p>
            <h2 class="mt-5 font-display text-4xl font-bold leading-[1.06] tracking-[-0.05em]">
              Ваша следующая игра начинается с одного предложения.
            </h2>
            <p class="mt-5 max-w-sm text-sm leading-6 text-white/45">
              BeeVibe собирает механику, код и визуал. Вам остаётся играть, оценивать и просить
              правки — без IDE и деплоя.
            </p>
          </div>

          <div class="relative grid grid-cols-3 gap-3">
            <div v-for="item in [
              { value: '9', label: 'этапов' },
              { value: '1', label: 'HTML-файл' },
              { value: '5', label: 'баллов' },
            ]" :key="item.label" class="rounded-2xl border border-white/10 bg-white/5 p-3">
              <p class="font-display text-xl font-bold text-bee">{{ item.value }}</p>
              <p class="mt-1 font-mono text-[8px] uppercase tracking-[0.12em] text-white/35">{{ item.label }}</p>
            </div>
          </div>

          <div class="relative mt-8 flex items-center gap-2 text-[10px] text-white/30">
            <AppIcon name="shield" :size="14" />
            HttpOnly-сессия, без токенов в localStorage
          </div>
        </aside>

        <div class="relative p-5 sm:p-8 lg:p-10">
          <div class="mx-auto max-w-md">
            <div class="flex items-center justify-between">
              <RouterLink to="/" class="inline-flex items-center gap-1.5 text-[11px] font-bold text-muted transition hover:text-ink">
                <AppIcon name="back" :size="14" />
                На главную
              </RouterLink>
              <span class="font-mono text-[9px] uppercase tracking-[0.14em] text-faint">
                {{ isRegister ? 'new account' : 'sign in' }}
              </span>
            </div>

            <div class="mt-8">
              <h1 class="display-title text-4xl">{{ title }}</h1>
              <p class="mt-3 text-sm leading-6 text-muted">{{ subtitle }}</p>
            </div>

            <div class="mt-6 grid grid-cols-2 gap-1 rounded-full border border-line bg-canvas/60 p-1">
              <button
                type="button"
                class="rounded-full py-2 text-xs font-bold transition"
                :class="mode === 'login' ? 'bg-ink text-canvas shadow-sm' : 'text-muted hover:text-ink'"
                @click="switchMode('login')"
              >
                Вход
              </button>
              <button
                type="button"
                class="rounded-full py-2 text-xs font-bold transition"
                :class="mode === 'register' ? 'bg-ink text-canvas shadow-sm' : 'text-muted hover:text-ink'"
                @click="switchMode('register')"
              >
                Регистрация
              </button>
            </div>

            <form class="mt-6 space-y-4" @submit.prevent="submit">
              <div>
                <label for="email" class="mono-label mb-2 block">Email</label>
                <div class="relative">
                  <AppIcon name="user" :size="16" class="pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 text-faint" />
                  <InputText
                    id="email"
                    v-model="email"
                    type="email"
                    class="h-12 w-full pl-10"
                    placeholder="you@example.com"
                    autocomplete="email"
                    required
                    @update:model-value="formError = ''"
                  />
                </div>
              </div>

              <div>
                <div class="mb-2 flex items-center justify-between">
                  <label for="password" class="mono-label">Пароль</label>
                  <span v-if="isRegister" class="text-[10px] text-faint">минимум 8 символов</span>
                </div>
                <div class="relative">
                  <AppIcon name="shield" :size="16" class="pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 text-faint" />
                  <InputText
                    id="password"
                    v-model="password"
                    :type="showPassword ? 'text' : 'password'"
                    class="h-12 w-full px-10"
                    placeholder="••••••••"
                    :autocomplete="isRegister ? 'new-password' : 'current-password'"
                    required
                    @update:model-value="formError = ''"
                  />
                  <button
                    type="button"
                    class="absolute right-2.5 top-1/2 grid size-8 -translate-y-1/2 place-items-center rounded-full text-faint transition hover:bg-hover-surface hover:text-ink"
                    :aria-label="showPassword ? 'Скрыть пароль' : 'Показать пароль'"
                    @click="showPassword = !showPassword"
                  >
                    <AppIcon :name="showPassword ? 'eye-off' : 'eye'" :size="15" />
                  </button>
                </div>
              </div>

              <div v-if="isRegister">
                <label for="confirm" class="mono-label mb-2 block">Повторите пароль</label>
                <div class="relative">
                  <AppIcon name="check" :size="16" class="pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 text-faint" />
                  <InputText
                    id="confirm"
                    v-model="confirm"
                    :type="showPassword ? 'text' : 'password'"
                    class="h-12 w-full pl-10"
                    placeholder="••••••••"
                    autocomplete="new-password"
                    required
                    @update:model-value="formError = ''"
                  />
                </div>
              </div>

              <div
                v-if="formError || auth.error"
                class="flex items-start gap-2.5 rounded-2xl border border-red-500/20 bg-red-500/10 px-4 py-3 text-xs leading-5 text-red-500"
              >
                <AppIcon name="alert" :size="16" class="mt-0.5" />
                {{ formError || auth.error }}
              </div>

              <Button type="submit" class="bee-button mt-2 w-full" size="large" :loading="submitting" :disabled="submitting">
                {{ isRegister ? 'Создать аккаунт' : 'Войти' }}
                <AppIcon name="arrow-right" :size="16" class="ml-2" />
              </Button>
            </form>

            <p class="mt-5 text-center text-[11px] leading-5 text-faint">
              Продолжая, вы соглашаетесь с правилами сервиса и обработкой cookie для сессии.
            </p>

            <div class="mt-7 flex items-center gap-3">
              <span class="h-px flex-1 bg-line" />
              <span class="font-mono text-[8px] uppercase tracking-[0.14em] text-faint">BeeVibe</span>
              <span class="h-px flex-1 bg-line" />
            </div>

            <p class="mt-4 text-center text-[11px] text-muted">
              {{ isRegister ? 'Уже есть аккаунт?' : 'Впервые здесь?' }}
              <button
                type="button"
                class="font-bold text-ink underline decoration-bee decoration-2 underline-offset-2"
                @click="switchMode(isRegister ? 'login' : 'register')"
              >
                {{ isRegister ? 'Войти' : 'Зарегистрироваться' }}
              </button>
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
