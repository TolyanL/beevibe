<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import type { ProfileGeneration } from '@/api/types'
import { formatDateTime, formatNumber, formatTokens } from '@/lib/format'
import { shortGameTitle, statusMeta, versionSourceLabel } from '@/lib/game'
import { generationSourceMeta } from '@/lib/profile'
import AppIcon from '@/components/ui/AppIcon.vue'

const props = defineProps<{ generation: ProfileGeneration }>()

const source = computed(() => generationSourceMeta(props.generation.source))
const title = computed(() =>
  shortGameTitle(props.generation.game_title, props.generation.game_prompt),
)
const gameStatus = computed(() => statusMeta(props.generation.game_status))
</script>

<template>
  <article class="hover-row group flex items-start gap-3 px-4 py-4 sm:px-5">
    <span
      class="grid size-10 shrink-0 place-items-center rounded-xl border font-mono text-[11px] font-bold tabular-nums"
      :class="source.tone"
      :title="`Версия #${generation.version_number}`"
    >
      #{{ generation.version_number }}
    </span>

    <div class="min-w-0 flex-1">
      <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
        <span
          class="inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 font-mono text-[8px] font-bold uppercase tracking-[0.08em]"
          :class="source.tone"
        >
          <AppIcon :name="source.icon" :size="10" :stroke="2.2" />
          {{ versionSourceLabel(generation.source) }}
        </span>
        <span class="font-mono text-[9px] uppercase tracking-[0.06em] text-faint">
          {{ formatDateTime(generation.created_at) }}
        </span>
      </div>

      <RouterLink
        :to="{ name: 'game', params: { id: generation.game_slug } }"
        class="mt-1.5 block truncate text-[13px] font-bold transition hover:text-hover-accent"
      >
        {{ title }}
      </RouterLink>

      <p
        v-if="generation.note"
        class="mt-0.5 line-clamp-1 text-[11px] leading-5 text-muted"
        :title="generation.note"
      >
        {{ generation.note }}
      </p>

      <div
        class="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 font-mono text-[9px] uppercase tracking-[0.06em] text-faint"
      >
        <span class="inline-flex items-center gap-1">
          <AppIcon name="bolt" :size="10" />
          {{ formatTokens(generation.tokens_cost) }} токенов
        </span>
        <span class="inline-flex items-center gap-1">
          <AppIcon name="layers" :size="10" />
          {{ formatNumber(generation.model_calls) }} вызовов
        </span>
        <span class="inline-flex items-center gap-1">
          <AppIcon name="gamepad" :size="10" />
          {{ gameStatus.short }}
        </span>
      </div>
    </div>

    <AppIcon
      name="arrow-right"
      :size="14"
      class="mt-1 shrink-0 text-faint transition group-hover:translate-x-0.5 group-hover:text-ink"
    />
  </article>
</template>
