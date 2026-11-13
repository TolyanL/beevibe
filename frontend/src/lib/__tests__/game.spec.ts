import { describe, expect, it } from 'vitest'

import {
  PIPELINE,
  pipelineFor,
  shortGameTitle,
  stageLabel,
  visibleGameDescription,
} from '@/lib/game'

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

describe('visibleGameDescription', () => {
  it('keeps a real description', () => {
    expect(visibleGameDescription('Собирайте кристаллы', 'идея игры', 'Кристаллы')).toBe(
      'Собирайте кристаллы',
    )
  })

  it('falls back to the prompt when the title is custom', () => {
    expect(visibleGameDescription(undefined, 'змейка в неоне', 'Змейка')).toBe('змейка в неоне')
  })

  it('does not repeat the prompt under a prompt-derived title', () => {
    const prompt = 'змейка в неоне, которая собирает звёзды и растёт'

    expect(visibleGameDescription(undefined, prompt, shortGameTitle(undefined, prompt))).toBe('')
  })

  it('does not repeat the title', () => {
    expect(visibleGameDescription('Кристаллы', 'кристаллы', 'Кристаллы')).toBe('')
  })
})
