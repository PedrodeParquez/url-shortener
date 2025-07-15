# URL Shortener

Сервис сокращения ссылок на Go с использованием Gin и PostgreSQL.

## Установка и запуск

1. Клонируйте репозиторий:

```bash
git clone <https://github.com/PedrodeParquez/url-shortener>
cd url-shortener
```

2. Установите зависимости:

```bash
go mod download
```

3. Создайте базу данных PostgreSQL:

```sql
CREATE DATABASE name_db;
CREATE USER name_user WITH PASSWORD 'password';
```

4. Настройте переменные окружения:

```bash
cp example.env .env

cd config

cp example.yaml .yaml
```

5. Запустите сервис:

```bash
go run cmd/url-shortener/main.go
```

## API Эндпоинты

### 1. Создание короткой ссылки

```bash
POST /api/save
Authorization: Basic

{
    "url": "https://example.com",
    "alias": "custom-alias"
}
```

### 2. Удаление ссылки

```bash
DELETE /api/link/{alias}
Authorization: Basic
```

### 3. Переход по короткой ссылке

```bash
GET /api/{alias}
```

## Тестирование

Запуск unit-тестов:

```bash
go test ./...
```

Запуск unit-тестов с покрытием:

```bash
go test -cover ./...
```

Запуск интеграционных тестов:

```bash
go test ./internal/config/tests/
```
