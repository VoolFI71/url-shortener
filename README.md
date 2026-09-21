# URL shortener

Небольшой сервис для сокращения ссылок. Короткий код всегда состоит из 10
символов: букв латинского алфавита, цифр и `_`.

Если отправить один и тот же URL несколько раз, сервис вернёт одну и ту же
короткую ссылку.

## Запуск в Windows PowerShell

```powershell
go run .\cmd\api
```

Сервис будет доступен на `http://localhost:8080`.

Создать короткую ссылку:

```powershell
$body = @{ url = 'https://example.com/article' } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri 'http://localhost:8080/api/v1/urls' -ContentType 'application/json' -Body $body
```

Новая ссылка возвращается с `201 Created`. Повторный запрос того же URL — с
`200 OK` и тем же кодом.

Переход по `GET /{code}` отвечает `302 Found` и перенаправляет на исходный
адрес. Подставь код из поля `short_url`:

```powershell
curl.exe -i http://localhost:8080/Abc123_Xyz
```

При запуске в памяти данные пропадут после остановки сервиса.

## PostgreSQL и Docker

Для запуска вместе с PostgreSQL:

```powershell
docker-compose up --build
```

Compose создаёт базу, применяет миграцию из `migrations` и запускает сервис на
порту `8080`.

Можно подключить свою базу:

```powershell
$env:STORAGE = 'postgres'
$env:DATABASE_URL = 'postgres://user:password@localhost:5432/shortener?sslmode=disable'
go run .\cmd\api
```

Переменные окружения:

- `ADDR` — адрес для HTTP-сервера, по умолчанию `:8080`;
- `PUBLIC_BASE_URL` — адрес, который будет добавляться к короткому коду;
- `STORAGE` — `memory` или `postgres`;
- `DATABASE_URL` — нужен только для PostgreSQL.

## Тесты

```powershell
go test .\...
```
