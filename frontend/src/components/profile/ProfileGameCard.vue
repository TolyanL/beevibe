<script setup lang="ts">
import { computed } from 'vue'
import type { CSSProperties } from 'vue'
import { RouterLink } from 'vue-router'
import type { ProfileGame } from '@/api/types'
import { formatDate, formatNumber, formatRelative, formatTokens, plural } from '@/lib/format'
import { genreLabel, shortGameTitle, statusMeta, visibleGameDescription } from '@/lib/game'
import AppIcon from '@/components/ui/AppIcon.vue'
import StatusPill from '@/components/ui/StatusPill.vue'

const props = defineProps<{ game: ProfileGame }>()

const palettes = [
  ['#ffcc00', '#7c3aed'],
  ['#22d3ee', '#2563eb'],
  ['#fb7185', '#f59e0b'],
  ['#84cc16', '#0d9488'],
  ['#f97316', '#db2777'],
  ['#a78bfa', '#22c55e'],
  ['#38bdf8', '#eab308'],
]

function hash(value: string): number {
  let result = 0
  for (let index = 0; index < value.length; index += 1) {
    result = (result * 31 + value.charCodeAt(index)) >>> 0
  }

  return result
}

const coverStyle = computed<CSSProperties>(() => {
  const palette = props.game.cover?.palette

  if (palette && palette.length >= 2) {
    return { '--thumb-a': palette[0], '--thumb-b': palette[palette.length - 1] } as CSSProperties
  }

  const [a, b] = palettes[hash(props.game.id) % palettes.length] ?? palettes[0]!

  return { '--thumb-a': a, '--thumb-b': b } as CSSProperties
})

const title = computed(() => shortGameTitle(props.game.title, props.game.prompt))
const description = computed(() =>
  visibleGameDescription(props.game.description, props.game.prompt, title.value),
)
const genre = computed(() => genreLabel(props.game.genre))
const meta = computed(() => statusMeta(props.game.status))
const isReady = computed(() => props.game.status === 'ready')
const isPublic = computed(() => props.game.visibility === 'public')
const fixesLabel = computed(
  () => `${props.game.fixes} ${plural(props.game.fixes, 'починка', 'починки', 'починок')}`,
)
</script>

<template>
  <article
    class="group flex min-w-0 flex-col overflow-hidden rounded-[1.2rem] border border-line bg-surface/75 transition duration-300 hover:-translate-y-0.5 hover:border-bee-deep/40 hover:shadow-lift"
  >
    <RouterLink
      :to="{ name: 'game', params: { id: game.slug } }"
      class="game-thumb relative block aspect-[16/9]"
      :style="coverStyle"
      :aria-label="`Открыть игру ${title}`"
    >
      <div class="dot-grid absolute inset-0 opacity-45" />

      <div class="absolute inset-x-3 top-3 flex items-start justify-between gap-2">
        <span
          class="rounded-full border border-white/15 bg-black/35 px-2.5 py-1 font-mono text-[9px] font-bold uppercase tracking-[0.12em] text-white/85 backdrop-blur"
        >
          {{ genre }}
        </span>
        <span
          v-if="game.speed === 'fast'"
          class="rounded-full border border-white/15 bg-black/35 px-2.5 py-1 font-mono text-[9px] font-bold uppercase tracking-[0.12em] text-bee backdrop-blur"
        >
          fast
        </span>
      </div>

      <div class="absolute inset-x-3 bottom-3 flex items-end justify-between gap-3">
        <div class="min-w-0">
          <p class="line-clamp-2 font-display text-[1.05rem] font-bold leading-tight text-white">
            {{ title }}
          </p>
          <p class="mt-1 font-mono text-[9px] uppercase tracking-[0.1em] text-white/60">
            {{ formatTokens(game.tokens_cost) }} токенов ·
            {{ formatNumber(game.model_calls) }} вызовов
          </p>
        </div>

        <span
          class="grid size-10 shrink-0 place-items-center rounded-full border border-white/20 bg-white/10 text-white opacity-0 backdrop-blur transition duration-300 group-hover:bg-bee group-hover:text-black group-hover:opacity-100"
        >
          <AppIcon :name="isReady ? 'play' : 'arrow-right'" :size="16" :stroke="2" />
        </span>
      </div>
    </RouterLink>

    <div class="flex flex-1 flex-col p-4">
      <div class="flex items-center justify-between gap-2">
        <StatusPill :status="game.status" compact />
        <span
          class="inline-flex items-center gap-1 font-mono text-[9px] uppercase tracking-[0.08em] text-faint"
          :title="meta.description"
        >
          <AppIcon :name="isPublic ? 'eye' : 'eye-off'" :size="12" />
          {{ isPublic ? 'в каталоге' : 'приватная' }}
        </span>
      </div>

      <p
        v-if="description && description !== title"
        class="mt-2 line-clamp-2 text-[11px] leading-5 text-muted"
        :title="description"
      >
        {{ description }}
      </p>
      <p v-else class="mt-2 line-clamp-2 text-[11px] leading-5 text-muted" :title="game.prompt">
        {{ meta.description }}
      </p>

      <div
        class="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1.5 border-t border-line pt-3 font-mono text-[9px] uppercase tracking-[0.06em] text-faint"
      >
        <span class="inline-flex items-center gap-1.5">
          <AppIcon name="clock" :size="11" />
          {{ formatDate(game.created_at) }}
        </span>
        <span class="inline-flex items-center gap-1.5">
          <AppIcon name="layers" :size="11" />
          {{ formatNumber(game.model_calls) }} вызовов
        </span>
        <span v-if="game.fixes" class="inline-flex items-center gap-1.5">
          <AppIcon name="refresh" :size="11" />
          {{ fixesLabel }}
        </span>
      </div>
    </div>

    <footer class="flex items-center justify-between gap-2 border-t border-line px-4 py-3">
      <span class="font-mono text-[9px] uppercase tracking-[0.08em] text-faint">
        {{ formatRelative(game.created_at) }}
      </span>

      <div class="flex items-center gap-2">
        <RouterLink
          v-if="isReady"
          :to="{ name: 'play', params: { id: game.slug } }"
          class="inline-flex items-center gap-1.5 rounded-full bg-bee px-3 py-1.5 text-[10px] font-extrabold text-graphite transition hover:bg-bee-deep"
        >
          <AppIcon name="play" :size="11" :stroke="2.2" />
          Играть
        </RouterLink>
        <RouterLink
          :to="{ name: 'game', params: { id: game.slug } }"
          class="inline-flex items-center gap-1 rounded-full border border-line px-3 py-1.5 text-[10px] font-bold text-muted transition hover:border-bee-deep hover:text-ink"
        >
          Открыть
          <AppIcon name="arrow-right" :size="11" />
        </RouterLink>
      </div>
    </footer>
  </article>
</template>
