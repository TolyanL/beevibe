<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getCatalog } from '@/api/catalog'
import type { CatalogCard } from '@/api/types'
import { formatNumber, formatRelative, plural } from '@/lib/format'
import { shortGameTitle } from '@/lib/game'
import { useAuthStore } from '@/stores/auth'
import AppIcon from '@/components/ui/AppIcon.vue'
import GameCard from '@/components/catalog/GameCard.vue'
import BuildPipeline from '@/components/home/BuildPipeline.vue'

const auth = useAuthStore()
const fresh = ref<CatalogCard[]>([])
const top = ref<CatalogCard[]>([])
const total = ref(0)
const loading = ref(true)

const freshCards = computed(() => fresh.value.slice(0, 3))

const lastGame = computed(() => {
  const games = auth.profile?.games ?? []

  return games.find((item) => item.status === 'ready') ?? games[0] ?? null
})

const activeGame = computed(() => {
  const games = auth.profile?.games ?? []
  const busy = [
    'queued',
    'analyzing',
    'planning',
    'coding',
    'assets',
    'reviewing',
    'assembling',
    'checking',
    'patching',
  ]

  return games.find((item) => busy.includes(item.status)) ?? null
})

const startLink = computed(() => (auth.isAuthenticated ? { name: 'create' } : { name: 'register' }))

const catalogCaption = computed(() =>
  total.value > 0
    ? `${formatNumber(total.value)} ${plural(total.value, 'игра', 'игры', 'игр')} в каталоге`
    : 'Пока ни одной игры',
)

const steps = [
  {
    num: '01',
    icon: 'wand',
    title: 'Идея',
    text: 'Одно предложение: герой, механика, настроение.',
  },
  {
    num: '02',
    icon: 'code',
    title: 'План и код',
    text: 'Модель придумывает правила и пишет игру одним файлом.',
  },
  {
    num: '03',
    icon: 'layers',
    title: 'Сборка',
    text: 'Всё складывается в один html-файл.',
  },
  {
    num: '04',
    icon: 'shield',
    title: 'Проверка',
    text: 'Быстрый прогон перед публикацией.',
  },
] as const

const ideas = [
  {
    label: 'лягушка под дождём',
    prompt:
      'Мультяшный платформер: лягушка прыгает по плавающим листьям и собирает светлячков под дождём.',
  },
  {
    label: 'кот-хакер',
    prompt:
      'Киберпанк-кликер: кот-хакер взламывает корпорации, каждая клика приносит кредиты и апгрейды.',
  },
  {
    label: 'змейка в космосе',
    prompt: 'Неоновая змейка в невесомости: собирай звёзды, уворачивайся от астероидов, ускоряйся.',
  },
] as const

const facts = [
  { value: '4', label: 'шага в быстрой сборке' },
  { value: '8', label: 'этапов в обычной сборке' },
  { value: '1', label: 'html-файл' },
  { value: '1 000 000', label: 'токенов на старт' },
] as const

const tickerPhrases = [
  'одна фраза',
  'восемь этапов',
  'ноль установок',
  'правки словами',
  'играй сразу',
] as const

const priceRows = computed(() => [
  {
    label: 'Быстрая сборка',
    hint: 'четыре шага, без ревью',
    value: auth.prices.generate_fast,
    hot: true,
  },
  {
    label: 'Обычная сборка',
    hint: 'восемь этапов, с ревью',
    value: auth.prices.generate,
    hot: false,
  },
  {
    label: 'Правка',
    hint: 'изменения словами',
    value: auth.prices.patch,
    hot: false,
  },
  {
    label: 'Название и описание',
    hint: 'придумает модель',
    value: auth.prices.meta,
    hot: false,
  },
])

const faq = [
  {
    q: 'Это действительно бесплатно?',
    a: 'При регистрации начисляется 1 000 000 токенов. Обычная сборка стоит 320 000, быстрая — 240 000, правки — от 40 000. Дальше токены можно докупить в профиле.',
  },
  {
    q: 'Сколько ждать игру?',
    a: 'Обычный режим проходит восемь этапов: очередь, разбор идеи, план, код, графику, ревью, сборку и проверку. Быстрый — четыре шага. Точное время зависит от очереди.',
  },
  {
    q: 'Какие игры получаются?',
    a: 'Небольшие браузерные игры: аркады, платформеры, змейки, кликеры и эксперименты. Каждая — один HTML-файл, который открывается в браузере без установки.',
  },
  {
    q: 'Как исправить игру?',
    a: 'Скажите словами, что поменять: скорость, сложность, персонажа, правила. Соберётся новая версия, а старые останутся в профиле — можно откатиться.',
  },
  {
    q: 'Где живут игры?',
    a: 'Готовая игра открывается по ссылке. В каталоге она появляется, если открыть её для всех; личные игры остаются в профиле.',
  },
  {
    q: 'Нужен ли аккаунт?',
    a: 'Чтобы играть в каталоге — нет. Чтобы собрать свою игру и получить токены — да, нужен аккаунт.',
  },
] as const

