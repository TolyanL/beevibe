<script setup lang="ts">
withDefaults(defineProps<{ compact?: boolean }>(), { compact: false })
</script>

<template>
  <div class="inline-flex items-center gap-2.5">
    <span
      class="relative grid size-10 place-items-center overflow-hidden rounded-[14px] bg-bee text-graphite shadow-[0_8px_24px_-12px_rgba(0,0,0,.8)]"
    >
      <svg viewBox="0 0 32 32" class="size-7" aria-hidden="true">
        <path
          d="M8 7h8c4.4 0 7 2.1 7 5.6 0 2.2-1.1 3.8-2.8 4.6 2.2.7 3.5 2.3 3.5 4.6 0 3.6-2.9 5.9-7.3 5.9H8V7Zm7 7.6c2.1 0 3.4-.8 3.4-2.4 0-1.6-1.3-2.3-3.4-2.3h-1.7v4.7H15Zm.4 8.8c2.3 0 3.7-.8 3.7-2.5s-1.4-2.5-3.7-2.5h-2.1v5h2.1Z"
          fill="currentColor"
        />
      </svg>
      <span class="absolute right-1.5 top-1.5 size-1.5 rounded-full bg-graphite/70" />
    </span>
    <span v-if="!compact" class="leading-none">
      <span class="block font-display text-[1.02rem] font-bold tracking-[-0.05em]">BeeVibe</span>
      <span class="mt-1 block font-mono text-[0.54rem] font-bold uppercase tracking-[0.18em] text-muted">
        game lab
      </span>
    </span>
  </div>
</template>
