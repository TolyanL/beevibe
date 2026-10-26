import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { getCatalog } from '@/api/catalog'
import { messageFromError } from '@/api/client'
import type { CatalogCard, CatalogSort, Genre } from '@/api/types'

const PAGE_SIZE = 12

export const useCatalogStore = defineStore('catalog', () => {
  const items = ref<CatalogCard[]>([])
  const total = ref(0)
  const search = ref('')
  const sort = ref<CatalogSort>('fresh')
  const genre = ref<Genre | ''>('')
  const loading = ref(false)
  const loadingMore = ref(false)
  const error = ref('')
  const hasMore = ref(true)
  const offset = ref(0)
  const requestId = ref(0)

  const isEmpty = computed(() => !loading.value && items.value.length === 0)

  async function load(reset = true): Promise<void> {
    const current = ++requestId.value

    if (reset) {
      loading.value = true
      offset.value = 0
      hasMore.value = true
      error.value = ''
    } else {
      if (loadingMore.value || !hasMore.value) return
      loadingMore.value = true
      error.value = ''
    }

    try {
      const page = await getCatalog({
        q: search.value || undefined,
        sort: sort.value,
        genre: genre.value,
        limit: PAGE_SIZE,
        offset: reset ? 0 : offset.value,
      })

      if (current !== requestId.value) return

      const next = page.items ?? []
      total.value = page.total ?? 0

      if (reset) items.value = next
      else items.value = [...items.value, ...next]

      offset.value = items.value.length
      hasMore.value = items.value.length < total.value
    } catch (err) {
      if (current === requestId.value) error.value = messageFromError(err)
    } finally {
      if (current === requestId.value) {
        loading.value = false
        loadingMore.value = false
      }
    }
  }

  async function loadMore(): Promise<void> {
    await load(false)
  }

  function setSearch(next: string): void {
    const query = next.trim()
    if (query === search.value) return

    search.value = query
    void load(true)
  }

  function setFilters(next: { sort?: CatalogSort; genre?: Genre | '' }): void {
    if (next.sort !== undefined) sort.value = next.sort
    if (next.genre !== undefined) genre.value = next.genre
    void load(true)
  }

  return {
    items,
    total,
    search,
    sort,
    genre,
    loading,
    loadingMore,
    error,
    hasMore,
    isEmpty,
    load,
    loadMore,
    setSearch,
    setFilters,
  }
})
