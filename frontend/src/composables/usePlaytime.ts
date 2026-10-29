import { onBeforeUnmount, ref, watch } from 'vue'
import { recordPlaytime } from '@/api/games'
import { useAuthStore } from '@/stores/auth'

// NOTE: the playtime endpoint needs a session, so guests only keep a local timer.
export function usePlaytime(gameId: () => string | undefined, active: () => boolean) {
  const auth = useAuthStore()
  const seconds = ref(0)
  const savedSeconds = ref(0)

  let timer: ReturnType<typeof setInterval> | null = null
  let flushTimer: ReturnType<typeof setInterval> | null = null
  let pageVisible = typeof document === 'undefined' ? true : !document.hidden

  function onVisibility(): void {
    pageVisible = !document.hidden
  }

  async function flush(): Promise<void> {
    const pending = seconds.value - savedSeconds.value
    const currentGame = gameId()

    if (!currentGame || !auth.isAuthenticated || pending < 10) return

    savedSeconds.value = seconds.value

    try {
      await recordPlaytime(currentGame, pending)
    } catch {
      savedSeconds.value -= pending
    }
  }

  function start(): void {
    if (timer) return
    document.addEventListener('visibilitychange', onVisibility)

    timer = setInterval(() => {
      if (pageVisible && active()) seconds.value += 1
    }, 1000)

    flushTimer = setInterval(() => void flush(), 60_000)
  }

  function stop(): void {
    if (timer) clearInterval(timer)
    if (flushTimer) clearInterval(flushTimer)
    timer = null
    flushTimer = null
    document.removeEventListener('visibilitychange', onVisibility)
    void flush()
  }

  watch(
    active,
    (value) => {
      if (value) start()
      else stop()
    },
    { immediate: true },
  )

  onBeforeUnmount(stop)

  return { seconds, flush }
}
