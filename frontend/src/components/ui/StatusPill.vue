<script setup lang="ts">
import { computed } from 'vue'
import { statusMeta } from '@/lib/game'

const props = withDefaults(
  defineProps<{
    status: string
    pulse?: boolean
    compact?: boolean
  }>(),
  { pulse: true, compact: false },
)

const meta = computed(() => statusMeta(props.status))

const toneClass = computed(() => {
  if (meta.value.tone === 'good') return 'border-emerald-500/25 bg-emerald-500/10 text-emerald-500'
  if (meta.value.tone === 'bad') return 'border-red-500/25 bg-red-500/10 text-red-500'
  if (meta.value.tone === 'live') return 'border-bee/40 bg-bee/10 text-bee-deep dark:text-bee'
  if (meta.value.tone === 'warn') return 'border-orange-500/25 bg-orange-500/10 text-orange-500'

  return 'border-line-strong bg-surface text-muted'
})
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 font-mono text-[10px] font-bold uppercase tracking-[0.08em]"
    :class="toneClass"
  >
    <span
      class="size-1.5 rounded-full"
      :class="[
        meta.tone === 'good' ? 'bg-emerald-500' : meta.tone === 'bad' ? 'bg-red-500' : meta.tone === 'live' ? 'bg-bee' : meta.tone === 'warn' ? 'bg-orange-500' : 'bg-faint',
        pulse && meta.tone === 'live' ? 'animate-pulse-soft' : '',
      ]"
    />
    {{ compact ? meta.short : meta.label }}
  </span>
</template>
