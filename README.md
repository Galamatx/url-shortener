# URL Shortener (Go + PostgreSQL)

Принимает длинный URL, выдаёт короткий код, редиректит по нему и считает переходы.
Идея основана на упражнении «URL Shortener» из [Gophercises](https://gophercises.com), расширена хранением в PostgreSQL.

Стек: Go 1.22, `net/http`, `database/sql` + `lib/pq`, PostgreSQL 16, Docker Compose.

## Запуск

```bash
docker compose up --build
```

Сервис: `http://localhost:8080`. Таблицы создаются из `db/init.sql` при первом старте базы
(чтобы пересоздать их с нуля: `docker compose down -v`).

## API

| Метод | Путь | Что делает | Коды |
|-------|------|-----------|------|
| `POST` | `/shorten` | создаёт короткую ссылку, тело: `{"url": "https://..."}` | `201`, `400` |
| `GET` | `/{code}` | редирект на оригинальный URL, записывает клик | `302`, `404` |
| `GET` | `/stats/{code}` | оригинальный URL и число кликов | `200`, `404` |

```bash
curl -X POST localhost:8080/shorten -d '{"url":"https://go.dev/doc"}'
# {"code":"aB3dE9","short_url":"http://localhost:8080/aB3dE9"}

curl -i localhost:8080/aB3dE9        # 302 + заголовок Location
curl localhost:8080/stats/aB3dE9     # {"clicks":1,"code":"aB3dE9","url":"https://go.dev/doc"}
```

## Файлы

```
main.go            запуск: читает переменные окружения, подключается к БД, стартует сервер
handlers.go        HTTP-обработчики и валидация URL
store.go           работа с PostgreSQL (интерфейс Store + реализация)
handlers_test.go   тесты обработчиков (с поддельным хранилищем, без базы)
db/init.sql        схема БД
```

## База данных

- `links(id, code UNIQUE, url, created_at)`
- `clicks(id, link_id → links.id, clicked_at)`
- индекс по `clicks.link_id` для JOIN в статистике; по `code` индекс создаёт `UNIQUE`.

## Что здесь важно

- **Транзакция** при редиректе: ссылка ищется и клик записывается как одна операция.
- **LEFT JOIN** в статистике, чтобы ссылка без кликов тоже возвращалась (с нулём).
- **Коллизии кодов:** код случайный, уникальность проверяет БД (`UNIQUE`); если код занят, берётся новый.
- **302, а не 301:** браузер кеширует 301, и повторные переходы не считались бы.

## Тесты

```bash
go test ./...
```

