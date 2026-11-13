<script setup lang="ts">
import AppIcon from './AppIcon.vue'

withDefaults(defineProps<{ title?: string; text: string }>(), { title: 'Не удалось загрузить' })

defineEmits<{ retry: [] }>()
</script>

<template>
  <div class="flex flex-col items-center gap-3 px-5 py-8 text-center">
    <span
      class="grid size-10 place-items-center rounded-xl border border-red-500/20 bg-red-500/10 text-red-500"
    >
      <AppIcon name="alert" :size="18" />
    </span>
    <div>
      <p class="text-xs font-bold">{{ title }}</p>
      <p class="mt-1 text-[11px] leading-5 text-muted">{{ text }}</p>
    </div>
    <button
      type="button"
      class="rounded-full border border-line-strong px-3.5 py-1.5 text-[10px] font-bold text-muted transition hover:border-bee-deep hover:text-ink"
      @click="$emit('retry')"
    >
      Повторить
    </button>
  </div>
</template>
