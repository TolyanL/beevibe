<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import type { CSSProperties } from 'vue'
import type { CatalogCard } from '@/api/types'
import { formatHours, formatRelative } from '@/lib/format'
import { gameDescription, genreLabel, mapSizeLabel, shortGameTitle, styleLabel } from '@/lib/game'
import AppIcon from '@/components/ui/AppIcon.vue'
import RatingBadge from '@/components/ui/RatingBadge.vue'

const props = defineProps<{ card: CatalogCard }>()

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

const style = computed<CSSProperties>(() => {
  const cover = props.card.cover?.palette

  if (cover && cover.length >= 2) {
    return { '--thumb-a': cover[0], '--thumb-b': cover[cover.length - 1] } as CSSProperties
  }

  const [a, b] = palettes[hash(props.card.id) % palettes.length] ?? palettes[0]!

  return { '--thumb-a': a, '--thumb-b': b } as CSSProperties
})

const title = computed(() => shortGameTitle(props.card.title, props.card.prompt))
const description = computed(() => gameDescription(props.card.description, props.card.prompt))
const genre = computed(() => genreLabel(props.card.genre, props.card.custom_genre))
const styleName = computed(() => styleLabel(props.card.style, props.card.custom_style))
const author = computed(() => props.card.author ?? '')
const mapSize = computed(() => mapSizeLabel(props.card.map_size))
const playtime = computed(() => formatHours(props.card.playtime?.hours ?? 0))
</script>

<template>
  <RouterLink
    :to="{ name: 'play', params: { id: card.slug } }"
    class="group flex flex-col overflow-hidden rounded-[1.35rem] border border-line bg-surface/75 transition duration-300 hover:-translate-y-1 hover:border-bee-deep/50 hover:shadow-lift"
  >
    <div class="game-thumb relative aspect-[16/10]" :style="style">
      <div class="dot-grid absolute inset-0 opacity-50" />

      <div class="absolute inset-0 flex flex-col justify-between p-4">
        <div class="flex items-center justify-between gap-2">
          <span
            v-if="genre"
            class="rounded-full border border-white/15 bg-black/35 px-2.5 py-1 font-mono text-[9px] font-bold uppercase tracking-[0.14em] text-white/80 backdrop-blur"
          >
            {{ genre }}
          </span>
          <span v-else class="size-0" />
          <RatingBadge
            :average="card.rating?.average ?? 0"
            :total="card.rating?.total ?? 0"
            size="sm"
          />
        </div>

        <div class="flex items-end justify-between gap-3">
          <div class="min-w-0">
            <p
              v-if="styleName"
              class="font-display text-[1.05rem] font-bold leading-tight text-white"
            >
              {{ styleName }}
            </p>
            <p
              class="mt-1 flex items-center gap-2 font-mono text-[9px] uppercase tracking-[0.12em] text-white/55"
            >
              <span v-if="mapSize">{{ mapSize }}</span>
              <span v-if="mapSize" class="size-0.5 rounded-full bg-white/40" />
              <span>{{ playtime }}</span>
            </p>
          </div>

          <span
            class="grid size-10 shrink-0 place-items-center rounded-full border border-white/20 bg-white/10 text-white opacity-0 backdrop-blur transition duration-300 group-hover:bg-bee group-hover:text-black group-hover:opacity-100"
          >
            <AppIcon name="play" :size="16" :stroke="2" />
          </span>
        </div>
      </div>
    </div>

    <div class="flex flex-1 flex-col p-4">
      <h3
        class="line-clamp-1 text-[15px] font-bold leading-snug tracking-[-0.015em]"
        :title="title"
      >
        {{ title }}
      </h3>
      <p
        v-if="description && description !== title"
        class="mt-1.5 line-clamp-2 text-[11px] leading-5 text-muted"
        :title="description"
      >
        {{ description }}
      </p>
      <div class="mt-auto flex items-center justify-between gap-3 pt-4">
        <span class="text-[11px] text-faint">
          <span v-if="author" class="text-muted">{{ author }} · </span
          >{{ formatRelative(card.created_at) }}
        </span>
        <span
          class="inline-flex items-center gap-1 font-mono text-[10px] font-bold uppercase tracking-[0.09em] text-muted"
        >
          Играть
          <AppIcon name="arrow-right" :size="13" />
        </span>
      </div>
    </div>
  </RouterLink>
</template>
