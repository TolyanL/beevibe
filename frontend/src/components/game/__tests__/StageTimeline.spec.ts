import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import StageTimeline from '@/components/game/StageTimeline.vue'

function stageRows(wrapper: ReturnType<typeof mount>): string[] {
  return wrapper.findAll('p.text-\\[12px\\]').map((node) => node.text())
}

describe('StageTimeline', () => {
  it('renders three stages for a fast game', () => {
    const wrapper = mount(StageTimeline, {
      props: { stages: [], status: 'coding', speed: 'fast' },
    })

    expect(stageRows(wrapper)).toEqual(['Очередь', 'Идея и план', 'Код игры', 'Проверка'])
    expect(wrapper.text()).toContain('20%')
  })

  it('renders the full pipeline for a normal game', () => {
    const wrapper = mount(StageTimeline, {
      props: { stages: [], status: 'coding', speed: 'normal' },
    })

    expect(stageRows(wrapper)).toHaveLength(8)
  })
})
