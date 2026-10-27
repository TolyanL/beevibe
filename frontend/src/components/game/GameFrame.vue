<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useToast } from 'openvue/usetoast'
import AppIcon from '@/components/ui/AppIcon.vue'
import StatusPill from '@/components/ui/StatusPill.vue'
import { statusMeta } from '@/lib/game'

const props = withDefaults(
  defineProps<{
    gameId: string
    slug?: string
    status: string
    versionToken?: number
  }>(),
  { versionToken: 0 },
)

const emit = defineEmits<{ loaded: [] }>()

const toast = useToast()
const frame = ref<HTMLIFrameElement | null>(null)
const shell = ref<HTMLElement | null>(null)
const isFullscreen = ref(false)
const loading = ref(true)
const manualReload = ref(0)

const gameBase =
  (import.meta.env.VITE_GAME_BASE_URL as string | undefined)?.replace(/\/$/, '') ?? ''
const src = computed(
  () =>
    `${gameBase}/g/${encodeURIComponent(props.gameId)}?v=${props.versionToken}-${manualReload.value}`,
)
const downloadUrl = computed(() => `${gameBase}/g/${encodeURIComponent(props.gameId)}/download`)
const publicUrl = computed(() => `${window.location.origin}/play/${props.slug || props.gameId}`)
function onFrameLoad(): void {
  loading.value = false
  emit('loaded')
}

const ready = computed(() => props.status === 'ready')
const meta = computed(() => statusMeta(props.status))

watch(
  () => props.versionToken,
  () => {
    loading.value = true
  },
)

function syncFullscreen(): void {
  isFullscreen.value = document.fullscreenElement === shell.value
}

onMounted(() => document.addEventListener('fullscreenchange', syncFullscreen))
onBeforeUnmount(() => document.removeEventListener('fullscreenchange', syncFullscreen))

async function toggleFullscreen(): Promise<void> {
  try {
    if (document.fullscreenElement) {
      await document.exitFullscreen()
    } else {
      await shell.value?.requestFullscreen()
    }
  } catch {
    toast.add({
      severity: 'error',
      summary: 'Fullscreen недоступен',
      detail: 'Браузер не разрешил полноэкранный режим.',
      life: 3000,
    })
  }
}

function reload(): void {
  loading.value = true
  manualReload.value += 1
}

async function share(): Promise<void> {
  const url = publicUrl.value
  try {
    if (navigator.share) {
      await navigator.share({
        title: 'BeeVibe — игра',
        text: 'Сыграй в игру, собранную BeeVibe',
        url,
      })
    } else {
      await navigator.clipboard.writeText(url)
      toast.add({ severity: 'success', summary: 'Ссылка скопирована', life: 2200 })
    }
  } catch {
    // A cancelled share dialog is not an error.
  }
}
</script>

<template>
  <div ref="shell" class="stage-shell">
    <div
      class="flex flex-wrap items-center justify-between gap-3 border-b border-white/10 px-3.5 py-2.5 sm:px-4"
    >
      <div class="flex items-center gap-2.5">
        <span class="flex gap-1.5" aria-hidden="true">
          <span class="size-2 rounded-full bg-white/20" />
          <span class="size-2 rounded-full bg-white/20" />
          <span class="size-2 rounded-full bg-white/20" />
        </span>
        <span class="font-mono text-[10px] uppercase tracking-[0.14em] text-white/40"
          >playable stage</span
        >
      </div>

      <div class="flex items-center gap-1">
        <StatusPill :status="status" compact class="hidden sm:inline-flex" />
        <button
          type="button"
          class="grid size-8 place-items-center rounded-lg text-white/55 transition hover:bg-white/10 hover:text-white"
          title="Перезапустить"
          aria-label="Перезапустить игру"
          @click="reload"
        >
          <AppIcon name="refresh" :size="15" />
        </button>
        <button
          type="button"
          class="grid size-8 place-items-center rounded-lg text-white/55 transition hover:bg-white/10 hover:text-white"
          title="Поделиться"
          aria-label="Поделиться игрой"
          @click="share"
        >
          <AppIcon name="share" :size="15" />
        </button>
        <a
          :href="downloadUrl"
          class="grid size-8 place-items-center rounded-lg text-white/55 transition hover:bg-white/10 hover:text-white"
          title="Скачать HTML"
          aria-label="Скачать игру"
          download
        >
          <AppIcon name="download" :size="15" />
        </a>
        <button
          type="button"
          class="grid size-8 place-items-center rounded-lg text-white/55 transition hover:bg-white/10 hover:text-white"
          :title="isFullscreen ? 'Выйти из fullscreen' : 'На весь экран'"
          :aria-label="isFullscreen ? 'Выйти из полноэкранного режима' : 'Открыть на весь экран'"
          @click="toggleFullscreen"
        >
          <AppIcon :name="isFullscreen ? 'compress' : 'expand'" :size="15" />
        </button>
      </div>
    </div>

    <div class="stage-frame-wrap relative bg-[#0b0b09]">
      <iframe
        v-if="ready"
        ref="frame"
        :key="`${versionToken}-${manualReload}`"
        :src="src"
        :title="`Игра ${gameId}`"
        class="stage-frame"
        sandbox="allow-scripts allow-pointer-lock allow-modals"
        allow="fullscreen; gamepad; autoplay; clipboard-write"
        referrerpolicy="no-referrer"
        @load="onFrameLoad"
      />

      <div v-else class="stage-frame flex flex-col items-center justify-center px-6 text-center">
        <span
          class="relative grid size-16 place-items-center rounded-2xl border border-bee/25 bg-bee/10 text-bee"
        >
          <span
            v-if="meta.tone === 'live'"
            class="absolute inset-0 animate-scan rounded-2xl bg-gradient-to-b from-transparent via-bee/10 to-transparent"
          />
          <AppIcon :name="meta.tone === 'bad' ? 'alert' : 'wand'" :size="28" />
        </span>
        <p class="mt-5 font-display text-lg font-bold text-white">{{ meta.label }}</p>
        <p class="mt-1.5 max-w-sm text-xs leading-5 text-white/45">{{ meta.description }}</p>
      </div>

      <div
        v-if="ready && loading"
        class="absolute inset-0 flex flex-col items-center justify-center bg-[#0b0b09]"
      >
        <div class="flex gap-1.5">
          <span class="size-2 animate-pulse-soft rounded-full bg-bee" />
          <span class="size-2 animate-pulse-soft rounded-full bg-bee [animation-delay:140ms]" />
          <span class="size-2 animate-pulse-soft rounded-full bg-bee [animation-delay:280ms]" />
        </div>
        <p class="mt-4 font-mono text-[10px] uppercase tracking-[0.16em] text-white/40">
          Загружаем игру
        </p>
      </div>
    </div>
  </div>
</template>
