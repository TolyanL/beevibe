<script setup lang="ts">
import AppIcon from '@/components/ui/AppIcon.vue'
</script>

<template>
  <div
    class="relative overflow-hidden rounded-[1.6rem] border border-white/10 bg-[#0c0c0a] p-3 shadow-[0_40px_100px_-40px_rgba(0,0,0,.9)] sm:p-4"
  >
    <div class="absolute -right-16 -top-16 size-52 rounded-full bg-bee/15 blur-3xl" />
    <div class="absolute -bottom-20 -left-10 size-56 rounded-full bg-violet-500/10 blur-3xl" />

    <div class="relative flex items-center justify-between gap-3 px-1 pb-3">
      <div class="flex items-center gap-2">
        <span class="grid size-6 place-items-center rounded-md bg-bee text-black">
          <AppIcon name="bolt" :size="13" :stroke="2.4" />
        </span>
        <span class="font-mono text-[9px] font-bold uppercase tracking-[0.16em] text-white/50"
          >beevibe://live</span
        >
      </div>
      <span
        class="flex items-center gap-1.5 rounded-full border border-emerald-400/20 bg-emerald-400/10 px-2 py-1 font-mono text-[8px] font-bold uppercase tracking-[0.12em] text-emerald-300"
      >
        <span class="size-1 animate-pulse-soft rounded-full bg-emerald-400" />
        playable
      </span>
    </div>

    <div class="relative overflow-hidden rounded-[1.1rem] border border-white/10 bg-[#0e100c]">
      <div class="dot-grid absolute inset-0 opacity-25" />

      <svg viewBox="0 0 640 360" class="relative block w-full" aria-hidden="true">
        <defs>
          <linearGradient id="snake-body" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%" stop-color="#ffe14d" />
            <stop offset="100%" stop-color="#ffcc00" />
          </linearGradient>
          <filter id="snake-glow">
            <feGaussianBlur stdDeviation="8" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        <g stroke="rgba(255,255,255,.055)" stroke-width="1">
          <path d="M80 0v360M160 0v360M240 0v360M320 0v360M400 0v360M480 0v360M560 0v360" />
          <path d="M0 60h640M0 120h640M0 180h640M0 240h640M0 300h640" />
        </g>

        <path
          d="M116 244H238V142H380V264H512"
          fill="none"
          stroke="rgba(255,204,0,.17)"
          stroke-width="34"
          stroke-linecap="round"
          stroke-linejoin="round"
          filter="url(#snake-glow)"
        />
        <path
          d="M116 244H238V142H380V264H512"
          fill="none"
          stroke="url(#snake-body)"
          stroke-width="20"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
        <g fill="#0e100c">
          <circle cx="470" cy="264" r="4" />
          <circle cx="430" cy="264" r="4" />
          <circle cx="390" cy="264" r="4" />
          <circle cx="380" cy="224" r="4" />
          <circle cx="380" cy="184" r="4" />
          <circle cx="340" cy="142" r="4" />
          <circle cx="300" cy="142" r="4" />
          <circle cx="260" cy="142" r="4" />
          <circle cx="238" cy="182" r="4" />
          <circle cx="238" cy="222" r="4" />
        </g>
        <circle cx="546" cy="264" r="12" fill="#fb7185" />
        <circle
          cx="546"
          cy="264"
          r="20"
          fill="none"
          stroke="rgba(251,113,133,.25)"
          stroke-width="2"
        />
        <circle cx="512" cy="264" r="9" fill="#fff8d6" />
        <circle cx="515" cy="261" r="2.2" fill="#141414" />
        <circle cx="523" cy="261" r="2.2" fill="#141414" />

        <g font-family="JetBrains Mono, monospace" fill="rgba(255,255,255,.55)">
          <text x="26" y="34" font-size="11" letter-spacing="2">SCORE</text>
          <text x="26" y="60" font-size="24" fill="#f6f4e8" font-weight="bold">04 820</text>
          <text x="552" y="34" font-size="11" letter-spacing="2" text-anchor="end">BEST</text>
          <text x="614" y="60" font-size="16" fill="#ffcc00" font-weight="bold" text-anchor="end">
            12 400
          </text>
        </g>
      </svg>

      <div class="absolute bottom-3 left-3 right-3 flex items-end justify-between gap-3">
        <div class="rounded-xl border border-white/10 bg-black/45 px-3 py-2 backdrop-blur">
          <p class="font-mono text-[8px] uppercase tracking-[0.15em] text-white/35">
            prompt → game
          </p>
          <p class="mt-0.5 max-w-[210px] text-[11px] font-bold leading-4 text-white/80">
            «неоновая змейка с ускоряющимся темпом»
          </p>
        </div>
        <span
          class="grid size-11 shrink-0 place-items-center rounded-full bg-bee text-black shadow-[0_0_30px_rgba(255,204,0,.35)]"
        >
          <AppIcon name="play" :size="18" :stroke="2.2" />
        </span>
      </div>
    </div>

    <div class="mt-3 flex items-center gap-2 px-1">
      <span
        class="rounded-full border border-white/10 bg-white/5 px-2.5 py-1 font-mono text-[8px] uppercase tracking-[0.12em] text-white/45"
      >
        responsive
      </span>
      <span
        class="rounded-full border border-white/10 bg-white/5 px-2.5 py-1 font-mono text-[8px] uppercase tracking-[0.12em] text-white/45"
      >
        html5
      </span>
      <span class="ml-auto font-mono text-[8px] uppercase tracking-[0.12em] text-white/25"
        >build 00:42</span
      >
    </div>
  </div>
</template>
