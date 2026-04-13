# Task Service

Сервис для управления задачами с HTTP API на Go.

## Реализованный функционал периодичности

Добавлена поддержка периодичности задач. Для описания повторений используется поле `recurrence` (тип `JSONB` в БД) и обязательная дата старта `start_date`. Опционально может быть задано `end_date` после которого задача считается неактивной

Поддерживаемые типы повторения:

- `none`: без повторений (задача активна только в `start_date`)
- `daily`: каждый `interval_days` день, начиная с `start_date`
- `monthly`: в указанное число месяца `day_of_month` (только `1..30`)
- `specific`: только в даты из `specific_dates`
- `parity`: по четности дня месяца (`parity_even=true` для четных, `false` для нечетных)

Добавлен маршрут `GET /api/v1/tasks/by-date?date=YYYY-MM-DD`
Валидация правил повторения вынесена в доменную модель

## Ключевые бизнес-правила

- `start_date` обязателен при `POST /tasks` и `PUT /tasks/{id}`
- для `monthly` разрешены только дни `1..30` (по условию задания)
- для `daily` `interval_days` должен быть больше `0`
- для `specific` список дат `specific_dates` не может быть пустым
- если задан `end_date`, задача считается неактивной после этой даты

## Примеры запросов

Создание ежедневной задачи (каждый 2-й день):

```json
{
  "title": "Обход пациентов",
  "description": "Палата 2 и 3",
  "status": "new",
  "start_date": "2026-04-01T00:00:00Z",
  "recurrence": {
    "type": "daily",
    "interval_days": 2
  }
}
```

Создание задачи по четным дням:

```json
{
  "title": "Формирование отчета",
  "description": "Только по четным дням",
  "status": "new",
  "start_date": "2026-04-01T00:00:00Z",
  "recurrence": {
    "type": "parity",
    "parity_even": true
  }
}
```

Получение задач на дату:

```text
GET /api/v1/tasks/by-date?date=2026-04-12
```

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Миграции из `migrations/0001_create_tasks.up.sql`, `migrations/0002_add_recurrence.up.sql` и `migrations/0004_add_start_date.up.sql` применяются при инициализации пустого volume через `docker-entrypoint-initdb.d`.

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`
- `GET /api/v1/tasks/by-date?date=YYYY-MM-DD`
