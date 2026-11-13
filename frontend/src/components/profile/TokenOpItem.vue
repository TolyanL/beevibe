<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import type { ProfileTokenOp } from '@/api/types'
import { formatDateTime, formatTokens } from '@/lib/format'
import { tokenOpMeta } from '@/lib/profile'
import AppIcon from '@/components/ui/AppIcon.vue'

const props = defineProps<{ op: ProfileTokenOp }>()

const meta = computed(() => tokenOpMeta(props.op.reason))
const positive = computed(() => props.op.delta >= 0)
const amount = computed(() => `${positive.value ? '+' : ''}${formatTokens(props.op.delta)}`)
const gameLabel = computed(() => (props.op.game_label ?? '').trim())
</script>

<template>
  <article class="hover-row flex items-start gap-3 px-4 py-3.5 sm:px-5">
    <span
      class="mt-0.5 grid size-8 shrink-0 place-items-center rounded-full border"
      :class="
        positive
          ? 'border-emerald-500/25 bg-emerald-500/10 text-emerald-500'
          : 'border-line-strong bg-canvas/50 text-muted'
      "
    >
      <AppIcon :name="meta.icon" :size="13" :stroke="2" />
    </span>

    <div class="min-w-0 flex-1">
      <p class="text-[12px] font-bold leading-5">{{ meta.label }}</p>

      <RouterLink
        v-if="gameLabel && op.game_slug"
        :to="{ name: 'game', params: { id: op.game_slug } }"
        class="mt-0.5 block truncate text-[10px] text-muted transition hover:text-hover-accent"
        :title="gameLabel"
      >
        {{ gameLabel }}
      </RouterLink>
      <p v-else-if="gameLabel" class="mt-0.5 truncate text-[10px] text-muted">{{ gameLabel }}</p>

      <p class="mt-0.5 font-mono text-[9px] uppercase tracking-[0.06em] text-faint">
        {{ formatDateTime(op.created_at) }}
      </p>
    </div>

    <span
      class="mt-0.5 shrink-0 font-mono text-[11px] font-bold tabular-nums"
      :class="positive ? 'text-emerald-500' : 'text-ink'"
    >
      {{ amount }}
    </span>
  </article>
</template>
