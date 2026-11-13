import { describe, expect, it } from 'vitest'

import { PIPELINE, pipelineFor, stageLabel } from '@/lib/game'

describe('pipelineFor', () => {
  it('keeps the whole pipeline in the normal mode', () => {
    expect(pipelineFor('normal')).toHaveLength(8)
    expect(pipelineFor(undefined)).toEqual(PIPELINE)
  })

  it('drops the review, assets and assembly stages in the fast mode', () => {
    const ids = pipelineFor('fast').map((stage) => stage.id)

    expect(ids).toEqual(['queued', 'analyzing', 'coding', 'checking'])
  })
})

describe('stageLabel', () => {
  it('labels the shared stages the same way in both modes', () => {
    expect(stageLabel('coding')).toBe('Код')
    expect(stageLabel('checking')).toBe('Проверка')
  })
})
