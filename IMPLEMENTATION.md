# Реализация функциональности периодичности задач

## Назначение

Добавлена возможность задавать периодичность (расписание) для задач в трекере. Это позволяет врачам и персоналу создавать повторяющиеся задачи с различными типами периодичности.

## Изменённые файлы

### 1. Domain Layer (`internal/domain/task/task.go`)

**Добавлены:**
- Тип `PeriodicityType` с константами:
  - `PeriodicityDaily` - каждый n-й день
  - `PeriodicityMonthly` - определённое число месяца (1-30)
  - `PeriodicitySpecificDate` - конкретные даты
  - `PeriodicityEvenDays` - чётные дни месяца
  - `PeriodicityOddDays` - нечётные дни месяца
  - `PeriodicityNone` - однократная задача (по умолчанию)

- Структура `PeriodicityConfig` с полями:
  - `Interval int` - интервал для ежедневных задач (каждый n-й день)
  - `MonthDay int` - день месяца для ежемесячных (1-30)
  - `SpecificDates []string` - массив дат в формате "YYYY-MM-DD"
  - `StartDate time.Time` - дата начала периодичности

- Методы валидации:
  - `(PeriodicityType).Valid()` - проверка корректности типа
  - Реализованы интерфейсы `Scan()` и `Value()` для работы с JSON в PostgreSQL

### 2. HTTP Layer

**`internal/transport/http/handlers/dto.go`:**
- Добавлены поля `PeriodicityType` и `PeriodicityConfig` в `taskMutationDTO` и `taskDTO`

**`internal/transport/http/handlers/task_handler.go`:**
- Обновлены методы `Create()` и `Update()` для передачи периодичности в usecase

### 3. Usecase Layer

**`internal/usecase/task/ports.go`:**
- Добавлены поля в `CreateInput` и `UpdateInput` для передачи периодичности

**`internal/usecase/task/service.go`:**
- Реализована функция `validatePeriodicityConfig()` с проверками:
  - Для `daily`: требуется положительный `interval` и `start_date`
  - Для `monthly`: требуется `month_day` (1-30) и `start_date`
  - Для `specific_dates`: требуется непустой массив дат в правильном формате
  - Для `even_days` / `odd_days`: требуется `start_date`
- Обновлены методы `validateCreateInput()` и `validateUpdateInput()` для проверки периодичности
- Обновлены `Create()` и `Update()` методы сервиса

### 4. Repository Layer

**`internal/repository/postgres/task_repository.go`:**
- Добавлена сериализация/десериализация `PeriodicityConfig` в JSON
- Обновлены все методы (Create, Update, GetByID, List) для работы с новыми полями
- Реализована функция `scanTask()` с парсингом JSON конфига периодичности

### 5. Database

**Новая миграция `migrations/0002_add_periodicity_to_tasks.up.sql`:**
```sql
ALTER TABLE tasks ADD COLUMN periodicity_type TEXT NOT NULL DEFAULT 'none';
ALTER TABLE tasks ADD COLUMN periodicity_config JSONB NOT NULL DEFAULT '{}';
CREATE INDEX idx_tasks_periodicity_type ON tasks (periodicity_type);
```

### 6. Documentation

**`README.md`:**
- Расширено описание функциональности периодичности
- Добавлены примеры использования для каждого типа периодичности
- Документированы требования к валидации
- Описаны архитектурные решения и предположения

## Архитектурные решения

### 1. Выбор формата хранения (JSONB)
- **Причина**: Гибкость - разные типы периодичности требуют разные наборы параметров
- **Альтернатива**: Отдельные столбцы для каждого параметра (было бы избыточно)

### 2. Значение по умолчанию
- По умолчанию `periodicity_type = "none"` (однократная задача)
- Это сохраняет обратную совместимость - старые задачи продолжают работать

### 3. Ограничение дней месяца (1-30)
- Избегает проблем с месяцами < 31 дней
- Февраль, апрель, июнь, сентябрь, ноябрь работают корректно

### 4. Формат дат
- Используется ISO 8601 (YYYY-MM-DD) для `specific_dates`
- Типовой стандарт, легко парсить и сравнивать

## Примеры использования API

### Ежедневная задача (каждый день)
```bash
POST /api/v1/tasks
{
  "title": "Обзвон пациентов",
  "status": "new",
  "periodicity_type": "daily",
  "periodicity_config": {
    "interval": 1,
    "start_date": "2024-01-15T00:00:00Z"
  }
}
```

### Ежемесячная задача (15-го числа каждого месяца)
```bash
POST /api/v1/tasks
{
  "title": "Отчет",
  "status": "new",
  "periodicity_type": "monthly",
  "periodicity_config": {
    "month_day": 15,
    "start_date": "2024-01-15T00:00:00Z"
  }
}
```

### На конкретные даты
```bash
POST /api/v1/tasks
{
  "title": "Совещания",
  "status": "new",
  "periodicity_type": "specific_dates",
  "periodicity_config": {
    "specific_dates": ["2024-01-15", "2024-02-10", "2024-03-20"]
  }
}
```

### На чётные дни месяца
```bash
POST /api/v1/tasks
{
  "title": "Инспекция",
  "status": "new",
  "periodicity_type": "even_days",
  "periodicity_config": {
    "start_date": "2024-01-02T00:00:00Z"
  }
}
```

## Обработка ошибок

При валидации периодичности возвращаются информативные сообщения об ошибках:
- **400 Bad Request** - при невалидных параметрах периодичности
- **400 Bad Request** - при отсутствии обязательных полей конфигурации

Примеры ошибок:
- "invalid periodicity type"
- "periodicity_config is required when periodicity_type is set"
- "month_day must be between 1 and 30 for monthly periodicity"
- "interval must be positive for daily periodicity"
- "invalid date format in specific_dates: 15-01-2024 (expected YYYY-MM-DD)"

## Граничные случаи

1. **Задача без периодичности** - автоматически устанавливается `periodicity_type: "none"`
2. **Изменение периодичности** - можно изменить через PUT запрос
3. **Удаление периодичности** - достаточно установить `periodicity_type: "none"`
4. **Февраль и 31-день месяцы** - не возникает проблем благодаря ограничению до 30 дней

## Обратная совместимость

- Существующие задачи получат `periodicity_type: "none"` и `periodicity_config: {}`
- Все старые API запросы продолжат работать (новые поля опциональны)
- Миграция БД использует `DEFAULT` значения для существующих строк
