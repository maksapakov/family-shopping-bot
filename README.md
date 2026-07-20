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

Добавление в чат: пиши названия продуктов (`Картошка`, `Морковь Соль` — несколько слов = несколько item). Сообщение после обработки redact'ится.

## Переменные окружения

| Переменная | Обязательна | По умолчанию | Описание |
|------------|-------------|--------------|----------|
| `CALLBACK_SECRET` | да | — | Пусто → бот не стартует |
| `CALLBACK_BASE_URL` | для ссылок в списке | пусто | База URL callback |
| `DATABASE_PATH` | нет | `shopping.db` | Путь к SQLite |
| `DEMO` | нет | выкл | `DEMO=1` — seed `demo-chat` + toggle Milk при старте |
| `MATRIX_*` | для Matrix | — | См. таблицу выше |
| `SHOPPING_CHAT_ID` | для Matrix listener | пусто | `chat_id` в БД для входящих сообщений |

### Важно про `DEMO=1`

При каждом старте `runDemo` **перезаписывает** список `demo-chat` (один item Milk). Накопленные в Matrix товары для этого `chat_id` пропадут. Для обыной работы **не ставь** `DEMO=1`, либо используй другой `SHOPPING_CHAT_ID`, либо доработай demo «seed только если списка нет».

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
