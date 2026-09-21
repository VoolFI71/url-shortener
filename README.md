# URL shortener

Небольшой сервис для сокращения ссылок. Короткий код всегда состоит из 10
символов: букв латинского алфавита, цифр и `_`.

Если отправить один и тот же URL несколько раз, сервис вернёт одну и ту же
короткую ссылку.

## Запуск в памяти

```bash
go run ./cmd/api
```

Сервис будет доступен на `http://localhost:8080`.

Создать короткую ссылку:

```bash
curl -i -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/article"}'
```

Новая ссылка возвращается с `201 Created`. Повторный запрос того же URL — с
`200 OK` и тем же кодом.

Переход по `GET /{code}` отвечает `302 Found` и перенаправляет на исходный
адрес.

При запуске в памяти данные пропадут после остановки сервиса.

## PostgreSQL и Docker

Для запуска вместе с PostgreSQL:

```bash
docker compose up --build
```

Compose создаёт базу, применяет миграцию из `migrations` и запускает сервис на
порту `8080`.

Можно подключить свою базу:

```bash
STORAGE=postgres DATABASE_URL='postgres://user:password@localhost:5432/shortener?sslmode=disable' go run ./cmd/api
```

Переменные окружения:

- `ADDR` — адрес для HTTP-сервера, по умолчанию `:8080`;
- `PUBLIC_BASE_URL` — адрес, который будет добавляться к короткому коду;
- `STORAGE` — `memory` или `postgres`;
- `DATABASE_URL` — нужен только для PostgreSQL.

## Тесты

```bash
go test ./...
```
