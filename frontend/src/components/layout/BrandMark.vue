<script setup lang="ts">
withDefaults(defineProps<{ compact?: boolean }>(), { compact: false })
</script>

<template>
  <div class="inline-flex items-center gap-2.5">
    <span
      class="grid size-10 shrink-0 place-items-center shadow-[0_8px_24px_-12px_rgba(0,0,0,.8)]"
    >
      <svg viewBox="0 0 64 64" class="size-10" aria-hidden="true">
        <rect width="64" height="64" rx="16" fill="#FFCC00" />
        <path
          d="M20 18h14c8 0 13 4 13 10 0 4-2 7-5 8 4 1 6 4 6 8 0 6-5 10-13 10H20V18Zm12 14c4 0 6-1 6-4s-2-4-6-4h-3v8h3Zm1 16c4 0 7-1 7-4s-3-4-7-4h-4v8h4Z"
          fill="#141414"
        />
        <circle cx="12" cy="12" r="3" fill="#141414" />
        <circle cx="52" cy="52" r="3" fill="#141414" />
      </svg>
    </span>
    <span v-if="!compact" class="leading-none">
      <span class="block font-display text-[1.02rem] font-bold tracking-[-0.05em]">BeeVibe</span>
      <span
        class="mt-1 block font-mono text-[0.54rem] font-bold uppercase tracking-[0.18em] text-muted"
      >
        игры из фразы
      </span>
    </span>
  </div>
</template>
