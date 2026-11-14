<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useThemeStore } from '@/stores/theme'
import AppIcon from '@/components/ui/AppIcon.vue'
import starryEyes from '@/assets/emoji/starry-eyes.svg'

const theme = useThemeStore()

const stages = [
  { label: 'Очередь', icon: 'clock' },
  { label: 'Разбор идеи', icon: 'search' },
  { label: 'План и правила', icon: 'layers' },
  { label: 'Код игры', icon: 'code' },
  { label: 'Графика', icon: 'spark' },
  { label: 'Ревью', icon: 'eye' },
  { label: 'Сборка', icon: 'layers' },
  { label: 'Проверка', icon: 'shield' },
] as const

const STEP_MS = 900
const FINISH_MS = 550
const PARTY_MS = 3000
const RESET_MS = 700

const activeIndex = ref(0)
const celebrating = ref(false)

let timer: ReturnType<typeof setTimeout> | undefined
let stopped = false

const statusTitle = computed(() =>
  celebrating.value ? 'Игра готова' : stages[activeIndex.value]!.label,
)

const statusIcon = computed(() => (celebrating.value ? 'check' : stages[activeIndex.value]!.icon))

const progress = computed(() =>
  celebrating.value ? 100 : Math.round(((activeIndex.value + 1) / stages.length) * 100),
)

function statusOf(index: number): 'done' | 'active' | 'pending' {
  if (celebrating.value || index < activeIndex.value) return 'done'
  if (index === activeIndex.value) return 'active'

  return 'pending'
}

const stageRows = computed(() =>
  stages.map((stage, index) => ({
    ...stage,
    num: String(index + 1).padStart(2, '0'),
    status: statusOf(index),
  })),
)

