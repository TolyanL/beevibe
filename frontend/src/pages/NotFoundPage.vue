<script setup lang="ts">
import { RouterLink } from 'vue-router'
import AppIcon from '@/components/ui/AppIcon.vue'
</script>

<template>
  <div class="pb-10 pt-16 sm:pt-24">
    <div class="app-container">
      <div class="mx-auto max-w-2xl text-center">
        <p class="font-mono text-[10px] font-bold uppercase tracking-[0.2em] text-bee-deep dark:text-bee">404 · out of bounds</p>
        <h1 class="display-title mt-5 text-[clamp(4rem,13vw,8.5rem)] leading-[0.82]">
          Стена.
        </h1>
        <p class="mx-auto mt-6 max-w-md text-sm leading-6 text-muted">
          Эта страница вышла за пределы карты. Вернитесь в генератор или найдите готовую игру в
          каталоге.
        </p>
        <div class="mt-7 flex flex-wrap justify-center gap-2.5">
          <RouterLink
            to="/"
            class="inline-flex items-center gap-2 rounded-full bg-bee px-5 py-3 text-xs font-extrabold text-graphite transition hover:bg-bee-deep"
          >
            <AppIcon name="wand" :size="15" />
            В генератор
          </RouterLink>
          <RouterLink
            to="/catalog"
            class="inline-flex items-center gap-2 rounded-full border border-line-strong px-5 py-3 text-xs font-bold transition hover:bg-surface"
          >
            <AppIcon name="grid" :size="15" />
            В каталог
          </RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>
