# GameGen — бэкенд

Сервис генерации браузерных игр по текстовому промпту: многоагентный LLM-пайплайн
(анализ → план → код → ассеты → ревью → сборка), версии, оценки, каталог, профиль с балансом токенов.

Документация API — `api/http/openapi.yaml`. План реализации — `docs/backend-plan.md` в корне репозитория.

## Стек

Go 1.27, chi v5, pgx v5 (Postgres), goose (миграции), zerolog, OpenRouter (LLM), mockery (моки).

## Запуск

```bash
docker compose -f ../docker-compose.dev.yaml up -d   # Postgres на localhost:5432
cp .env.example .env                 # заполнить OPENROUTER_API_KEY, APP_NAME, APP_VERSION
make migrate-up                      # применить миграции
make run                             # поднять сервис на :8080
```

Конфиг читается из переменных окружения (envconfig). Каждый параметр либо обязателен, либо имеет значение
по умолчанию; полный список — `.env.example`.

## Команды

| Команда | Что делает |
|---|---|
| `make build` | собрать бинарник в `bin/gamegen` |
| `make run` | запустить сервис |
| `make lint` | `golangci-lint run` |
| `make test` | юнит-тесты |
| `make integration-test` | интеграционные тесты (нужен поднятый Postgres) |
| `make migrate-up` / `make migrate-down` / `make migrate-status` | миграции goose |
| `make mocks` | перегенерировать моки портов (mockery) |

## Структура

```
cmd/gamegen        — точка входа сервиса
cmd/migrate        — CLI миграций
config             — агрегат конфигов слоёв
internal/app       — composition root, воркер генерации, порядок остановки
internal/domain    — сущности, инварианты, статусная машина игры
internal/dto       — контракты операций (request/response)
internal/usecase   — порты (usecase.go) и операции (файл = операция)
internal/adapter   — postgres (store), llm (OpenRouter), validator (контракт игры), auth (bcrypt)
internal/controller/http — HTTP-роуты: /api/* (v1) и публичные /g/:id (public)
pkg/               — transaction, postgres, render, httpserver, logger, tracer
migration/         — SQL-миграции goose (+ embed)
assets/prompts/    — системные промпты агентов (embed)
api/http/          — OpenAPI-контракт
test/integration/  — интеграционные тесты (//go:build integration)
```

## Конвейер генерации

`POST /api/games` списывает токены и ставит игру в очередь. Воркер (`internal/app/worker.go`) забирает
игру из БД (`FOR UPDATE SKIP LOCKED`) и прогоняет 6 этапов, публикуя события в `gen_events`:
`stage` → `artifact` (spec, config) → `done`. Фронт читает их через `GET /api/games/:id/stream` (SSE)
и может переподключиться с заголовком `Last-Event-ID` — пропущенные события доигрываются.

При сбое генерации токены возвращаются пользователю (`token_ops.reason = refund`).

Правки: `POST /api/games/:id/patch` (текст → LLM → валидация → версия, списывает токены),
`PATCH /api/games/:id/config` (правка блока `CONFIG` без LLM, бесплатно),
`POST /api/games/:id/rollback`, `POST /api/games/:id/fix` (repair по ошибке из превью).

## Идемпотентность

`POST /api/games` принимает заголовок `Idempotency-Key`: повторный запрос с тем же ключом возвращает
ту же игру и не списывает токены второй раз; пока операция выполняется — `409 idempotency_in_progress`.

## Тесты

```bash
make test               # домен, юзкейсы (моки портов), адаптеры llm/validator
make integration-test   # полный срез: register → generate → stream → patch → rate → catalog
```

Интеграционные тесты поднимают фейковый OpenRouter (`httptest`), поэтому ключ API и деньги не нужны.
Они работают в отдельной БД (`POSTGRES_TEST_DSN`, по умолчанию `bee_test`, создаётся автоматически) —
чтобы воркер запущенного dev-сервера не забирал тестовые игры.
