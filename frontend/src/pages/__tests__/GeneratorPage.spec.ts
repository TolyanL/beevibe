import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import GeneratorPage from '@/pages/GeneratorPage.vue'

const mocks = vi.hoisted(() => ({
  push: vi.fn(),
  routeQuery: {} as Record<string, string>,
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: mocks.routeQuery }),
  useRouter: () => ({ push: mocks.push }),
  RouterLink: { name: 'RouterLink', template: '<a><slot /></a>' },
}))

vi.mock('openvue/usetoast', () => ({
  useToast: () => ({ add: vi.fn() }),
}))

function buttonByText(wrapper: ReturnType<typeof mount>, text: string) {
  const button = wrapper.findAll('button').find((node) => node.text().includes(text))
  if (!button) throw new Error(`button ${text} not found`)

  return button
}

describe('GeneratorPage speed switch', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mocks.routeQuery = {}
  })

  it('offers both modes with their own price', () => {
    const wrapper = mount(GeneratorPage, { shallow: true })

    expect(buttonByText(wrapper, 'Быстро').text()).toContain('240')
    expect(buttonByText(wrapper, 'Обычный').text()).toContain('320')
    expect(wrapper.text()).toContain('≈ 4–6 минут')
    expect(wrapper.text()).toContain('≈ 15–20 минут')
  })

  it('switches the price of the submit row to the fast one', async () => {
    const wrapper = mount(GeneratorPage, { shallow: true })

    expect(wrapper.text()).toContain('Стоимость: 320K')

    await buttonByText(wrapper, 'Быстро').trigger('click')

    expect(wrapper.text()).toContain('Стоимость: 240K')
    expect(wrapper.text()).toContain('Генерация · 240K')
  })

  it('prefills the prompt from the landing idea query', () => {
    mocks.routeQuery = { idea: 'Киберпанк-кликер про кота-хакера' }

    const wrapper = mount(GeneratorPage, { shallow: true })

    expect(wrapper.text()).toContain('32 / 2000')
  })
})
