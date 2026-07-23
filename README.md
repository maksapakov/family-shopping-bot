# Family Shopping Bot

Семейный список покупок: один список в чате, toggle/undo, добавление по тексту (Matrix).

## Требования

- Go 1.22+ (в проекте: 1.26.4)
- SQLite (файл создаётся автоматически)

## Быстрый старт (локально, без Matrix)

HTTP на `:8181`, messenger — заглушка (FakeMessenger).

```bash
cd /path/to/family-shopping-bot

CALLBACK_SECRET=dev-secret \
CALLBACK_BASE_URL=http://localhost:8181 \
DEMO=1 \
go run ./cmd/bot/
```

Проверка:

```bash
curl http://localhost:8181/health
```

В логе при `DEMO=1` будут подписанные URL для `/toggle` и `/undo` (префикс `http://localhost:8181` добавь сам).

## Запуск с Matrix

Нужны все три переменные — иначе используется FakeMessenger.

| Переменная | Пример | Описание |
|------------|--------|----------|
| `MATRIX_HOMESERVER` | `https://myserver.ru` | URL homeserver без `/_matrix/...` |
| `MATRIX_ACCESS_TOKEN` | `syt_...` | Access token бота |
| `MATRIX_ROOM_ID` | `!abc:myserver.ru` | Internal room ID из Element → Settings → Advanced |
| `SHOPPING_CHAT_ID` | `demo-chat` | Ключ списка в SQLite (должен совпадать с тем, что в БД) |
| `CALLBACK_SECRET` | произвольная строка | Секрет для подписи ссылок toggle/undo |
| `CALLBACK_BASE_URL` | `http://localhost:8181` | Базовый URL для ссылок в сообщении списка |

Пример:

```bash
CALLBACK_SECRET=dev-secret \
CALLBACK_BASE_URL=http://localhost:8181 \
SHOPPING_CHAT_ID=demo-chat \
MATRIX_HOMESERVER=https://myserver.ru \
MATRIX_ACCESS_TOKEN='your-token' \
MATRIX_ROOM_ID='!your-room-id:myserver.ru' \
go run ./cmd/bot/
```

Комната должна быть **без E2EE** (или бот не прочитает текст). Боту нужны права **Moderator** (или redact чужих сообщений), иначе ввод семьи не удалится из чата.

Добавление в чат: пиши названия продуктов (`Картошка, Морковь; Соль` — несколько item через `,` `;` или перевод строки; пробел внутри имени сохраняется). Сообщение после обработки redact'ится.

## Переменные окружения

| Переменная | Обязательна | По умолчанию | Описание |
|------------|-------------|--------------|----------|
| `CALLBACK_SECRET` | да | — | Пусто → бот не стартует |
| `CALLBACK_BASE_URL` | для ссылок в списке | пусто | База URL callback |
| `DATABASE_PATH` | нет | `shopping.db` | Путь к SQLite |
| `DEMO` | нет | выкл | `DEMO=1` — seed `demo-chat` (Milk), только если списка ещё нет |
| `MATRIX_*` | для Matrix | — | См. таблицу выше |
| `SHOPPING_CHAT_ID` | для Matrix listener | пусто | `chat_id` в БД для входящих сообщений |

### Важно про `DEMO=1`

При старте `runDemo` создаёт `demo-chat` **только если списка ещё нет** (seed Milk + один toggle). Если список уже есть — не трогает. На проде **не ставь** `DEMO=1`, либо используй другой `SHOPPING_CHAT_ID`.

## HTTP API

| Метод | Путь | Параметры |
|-------|------|-----------|
| GET | `/health` | — |
| GET | `/toggle` | `chat`, `item`, `sig` |
| GET | `/undo` | `chat`, `sig` |

Без валидной подписи `sig` → 403.

## Тесты

```bash
go test ./internal/domain/ ./internal/app/ ./internal/adapters/matrix/ -v
```

