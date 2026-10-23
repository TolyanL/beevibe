# GameGen — бэкенд

Сервис генерации браузерных игр по текстовому промпту: многоагентный LLM-пайплайн
(анализ → план → код → ассеты → ревью → сборка), версии, оценки, каталог, профиль с балансом токенов.

Документация API — `api/http/openapi.yaml`. План реализации — `docs/backend-plan.md` в корне репозитория.

## Стек

Go 1.27, chi v5, pgx v5 (Postgres), goose (миграции), zerolog, OpenRouter (LLM), mockery (моки).

## Запуск

```bash
make up                                     # Postgres + Kafka + Kafka UI (docker-compose.dev.yaml)
cp configs/config.example.yaml configs/config.yaml   # заполнить llm.api_key
make migrate-up                             # применить миграции (goose CLI)
make run                                    # поднять сервис на :8080
```

Путь к файлу конфигурации можно переопределить переменной `APP_CONFIG_PATH`.

Миграции — goose CLI по каталогу `migrations/`: `make migrate-up`, `make migrate-down`,
`make migrate-status`, `make migrate-create name=<имя>`.

Конфиг читается из YAML-файла `configs/config.yaml` (шаблон — `configs/config.example.yaml`,
рабочая копия в `.gitignore`). Отсутствующие поля берутся из значений по умолчанию,
некорректные значения (уровень рассуждения, диалект провайдера) приводят к отказу на старте.
Файл `.env` оставлен только для переменных docker compose и команд миграций.

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
cmd/app            — точка входа сервиса
cmd/migrate        — CLI миграций
configs/           — конфигурация сервиса (config.yaml + шаблон config.example.yaml)
internal/app       — composition root, воркер генерации, порядок остановки
internal/domain    — сущности, инварианты, статусная машина игры
internal/dto       — контракты операций (request/response)
internal/usecase   — агрегат UseCase (usecase.go) + домены:
    user/ game/ catalog/ outbox/ event/
internal/adapter   — postgres (агрегат + user/ game/ catalog/ outbox/), kafka,
                     validator (контракт игры), auth (bcrypt)
internal/config    — агрегат конфигов слоёв, чтение YAML
internal/controller — http (router, v1, public), kafka_consumer
pkg/               — llm (клиент, порт, промпты), postgres, transaction, render,
                     httpserver, logger, tracer
migrations/        — SQL-миграции goose (каталог, goose CLI)
api/http/          — OpenAPI-контракт
test/integration/  — интеграционные тесты (//go:build integration)
```

Пакеты по доменам — как в референсном проекте `vpn-proxy`: агрегат в корне пакета
собирает под-пакеты доменов, каждый домен объявляет свои узкие порты.

## События и Kafka (outbox)

События жизненного цикла игры (`game.created`, `game.ready`, `game.failed`, `game.patched`)
пишутся в таблицу `outbox` **в одной транзакции** с доменной правкой. Фоновый воркер
(`internal/app/outbox_worker.go`) читает неотправленные события и публикует их в топик
`game.events` (продюсер — `internal/adapter/kafka`). Консьюмер
(`internal/controller/kafka_consumer`) складывает события в проекцию `game_event_log`
идемпотентно по `event_id` — повторная доставка не создаёт дубль.

Адрес брокера — `kafka.brokers` в конфиге (по умолчанию `localhost:9094`, внешний listener из compose);
топик создаётся автоматически (`KAFKA_AUTO_CREATE_TOPICS_ENABLE=true`).

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