onMounted(async () => {
  if (auth.isAuthenticated && !auth.profile) void auth.refreshProfile()

  try {
    const [freshPage, topPage] = await Promise.all([
      getCatalog({ sort: 'fresh', limit: 6 }),
      getCatalog({ sort: 'rating', limit: 3 }),
    ])

    fresh.value = freshPage.items ?? []
    total.value = freshPage.total ?? fresh.value.length
    top.value = (topPage.items ?? []).filter(
      (card) => (card.rating?.total ?? 0) > 0 && !fresh.value.some((item) => item.id === card.id),
    )
  } catch {
    fresh.value = []
    top.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="home-page">
    <section class="hero" aria-labelledby="hero-title">
      <div class="hero-ornament" aria-hidden="true">
        <span class="hero-sun"></span>
        <span class="hero-dots"></span>
      </div>

      <div class="app-container">
        <div class="hero-grid">
          <div class="hero-intro">
            <p class="hero-kicker">
              <span class="hero-kicker-dot" aria-hidden="true"></span>
              BeeVibe · игры без установки
            </p>

            <h1 id="hero-title" class="hero-title" aria-label="Игра из одной фразы">
              <span>Игра</span>
              <span>из одной</span>
              <span class="hero-title-mark">фразы</span>
            </h1>
          </div>

          <div class="hero-copy">
            <p class="hero-lead">
              Нейросети придумают правила, напишут код и нарисуют графику. Одна фраза — и можно
              играть. Правки — словами.
            </p>

            <div class="hero-actions">
              <RouterLink :to="{ name: 'create' }" class="bb-btn bb-btn--ink">
                <AppIcon name="spark" :size="16" />
                Создать игру
                <AppIcon name="arrow-right" :size="16" />
              </RouterLink>
              <RouterLink :to="{ name: 'catalog' }" class="bb-btn bb-btn--line">
                Смотреть каталог
                <AppIcon name="arrow-right" :size="16" />
              </RouterLink>
            </div>

            <div v-if="activeGame" class="hero-status">
              <span class="hero-status-dot" aria-hidden="true"></span>
              <span class="hero-status-label">Идёт сборка</span>
              <RouterLink
                :to="{ name: 'game', params: { id: activeGame.slug } }"
                class="hero-status-link"
              >
                {{ shortGameTitle(activeGame.title, activeGame.prompt) }}
              </RouterLink>
            </div>
            <div v-else-if="lastGame" class="hero-status hero-status--quiet">
              <span class="hero-status-label">Последняя игра</span>
              <RouterLink
                :to="{ name: 'game', params: { id: lastGame.slug } }"
                class="hero-status-link"
              >
                {{ shortGameTitle(lastGame.title, lastGame.prompt) }}
              </RouterLink>
            </div>

            <div class="hero-ideas">
              <span class="hero-ideas-label">или начните с:</span>
              <RouterLink
                v-for="idea in ideas"
                :key="idea.label"
                :to="{ name: 'create', query: { idea: idea.prompt } }"
                class="hero-idea"
              >
                {{ idea.label }}
              </RouterLink>
            </div>
          </div>
        </div>

        <div class="hero-facts">
          <div v-for="fact in facts" :key="fact.label" class="hero-fact">
            <span class="hero-fact-value">{{ fact.value }}</span>
            <span class="hero-fact-label">{{ fact.label }}</span>
          </div>
        </div>
      </div>
    </section>

    <div class="ticker" aria-hidden="true">
      <div class="ticker-track">
        <div v-for="copy in 2" :key="copy" class="ticker-group">
          <span v-for="phrase in tickerPhrases" :key="`${copy}-${phrase}`" class="ticker-item">
            {{ phrase }}
            <span class="ticker-star">✦</span>
          </span>
        </div>
      </div>
    </div>

    <section class="how" aria-labelledby="how-title">
      <div class="app-container">
        <div class="section-head">
          <div>
            <p class="bb-eyebrow bb-eyebrow--yellow">два режима</p>
            <h2 id="how-title" class="bb-title bb-title--light">
              Быстро — 4 шага<br />Обычно — 8 этапов
            </h2>
            <p class="how-lead">
              Быстрая — за несколько минут и на 25% дешевле. Обычная — с графикой, ревью и
              проверкой.
            </p>
          </div>
          <RouterLink :to="{ name: 'create' }" class="bb-btn bb-btn--yellow">
            Попробовать
            <AppIcon name="arrow-right" :size="16" />
          </RouterLink>
        </div>

        <div class="how-modes">
          <div class="how-mode">
            <div class="how-mode-head">
              <span class="how-mode-kicker">быстрая сборка</span>
              <span class="how-mode-meta">4 шага · ≈ 4–6 мин</span>
            </div>
            <ol class="how-fast-grid">
              <li v-for="(step, index) in steps" :key="step.title" class="how-card">
                <div class="how-card-head">
                  <span class="how-card-num">{{ step.num }}</span>
                  <span class="how-card-icon">
                    <AppIcon :name="step.icon" :size="18" />
                  </span>
                </div>
                <h3 class="how-card-title">{{ step.title }}</h3>
                <p class="how-card-text">{{ step.text }}</p>
                <span class="how-card-progress" aria-hidden="true">
                  <i :style="{ width: `${(index + 1) * 25}%` }"></i>
                </span>
              </li>
            </ol>
          </div>

          <div class="how-mode">
            <div class="how-mode-head">
              <span class="how-mode-kicker">обычная сборка</span>
              <span class="how-mode-meta">8 этапов</span>
            </div>
            <div class="how-pipeline-panel">
              <BuildPipeline />
            </div>
          </div>
        </div>

        <p class="how-note">
          <AppIcon name="bolt" :size="15" />
          В обоих режимах игру можно править словами — версии сохраняются.
          <RouterLink :to="{ name: 'create' }" class="how-note-link">Собрать игру</RouterLink>
        </p>
      </div>
    </section>

    <section class="catalog" aria-labelledby="catalog-title">
      <div class="app-container">
        <div class="section-head">
          <div>
            <p class="bb-eyebrow">каталог</p>
            <h2 id="catalog-title" class="bb-title">Что уже собрали</h2>
            <p class="section-sub">{{ catalogCaption }}</p>
          </div>
          <RouterLink :to="{ name: 'catalog' }" class="bb-btn bb-btn--line">
            Весь каталог
            <AppIcon name="arrow-right" :size="16" />
          </RouterLink>
        </div>

        <div v-if="loading" class="catalog-grid">
          <div v-for="index in 3" :key="index" class="catalog-skeleton">
            <div class="catalog-skeleton-thumb"></div>
            <div class="catalog-skeleton-body">
              <span></span>
              <span></span>
            </div>
          </div>
        </div>

        <div v-else-if="freshCards.length" class="catalog-grid">
          <GameCard v-for="card in freshCards" :key="card.id" :card="card" />
          <RouterLink
            v-if="freshCards.length < 3"
            :to="{ name: 'create' }"
            class="catalog-slot"
            :class="{ 'catalog-slot--wide': freshCards.length === 1 }"
          >
            <div class="catalog-slot-top">
              <span class="catalog-slot-mark" aria-hidden="true">+</span>
              <span class="catalog-slot-kicker">свободное место</span>
            </div>
            <div>
              <p class="catalog-slot-title">Здесь может быть ваша игра</p>
              <p class="catalog-slot-text">Одна фраза — и карточка появится в каталоге.</p>
            </div>
            <span class="catalog-slot-go">
              Собрать игру
              <AppIcon name="arrow-right" :size="14" />
            </span>
          </RouterLink>
        </div>

        <div v-else class="catalog-empty">
          <div>
            <p class="catalog-empty-title">Пока пусто — но это ненадолго</p>
            <p class="catalog-empty-text">
              Соберите первую игру: она появится здесь, если открыть её для всех.
            </p>
          </div>
          <RouterLink :to="{ name: 'create' }" class="bb-btn bb-btn--ink">
            <AppIcon name="spark" :size="16" />
            Создать игру
          </RouterLink>
        </div>

        <div v-if="top.length" class="catalog-top">
          <div class="catalog-top-head">
            <h3 class="catalog-top-title">Выбор игроков</h3>
            <RouterLink
              :to="{ name: 'catalog', query: { sort: 'rating' } }"
              class="catalog-top-link"
            >
              По рейтингу
              <AppIcon name="arrow-right" :size="14" />
            </RouterLink>
          </div>
          <div class="catalog-grid">
            <GameCard v-for="card in top" :key="card.id" :card="card" />
          </div>
        </div>

        <p v-if="fresh.length" class="catalog-updated">
          Последняя сборка · {{ formatRelative(fresh[0]!.created_at) }}
        </p>
      </div>
    </section>

    <section class="offer" aria-labelledby="offer-title">
      <div class="app-container offer-grid">
        <div class="offer-copy">
          <p class="bb-eyebrow bb-eyebrow--ink">на старт</p>
          <h2 id="offer-title" class="bb-title bb-title--ink">Первая игра<br />за наш счёт</h2>
          <p class="offer-lead">
            При регистрации — миллион токенов. Хватит на три обычные сборки или четыре быстрые.
          </p>

          <div class="offer-amount">
            <span class="offer-amount-value">1 000 000</span>
            <span class="offer-amount-unit">токенов</span>
          </div>

          <RouterLink :to="startLink" class="bb-btn bb-btn--ink bb-btn--lg">
            {{ auth.isAuthenticated ? 'Создать игру' : 'Забрать миллион' }}
            <AppIcon name="arrow-right" :size="16" />
          </RouterLink>

          <p v-if="auth.isAuthenticated" class="offer-note">
            На балансе: {{ formatNumber(auth.balance) }} токенов
          </p>
          <p v-else class="offer-note">Токены начисляются сразу после регистрации</p>
        </div>

        <div class="price-card">
          <div class="price-head">
            <span class="price-head-title">Прайс</span>
            <span class="price-head-unit">в токенах</span>
          </div>
          <ul class="price-list">
            <li v-for="row in priceRows" :key="row.label" class="price-row">
              <div class="price-row-info">
                <span class="price-row-label">
                  {{ row.label }}
                  <span v-if="row.hot" class="price-row-tag">−25%</span>
                </span>
                <span class="price-row-hint">{{ row.hint }}</span>
              </div>
              <span class="price-row-value">{{ formatNumber(row.value) }}</span>
            </li>
          </ul>
          <p class="price-foot">
            <AppIcon name="bolt" :size="14" />
            Миллиона на старте хватает примерно на три обычные сборки
          </p>
        </div>
      </div>
    </section>

    <section class="faq" aria-labelledby="faq-title">
      <div class="app-container faq-grid">
        <div class="faq-intro">
          <p class="bb-eyebrow">вопросы</p>
          <h2 id="faq-title" class="bb-title">Коротко<br />о главном</h2>
          <p class="faq-lead">
            Если чего-то не хватает, загляните в каталог: там видно, как игры выглядят вживую.
          </p>
          <RouterLink :to="{ name: 'catalog' }" class="faq-link">
            Открыть каталог
            <AppIcon name="arrow-right" :size="14" />
          </RouterLink>
        </div>

        <div class="faq-list">
          <details v-for="item in faq" :key="item.q" class="faq-item">
            <summary class="faq-summary">
              <span>{{ item.q }}</span>
              <span class="faq-plus" aria-hidden="true"></span>
            </summary>
            <p class="faq-answer">{{ item.a }}</p>
          </details>
        </div>
      </div>
    </section>

    <section class="final" aria-labelledby="final-title">
      <div class="final-ornament" aria-hidden="true"></div>
      <div class="app-container final-inner">
        <div>
          <p class="bb-eyebrow bb-eyebrow--yellow">поехали</p>
          <h2 id="final-title" class="final-title">Ну что,<br /><span>соберём?</span></h2>
          <p class="final-lead">Одна фраза — и у вас своя игра. Первый миллион токенов уже ждёт.</p>
        </div>

        <div class="final-actions">
          <RouterLink :to="startLink" class="bb-btn bb-btn--yellow bb-btn--lg">
            {{ auth.isAuthenticated ? 'Создать игру' : 'Начать бесплатно' }}
            <AppIcon name="arrow-right" :size="16" />
          </RouterLink>
          <RouterLink :to="{ name: 'catalog' }" class="final-link">
            Или посмотреть, что уже собрали
          </RouterLink>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.home-page {
  overflow-x: clip;
}

.bb-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
  font-family: var(--font-mono);
  font-size: 0.64rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  line-height: 1;
  text-transform: uppercase;
  color: var(--bv-muted);
}

.bb-eyebrow::before {
  width: 24px;
  height: 5px;
  border-radius: 1px;
  content: '';
  background: repeating-linear-gradient(90deg, #141414 0 4px, #ffcc00 4px 8px);
}

.bb-eyebrow--yellow {
  color: #ffcc00;
}

.bb-eyebrow--yellow::before {
  background: repeating-linear-gradient(90deg, #ffcc00 0 4px, #141414 4px 8px);
}

.bb-eyebrow--ink {
  color: rgba(20, 20, 20, 0.62);
}

.bb-title {
  margin-top: 0.75rem;
  font-family: var(--font-display);
  font-size: clamp(1.9rem, 4vw, 3.4rem);
  font-weight: 800;
  letter-spacing: -0.055em;
  line-height: 0.92;
  text-transform: uppercase;
  color: var(--bv-text);
}

.bb-title--light {
  color: #f6f4e8;
}

.bb-title--ink {
  color: #141414;
}

.bb-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 3rem;
  padding: 0.8rem 1.35rem;
  border: 2px solid #141414;
  border-radius: 999px;
  font-family: var(--font-sans);
  font-size: 0.84rem;
  font-weight: 800;
  line-height: 1;
  text-decoration: none;
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease,
    background-color 0.18s ease,
    color 0.18s ease,
    border-color 0.18s ease;
}

.bb-btn--ink {
  background: #141414;
  color: #ffcc00;
  box-shadow: 0 4px 0 rgba(20, 20, 20, 0.24);
}

.bb-btn--ink:hover {
  transform: translateY(-2px);
  background: #000;
  box-shadow: 0 6px 0 rgba(20, 20, 20, 0.2);
}

.bb-btn--line {
  background: transparent;
  color: #141414;
}

.bb-btn--line:hover {
  transform: translateY(-2px);
  background: #141414;
  color: #ffcc00;
}

.bb-btn--yellow {
  background: #ffcc00;
  border-color: #ffcc00;
  color: #141414;
  box-shadow: 0 4px 0 rgba(0, 0, 0, 0.35);
}

.bb-btn--yellow:hover {
  transform: translateY(-2px);
  background: #141414;
  border-color: #ffcc00;
  color: #ffcc00;
}

.bb-btn--lg {
  min-height: 3.5rem;
  padding: 1rem 1.7rem;
  font-size: 0.95rem;
}

.hero {
  position: relative;
  overflow: hidden;
  padding-top: 3.2rem;
  background: #ffcc00;
  color: #141414;
}

.hero::after {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 10px;
  content: '';
  background: repeating-linear-gradient(135deg, #141414 0 14px, #ffcc00 14px 28px);
}

.hero-ornament {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.hero-sun {
  position: absolute;
  top: -260px;
  right: -190px;
  width: 640px;
  height: 640px;
  border-radius: 999px;
  background: repeating-linear-gradient(135deg, #141414 0 17px, transparent 17px 34px);
  opacity: 0.08;
}

.hero-dots {
  position: absolute;
  bottom: 4rem;
  left: -2rem;
  width: 240px;
  height: 240px;
  background-image: radial-gradient(rgba(20, 20, 20, 0.24) 1px, transparent 1px);
  background-size: 15px 15px;
  mask-image: radial-gradient(circle at 50% 50%, black, transparent 72%);
}

.hero-grid {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  align-items: center;
  gap: 2.5rem;
}

.hero-intro,
.hero-copy {
  min-width: 0;
}

.hero-copy {
  align-self: end;
}

.hero-kicker {
  display: inline-flex;
  align-items: center;
  gap: 0.55rem;
  padding: 0.45rem 0.8rem;
  border: 2px solid #141414;
  border-radius: 999px;
  background: #fffdf7;
  font-family: var(--font-mono);
  font-size: 0.6rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  line-height: 1;
  text-transform: uppercase;
  animation: bb-rise 0.5s 0.04s both;
}

.hero-kicker-dot {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: #141414;
}

.hero-title {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  margin-top: 1.35rem;
  font-family: var(--font-display);
  font-size: clamp(2.7rem, 6.2vw, 6.1rem);
  font-weight: 800;
  letter-spacing: -0.06em;
  line-height: 0.88;
  text-transform: uppercase;
  animation: bb-rise 0.62s 0.1s both;
}

.hero-title-mark {
  display: inline-block;
  margin-top: 0.08em;
  margin-left: -0.04em;
  padding: 0.02em 0.16em 0.06em;
  background: #141414;
  color: #ffcc00;
  box-shadow: 6px 6px 0 rgba(20, 20, 20, 0.18);
  transform: rotate(-1.2deg);
}

.hero-lead {
  max-width: 34rem;
  margin-top: 1.45rem;
  font-size: 0.98rem;
  font-weight: 500;
  line-height: 1.65;
  color: rgba(20, 20, 20, 0.76);
  animation: bb-rise 0.62s 0.16s both;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.7rem;
  margin-top: 1.7rem;
  animation: bb-rise 0.62s 0.22s both;
}

.hero-status {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  max-width: 100%;
  margin-top: 1.2rem;
  padding: 0.55rem 0.85rem;
  border: 2px solid #141414;
  border-radius: 999px;
  background: rgba(255, 253, 247, 0.62);
  font-size: 0.7rem;
  font-weight: 700;
}

.hero-status--quiet {
  background: transparent;
}

.hero-status-dot {
  flex-shrink: 0;
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: #141414;
  animation: bb-pulse 1.6s ease-in-out infinite;
}

.hero-status-label {
  flex-shrink: 0;
  font-family: var(--font-mono);
  font-size: 0.56rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: rgba(20, 20, 20, 0.7);
}

.hero-status-link {
  min-width: 0;
  overflow: hidden;
  color: #141414;
  font-weight: 800;
  text-decoration: underline;
  text-underline-offset: 3px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hero-status-link:hover {
  color: #6b5300;
}

.hero-ideas {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem;
  margin-top: 1.1rem;
  animation: bb-rise 0.62s 0.28s both;
}

.hero-ideas-label {
  margin-right: 0.15rem;
  font-family: var(--font-mono);
  font-size: 0.56rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: rgba(20, 20, 20, 0.7);
}

.hero-idea {
  padding: 0.38rem 0.7rem;
  border: 1.5px solid rgba(20, 20, 20, 0.32);
  border-radius: 999px;
  color: #141414;
  font-size: 0.7rem;
  font-weight: 700;
  text-decoration: none;
  transition:
    background-color 0.15s ease,
    border-color 0.15s ease,
    color 0.15s ease;
}

.hero-idea:hover {
  border-color: #141414;
  background: #141414;
  color: #ffcc00;
}

.hero-facts {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin-top: 2.6rem;
  border-top: 2px solid #141414;
  animation: bb-rise 0.62s 0.32s both;
}

.hero-fact {
  padding: 1rem 0.4rem 1.3rem;
  border-bottom: 1px solid rgba(20, 20, 20, 0.18);
}

.hero-fact-value {
  display: block;
  font-family: var(--font-display);
  font-size: clamp(1.3rem, 2.2vw, 1.9rem);
  font-weight: 800;
  letter-spacing: -0.04em;
  line-height: 1;
  white-space: nowrap;
}

.hero-fact-label {
  display: block;
  margin-top: 0.35rem;
  font-family: var(--font-mono);
  font-size: 0.58rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: rgba(20, 20, 20, 0.7);
}

.ticker {
  overflow: hidden;
  padding: 0.85rem 0;
  border-bottom: 2px solid #141414;
  background: #141414;
  color: #ffcc00;
}

.ticker-track {
  display: flex;
  width: max-content;
  animation: bb-marquee 30s linear infinite;
}

.ticker-group {
  display: flex;
  align-items: center;
}

.ticker-item {
  display: inline-flex;
  align-items: center;
  gap: 1rem;
  padding-right: 1rem;
  font-family: var(--font-display);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  white-space: nowrap;
}

.ticker-star {
  font-size: 0.85rem;
}

.section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1.25rem;
  margin-bottom: 1.9rem;
}

.section-sub {
  margin-top: 0.65rem;
  font-family: var(--font-mono);
  font-size: 0.62rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--bv-muted);
}

.how {
  position: relative;
  overflow: hidden;
  padding: 4rem 0 4.5rem;
  background: #0f0f0e;
  color: #f6f4e8;
}

.how::before {
  position: absolute;
  top: -220px;
  right: -160px;
  width: 520px;
  height: 520px;
  border-radius: 999px;
  content: '';
  background: repeating-linear-gradient(135deg, #ffcc00 0 14px, transparent 14px 28px);
  opacity: 0.07;
}

.how-lead {
  max-width: 38rem;
  margin-top: 1rem;
  font-size: 0.9rem;
  line-height: 1.65;
  color: rgba(246, 244, 232, 0.6);
}

.how-modes {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1.75rem;
}

.how-mode {
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
  min-width: 0;
}

.how-mode-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  min-height: 1rem;
}

.how-mode-kicker {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-family: var(--font-mono);
  font-size: 0.6rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: #ffcc00;
}

.how-mode-kicker::before {
  width: 20px;
  height: 4px;
  content: '';
  background: repeating-linear-gradient(90deg, #ffcc00 0 4px, transparent 4px 8px);
}

.how-mode-meta {
  flex-shrink: 0;
  font-family: var(--font-mono);
  font-size: 0.56rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: rgba(246, 244, 232, 0.45);
}

.how-fast-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.85rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.how-pipeline-panel {
  overflow: hidden;
  padding: 1.1rem;
  border: 2px solid #141414;
  border-radius: 1.75rem;
  background: #ffcc00;
  box-shadow: 0 24px 48px -34px rgba(0, 0, 0, 0.9);
}

.how-card {
  position: relative;
  display: flex;
  flex-direction: column;
  min-height: 12.5rem;
  padding: 1.35rem;
  border: 1px solid rgba(246, 244, 232, 0.16);
  border-radius: 1.5rem;
  background: #151513;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    transform 0.2s ease;
}

.how-card:hover {
  border-color: #ffcc00;
  background: #1a1a17;
  transform: translateY(-4px);
}

.how-card-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.how-card-num {
  font-family: var(--font-display);
  font-size: 3.2rem;
  font-weight: 800;
  letter-spacing: -0.06em;
  line-height: 0.8;
  color: transparent;
  -webkit-text-stroke: 1.5px #ffcc00;
}

.how-card-icon {
  display: grid;
  flex-shrink: 0;
  place-items: center;
  width: 2.4rem;
  height: 2.4rem;
  border: 1px solid rgba(255, 204, 0, 0.4);
  border-radius: 999px;
  color: #ffcc00;
}

.how-card-title {
  margin-top: auto;
  padding-top: 2rem;
  font-family: var(--font-display);
  font-size: 1.1rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  text-transform: uppercase;
}

.how-card-text {
  margin-top: 0.5rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: rgba(246, 244, 232, 0.62);
}

.how-card-progress {
  display: block;
  height: 3px;
  margin-top: 1.15rem;
  border-radius: 999px;
  background: rgba(246, 244, 232, 0.12);
}

.how-card-progress i {
  display: block;
  height: 100%;
  border-radius: 999px;
  background: #ffcc00;
}

.how-note {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-top: 2rem;
  font-size: 0.78rem;
  color: rgba(246, 244, 232, 0.55);
}

.how-note svg {
  flex-shrink: 0;
  color: #ffcc00;
}

.how-note-link {
  color: #ffcc00;
  font-weight: 700;
  text-decoration: none;
  border-bottom: 1px solid rgba(255, 204, 0, 0.5);
}

.how-note-link:hover {
  border-color: #fff;
  color: #fff;
}

.catalog {
  padding: 4rem 0;
  background: var(--bv-bg);
  color: var(--bv-text);
}

.catalog-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1rem;
}

.catalog-skeleton {
  overflow: hidden;
  border: 1px solid var(--bv-line);
  border-radius: 1.55rem;
  background: var(--bv-surface);
}

.catalog-skeleton-thumb {
  aspect-ratio: 16 / 10;
  background: var(--bv-line);
  animation: bb-skeleton 1.6s ease-in-out infinite;
}

.catalog-skeleton-body {
  display: grid;
  gap: 0.6rem;
  padding: 1rem;
}

.catalog-skeleton-body span {
  height: 0.8rem;
  border-radius: 999px;
  background: var(--bv-line);
  animation: bb-skeleton 1.6s ease-in-out infinite;
}

.catalog-skeleton-body span:first-child {
  width: 76%;
}

.catalog-skeleton-body span:last-child {
  width: 42%;
}

.catalog-empty {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1.2rem;
  padding: 1.8rem;
  border: 2px solid var(--bv-line-strong);
  border-radius: 1.7rem;
  background: color-mix(in srgb, var(--bv-surface) 82%, transparent);
}

.catalog-empty-title {
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 800;
  letter-spacing: -0.035em;
}

.catalog-empty-text {
  margin-top: 0.4rem;
  max-width: 30rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--bv-muted);
}

.catalog-slot {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 1.5rem;
  min-height: 21rem;
  overflow: hidden;
  padding: 1.5rem;
  border: 2px solid color-mix(in srgb, var(--bv-text) 22%, #141414);
  border-radius: 1.55rem;
  background: #141414;
  color: #ffcc00;
  text-decoration: none;
  transition:
    background-color 0.25s ease,
    transform 0.25s ease,
    box-shadow 0.25s ease;
}

.catalog-slot::before {
  position: absolute;
  top: -130px;
  right: -110px;
  width: 320px;
  height: 320px;
  border-radius: 999px;
  content: '';
  background: repeating-linear-gradient(135deg, #ffcc00 0 13px, transparent 13px 26px);
  opacity: 0.14;
}

.catalog-slot:hover {
  background: #0c0c0b;
  box-shadow: 0 24px 48px -34px rgba(20, 20, 20, 0.8);
  transform: translateY(-4px);
}

.catalog-slot-top {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.catalog-slot-mark {
  display: grid;
  flex-shrink: 0;
  place-items: center;
  width: 2.8rem;
  height: 2.8rem;
  border-radius: 999px;
  background: #ffcc00;
  color: #141414;
  font-family: var(--font-display);
  font-size: 1.5rem;
  font-weight: 800;
  line-height: 1;
}

.catalog-slot-kicker {
  font-family: var(--font-mono);
  font-size: 0.56rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: rgba(255, 204, 0, 0.6);
}

.catalog-slot-title {
  position: relative;
  font-family: var(--font-display);
  font-size: 1.35rem;
  font-weight: 800;
  letter-spacing: -0.04em;
  line-height: 1.05;
  color: #fffdf7;
}

.catalog-slot-text {
  position: relative;
  max-width: 24rem;
  margin-top: 0.5rem;
  font-size: 0.78rem;
  line-height: 1.55;
  color: rgba(255, 253, 247, 0.55);
}

.catalog-slot-go {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-family: var(--font-mono);
  font-size: 0.6rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: #ffcc00;
}

.catalog .bb-btn--line {
  border-color: var(--bv-line-strong);
  color: var(--bv-text);
}

.catalog .bb-btn--line:hover {
  border-color: var(--bv-text);
  background: var(--bv-text);
  color: var(--bv-bg);
}

.catalog-top {
  margin-top: 3.6rem;
  padding-top: 2.4rem;
  border-top: 2px solid var(--bv-line-strong);
}

.catalog-top-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.2rem;
}

.catalog-top-title {
  font-family: var(--font-display);
  font-size: 1.35rem;
  font-weight: 800;
  letter-spacing: -0.04em;
  text-transform: uppercase;
}

.catalog-top-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-family: var(--font-mono);
  font-size: 0.6rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--bv-muted);
  text-decoration: none;
}

.catalog-top-link:hover {
  color: var(--bv-text);
}

.catalog-updated {
  margin-top: 1.5rem;
  font-family: var(--font-mono);
  font-size: 0.56rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--bv-faint);
}

.offer {
  position: relative;
  overflow: hidden;
  padding: 4rem 0;
  border-top: 2px solid #141414;
  border-bottom: 2px solid #141414;
  background: #ffcc00;
  color: #141414;
}

.offer::before {
  position: absolute;
  bottom: -260px;
  left: -170px;
  width: 560px;
  height: 560px;
  border-radius: 999px;
  content: '';
  background: repeating-linear-gradient(135deg, #141414 0 16px, transparent 16px 32px);
  opacity: 0.07;
}

.offer-grid {
  position: relative;
  display: grid;
  gap: 2.5rem;
}

.offer-lead {
  max-width: 32rem;
  margin-top: 1rem;
  font-size: 0.92rem;
  line-height: 1.65;
  color: rgba(20, 20, 20, 0.74);
}

.offer-amount {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.6rem;
  margin-top: 1.6rem;
}

.offer-amount-value {
  font-family: var(--font-display);
  font-size: clamp(2.5rem, 6vw, 4.5rem);
  font-weight: 800;
  letter-spacing: -0.06em;
  line-height: 0.85;
  white-space: nowrap;
}

.offer-amount-unit {
  padding: 0.32rem 0.58rem;
  border: 2px solid #141414;
  border-radius: 999px;
  font-family: var(--font-mono);
  font-size: 0.6rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.offer .bb-btn {
  margin-top: 1.6rem;
}

.offer-note {
  margin-top: 0.85rem;
  font-family: var(--font-mono);
  font-size: 0.58rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgba(20, 20, 20, 0.68);
}

.price-card {
  padding: 1.35rem;
  border: 2px solid #141414;
  border-radius: 1.75rem;
  background: #141414;
  color: #f6f4e8;
  box-shadow: 10px 10px 0 rgba(20, 20, 20, 0.2);
}

.price-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding-bottom: 0.9rem;
  border-bottom: 1px solid rgba(246, 244, 232, 0.16);
}

.price-head-title {
  font-family: var(--font-display);
  font-size: 1.05rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  text-transform: uppercase;
  color: #ffcc00;
}

.price-head-unit {
  font-family: var(--font-mono);
  font-size: 0.55rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: rgba(246, 244, 232, 0.56);
}

.price-list {
  display: flex;
  flex-direction: column;
  margin: 0;
  padding: 0;
  list-style: none;
}

.price-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.9rem 0;
  border-bottom: 1px dashed rgba(246, 244, 232, 0.16);
}

.price-row:last-child {
  border-bottom: 0;
}

.price-row-info {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.2rem;
}

.price-row-label {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.82rem;
  font-weight: 700;
}

.price-row-tag {
  padding: 0.14rem 0.38rem;
  border-radius: 999px;
  background: #ffcc00;
  color: #141414;
  font-family: var(--font-mono);
  font-size: 0.54rem;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.price-row-hint {
  font-family: var(--font-mono);
  font-size: 0.56rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgba(246, 244, 232, 0.56);
}

.price-row-value {
  flex-shrink: 0;
  font-family: var(--font-display);
  font-size: 1.1rem;
  font-weight: 800;
  letter-spacing: -0.03em;
  color: #ffcc00;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.price-foot {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(246, 244, 232, 0.16);
  font-size: 0.68rem;
  line-height: 1.45;
  color: rgba(246, 244, 232, 0.55);
}

.price-foot svg {
  flex-shrink: 0;
  color: #ffcc00;
}

.faq {
  padding: 4rem 0 5rem;
  background: var(--bv-bg);
  color: var(--bv-text);
}

.faq-grid {
  display: grid;
  gap: 2.5rem;
}

.faq-lead {
  max-width: 24rem;
  margin-top: 1rem;
  font-size: 0.85rem;
  line-height: 1.65;
  color: var(--bv-muted);
}

.faq-link {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  margin-top: 1.2rem;
  padding-bottom: 0.15rem;
  border-bottom: 2px solid #ffcc00;
  color: var(--bv-text);
  font-size: 0.76rem;
  font-weight: 800;
  text-decoration: none;
}

.faq-link:hover {
  color: var(--bv-hover-accent);
}

.faq-list {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.faq-item {
  padding: 0 1.15rem;
  border: 2px solid var(--bv-line-strong);
  border-radius: 1.2rem;
  background: color-mix(in srgb, var(--bv-surface) 72%, transparent);
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease;
}

.faq-item:hover,
.faq-item[open] {
  border-color: #ffcc00;
}

.faq-item[open] {
  background: var(--bv-surface);
}

.faq-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.05rem 0;
  font-family: var(--font-display);
  font-size: 0.92rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  cursor: pointer;
  list-style: none;
}

.faq-summary::-webkit-details-marker {
  display: none;
}

.faq-summary::marker {
  content: '';
}

.faq-plus {
  position: relative;
  flex-shrink: 0;
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 999px;
  background: #ffcc00;
  transition: transform 0.2s ease;
}

.faq-plus::before,
.faq-plus::after {
  position: absolute;
  top: 50%;
  left: 50%;
  content: '';
  background: #141414;
  transform: translate(-50%, -50%);
}

.faq-plus::before {
  width: 10px;
  height: 2px;
}

.faq-plus::after {
  width: 2px;
  height: 10px;
}

.faq-item[open] .faq-plus {
  transform: rotate(45deg);
}

.faq-answer {
  margin: 0;
  padding: 0 2.5rem 1.2rem 0;
  font-size: 0.82rem;
  line-height: 1.65;
  color: var(--bv-muted);
}

.final {
  position: relative;
  overflow: hidden;
  padding: 4rem 0 4.5rem;
  background: #0f0f0e;
  color: #f6f4e8;
}

.final::after {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 10px;
  content: '';
  background: repeating-linear-gradient(135deg, #ffcc00 0 14px, #0f0f0e 14px 28px);
}

.final-ornament {
  position: absolute;
  top: -240px;
  right: -200px;
  width: 620px;
  height: 620px;
  border-radius: 999px;
  background: repeating-linear-gradient(135deg, #ffcc00 0 16px, transparent 16px 32px);
  opacity: 0.12;
}

.final-inner {
  position: relative;
  display: grid;
  gap: 2rem;
}

.final-title {
  margin-top: 0.9rem;
  font-family: var(--font-display);
  font-size: clamp(2.6rem, 7vw, 5.4rem);
  font-weight: 800;
  letter-spacing: -0.06em;
  line-height: 0.86;
}

.final-title span {
  color: #ffcc00;
}

.final-lead {
  max-width: 28rem;
  margin-top: 1.1rem;
  font-size: 0.88rem;
  line-height: 1.6;
  color: rgba(246, 244, 232, 0.6);
}

.final-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.9rem;
}

.final-link {
  color: rgba(246, 244, 232, 0.6);
  font-size: 0.76rem;
  text-decoration: underline;
  text-underline-offset: 4px;
}

.final-link:hover {
  color: #ffcc00;
}

.hero :focus-visible,
.offer :focus-visible {
  outline-color: #141414;
}

.how :focus-visible,
.ticker :focus-visible,
.final :focus-visible {
  outline-color: #ffcc00;
}

@keyframes bb-rise {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes bb-marquee {
  from {
    transform: translateX(0);
  }
  to {
    transform: translateX(-50%);
  }
}

@keyframes bb-pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.35;
  }
}

@keyframes bb-skeleton {
  0%,
  100% {
    opacity: 0.55;
  }
  50% {
    opacity: 0.25;
  }
}

@media (min-width: 640px) {
  .hero {
    padding-top: 4rem;
  }

  .hero-facts {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .hero-fact {
    padding: 1.15rem 1rem 1.5rem;
    border-bottom: 0;
  }

  .hero-fact + .hero-fact {
    border-left: 1px solid rgba(20, 20, 20, 0.2);
  }

  .how,
  .catalog,
  .offer,
  .faq,
  .final {
    padding-top: 5rem;
    padding-bottom: 5rem;
  }

  .how-fast-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .catalog-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .catalog-empty {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    padding: 2rem;
  }

  .catalog-slot {
    grid-column: span 2;
    min-height: 18rem;
  }
}

@media (min-width: 1024px) {
  .hero {
    padding-top: 4.5rem;
  }

  .hero-grid {
    grid-template-columns: minmax(0, 1.08fr) minmax(0, 0.92fr);
    align-items: end;
    gap: 2.5rem;
  }

  .hero-copy {
    padding-bottom: 0.25rem;
  }

  .hero-facts {
    margin-top: 3.2rem;
  }

  .how-modes {
    grid-template-columns: minmax(0, 1.08fr) minmax(0, 0.92fr);
    align-items: start;
    gap: 1.5rem;
  }

  .catalog-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .catalog-slot {
    grid-column: span 1;
  }

  .catalog-slot--wide {
    grid-column: span 2;
  }

  .offer-grid {
    grid-template-columns: minmax(0, 1.05fr) minmax(0, 0.95fr);
    align-items: center;
    gap: 3rem;
  }

  .faq-grid {
    grid-template-columns: minmax(0, 0.85fr) minmax(0, 1.15fr);
    align-items: start;
    gap: 4rem;
  }

  .faq-intro {
    position: sticky;
    top: 100px;
  }

  .final-inner {
    grid-template-columns: minmax(0, 1.2fr) minmax(0, 0.8fr);
    align-items: center;
    gap: 2rem;
  }

  .final-actions {
    align-items: flex-end;
  }
}

@media (max-width: 639px) {
  .hero-actions .bb-btn {
    flex: 1 1 100%;
  }

  .section-head .bb-btn {
    width: 100%;
  }

  .price-card {
    box-shadow: 6px 6px 0 rgba(20, 20, 20, 0.18);
  }
}

@media (prefers-reduced-motion: reduce) {
  .ticker-track {
    animation: none;
  }

  .hero-kicker,
  .hero-title,
  .hero-lead,
  .hero-actions,
  .hero-status,
  .hero-ideas,
  .hero-facts {
    animation: none;
  }
}
</style>
