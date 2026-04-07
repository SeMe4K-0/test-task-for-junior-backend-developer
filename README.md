# Task Service

Сервис для управления задачами с HTTP API на Go.

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

Причина в том, что SQL-файлы из `migrations/` монтируются в `docker-entrypoint-initdb.d` и применяются только при инициализации пустого data volume.

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

## Периодичность задач

Сервис поддерживает установку периодичности для задач. Врачи могут настраивать задачи на повторение с определённой частотой.

### Типы периодичности

1. **none** - однократная задача (по умолчанию)
2. **daily** - ежедневная задача каждый n-й день
3. **monthly** - ежемесячная задача на конкретное число (1-30)
4. **even_days** - только на чётные дни месяца
5. **odd_days** - только на нечётные дни месяца
6. **specific_dates** - только на указанные конкретные даты

### Примеры использования

#### Создание ежедневной задачи

```bash
POST /api/v1/tasks

{
  "title": "Обзвон пациентов",
  "description": "Ежедневный обзвон пациентов",
  "status": "new",
  "periodicity_type": "daily",
  "periodicity_config": {
    "interval": 1,
    "start_date": "2024-01-15T00:00:00Z"
  }
}
```

#### Создание ежемесячной задачи

```bash
POST /api/v1/tasks

{
  "title": "Формирование отчётности",
  "description": "Формирование ежемесячного отчёта",
  "status": "new",
  "periodicity_type": "monthly",
  "periodicity_config": {
    "month_day": 15,
    "start_date": "2024-01-15T00:00:00Z"
  }
}
```

#### Создание задачи на чётные дни

```bash
POST /api/v1/tasks

{
  "title": "Инспекция на чётные дни",
  "description": "Проверка оборудования только на чётные дни",
  "status": "new",
  "periodicity_type": "even_days",
  "periodicity_config": {
    "start_date": "2024-01-01T00:00:00Z"
  }
}
```

#### Создание задачи на конкретные даты

```bash
POST /api/v1/tasks

{
  "title": "Совещание руководства",
  "description": "Встреча в определённые дни",
  "status": "new",
  "periodicity_type": "specific_dates",
  "periodicity_config": {
    "specific_dates": ["2024-01-15", "2024-02-10", "2024-03-20"]
  }
}
```

### Структура PeriodicityConfig

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|---------|
| `interval` | int | Для daily | Интервал в днях (n > 0) |
| `month_day` | int | Для monthly | День месяца (1-30) |
| `specific_dates` | []string | Для specific_dates | Массив дат в формате "YYYY-MM-DD" |
| `start_date` | timestamp | Для daily/monthly/even_days/odd_days | Дата начала периодичности |

### Валидация

- Все задачи с периодичностью (кроме `none` и `specific_dates`) требуют `start_date`
- Для `monthly`: `month_day` должен быть в диапазоне 1-30
- Для `daily`: `interval` должен быть положительным числом
- Для `specific_dates`: массив не должен быть пустым, даты должны быть в формате "YYYY-MM-DD"
- В случае невалидных параметров возвращается HTTP 400 с описанием ошибки

## Архитектура

Проект следует архитектурным принципам:

- **Domain Layer** (`internal/domain`) - модели и бизнес-логика
- **Usecase Layer** (`internal/usecase`) - сценарии использования
- **Repository Pattern** (`internal/repository`) - абстракция БД
- **HTTP Transport** (`internal/transport/http`) - REST API

## Предположения и решения

1. **Хранение конфига периодичности**: Используется JSONB в PostgreSQL для гибкого хранения различных типов параметров периодичности.

2. **Значение по умолчанию**: Если `periodicity_type` не указан, используется `none` (однократная задача).

3. **Конкретные даты**: Дата начала не требуется для `specific_dates`, так как сами даты определяют расписание.

4. **Редактирование периодичности**: После создания периодичность задачи можно изменить через PUT запрос, включая смену типа периодичности.

5. **Ограничение дней месяца**: Максимум 30 дней для `monthly` типа, чтобы избежать проблем с месяцами < 31 дней.

