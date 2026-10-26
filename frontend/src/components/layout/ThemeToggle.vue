<script setup lang="ts">
import { useThemeStore } from '@/stores/theme'
import AppIcon from '@/components/ui/AppIcon.vue'

const theme = useThemeStore()
</script>

<template>
  <button
    type="button"
    class="grid size-10 place-items-center rounded-full border border-line bg-surface/70 text-ink transition hover:border-bee-deep hover:text-bee-deep"
    :aria-label="theme.isDark ? 'Включить светлую тему' : 'Включить тёмную тему'"
    :title="theme.isDark ? 'Светлая тема' : 'Тёмная тема'"
    @click="theme.toggle()"
  >
    <AppIcon :name="theme.isDark ? 'sun' : 'moon'" :size="18" />
  </button>
</template>
