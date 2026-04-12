# Task Service

Сервис для управления задачами с HTTP API на Go. Поддерживает разовые задачи и периодические шаблоны с автоматической материализацией экземпляров.

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

Оба SQL-файла из `migrations/` монтируются в `docker-entrypoint-initdb.d` и применяются только при инициализации пустого data volume.

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

## Периодические задачи (Recurrence)

### Типы периодичности

| `recurrence_type` | Обязательный параметр | Описание |
|---|---|---|
| `daily` | `recurrence_daily_interval` (int ≥ 1) | Каждые N дней, отсчёт от эпохи 2000-01-01 UTC |
| `monthly_days` | `recurrence_monthly_days` (int[] 1–30) | Перечисленные числа каждого месяца; дни, которых нет в месяце, пропускаются |
| `specific_dates` | `recurrence_specific_dates` (string[] YYYY-MM-DD) | Конкретные даты |
| `day_parity` | `recurrence_day_parity` (`"even"` / `"odd"`) | Чётные или нечётные числа месяца; день 31 всегда пропускается |

### Примеры

**Создать ежедневный шаблон (каждые 2 дня):**
```bash
curl -s -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Ежедневный обзор","recurrence_type":"daily","recurrence_daily_interval":2}'
```

**Создать шаблон по числам 1 и 15 каждого месяца:**
```bash
curl -s -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Двухнедельная задача","recurrence_type":"monthly_days","recurrence_monthly_days":[1,15]}'
```

**Создать шаблон на конкретные даты:**
```bash
curl -s -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Квартальный отчёт","recurrence_type":"specific_dates","recurrence_specific_dates":["2026-06-01","2026-09-01","2026-12-01"]}'
```

**Создать шаблон по чётным числам месяца:**
```bash
curl -s -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Чётные дни","recurrence_type":"day_parity","recurrence_day_parity":"even"}'
```

**Получить список задач (включая шаблоны):**
```bash
curl -s "http://localhost:8080/api/v1/tasks?include=templates"
```

**Обновить шаблон (заменить правило):**
```bash
curl -s -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Ежедневный обзор","recurrence_type":"daily","recurrence_daily_interval":3}'
```

**Удалить шаблон вместе со всеми экземплярами:**
```bash
curl -s -X DELETE http://localhost:8080/api/v1/tasks/1
```

## Принятые решения

Обоснование архитектурных и технических решений по реализации периодических задач — в файле [decisions.md](decisions.md).
