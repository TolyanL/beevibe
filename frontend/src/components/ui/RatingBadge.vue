<script setup lang="ts">
import { computed } from 'vue'
import { ratingTone } from '@/lib/rating'
import { plural } from '@/lib/format'
import AppIcon from './AppIcon.vue'

const props = withDefaults(
  defineProps<{
    average: number
    total?: number
    size?: 'sm' | 'md' | 'lg'
  }>(),
  { total: 0, size: 'md' },
)

const tone = computed(() => ratingTone(props.average))
const score = computed(() => (props.total > 0 ? props.average.toFixed(1).replace('.', ',') : '—'))
const sizeClass = computed(() => {
  if (props.size === 'lg') return 'px-3 py-1.5 text-sm'
  if (props.size === 'sm') return 'px-2 py-0.5 text-[11px]'

  return 'px-2.5 py-1 text-xs'
})
const votesLabel = computed(
  () => `${props.total} ${plural(props.total, 'оценка', 'оценки', 'оценок')}`,
)
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-full border font-mono font-bold tabular-nums"
    :class="[tone.background, tone.border, tone.text, sizeClass]"
    :title="total ? `${votesLabel}, средняя ${score} из 5` : 'Пока нет оценок'"
  >
    <AppIcon name="star" :size="size === 'lg' ? 15 : size === 'sm' ? 11 : 13" :stroke="2.2" />
    <span>{{ score }}</span>
    <span v-if="total" class="opacity-60">/ 5</span>
  </span>
</template>