function prefersReducedMotion(): boolean {
  if (theme.reducedMotion) return true
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false

  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

function schedule(callback: () => void, delay: number): void {
  timer = setTimeout(() => {
    timer = undefined
    if (!stopped) callback()
  }, delay)
}

function runStage(): void {
  if (stopped) return

  if (activeIndex.value < stages.length - 1) {
    schedule(() => {
      activeIndex.value += 1
      runStage()
    }, STEP_MS)

    return
  }

  schedule(() => {
    celebrating.value = true

    schedule(() => {
      celebrating.value = false

      schedule(() => {
        activeIndex.value = 0
        runStage()
      }, RESET_MS)
    }, PARTY_MS)
  }, FINISH_MS)
}

onMounted(() => {
  if (prefersReducedMotion()) {
    activeIndex.value = stages.length - 1
    celebrating.value = true

    return
  }

  runStage()
})

onBeforeUnmount(() => {
  stopped = true
  if (timer) clearTimeout(timer)
})
</script>

<template>
  <div class="build-pipeline">
    <span class="pipe-sticker pipe-sticker--time" aria-hidden="true">≈ 15 мин</span>
    <span class="pipe-sticker pipe-sticker--file" aria-hidden="true">1 файл</span>

    <div class="pipe-card">
      <div class="pipe-bar">
        <span class="pipe-dots" aria-hidden="true"><i></i><i></i><i></i></span>
        <span class="pipe-address">beevibe://build</span>
        <span class="pipe-live" :class="{ 'is-done': celebrating }">
          <i aria-hidden="true"></i>
          {{ celebrating ? 'готово' : 'в эфире' }}
        </span>
      </div>

      <div class="pipe-status" :class="{ 'is-done': celebrating }">
        <div class="pipe-status-main">
          <span class="pipe-status-icon" aria-hidden="true">
            <AppIcon :name="statusIcon" :size="15" :stroke="2.2" />
          </span>
          <div class="pipe-status-copy">
            <p class="pipe-status-kicker">конвейер сборки</p>
            <p class="pipe-status-title">{{ statusTitle }}</p>
          </div>
          <span class="pipe-status-pct">{{ progress }}%</span>
        </div>
        <div
          class="pipe-progress"
          role="progressbar"
          aria-label="Прогресс сборки игры"
          aria-valuemin="0"
          aria-valuemax="100"
          :aria-valuenow="progress"
        >
          <i :style="{ width: `${progress}%` }"></i>
        </div>
      </div>

      <div class="pipe-stage-wrap">
        <ol class="pipe-stages">
          <li
            v-for="stage in stageRows"
            :key="stage.label"
            class="pipe-stage"
            :class="`is-${stage.status}`"
          >
            <span class="pipe-node" aria-hidden="true">
              <AppIcon v-if="stage.status === 'done'" name="check" :size="9" :stroke="3.2" />
            </span>
            <span class="pipe-stage-label">{{ stage.label }}</span>
            <span class="pipe-stage-num" aria-hidden="true">{{ stage.num }}</span>
          </li>
        </ol>

        <Transition name="pipe-party">
          <div v-if="celebrating" class="pipe-party" aria-hidden="true">
            <div class="pipe-fireworks">
              <i v-for="n in 24" :key="n" :style="{ '--i': n - 1 }"></i>
              <span class="pipe-burst-ring"></span>
              <span class="pipe-flash"></span>
            </div>
            <div class="pipe-sparkles">
              <span v-for="n in 12" :key="n" :style="{ '--s': n - 1 }"></span>
            </div>
            <img class="pipe-emoji" :src="starryEyes" alt="" />
            <p class="pipe-party-title">Игра готова</p>
            <p class="pipe-party-text">8 этапов · один html-файл</p>
            <span class="pipe-party-tag">без установки</span>
          </div>
        </Transition>
      </div>
    </div>
  </div>
</template>

<style scoped>
.build-pipeline {
  position: relative;
  width: 100%;
  max-width: 470px;
  margin-inline: auto;
}

.pipe-card {
  padding: 10px;
  border: 2px solid #141414;
  border-radius: 28px;
  background: #101010;
  color: #fffdf7;
  box-shadow:
    0 30px 60px -30px rgba(17, 17, 17, 0.75),
    0 5px 0 rgba(17, 17, 17, 0.28);
  transform: rotate(-1.6deg);
}

.pipe-bar {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.15rem 0.35rem 0.7rem;
  font-family: var(--font-mono);
  font-size: 0.5rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.pipe-dots {
  display: flex;
  flex-shrink: 0;
  gap: 1px;
}

.pipe-dots i {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: #ffcc00;
}

.pipe-dots i:nth-child(2) {
  background: rgba(255, 253, 247, 0.42);
}

.pipe-dots i:nth-child(3) {
  background: rgba(255, 253, 247, 0.16);
}

.pipe-address {
  min-width: 0;
  overflow: hidden;
  color: rgba(255, 253, 247, 0.5);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pipe-live {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  gap: 0.3rem;
  margin-left: auto;
  color: rgba(255, 253, 247, 0.66);
}

.pipe-live.is-done {
  color: #ffcc00;
}

.pipe-live i {
  width: 5px;
  height: 5px;
  border-radius: 999px;
  background: #ffcc00;
  animation: pipe-blink 2.4s ease-in-out infinite;
}

.pipe-status {
  padding: 0.6rem 0.75rem 0.7rem;
  border: 1px solid rgba(255, 204, 0, 0.26);
  border-radius: 0.9rem;
  background: #1a1a17;
  transition:
    background-color 0.3s ease,
    border-color 0.3s ease,
    color 0.3s ease;
}

.pipe-status.is-done {
  border-color: #ffcc00;
  background: #ffcc00;
  color: #141414;
}

.pipe-status-main {
  display: flex;
  align-items: center;
  gap: 0.55rem;
}

.pipe-status-icon {
  display: grid;
  flex-shrink: 0;
  place-items: center;
  width: 1.9rem;
  height: 1.9rem;
  border-radius: 0.55rem;
  background: #ffcc00;
  color: #141414;
  transition:
    background-color 0.3s ease,
    color 0.3s ease;
}

.pipe-status.is-done .pipe-status-icon {
  background: #141414;
  color: #ffcc00;
}

.pipe-status-copy {
  min-width: 0;
  flex: 1;
}

.pipe-status-kicker {
  font-family: var(--font-mono);
  font-size: 0.46rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: rgba(255, 253, 247, 0.45);
}

.pipe-status.is-done .pipe-status-kicker {
  color: rgba(20, 20, 20, 0.55);
}

.pipe-status-title {
  margin-top: 0.1rem;
  overflow: hidden;
  font-family: var(--font-display);
  font-size: 0.98rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  line-height: 1.1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pipe-status-pct {
  flex-shrink: 0;
  font-family: var(--font-display);
  font-size: 0.92rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  color: #ffcc00;
  font-variant-numeric: tabular-nums;
}

.pipe-status.is-done .pipe-status-pct {
  color: #141414;
}

.pipe-progress {
  height: 3px;
  margin-top: 0.55rem;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(255, 253, 247, 0.12);
  transition: background-color 0.3s ease;
}

.pipe-status.is-done .pipe-progress {
  background: rgba(20, 20, 20, 0.18);
}

.pipe-progress i {
  display: block;
  height: 100%;
  border-radius: 999px;
  background: #ffcc00;
  transition:
    width 0.5s cubic-bezier(0.2, 0.8, 0.2, 1),
    background-color 0.3s ease;
}

.pipe-status.is-done .pipe-progress i {
  background: #141414;
}

.pipe-stage-wrap {
  position: relative;
  margin-top: 0.25rem;
  overflow: hidden;
  border-radius: 0.9rem;
}

.pipe-stages {
  position: relative;
  display: flex;
  flex-direction: column;
  margin: 0;
  padding: 0.4rem 0.35rem 0.4rem 0;
  list-style: none;
}

.pipe-stages::before {
  position: absolute;
  top: 0.9rem;
  bottom: 0.9rem;
  left: 1.06rem;
  width: 2px;
  content: '';
  background: repeating-linear-gradient(
    to bottom,
    rgba(255, 204, 0, 0.32) 0 5px,
    transparent 5px 10px
  );
  background-size: 2px 20px;
  animation: pipe-belt 1.1s linear infinite;
}

.pipe-stage {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.55rem;
  min-height: 2rem;
  padding: 0.15rem 0.5rem 0.15rem 0.5rem;
  border-radius: 0.55rem;
  transition: background-color 0.25s ease;
}

.pipe-stage.is-active {
  background: linear-gradient(90deg, rgba(255, 204, 0, 0.16), rgba(255, 204, 0, 0));
}

.pipe-node {
  position: relative;
  z-index: 1;
  display: grid;
  flex-shrink: 0;
  place-items: center;
  width: 1.15rem;
  height: 1.15rem;
  border: 1px solid rgba(255, 253, 247, 0.2);
  border-radius: 999px;
  background: #101010;
  transition:
    background-color 0.25s ease,
    border-color 0.25s ease,
    box-shadow 0.25s ease;
}

.pipe-stage.is-done .pipe-node {
  border-color: #ffcc00;
  background: #ffcc00;
  color: #141414;
}

.pipe-stage.is-active .pipe-node {
  border-color: #ffcc00;
  background: #ffcc00;
  box-shadow: 0 0 0 4px rgba(255, 204, 0, 0.14);
  animation: pipe-pulse 1.3s ease-in-out infinite;
}

.pipe-stage.is-active .pipe-node::after {
  width: 5px;
  height: 5px;
  border-radius: 999px;
  content: '';
  background: #141414;
}

.pipe-stage-label {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  font-size: 0.74rem;
  font-weight: 700;
  letter-spacing: 0.01em;
  line-height: 1.2;
  white-space: nowrap;
  text-overflow: ellipsis;
  color: rgba(255, 253, 247, 0.92);
  transition: color 0.25s ease;
}

.pipe-stage.is-done .pipe-stage-label {
  color: rgba(255, 253, 247, 0.68);
}

.pipe-stage.is-pending .pipe-stage-label {
  color: rgba(255, 253, 247, 0.44);
}

.pipe-stage-num {
  flex-shrink: 0;
  font-family: var(--font-mono);
  font-size: 0.46rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: rgba(255, 253, 247, 0.32);
  transition: color 0.25s ease;
}

.pipe-stage.is-active .pipe-stage-num {
  color: #ffcc00;
}

.pipe-stage.is-done .pipe-stage-num {
  color: rgba(255, 204, 0, 0.5);
}

.pipe-party {
  position: absolute;
  inset: 0;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.1rem;
  overflow: hidden;
  border-radius: 0.9rem;
  background: rgba(10, 10, 9, 0.91);
  text-align: center;
  backdrop-filter: blur(3px);
}

.pipe-party-enter-active,
.pipe-party-leave-active {
  transition: opacity 0.3s ease;
}

.pipe-party-enter-from,
.pipe-party-leave-to {
  opacity: 0;
}

.pipe-emoji {
  width: 3.8rem;
  height: 3.8rem;
  filter: drop-shadow(0 7px 12px rgba(0, 0, 0, 0.42));
  animation: pipe-emoji-pop 0.55s 0.04s cubic-bezier(0.2, 1.35, 0.4, 1) both;
}

.pipe-party-title {
  margin-top: 0.25rem;
  font-family: var(--font-display);
  font-size: 1.12rem;
  font-weight: 800;
  letter-spacing: -0.04em;
  line-height: 1;
  color: #ffcc00;
}

.pipe-party-text {
  margin-top: 0.25rem;
  font-family: var(--font-mono);
  font-size: 0.48rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: rgba(255, 253, 247, 0.5);
}

.pipe-party-tag {
  margin-top: 0.45rem;
  padding: 0.2rem 0.5rem;
  border: 1px solid rgba(255, 204, 0, 0.38);
  border-radius: 999px;
  font-family: var(--font-mono);
  font-size: 0.46rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: #ffcc00;
}

.pipe-fireworks {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.pipe-fireworks i {
  position: absolute;
  top: 45%;
  left: 50%;
  width: 4px;
  height: 15px;
  border-radius: 999px;
  background: #ffcc00;
  transform-origin: 50% 100%;
  animation: pipe-burst 1.7s cubic-bezier(0.16, 0.84, 0.3, 1) both;
  animation-delay: calc(var(--i) * 10ms);
}

.pipe-fireworks i:nth-child(3n + 2) {
  width: 2px;
  background: #fffdf7;
}

.pipe-fireworks i:nth-child(3n) {
  height: 8px;
  background: #ff8a00;
}

.pipe-burst-ring {
  position: absolute;
  top: 45%;
  left: 50%;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 204, 0, 0.62);
  border-radius: 999px;
  transform: translate(-50%, -50%);
  animation: pipe-ring 0.9s ease-out both;
}

.pipe-sparkles {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.pipe-sparkles span {
  position: absolute;
  width: 8px;
  height: 8px;
  background: #ffcc00;
  clip-path: polygon(
    50% 0,
    61% 35%,
    98% 35%,
    68% 57%,
    79% 91%,
    50% 70%,
    21% 91%,
    32% 57%,
    2% 35%,
    39% 35%
  );
  animation: pipe-twinkle 1.6s ease-in-out infinite;
  animation-delay: calc(var(--s) * -0.17s);
}

.pipe-sparkles span:nth-child(3n) {
  background: #fffdf7;
}

.pipe-sparkles span:nth-child(1) {
  top: 9%;
  left: 17%;
  width: 11px;
  height: 11px;
}

.pipe-sparkles span:nth-child(2) {
  top: 12%;
  right: 13%;
  width: 7px;
  height: 7px;
}

.pipe-sparkles span:nth-child(3) {
  top: 33%;
  left: 6%;
  width: 6px;
  height: 6px;
}

.pipe-sparkles span:nth-child(4) {
  top: 26%;
  right: 25%;
  width: 5px;
  height: 5px;
}

.pipe-sparkles span:nth-child(5) {
  top: 56%;
  left: 12%;
  width: 8px;
  height: 8px;
}

.pipe-sparkles span:nth-child(6) {
  top: 64%;
  right: 8%;
  width: 12px;
  height: 12px;
}

.pipe-sparkles span:nth-child(7) {
  bottom: 14%;
  left: 27%;
  width: 6px;
  height: 6px;
}

.pipe-sparkles span:nth-child(8) {
  bottom: 10%;
  right: 31%;
  width: 8px;
  height: 8px;
}

.pipe-sparkles span:nth-child(9) {
  top: 5%;
  left: 44%;
  width: 5px;
  height: 5px;
}

.pipe-sparkles span:nth-child(10) {
  bottom: 25%;
  right: 19%;
  width: 5px;
  height: 5px;
}

.pipe-sparkles span:nth-child(11) {
  top: 45%;
  right: 3%;
  width: 4px;
  height: 4px;
}

.pipe-sparkles span:nth-child(12) {
  top: 43%;
  left: 3%;
  width: 4px;
  height: 4px;
}

.pipe-flash {
  position: absolute;
  top: 45%;
  left: 50%;
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: #ffcc00;
  transform: translate(-50%, -50%);
  animation: pipe-flash 0.9s ease-out both;
}

.pipe-sticker {
  position: absolute;
  z-index: 3;
  padding: 0.45rem 0.7rem;
  border: 2px solid #141414;
  border-radius: 10px;
  font-family: var(--font-display);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  line-height: 1;
  text-transform: uppercase;
  box-shadow: 0 4px 0 rgba(17, 17, 17, 0.22);
}

.pipe-sticker--time {
  top: -0.9rem;
  left: -0.35rem;
  background: #fffdf7;
  color: #141414;
  transform: rotate(-7deg);
}

.pipe-sticker--file {
  right: -0.35rem;
  bottom: -0.85rem;
  background: #141414;
  color: #ffcc00;
  transform: rotate(4deg);
}

@keyframes pipe-belt {
  from {
    background-position-y: 0;
  }
  to {
    background-position-y: 20px;
  }
}

@keyframes pipe-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 4px rgba(255, 204, 0, 0.14);
  }
  50% {
    box-shadow: 0 0 0 7px rgba(255, 204, 0, 0.06);
  }
}

@keyframes pipe-blink {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.3;
  }
}

@keyframes pipe-emoji-pop {
  0% {
    opacity: 0;
    transform: scale(0.3) rotate(-12deg);
  }
  100% {
    opacity: 1;
    transform: scale(1) rotate(0deg);
  }
}

@keyframes pipe-burst {
  0% {
    opacity: 0;
    transform: rotate(calc(var(--i) * 15deg)) translateY(-52px) scaleY(0.25);
  }
  12% {
    opacity: 1;
  }
  68% {
    opacity: 1;
  }
  100% {
    opacity: 0;
    transform: rotate(calc(var(--i) * 15deg)) translateY(-96px) scaleY(1);
  }
}

@keyframes pipe-flash {
  0% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.2);
  }
  30% {
    opacity: 0.9;
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(3);
  }
}

@keyframes pipe-ring {
  0% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.2);
  }
  20% {
    opacity: 1;
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(5);
  }
}

@keyframes pipe-twinkle {
  0%,
  100% {
    opacity: 0.25;
    transform: scale(0.7) rotate(0deg);
  }
  50% {
    opacity: 1;
    transform: scale(1.15) rotate(20deg);
  }
}

@media (max-width: 420px) {
  .pipe-stage {
    min-height: 1.9rem;
  }

  .pipe-party-title {
    font-size: 1rem;
  }

  .pipe-sticker--time {
    left: 0.25rem;
  }

  .pipe-sticker--file {
    right: 0.25rem;
  }
}
</style>
