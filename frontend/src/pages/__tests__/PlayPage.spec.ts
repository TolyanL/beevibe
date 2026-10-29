import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia, type Pinia } from 'pinia'
import { defineComponent, h } from 'vue'

import PlayPage from '@/pages/PlayPage.vue'
import { ApiError } from '@/api/client'
import { getCatalog, getGamePublicCard } from '@/api/catalog'
import { rateGame } from '@/api/games'
import { useAuthStore } from '@/stores/auth'
import type { GamePublicCard } from '@/api/types'

vi.mock('@/api/catalog', () => ({
  getGamePublicCard: vi.fn(),
  getCatalog: vi.fn(async () => ({ items: [], total: 0 })),
}))

vi.mock('@/api/games', () => ({
  rateGame: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { id: 'prostoy-2d-platformer-da2f' } }),
  RouterLink: {
    name: 'RouterLink',
    props: { to: { type: [String, Object], default: '' } },
    setup:
      (_props: unknown, { slots }: { slots: Record<string, () => unknown> }) =>
      () =>
        h('a', slots.default ? (slots.default() as never) : []),
  },
}))

vi.mock('openvue/usetoast', () => ({
  useToast: () => ({ add: vi.fn() }),
}))

const RatingStub = defineComponent({
  name: 'RatingStub',
  props: { modelValue: { type: Number, default: 0 }, disabled: { type: Boolean, default: false } },
  emits: ['update:modelValue'],
  template: '<div class="rating-stub" />',
})

const card: GamePublicCard = {
  id: 'da2f3fb9-52d2-4183-9b47-4318aea16924',
  slug: 'prostoy-2d-platformer-da2f',
  prompt: 'простой 2D платформер',
  title: 'Платформер',
  description: '',
  genre: 'platformer',
  style: 'pixel',
  map_size: 'medium',
  created_at: '2026-10-27T18:00:00Z',
  author: 'Автор',
  rating: { average: 0, total: 0, counts: [0, 0, 0, 0, 0] },
  playtime: { sessions: 0, seconds: 0, hours: 0 },
  my_rating: 0,
  is_owner: false,
}

function mountPage(pinia: Pinia = createPinia()) {
  return mount(PlayPage, {
    global: {
      plugins: [pinia],
      stubs: { Rating: RatingStub },
    },
  })
}

function signIn(): Pinia {
  const pinia = createPinia()

  setActivePinia(pinia)
  useAuthStore().user = { id: 'user-1', email: 'player@example.com', display_name: '', balance: 0 }

  return pinia
}

beforeEach(() => {
  vi.mocked(getGamePublicCard).mockResolvedValue(structuredClone(card))
  vi.mocked(getCatalog).mockResolvedValue({ items: [], total: 0 })
  vi.mocked(rateGame).mockResolvedValue({
    rating: 5,
    aggregate: { average: 5, total: 1, counts: [0, 0, 0, 0, 1] },
  })
})

describe('PlayPage rating', () => {
  it('posts the rating under the slug, never under the database id', async () => {
    const pinia = signIn()

    const wrapper = mountPage(pinia)
    await flushPromises()

    wrapper.findComponent(RatingStub).vm.$emit('update:modelValue', 5)
    await flushPromises()

    expect(rateGame).toHaveBeenCalledTimes(1)
    expect(vi.mocked(rateGame).mock.calls[0]).toEqual(['prostoy-2d-platformer-da2f', 5])
  })

  it('shows the aggregate returned by the server', async () => {
    const pinia = signIn()

    const wrapper = mountPage(pinia)
    await flushPromises()

    wrapper.findComponent(RatingStub).vm.$emit('update:modelValue', 5)
    await flushPromises()

    expect(wrapper.text()).toContain('5,0')
  })

  it('hides the rating control from the author', async () => {
    const pinia = signIn()
    vi.mocked(getGamePublicCard).mockResolvedValue({ ...structuredClone(card), is_owner: true })

    const wrapper = mountPage(pinia)
    await flushPromises()

    expect(wrapper.find('[data-test="rating-input"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('Это ваша игра')
  })
})

describe('PlayPage states', () => {
  it('renders the not-found state instead of the frame when the game is gone', async () => {
    vi.mocked(getGamePublicCard).mockRejectedValue(
      new ApiError(404, 'not_found', 'request failed: not found'),
    )

    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.text()).toContain('Игра не найдена')
    expect(wrapper.find('iframe').exists()).toBe(false)
  })

  it('offers a retry when the card failed to load for another reason', async () => {
    vi.mocked(getGamePublicCard).mockRejectedValue(
      new ApiError(500, 'internal', 'Сервер не смог выполнить запрос.'),
    )

    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.text()).toContain('Не удалось загрузить игру')
    expect(wrapper.text()).toContain('Повторить')
    expect(wrapper.find('iframe').exists()).toBe(false)
  })
})
