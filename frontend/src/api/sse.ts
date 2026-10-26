import { apiUrl } from './client'
import type {
  ArtifactEventPayload,
  DoneEventPayload,
  ErrorEventPayload,
  StageEventPayload,
} from './types'

export interface GameStreamHandlers {
  onStage: (payload: StageEventPayload) => void
  onArtifact: (payload: ArtifactEventPayload) => void
  onDone: (payload: DoneEventPayload) => void
  onError: (payload: ErrorEventPayload) => void
  onConnectionError?: () => void
  onOpen?: () => void
}

function parse<T>(event: MessageEvent<string>): T | null {
  try {
    return JSON.parse(event.data) as T
  } catch {
    return null
  }
}

export function openGameStream(gameId: string, handlers: GameStreamHandlers): () => void {
  const source = new EventSource(apiUrl(`/games/${encodeURIComponent(gameId)}/stream`), {
    withCredentials: true,
  })

  source.addEventListener('open', () => handlers.onOpen?.())

  source.addEventListener('stage', (event) => {
    const payload = parse<StageEventPayload>(event as MessageEvent<string>)
    if (payload) handlers.onStage(payload)
  })

  source.addEventListener('artifact', (event) => {
    const payload = parse<ArtifactEventPayload>(event as MessageEvent<string>)
    if (payload) handlers.onArtifact(payload)
  })

  source.addEventListener('done', (event) => {
    const payload = parse<DoneEventPayload>(event as MessageEvent<string>)
    if (payload) handlers.onDone(payload)
    source.close()
  })

  source.addEventListener('error', (event) => {
    // A server-sent `error` event has data; a transport error does not.
    if (event instanceof MessageEvent && event.data) {
      const payload = parse<ErrorEventPayload>(event as MessageEvent<string>)
      if (payload) handlers.onError(payload)
      source.close()
      return
    }

    handlers.onConnectionError?.()
  })

  return () => source.close()
}
