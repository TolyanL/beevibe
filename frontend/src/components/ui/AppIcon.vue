<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    name: IconName
    size?: number
    stroke?: number
  }>(),
  { size: 20, stroke: 1.8 },
)

type IconName =
  | 'arrow-right'
  | 'arrow-left'
  | 'play'
  | 'expand'
  | 'compress'
  | 'share'
  | 'download'
  | 'spark'
  | 'gamepad'
  | 'grid'
  | 'user'
  | 'logout'
  | 'plus'
  | 'check'
  | 'clock'
  | 'star'
  | 'flame'
  | 'search'
  | 'filter'
  | 'chevron-down'
  | 'alert'
  | 'refresh'
  | 'code'
  | 'layers'
  | 'wand'
  | 'external'
  | 'x'
  | 'menu'
  | 'copy'
  | 'back'
  | 'coin'
  | 'bolt'
  | 'shield'
  | 'sun'
  | 'moon'
  | 'eye'
  | 'eye-off'
  | 'settings'
  | 'chart'
  | 'book'
  | 'trophy'
  | 'heart'

const paths: Record<IconName, string[]> = {
  'arrow-right': ['M4 12h15', 'm13 6 6 6-6 6'],
  'arrow-left': ['M20 12H5', 'm11 18-6-6 6-6'],
  play: ['M8 5.2v13.6a.7.7 0 0 0 1.07.6l11.1-6.8a.7.7 0 0 0 0-1.2L9.07 4.6A.7.7 0 0 0 8 5.2Z'],
  expand: [
    'M8 3H5a2 2 0 0 0-2 2v3',
    'M16 3h3a2 2 0 0 1 2 2v3',
    'M8 21H5a2 2 0 0 1-2-2v-3',
    'M16 21h3a2 2 0 0 0 2-2v-3',
  ],
  compress: [
    'M3 8h3a2 2 0 0 0 2-2V3',
    'M21 8h-3a2 2 0 0 1-2-2V3',
    'M3 16h3a2 2 0 0 1 2 2v3',
    'M21 16h-3a2 2 0 0 0-2 2v3',
  ],
  share: ['M12 3v12', 'm8 7 4-4 4 4', 'M5 14v5a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-5'],
  download: ['M12 3v12', 'm8 11 4 4 4-4', 'M5 21h14'],
  spark: ['m12 3 2.2 5.8L20 11l-5.8 2.2L12 19l-2.2-5.8L4 11l5.8-2.2L12 3Z'],
  gamepad: [
    'M7 8h10a5 5 0 0 1 0 10H7a5 5 0 0 1 0-10Z',
    'M8 11v4',
    'M6 13h4',
    'M16 11h.01',
    'M18 13h.01',
    'M16 15h.01',
  ],
  grid: ['M4 4h6v6H4z', 'M14 4h6v6h-6z', 'M4 14h6v6H4z', 'M14 14h6v6h-6z'],
  user: ['M8 7a4 4 0 1 0 8 0 4 4 0 0 0-8 0Z', 'M6 21v-2a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v2'],
  logout: ['M10 4H6a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h4', 'M15 8l4 4-4 4', 'M19 12H9'],
  plus: ['M12 5v14', 'M5 12h14'],
  check: ['m5 12 5 5L20 7'],
  clock: ['M12 7v5l3 2', 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Z'],
  star: ['m12 3 2.8 5.7 6.2.9-4.5 4.4 1.1 6.2-5.6-2.9-5.6 2.9 1.1-6.2L3 9.6l6.2-.9L12 3Z'],
  flame: ['M12 3c1 4-4 5-4 9a4 4 0 0 0 8 0c0-2-1-3-1-3s4 2 4 6a7 7 0 1 1-14 0c0-5 5-7 7-12Z'],
  search: ['M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16Z', 'm21 21-4.35-4.35'],
  filter: ['M4 6h16', 'M7 12h10', 'M10 18h4'],
  'chevron-down': ['m6 9 6 6 6-6'],
  alert: [
    'M12 8v5',
    'M12 17h.01',
    'M10.3 3.9 2.6 18a2 2 0 0 0 1.7 3h15.4a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0Z',
  ],
  refresh: [
    'M20 7v5h-5',
    'M4 17v-5h5',
    'M6.1 9a7 7 0 0 1 11-2.7L20 9',
    'M18 15a7 7 0 0 1-11 2.7L4 15',
  ],
  code: ['m8 8-4 4 4 4', 'm16 8 4 4-4 4', 'm14 5-4 14'],
  layers: ['m12 3 9 5-9 5-9-5 9-5Z', 'm3 12 9 5 9-5', 'm3 16 9 5 9-5'],
  wand: ['m15 4 5 5', 'M4 20 17 7', 'm6 4 .8 2.2L9 7l-2.2.8L6 10l-.8-2.2L3 7l2.2-.8L6 4Z'],
  external: [
    'M14 4h6v6',
    'M20 4 10 14',
    'M20 14v4a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h4',
  ],
  x: ['M18 6 6 18', 'm6 6 12 12'],
  menu: ['M4 7h16', 'M4 12h16', 'M4 17h16'],
  copy: [
    'M8 8h10a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V10a2 2 0 0 1 2-2Z',
    'M16 8V4H6a2 2 0 0 0-2 2v10h4',
  ],
  back: ['M19 12H5', 'm11 18-6-6 6-6'],
  coin: [
    'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Z',
    'M15 9.5c-.7-.7-1.7-1-3-1-1.7 0-3 .8-3 2s1.3 1.8 3 2c1.7.2 3 .8 3 2s-1.3 2-3 2c-1.3 0-2.3-.3-3-1',
  ],
  bolt: ['m13 2-9 12h7l-1 8 9-12h-7l1-8Z'],
  shield: ['M12 3 4 6v6c0 5 8 9 8 9s8-4 8-9V6l-8-3Z', 'm9 12 2 2 4-4'],
  sun: [
    'M12 17a5 5 0 1 0 0-10 5 5 0 0 0 0 10Z',
    'M12 2v2',
    'M12 20v2',
    'm4.9 4.9 1.4 1.4',
    'm17.7 17.7 1.4 1.4',
    'M2 12h2',
    'M20 12h2',
    'm4.9 19.1 1.4-1.4',
    'm17.7 6.3 1.4-1.4',
  ],
  moon: ['M21 12.8A8.5 8.5 0 1 1 11.2 3a6.5 6.5 0 0 0 9.8 9.8Z'],
  eye: ['M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7Z', 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z'],
  'eye-off': [
    'm3 3 18 18',
    'M10.6 5.2A10 10 0 0 1 12 5c6.5 0 10 7 10 7a18 18 0 0 1-3.2 4.2',
    'M6.6 6.6A18 18 0 0 0 2 12s3.5 7 10 7a9 9 0 0 0 4.4-1.1',
    'M10.6 10.6a2 2 0 0 0 2.8 2.8',
  ],
  settings: [
    'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z',
    'M19.4 15a1.7 1.7 0 0 0 .3 1.9l.1.1-2.8 2.8-.1-.1a1.7 1.7 0 0 0-1.9-.3 1.7 1.7 0 0 0-1 1.6v.2h-4v-.2a1.7 1.7 0 0 0-1-1.6 1.7 1.7 0 0 0-1.9.3l-.1.1L4.2 17l.1-.1a1.7 1.7 0 0 0 .3-1.9 1.7 1.7 0 0 0-1.6-1H2.8v-4H3a1.7 1.7 0 0 0 1.6-1 1.7 1.7 0 0 0-.3-1.9L4.2 7 7 4.2l.1.1A1.7 1.7 0 0 0 9 4.6a1.7 1.7 0 0 0 1-1.6v-.2h4V3a1.7 1.7 0 0 0 1 1.6 1.7 1.7 0 0 0 1.9-.3l.1-.1L19.8 7l-.1.1a1.7 1.7 0 0 0-.3 1.9 1.7 1.7 0 0 0 1.6 1h.2v4H21a1.7 1.7 0 0 0-1.6 1Z',
  ],
  chart: ['M4 20V10', 'M10 20V4', 'M16 20v-7', 'M22 20H2'],
  book: [
    'M4 4h6a3 3 0 0 1 3 3v13a2 2 0 0 0-2-2H4V4Z',
    'M20 4h-4a3 3 0 0 0-3 3v13a2 2 0 0 1 2-2h5V4Z',
  ],
  trophy: [
    'M8 4h8v5a4 4 0 0 1-8 0V4Z',
    'M8 6H4v2a4 4 0 0 0 4 4',
    'M16 6h4v2a4 4 0 0 1-4 4',
    'M12 13v5',
    'M8 21h8',
    'M9 18h6',
  ],
  heart: [
    'M20.8 4.6a5.5 5.5 0 0 0-7.8 0L12 5.7l-1.1-1.1a5.5 5.5 0 0 0-7.8 7.8l1.1 1.1L12 21.3l7.8-7.8 1.1-1.1a5.5 5.5 0 0 0-.1-7.8Z',
  ],
}

const iconPaths = computed(() => paths[props.name])
</script>

<template>
  <svg
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    :stroke-width="stroke"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
    class="shrink-0"
  >
    <path v-for="(path, index) in iconPaths" :key="index" :d="path" />
  </svg>
</template>
