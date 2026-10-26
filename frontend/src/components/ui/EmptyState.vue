<script setup lang="ts">
import AppIcon from './AppIcon.vue'

withDefaults(
  defineProps<{
    icon?: 'gamepad' | 'search' | 'alert' | 'clock' | 'spark' | 'wand'
    title: string
    text?: string
  }>(),
  { icon: 'gamepad' },
)
</script>

<template>
  <div class="flex flex-col items-center justify-center rounded-[1.35rem] border border-dashed border-line-strong px-6 py-14 text-center">
    <span class="grid size-14 place-items-center rounded-2xl border border-line bg-surface text-bee-deep">
      <AppIcon :name="icon" :size="26" />
    </span>
    <h3 class="mt-5 text-xl font-bold">{{ title }}</h3>
    <p v-if="text" class="mt-2 max-w-md text-sm leading-6 text-muted">{{ text }}</p>
    <div v-if="$slots.default" class="mt-5">
      <slot />
    </div>
  </div>
</template>