SQLite-тесты (нужен доступ к модулю `modernc.org/sqlite`):

```bash
go test ./internal/adapters/sqlite/ -v
```

## Структура

```
cmd/bot/           — точка входа, HTTP, wiring
internal/domain/   — список, item, undo
internal/app/      — use cases (Toggle, Undo, AddItem)
internal/adapters/ — sqlite, matrix, memory
internal/port/     — интерфейсы
```

## Разработка и релизы

Основная ветка: `main` (через PR). Semver считает [Release Please](https://github.com/googleapis/release-please) по [Conventional Commits](https://www.conventionalcommits.org/).

### Conventional Commits → semver

| Префикс | Версия | Пример |
|---------|--------|--------|
| `fix:` | patch (0.1.0 → 0.1.1) | `fix: не redact'ить при пустом вводе` |
| `feat:` | minor (0.1.0 → 0.2.0) | `feat: toggle по команде в Matrix` |
| `feat!:` или `BREAKING CHANGE:` в теле | major (0.1.0 → 1.0.0) | `feat!: смена API callback` |
| `docs:`, `chore:`, `refactor:`, `test:` | без bump | `docs: описать env` |

### Workflow

1. Ветка от `main` → коммиты как выше → PR
2. CI (`go test ./...`) зелёный → merge в `main`
3. Release Please откроет/обновит PR «Release v0.x.x»
4. Merge Release PR → тег + GitHub Release + `CHANGELOG.md`

## Docker

Образ: [`ghcr.io/maksapakov/family-shopping-bot`](https://github.com/maksapakov/family-shopping-bot/pkgs/container/family-shopping-bot)

Теги: `latest`, semver (`0.4.0`, `0.4`, …) — публикуются workflow при теге `v*` или вручную (`workflow_dispatch` → тег `manual`).

### Переменные окружения

```bash
cp env.example .env
```

Заполни значения в `.env`. Файл `.env` в git не коммитится.

| Переменная | Обязательна | Описание |
|------------|-------------|----------|
| `CALLBACK_SECRET` | да | Секрет подписи `/toggle` и `/undo` |
| `CALLBACK_BASE_URL` | да для ссылок в списке | Публичный базовый URL бота, напр. `https://bot.example.com` |
| `DATABASE_PATH` | нет | Путь к SQLite. В compose по умолчанию `/data/shopping.db` |
| `SHOPPING_CHAT_ID` | для Matrix | Ключ списка в БД |
| `MATRIX_HOMESERVER` | для Matrix | Homeserver без `/_matrix/...` |
| `MATRIX_ACCESS_TOKEN` | для Matrix | Access token бота |
| `MATRIX_ROOM_ID` | для Matrix | Internal room ID |
| `DEMO` | нет | `1` — seed `demo-chat`, только если списка ещё нет. На проде не включать |

### Compose (сервер, образ из GHCR)

Создай `compose.yaml` рядом с `.env`:

```yaml
services:
  bot:
    image: ghcr.io/maksapakov/family-shopping-bot:latest
    # или semver :0.4.0
    ports:
      - "8181:8181"
    env_file:
      - .env
    environment:
      DATABASE_PATH: /data/shopping.db
    volumes:
      - bot-data:/data
    restart: unless-stopped

volumes:
  bot-data:
```

Запуск:

```bash
docker compose pull
docker compose up -d
curl https://example.server.com/health
```

SQLite хранится в volume `bot-data` → `/data/shopping.db`.

Если package private:

```bash
echo "$GITHUB_TOKEN" | docker login ghcr.io -u USERNAME --password-stdin
```

Нужен token с `read:packages` (или `gh auth token` после `gh auth login`).

### Локальная сборка из репозитория

В репозитории есть `compose.yaml` с `build: .`.

```bash
cp env.example .env
# заполни секреты
docker compose up --build -d
curl https://example.server.com/health
```
