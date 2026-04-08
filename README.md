# Task Service

Сервис управления задачами с HTTP API на Go.

## Что изменено

Проект переведен на модель:

- `task template` — шаблон задачи с настройками периодичности;
- `task occurrence` — конкретное выполнение задачи на определенную дату со своим статусом.

Это убирает главный дефект прошлой версии: теперь отметка `done` относится к конкретному выполнению, а не ко всей серии целиком.

## Ключевые решения

- `GET /api/v1/tasks` работает с конкретными выполнениями задач.
- `GET /api/v1/task-templates` работает с шаблонами серий.
- `POST /api/v1/tasks` сохранен для совместимости и создает шаблон плюс первое выполнение.
- Для периодических задач новые выполнения создаются лениво:
  - при создании шаблона;
  - при обновлении шаблона;
  - при чтении списка задач.
- Горизонт автогенерации по умолчанию: 30 дней вперед.
- Периодические правила считаются в `schedule.timezone`, если она указана, иначе в `UTC`.
- Для управления расписанием периодической серии нужно использовать `/api/v1/task-templates/{id}`.
- `DELETE /api/v1/tasks/{id}` удаляет только конкретное выполнение.
- `DELETE /api/v1/task-templates/{id}` удаляет всю серию вместе со всеми выполнениями.

## Ограничения и допущения

- Для one-off задачи отдельного due-date нет: первое выполнение создается на дату создания задачи.
- Перевод периодической серии обратно в one-off через template endpoint запрещен, чтобы не терять историю и не вводить неочевидную семантику.
- При изменении шаблона серии будущие выполнения пересобираются с текущего локального дня шаблона.
- Исторические выполнения остаются неизменными.
- Пагинация идет через query-параметры `limit` и `offset`, а метаданные возвращаются в headers:
  - `X-Total-Count`
  - `X-Limit`
  - `X-Offset`

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Запуск

```bash
docker compose up --build
```

Сервис будет доступен по адресу `http://localhost:8080`.

## Миграции

Миграции больше не завязаны на `docker-entrypoint-initdb.d`.

Теперь приложение само:

- подключается к PostgreSQL;
- создает таблицу `schema_migrations`;
- применяет все `*.up.sql` из каталога `migrations`.

Если у вас остался старый volume с таблицей `tasks`, миграция `0006_migrate_legacy_tasks.up.sql` перенесет данные в новую схему.

## Healthcheck

- `GET /healthz`

## Swagger

- UI: `http://localhost:8080/swagger/`
- OpenAPI JSON: `http://localhost:8080/swagger/openapi.json`

## API

Базовый префикс: `/api/v1`

### Tasks

Вхождения задач:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

Фильтры списка задач:

- `limit`
- `offset`
- `status`
- `template_id`
- `date_from`
- `date_to`

### Task Templates

Шаблоны серий:

- `POST /api/v1/task-templates`
- `GET /api/v1/task-templates`
- `GET /api/v1/task-templates/{id}`
- `PUT /api/v1/task-templates/{id}`
- `DELETE /api/v1/task-templates/{id}`

Фильтры списка шаблонов:

- `limit`
- `offset`
- `search`
- `schedule_type`

## Пример создания периодической задачи

```json
{
  "title": "Обзвон пациентов",
  "description": "Ежедневный контроль после выписки",
  "status": "new",
  "schedule": {
    "type": "daily",
    "every_days": 2,
    "start_date": "2026-04-08",
    "timezone": "Europe/Moscow"
  }
}
```

## Пример создания шаблона серии

```json
{
  "title": "Инвентаризация",
  "description": "Еженедельная проверка расходников",
  "schedule": {
    "type": "monthly",
    "day_of_month": 15,
    "timezone": "Europe/Moscow"
  }
}
```

## Тесты

Юнит- и HTTP-тесты:

```bash
go test ./...
```

Интеграционный тест репозитория можно прогнать с отдельной БД:

```bash
set TEST_DATABASE_DSN=postgres://postgres:postgres@localhost:5432/taskservice_test?sslmode=disable
go test ./internal/repository/postgres -run TestRepositoryTemplateAndOccurrenceLifecycle -v
```
