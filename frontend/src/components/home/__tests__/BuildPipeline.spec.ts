import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

import BuildPipeline from '@/components/home/BuildPipeline.vue'

describe('BuildPipeline', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.stubGlobal(
      'matchMedia',
      vi.fn(() => ({
        matches: false,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
      })),
    )
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('moves through stages and celebrates when the last one is done', async () => {
    const wrapper = mount(BuildPipeline, {
      global: {
        stubs: {
          RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
        },
      },
    })

    const stages = wrapper.findAll('.pipe-stage')
    expect(stages).toHaveLength(8)
    expect(stages[0]!.classes()).toContain('is-active')
    expect(wrapper.find('.pipe-status-title').text()).toBe('Очередь')

    await vi.advanceTimersByTimeAsync(900)
    await nextTick()
    expect(wrapper.findAll('.pipe-stage')[1]!.classes()).toContain('is-active')
    expect(wrapper.find('.pipe-status-title').text()).toBe('Разбор идеи')

    await vi.advanceTimersByTimeAsync(6 * 900 + 600)
    await nextTick()
    expect(wrapper.find('.pipe-party').exists()).toBe(true)
    expect(wrapper.find('.pipe-party-title').text()).toBe('Игра готова')
    expect(wrapper.find('.pipe-emoji').attributes('src')).toBeTruthy()
    expect(wrapper.find('.pipe-status').classes()).toContain('is-done')

    wrapper.unmount()
  })
})
