# Task Service

Сервис для управления задачами с HTTP API на Go.


## Изменения
- Добавлена логика повторов задач:
    Устанавливаются обязательный пар-р начала повторения(start_date) и необязательный конца повторений (end_date)
    - Ежедневный:
        Режим повторения: `daily`, параметр `interval` - интервал повторений (1 - каждый день, 2 - через день и тд)
    - Ежемесячный:
        Режим повторения: `monthly`, п-р `month_day` - день месяца, в который будет повторятся задача
    - По чет/нечет дням:
        Режим повторения: `parity`, п-р `parity` - true, если повтор по чет. дням, иначе false
    - По конкретным датам:
        Режим повторения: `cpecific`, п-р `spec_days` - список дат, в которые будет повторяться задача
    - Без повторения:
        Режим повторения: `none`


- Добавлен маршрут:
    `GET /api/v1/tasks/{date}`
    - Получение списка задач на определенную дату, с фильтрацией по start_date и end_date

- Изменена структура БД
    - Структура повторений описана в JSONB для максимальной гибкости

- Интегрирован CORS для взаимодействия с фронтенд-приложениями
    В том числе с работой в swagger напрямую

- Починил маршрут `GET /swagger` для отображения Swagger UI



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

Причина в том, что SQL-файл из `migrations/0001_create_tasks.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

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
