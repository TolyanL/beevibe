import { apiUrl } from './client'
import type {
  ArtifactEventPayload,
  DoneEventPayload,
  ErrorEventPayload,
  StageEventPayload,
} from './types'

export interface GameStreamHandlers {
  onStage: (payload: StageEventPayload, seq: number) => void
  onArtifact: (payload: ArtifactEventPayload, seq: number) => void
  onDone: (payload: DoneEventPayload, seq: number) => void
  onError: (payload: ErrorEventPayload, seq: number) => void
  onConnectionError?: () => void
  onOpen?: () => void
}

function seqOf(event: MessageEvent<string>): number {
  const seq = Number(event.lastEventId)

  return Number.isFinite(seq) ? seq : 0
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
    if (payload) handlers.onStage(payload, seqOf(event))
  })

  source.addEventListener('artifact', (event) => {
    const payload = parse<ArtifactEventPayload>(event as MessageEvent<string>)
    if (payload) handlers.onArtifact(payload, seqOf(event))
  })

  source.addEventListener('done', (event) => {
    const payload = parse<DoneEventPayload>(event as MessageEvent<string>)
    if (payload) handlers.onDone(payload, seqOf(event))
    source.close()
  })

  source.addEventListener('error', (event) => {
    // NOTE: a server-sent error event carries data; a transport error does not.
    if (event instanceof MessageEvent && event.data) {
      const payload = parse<ErrorEventPayload>(event as MessageEvent<string>)
      if (payload) handlers.onError(payload, seqOf(event))
      source.close()
      return
    }

    handlers.onConnectionError?.()
  })

  return () => source.close()
}
