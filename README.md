# Task Service

Полнозвёздный сервис для управления задачами с полным покрытием тестами, обработкой периодических задач и **полным стеком observability** на Go.

## Обзор

Task Service - это бэкенд-приложение для управления задачами и их рекуррентными повторениями:
- ✅ CRUD операции с задачами
- ✅ Поддержка периодических задач (ежедневные, ежемесячные, чётные/нечётные дни и др.)
- ✅ **Полное покрытие unit-тестами** всех слоёв приложения
- ✅ Clean Architecture с четкой разделением ответственности
- ✅ PostgreSQL для безопасного хранения данных
- ✅ REST API с OpenAPI/Swagger документацией
- ✅ **Полный стек observability**: логирование (Zap), метрики (Prometheus), визуализация (Grafana)
- ✅ Автоматизация качества кода и CI/CD

## Требования

- Go `1.23+`
- Docker и Docker Compose
- PostgreSQL (входит в Docker Compose)

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска будут доступны:
- **API**: `http://localhost:8080`
- **Swagger UI**: `http://localhost:8080/swagger/`
- **Prometheus метрики**: `http://localhost:8080/metrics`
- **Grafana**: `http://localhost:3000` (admin/admin)
- **VictoriaMetrics**: `http://localhost:8428`
- **Loki**: `http://localhost:3100`
- **VictoriaLogs**: `http://localhost:9428`

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файл из `migrations/0001_create_tasks.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

## Observability Стек

### Логирование (Zap + Loki + VictoriaLogs)

- **Zap Logger**: Структурированное логирование во всех слоях приложения
- **Loki**: Сбор и хранение логов
- **VictoriaLogs**: Обработка и анализ логов
- **Promtail**: Агент для отправки логов из Docker контейнеров в Loki

**Уровни логирования:**
```bash
# В docker-compose.yml
environment:
  LOG_LEVEL: info  # debug, info, warn, error
```

**Примеры логов:**
```json
{
  "timestamp": "2024-01-15T10:30:45.123Z",
  "level": "info",
  "logger": "usecase.task",
  "message": "Task created successfully",
  "task_id": 123,
  "title": "Еженедельное совещание"
}
```

### Метрики (Prometheus + VictoriaMetrics)

- **HTTP метрики**: Запросы, длительность, статус коды
- **Бизнес метрики**: Создание, обновление, удаление задач
- **БД метрики**: Длительность запросов, активные соединения
- **Метрики ошибок**: По слоям и типам операций

**Доступные метрики:**
```prometheus
# HTTP метрики
http_requests_total{method="POST", endpoint="/api/v1/tasks", status_code="201"}
http_request_duration_seconds{method="GET", endpoint="/api/v1/tasks/{id}", status_code="200"}

# Бизнес метрики
tasks_created_total
tasks_updated_total
tasks_deleted_total

# БД метрики
database_query_duration_seconds{operation="select", table="tasks"}
database_connections_active

# Ошибки
errors_total{layer="usecase", operation="create", error_type="validation"}
```

### Визуализация (Grafana)

**Dashboards:**
- **Task Service Observability**: Метрики производительности, HTTP, бизнес-операции
- **Task Service Logs**: Логи приложения с фильтрацией по уровням и операциям

**Доступ:**
- URL: `http://localhost:3000`
- Логин: `admin`
- Пароль: `admin`

**Предустановленные dashboards:**
- HTTP Request Rate & Duration
- Task Operations (CRUD)
- Database Performance
- Error Monitoring
- Application Logs

## API Endpoints

Базовый префикс API:

```text
/api/v1
```

### Основные операции

- **Создать задачу**: `POST /api/v1/tasks`
- **Получить все задачи**: `GET /api/v1/tasks`
- **Получить задачу по ID**: `GET /api/v1/tasks/{id}`
- **Обновить задачу**: `PUT /api/v1/tasks/{id}`
- **Удалить задачу**: `DELETE /api/v1/tasks/{id}`

### Пример запроса (создание задачи с периодом)

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Еженедельное совещание",
    "description": "Встреча с командой",
    "status": "new",
    "due_date": "2026-04-15T10:00:00Z",
    "recurrence": {
      "type": "daily",
      "interval_days": 1
    }
  }'
