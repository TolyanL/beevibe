<script setup lang="ts">
import { computed } from 'vue'
import { formatNumber } from '@/lib/format'
import AppIcon from './AppIcon.vue'

const props = withDefaults(
  defineProps<{
    page: number
    pageSize: number
    total: number
    loading?: boolean
    label?: string
  }>(),
  { loading: false, label: 'записей' },
)

const emit = defineEmits<{ 'update:page': [page: number] }>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const from = computed(() => (props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1))
const to = computed(() => Math.min(props.page * props.pageSize, props.total))
const canPrev = computed(() => props.page > 1 && !props.loading)
const canNext = computed(() => props.page < totalPages.value && !props.loading)

type PageItem = number | 'gap'

const pages = computed<PageItem[]>(() => {
  const last = totalPages.value

  if (last <= 7) return Array.from({ length: last }, (_, index) => index + 1)

  const values = new Set<number>([1, last, props.page - 1, props.page, props.page + 1])

  if (props.page <= 4) [2, 3, 4, 5].forEach((value) => values.add(value))
  if (props.page >= last - 3) {
    ;[last - 1, last - 2, last - 3, last - 4].forEach((value) => values.add(value))
  }

  const sorted = [...values].filter((value) => value >= 1 && value <= last).sort((a, b) => a - b)
  const result: PageItem[] = []

  sorted.forEach((value, index) => {
    const previous = sorted[index - 1]
    if (index > 0 && previous !== undefined && value - previous > 1) result.push('gap')
    result.push(value)
  })

  return result
})

function go(page: number): void {
  if (props.loading) return
  const next = Math.min(Math.max(1, page), totalPages.value)
  if (next !== props.page) emit('update:page', next)
}
</script>

<template>
  <nav
    v-if="total > 0"
    aria-label="Пагинация"
    class="flex flex-col gap-3 border-t border-line px-4 py-3.5 sm:flex-row sm:items-center sm:justify-between sm:px-5"
  >
    <p
      class="flex items-center gap-2 font-mono text-[9px] uppercase tracking-[0.1em] text-faint"
      aria-live="polite"
    >
      <span v-if="total">{{ from }}–{{ to }} из {{ formatNumber(total) }} {{ label }}</span>
      <span v-else>0 {{ label }}</span>
      <span v-if="loading" class="inline-flex items-center gap-1.5 text-muted">
        <span class="size-2.5 animate-spin rounded-full border border-line border-t-bee-deep" />
        обновляем
      </span>
    </p>

    <div v-if="totalPages > 1" class="flex items-center gap-1">
      <button
        type="button"
        class="grid size-8 place-items-center rounded-lg border border-line text-muted transition hover:border-line-strong hover:bg-hover-surface hover:text-ink disabled:cursor-not-allowed disabled:opacity-35"
        aria-label="Предыдущая страница"
        :disabled="!canPrev"
        @click="go(page - 1)"
      >
        <AppIcon name="arrow-left" :size="14" />
      </button>

      <template v-for="(item, index) in pages" :key="`${item}-${index}`">
        <span
          v-if="item === 'gap'"
          class="grid size-8 place-items-center font-mono text-[10px] text-faint"
        >
          …
        </span>
        <button
          v-else
          type="button"
          class="grid size-8 place-items-center rounded-lg border font-mono text-[11px] font-bold transition"
          :class="
            item === page
              ? 'border-transparent bg-bee text-graphite'
              : 'border-line text-muted hover:border-line-strong hover:bg-hover-surface hover:text-ink'
          "
          :aria-label="`Страница ${item}`"
          :aria-current="item === page ? 'page' : undefined"
          :disabled="loading"
          @click="go(item)"
        >
          {{ item }}
        </button>
      </template>

      <button
        type="button"
        class="grid size-8 place-items-center rounded-lg border border-line text-muted transition hover:border-line-strong hover:bg-hover-surface hover:text-ink disabled:cursor-not-allowed disabled:opacity-35"
        aria-label="Следующая страница"
        :disabled="!canNext"
        @click="go(page + 1)"
      >
        <AppIcon name="arrow-right" :size="14" />
      </button>
    </div>
  </nav>
</template>
