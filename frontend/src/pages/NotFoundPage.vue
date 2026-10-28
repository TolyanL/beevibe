<script setup lang="ts">
import { RouterLink } from 'vue-router'
import AppIcon from '@/components/ui/AppIcon.vue'
</script>

<template>
  <div class="pb-10 pt-16 sm:pt-24">
    <div class="app-container">
      <div class="mx-auto max-w-2xl text-center">
        <p
          class="font-mono text-[10px] font-bold uppercase tracking-[0.2em] text-bee-deep dark:text-bee"
        >
          404 · out of bounds
        </p>
        <h1 class="display-title mt-5 text-[clamp(4rem,13vw,8.5rem)] leading-[0.82]">Стена.</h1>
        <p class="mx-auto mt-6 max-w-md text-sm leading-6 text-muted">
          Эта страница вышла за пределы карты. Вернитесь на главную, создайте новую игру или найдите
          готовую в каталоге.
        </p>
        <div class="mt-7 flex flex-wrap justify-center gap-2.5">
          <RouterLink
            :to="{ name: 'home' }"
            class="inline-flex items-center gap-2 rounded-full border border-line-strong px-5 py-3 text-xs font-bold transition hover:bg-hover-surface-strong"
          >
            <AppIcon name="back" :size="15" />
            На главную
          </RouterLink>
          <RouterLink
            :to="{ name: 'create' }"
            class="inline-flex items-center gap-2 rounded-full bg-bee px-5 py-3 text-xs font-extrabold text-graphite transition hover:bg-bee-deep"
          >
            <AppIcon name="wand" :size="15" />
            Создать игру
          </RouterLink>
          <RouterLink
            to="/catalog"
            class="inline-flex items-center gap-2 rounded-full border border-line-strong px-5 py-3 text-xs font-bold transition hover:bg-hover-surface-strong"
          >
            <AppIcon name="grid" :size="15" />
            В каталог
          </RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>