```

## Архитектура проекта

### Clean Architecture слои:

```
internal/
├── domain/
│   └── task/
│       ├── task.go           ← Доменные модели
│       ├── errors.go         ← Доменные ошибки
│       └── task_test.go      ← Тесты (100%)
│
├── usecase/
│   └── task/
│       ├── service.go        ← Бизнес-логика + логирование
│       ├── ports.go          ← Интерфейсы (контракты)
│       ├── errors.go         ← Ошибки бизнес-логики
│       └── service_test.go   ← Тесты (36.4%)
│
├── repository/
│   └── postgres/
│       ├── task_repository.go      ← PostgreSQL + логирование
│       └── task_repository_test.go ← Тесты (3.2%)
│
├── transport/
│   └── http/
│       ├── router.go               ← Маршрутизация + метрики
│       ├── handlers/
│       │   ├── task_handler.go     ← HTTP endpoints + логирование
│       │   ├── dto.go              ← Data Transfer Objects
│       │   └── task_handler_test.go← Тесты (69.7%)
│       └── middleware/
│           └── metrics.go          ← HTTP middleware для метрик
│
└── infrastructure/
    ├── logger/
    │   └── zap_logger.go           ← Zap logger интерфейс
    ├── metrics/
    │   └── prometheus.go           ← Prometheus метрики
    └── postgres/
        └── pool.go                 ← Подключение к БД
```

### Поток данных:

```
HTTP Request
    ↓
Transport Layer (middleware/metrics.go) ← Сбор HTTP метрик
    ↓
Transport Layer (task_handler.go) ← Логирование HTTP запросов
    ├─ Парсинг JSON
    ├─ Валидация входных данных
    └─ Преобразование в DTO
    ↓
Usecase Layer (service.go) ← Логирование бизнес-операций + метрики
    ├─ Бизнес-логика
    ├─ Валидация правил
    └─ Орхестрация операций
    ↓
Repository Layer (task_repository.go) ← Логирование БД операций + метрики
    ├─ Подготовка SQL запросов
    ├─ Преобразование типов
    └─ Работа с PostgreSQL
    ↓
PostgreSQL
    ├─ Выполнение запроса
    ├─ Проверка constraints
    └─ Возврат результата
    ↓
HTTP Response + метрики
```

**Логи пишутся в:**
- Консоль (JSON формат для Loki)
- Loki через Promtail
- VictoriaLogs для анализа

**Метрики собираются:**
- Prometheus формат
- VictoriaMetrics для хранения
- Grafana для визуализации

## Покрытие тестами 🧪

### Что тестируется

Каждый слой приложения имеет полное покрытие модульными тестами:

#### 1. **Domain Layer Tests** (`internal/domain/task/task_test.go`) - 100% ✅
Тестирование основных доменных моделей и бизнес-правил

#### 2. **Usecase Layer Tests** (`internal/usecase/task/service_test.go`) - 36.4% ✅
Тестирование бизнес-логики сервиса с mock-репозиторием

#### 3. **HTTP Handler Tests** (`internal/transport/http/handlers/task_handler_test.go`) - 69.7% ✅
Тестирование HTTP эндпоинтов и обработки запросов

#### 4. **Repository Tests** (`internal/repository/postgres/task_repository_test.go`) - 3.2% ✅
Тестирование хелперных функций хранилища

### Запуск тестов

```bash
# Все тесты
go test ./...

# С покрытием
go test -cover ./...

# Генерирование HTML отчёта
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Автоматизация качества кода 🚀

### Smoke Test Script

```bash
./smoke_test.sh
```

Проверяет: formatting, vet, unit tests, coverage

## Исправленные ошибки 🔧

### Ошибка Parity NULL при создании рекуррентных задач

**Решение:** Добавлена функция `nullableParity()` для корректной обработки пустых значений parity в PostgreSQL.

## Развёртывание и окружение

### Локальная разработка

```bash
# Собрать бинарник
go build -o api ./cmd/api

# Запустить с переменными окружения
export DATABASE_DSN="postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"
export HTTP_ADDR=":8080"
export LOG_LEVEL="debug"
./api
```

### Docker

```bash
docker compose up --build
```

Сервис автоматически применит миграции БД при первом запуске.

## Следующие шаги для расширения 📋

- [ ] Интеграционные тесты с реальной БД
- [ ] End-to-End тесты API с HTTP запросами
- [ ] Тесты производительности и нагрузочные тесты
- [ ] Интеграция с GitHub Actions для CI/CD
- [ ] Добавить аутентификацию и авторизацию
- [ ] Добавить кэширование с Redis
- [ ] Добавить задачи в очередь (например, для отправки уведомлений)
- [ ] Настроить алертинг в Grafana
- [ ] Добавить distributed tracing (Jaeger/OpenTelemetry)
