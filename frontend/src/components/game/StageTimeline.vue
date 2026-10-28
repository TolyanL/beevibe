<script setup lang="ts">
import { computed } from 'vue'
import type { GameStage } from '@/api/types'
import { PIPELINE, statusMeta } from '@/lib/game'
import AppIcon from '@/components/ui/AppIcon.vue'

const props = withDefaults(
  defineProps<{
    stages: GameStage[]
    status: string
    compact?: boolean
  }>(),
  { compact: false },
)

const stageMap = computed(() => {
  const map = new Map<string, GameStage>()
  for (const stage of props.stages) map.set(stage.stage, stage)
  return map
})

const currentIndex = computed(() => {
  if (props.status === 'ready') return PIPELINE.length
  const index = PIPELINE.findIndex((item) => item.id === props.status)
  return index < 0 ? 0 : index
})

const progress = computed(() => {
  if (props.status === 'ready') return 100
  if (props.status === 'failed') {
    const done = props.stages.filter((stage) => stage.status === 'done').length
    return Math.round((done / PIPELINE.length) * 100)
  }

  const done = props.stages.filter((stage) => stage.status === 'done').length
  return Math.max(Math.round((done / PIPELINE.length) * 100), currentIndex.value * 10)
})

type StageState = 'done' | 'active' | 'pending' | 'failed'

function stateOf(id: string, index: number): StageState {
  const stage = stageMap.value.get(id)
  if (stage?.status === 'done') return 'done'
  if (props.status === 'ready') return 'done'
  if (props.status === 'failed') {
    if (stage?.status === 'running') return 'failed'
    return index < currentIndex.value ? 'done' : 'pending'
  }
  if (stage?.status === 'running') return 'active'
  if (id === props.status) return 'active'
  if (index < currentIndex.value) return 'done'
  return 'pending'
}

function duration(ms?: number): string {
  if (!ms) return ''
  if (ms < 1000) return `${ms} мс`
  return `${(ms / 1000).toFixed(ms >= 10_000 ? 0 : 1)} с`
}

const meta = computed(() => statusMeta(props.status))
</script>

<template>
  <div>
    <div class="mb-4 flex items-center justify-between gap-4">
      <div>
        <p class="mono-label">Этапы сборки</p>
        <p class="mt-1 text-sm font-bold">{{ meta.label }}</p>
      </div>
      <span class="font-mono text-xs font-bold tabular-nums text-muted">{{ progress }}%</span>
    </div>

    <div
      class="mb-5 h-1.5 overflow-hidden rounded-full bg-line"
      :style="{ '--pipeline-progress': `${progress}%` }"
    >
      <div
        class="pipeline-line h-full rounded-full transition-[width] duration-500"
        :style="{ width: `${progress}%` }"
      />
    </div>

    <div class="space-y-0.5">
      <div
        v-for="(stage, index) in PIPELINE"
        :key="stage.id"
        class="relative flex items-start gap-3 rounded-xl px-2 py-2.5 transition"
        :class="stateOf(stage.id, index) === 'active' ? 'bg-bee/10' : ''"
      >
        <span
          class="relative z-10 mt-0.5 grid size-5 shrink-0 place-items-center rounded-full border text-[9px] font-bold"
          :class="{
            'border-emerald-500/30 bg-emerald-500 text-black': stateOf(stage.id, index) === 'done',
            'border-bee bg-bee text-black': stateOf(stage.id, index) === 'active',
            'border-red-500/30 bg-red-500 text-white': stateOf(stage.id, index) === 'failed',
            'border-line-strong bg-surface text-faint': stateOf(stage.id, index) === 'pending',
          }"
        >
          <AppIcon v-if="stateOf(stage.id, index) === 'done'" name="check" :size="10" :stroke="3" />
          <span
            v-else-if="stateOf(stage.id, index) === 'active'"
            class="size-1.5 animate-pulse-soft rounded-full bg-black"
          />
          <span v-else-if="stateOf(stage.id, index) === 'failed'">!</span>
          <span v-else>{{ index + 1 }}</span>
        </span>

        <div class="min-w-0 flex-1">
          <div class="flex items-baseline justify-between gap-2">
            <p
              class="text-[12px] font-bold"
              :class="stateOf(stage.id, index) === 'pending' ? 'text-faint' : 'text-ink'"
            >
              {{ stage.label }}
            </p>
            <span
              v-if="stageMap.get(stage.id)?.durationMs"
              class="font-mono text-[9px] tabular-nums text-faint"
            >
              {{ duration(stageMap.get(stage.id)?.durationMs) }}
            </span>
          </div>
          <p v-if="!compact && stageMap.get(stage.id)?.issues" class="mt-0.5 text-[10px] text-warn">
            {{ stageMap.get(stage.id)?.issues }} замечаний
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
