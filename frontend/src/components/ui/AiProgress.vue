<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import AppIcon from '@/components/ui/AppIcon.vue'

const props = withDefaults(
  defineProps<{
    startedAt?: string
    label?: string
    hint?: string
    expectedSeconds?: number
    compact?: boolean
  }>(),
  { label: 'Нейросеть работает', hint: '', expectedSeconds: 120, compact: false },
)

const now = ref(Date.now())
const timer = setInterval(() => {
  now.value = Date.now()
}, 1000)

onBeforeUnmount(() => clearInterval(timer))

const startedMs = computed(() => {
  if (!props.startedAt) return now.value - 1000

  const parsed = Date.parse(props.startedAt)

  return Number.isFinite(parsed) ? parsed : now.value - 1000
})

const elapsedSeconds = computed(() => Math.max(0, Math.round((now.value - startedMs.value) / 1000)))

const elapsedLabel = computed(() => {
  const minutes = Math.floor(elapsedSeconds.value / 60)
  const seconds = elapsedSeconds.value % 60

  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
})

// NOTE: an LLM call has no real progress, so the bar eases to 95% and waits there.
const percent = computed(() => {
  const expected = Math.max(10, props.expectedSeconds)
  const ratio = elapsedSeconds.value / expected

  return Math.min(95, Math.round((1 - Math.exp(-1.6 * ratio)) * 95))
})

const stage = computed(() => {
  if (elapsedSeconds.value < 5) return 'Отправляем запрос модели'
  if (elapsedSeconds.value < 25) return 'Модель обдумывает ответ'
  if (elapsedSeconds.value < 70) return 'Собираем результат'

  return 'Осталось чуть-чуть — ответ вот-вот придёт'
})
</script>

<template>
  <div
    class="rounded-2xl border border-bee-deep/25 bg-bee/[0.06] px-4 py-3"
    :class="compact ? '' : 'sm:px-5 sm:py-4'"
    role="status"
    aria-live="polite"
  >
    <div class="flex flex-wrap items-center justify-between gap-2">
      <p class="inline-flex items-center gap-2 text-xs font-bold text-ink">
        <AppIcon name="spark" :size="14" class="text-bee-deep dark:text-bee" />
        {{ label }}
      </p>
      <span class="font-mono text-[10px] tabular-nums text-muted"
        >{{ elapsedLabel }} · {{ percent }}%</span
      >
    </div>

    <div class="mt-2.5 h-1.5 overflow-hidden rounded-full bg-line">
      <div
        class="ai-progress-fill h-full rounded-full bg-bee transition-[width] duration-700 ease-out"
        :style="{ width: `${percent}%` }"
      />
    </div>

    <p v-if="!compact" class="mt-2 text-[10px] leading-4 text-muted">
      {{ stage }}<span v-if="hint"> · {{ hint }}</span>
    </p>
  </div>
</template>

<style scoped>
.ai-progress-fill {
  background-image: linear-gradient(
    100deg,
    transparent 0%,
    rgba(255, 255, 255, 0.45) 20%,
    transparent 40%
  );
  background-size: 220% 100%;
  animation: ai-progress-shimmer 1.6s linear infinite;
}

@keyframes ai-progress-shimmer {
  from {
    background-position: 120% 0;
  }
  to {
    background-position: -120% 0;
  }
}
</style>
