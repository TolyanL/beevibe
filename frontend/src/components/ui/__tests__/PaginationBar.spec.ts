import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import PaginationBar from '../PaginationBar.vue'

describe('PaginationBar', () => {
  it('is hidden when there is nothing to page', () => {
    const wrapper = mount(PaginationBar, { props: { page: 1, pageSize: 6, total: 0 } })

    expect(wrapper.find('nav').exists()).toBe(false)
  })

  it('shows the visible range and emits the next page', async () => {
    const wrapper = mount(PaginationBar, {
      props: { page: 1, pageSize: 6, total: 12, label: 'игр' },
    })

    expect(wrapper.text()).toContain('1–6 из 12 игр')

    await wrapper.get('button[aria-label="Следующая страница"]').trigger('click')

    expect(wrapper.emitted('update:page')?.[0]).toEqual([2])
  })

  it('marks the current page and collapses long ranges', () => {
    const wrapper = mount(PaginationBar, { props: { page: 9, pageSize: 10, total: 200 } })

    expect(wrapper.get('[aria-current="page"]').text()).toBe('9')
    expect(wrapper.text()).toContain('…')
  })

  it('does not go beyond the last page', () => {
    const wrapper = mount(PaginationBar, { props: { page: 4, pageSize: 6, total: 24 } })

    expect(
      wrapper.get('button[aria-label="Следующая страница"]').attributes('disabled'),
    ).toBeDefined()
  })
})
