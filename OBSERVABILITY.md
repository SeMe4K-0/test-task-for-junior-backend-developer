# Observability и Логирование в Task Service

Полная документация по внедрённому решению для наблюдаемости приложения.

---

## 📊 Обзор

В проект добавлена полная трёхуровневая система observability (наблюдаемости):

1. **Логирование** - Zap + Loki + Grafana
2. **Метрики** - Prometheus + VictoriaMetrics + Grafana
3. **Middleware** - HTTP метрики для всех запросов

Все компоненты автоматически развёртываются через Docker Compose.

---

## 🛠️ Архитектура

### Компоненты системы

```
┌─────────────────────────────────────────────────────────────┐
│                                                               │
│  Task Service App (Port 8080)                               │
│  ├─ Zap Logger (структурированные логи в JSON)             │
│  ├─ Prometheus Metrics (эндпоинт /metrics)                 │
│  └─ HTTP Middleware (автоматический сбор метрик)           │
│                                                               │
└─────────────────────────────────────────────────────────────┘
          │                          │
          ▼                          ▼
    ┌──────────────┐       ┌──────────────────┐
    │ Promtail     │       │ Prometheus       │
    │ (логи)       │       │ (метрики)        │
    └──────┬───────┘       └────────┬─────────┘
           │                        │
           ▼                        ▼
    ┌──────────────┐       ┌──────────────────┐
    │ Loki         │       │ VictoriaMetrics  │
    │ (хранилище   │       │ (временные ряды) │
    │  логов)      │       │                  │
    └──────┬───────┘       └────────┬─────────┘
           │                        │
           ├────────────┬────────────┤
           │            │            │
           ▼            ▼            ▼
    ┌─────────────────────────────────────┐
    │        Grafana (Port 3000)          │
    │  - Визуализация логов               │
    │  - Визуализация метрик              │
    │  - Dashboard с KPI                  │
    │  - Alerting                         │
    └─────────────────────────────────────┘
```

---

## 🔍 Зап Логирование

### Что было добавлено

**Файл:** `internal/infrastructure/logger/zap_logger.go`

Структурированное логирование с использованием Zap библиотеки:

```go
type Logger interface {
    Debug(msg string, fields ...zap.Field)
    Info(msg string, fields ...zap.Field)
    Warn(msg string, fields ...zap.Field)
    Error(msg string, fields ...zap.Field)
    Fatal(msg string, fields ...zap.Field)
    With(fields ...zap.Field) Logger
}
```

### Форматы логов

Логи выводятся в **JSON** формате для удобства обработки:

```json
{
  "timestamp": "2026-04-08T11:30:45.123Z",
  "level": "info",
  "caller": "usecase/task/service.go:45",
  "message": "Task created successfully",
  "task_id": 123,
  "title": "Update documentation"
}
```

### Уровни логирования

- **debug** - Детальная информация для отладки
- **info** - Информационные сообщения о важных событиях
- **warn** - Предупреждения о потенциальных проблемах
- **error** - Ошибки в операциях
- **fatal** - Критические ошибки

### Использование

```bash
# Запуск с разными уровнями логирования
export LOG_LEVEL=debug
export LOG_LEVEL=info    # по умолчанию
export LOG_LEVEL=warn
export LOG_LEVEL=error
```

### Где добавлено логирование

#### 1. **Domain Layer** (не требует логирования)
- Чистая бизнес-логика без побочных эффектов

#### 2. **Usecase Layer** - `internal/usecase/task/service.go`

```go
// Create операция
s.logger.Info("Creating new task", zap.String("title", title))
// ... выполнение...
s.logger.Info("Task created successfully", zap.Int64("task_id", id))
s.logger.Error("Failed to create task", zap.Error(err))

// GetByID операция
s.logger.Debug("Getting task by ID", zap.Int64("task_id", id))
s.logger.Warn("Task not found", zap.Int64("task_id", id))

// Update операция
s.logger.Info("Updating task", zap.Int64("task_id", id))

// Delete операция
s.logger.Info("Task deleted successfully", zap.Int64("task_id", id))

// List операция
s.logger.Debug("Listing all tasks")
s.logger.Debug("Tasks listed", zap.Int("count", len(tasks)))
```

#### 3. **Repository Layer** - `internal/repository/postgres/task_repository.go`

```go
// Create
r.logger.Debug("Creating task in database", zap.String("title", title))
r.logger.Info("Task created successfully", 
    zap.Int64("task_id", id), 
    zap.Duration("duration", duration))

// GetByID
r.logger.Debug("Getting task by ID", zap.Int64("task_id", id))
r.logger.Error("Failed to get task", zap.Error(err))

// Метрики БД
metrics.RecordDatabaseQuery("select", "tasks", duration)
```

#### 4. **HTTP Handlers** - `internal/transport/http/handlers/task_handler.go`

```go
// Create handler
h.logger.Info("HTTP request: Create task", 
    zap.String("method", r.Method),
    zap.String("path", r.URL.Path))
h.logger.Error("Failed to decode JSON", zap.Error(err))
h.logger.Info("Task created via HTTP", zap.Int64("task_id", id))
```

---

## 📈 Prometheus Метрики

### Что было добавлено

**Файл:** `internal/infrastructure/metrics/prometheus.go`

Сбор метрик в формате Prometheus:

#### HTTP Метрики

```go
// Длительность HTTP запросов
http_request_duration_seconds{method="POST", endpoint="/api/v1/tasks", status_code="201"}

// Количество HTTP запросов
http_requests_total{method="GET", endpoint="/api/v1/{id}", status_code="200"}
```

#### Бизнес Метрики

```go
// Количество созданных задач
tasks_created_total

// Количество обновлённых задач  
tasks_updated_total

// Количество удалённых задач
tasks_deleted_total
```

#### Метрики БД

```go
// Длительность запросов к БД
database_query_duration_seconds{operation="select", table="tasks"}

// Количество активных соединений
database_connections_active
```

#### Метрики Ошибок

```go
// Общее количество ошибок по слоям
errors_total{layer="usecase", operation="create", error_type="validation"}
errors_total{layer="repository", operation="query", error_type="timeout"}
```

### Эндпоинт метрик

```
http://localhost:8080/metrics
```

Возвращает текстовый формат Prometheus:

```
# HELP http_request_duration_seconds Duration of HTTP requests in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="POST",endpoint="/api/v1/tasks",status_code="201",le="0.005"} 5
http_request_duration_seconds_bucket{method="POST",endpoint="/api/v1/tasks",status_code="201",le="0.01"} 8
...
```

---

## 🚦 HTTP Middleware для метрик

**Файл:** `internal/transport/http/middleware/metrics.go`

Автоматический сбор метрик для всех HTTP запросов:

```go
// Middleware автоматически:
// 1. Записывает время начала запроса
// 2. Оборачивает ResponseWriter для захвата статус кода
// 3. Вызывает следующий обработчик
// 4. Записывает метрики о длительности и статусе

func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        // ... обработка ...
        duration := time.Since(start)
        metrics.RecordHTTPRequest(r.Method, endpoint, rw.statusCode, duration)
    })
}
```

### Регистрация в маршрутизаторе

**Файл:** `internal/transport/http/router.go`

```go
router.Use(middleware.MetricsMiddleware)
router.Handle("/metrics", promhttp.Handler())
```

---

## 🗂️ Docker Compose Стек

### Компоненты

| Сервис | Образ | Порт | Назначение |
|--------|-------|------|-----------|
| postgres | postgres:16-alpine | 5433 | База данных |
| app | custom | 8080 | Основное приложение |
| loki | grafana/loki:2.8.2 | 3100 | Сбор логов |
| victoriametrics | victoriametrics/victoria-metrics:v1.101.0 | 8428 | Хранилище метрик |
| vmagent | victoriametrics/vmagent:v1.101.0 | - | Scraper для метрик |
| promtail | grafana/promtail:3.0.0 | - | Агент отправки логов |
| grafana | grafana/grafana:10.4.0 | 3000 | Визуализация |

> Для управления PostgreSQL через веб-интерфейс доступен также сервис pgAdmin: `http://localhost:5050`. См. `PGADMIN.md` для пошаговой инструкции.

### Запуск

```bash
# Запустить весь стек
docker compose up --build

# Запустить в фоне
docker compose up --build -d

# Остановить
docker compose down

# Очистить volumes (удалить данные)
docker compose down -v
```

### Проверка статуса

```bash
# Показать все контейнеры
docker compose ps

# Посмотреть логи конкретного сервиса
docker compose logs -f app
docker compose logs -f grafana
docker compose logs -f loki
```

---

## 🔌 Конфигурация сервисов

### Loki Configuration

**Файл:** `observability/loki-config.yml`

```yaml
auth_enabled: false

server:
  http_listen_port: 3100

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
```

### Promtail Configuration

**Файл:** `observability/promtail-config.yml`

```yaml
clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
```

Promtail автоматически собирает логи из Docker контейнеров и отправляет в Loki.

---

## 💻 Grafana

### Доступ

```
URL: http://localhost:3000
Username: admin
Password: admin
```

### Data Sources (настроены автоматически)

1. **Loki** - http://loki:3100
   - Источник логов
   - Язык запросов: LogQL

2. **VictoriaMetrics** - http://victoriametrics:8428
   - Источник метрик
   - Язык запросов: MetricsQL (совместим с PromQL)

3. **Loki** - http://loki:3100
   - Источник логов
   - Язык запросов: LogQL

### Provisioning

**Директория:** `observability/grafana/provisioning/`

Автоматическая настройка:
- Data sources
- Dashboards
- Alerts

---

## 🔍 Как использовать

### 1. Просмотр логов в Grafana

#### LogQL запросы

```sql
-- Все логи приложения
{container="app"}

-- Только ошибки
{container="app"} | json | level="error"

-- Логи создания задач
{container="app"} | json | message="Task created successfully"

-- Логи за последний час
{container="app"} | json | timestamp>now-1h
```

#### Фильтры и анализ

```sql
-- Количество ошибок по операции
{container="app"} | json | level="error" | stats count() by operation

-- Средняя длительность запросов к БД
{container="app"} | json | message="Task created successfully" 
  | json duration_ms | stats avg(duration_ms) by operation
```

### 2. Просмотр метрик в Grafana

#### PromQL/MetricsQL запросы

```sql
-- Количество запросов в секунду
rate(http_requests_total[5m])

-- Средняя длительность запросов
histogram_quantile(0.95, http_request_duration_seconds)

-- Созданные задачи за последний час
increase(tasks_created_total[1h])

-- Ошибки по типам
sum(errors_total) by (error_type)

-- Процент успешных запросов
(http_requests_total{status_code=~"2.."}) / (http_requests_total) * 100
```

### 3. Просмотр raw метрик

```bash
# Через curl
curl http://localhost:8080/metrics

# Через prometheus UI
http://localhost:9090

# Через victoria metrics ui
http://localhost:8428
```

### 4. Просмотр raw логов

Логи доступны через Grafana и Loki.

- Loki UI: `http://localhost:3100`
- Grafana: `http://localhost:3000`

Поиск и группировка логов выполняйте через LogQL в Grafana или прямой запрос к Loki.

---

## 📊 Пример Dashboard

### Ключевые KPI для мониторинга

1. **HTTP метрики**
   - Requests per second (RPS)
   - Error rate (%)
   - P50, P95, P99 latency

2. **Бизнес метрики**
   - Tasks created/updated/deleted per hour
   - Error distribution by type

3. **Система**
   - DB connection pool size
   - DB query latency
   - Memory usage

### SQL для графиков

```sql
-- Requests per second по эндпоинтам
rate(http_requests_total[1m]) group by (endpoint)

-- Error rate
(sum(rate(http_requests_total{status_code=~"5.."}[1m]))) / 
(sum(rate(http_requests_total[1m]))) * 100

-- Tasks created per hour
increase(tasks_created_total[1h])

-- Database query latency
histogram_quantile(0.95, rate(database_query_duration_seconds_bucket[5m]))
```

---

## 🚨 Алёрты

### Где их настраивать

В Grafana:
1. Перейти в Alert Rulesadd
2. Добавить правило мониторинга
3. Установить threshold
4. Выбрать notification channel (email, slack, webhook)

### Примеры правил

**High Error Rate**
```
Condition: errors_total / http_requests_total > 0.1
Duration: 5 minutes
Message: "Error rate exceeded 10%"
```

**High Latency**
```
Condition: histogram_quantile(0.95, http_request_duration_seconds) > 1
Duration: 5 minutes
Message: "P95 latency exceeded 1 second"
```

**Database Issues**
```
Condition: database_connections_active > 20
Duration: 2 minutes
Message: "Too many active DB connections"
```

---

## 🔌 Интеграция в коде

### Инициализация логгера

**Файл:** `cmd/api/main.go`

```go
// Создание логгера
zapLogger, err := logger.NewZapLogger(getEnv("LOG_LEVEL", "info"))
if err != nil {
    slog.Error("Failed to initialize logger", "error", err)
    os.Exit(1)
}

// Передача во все слои
taskRepo := postgresrepo.New(pool, zapLogger)
taskUsecase := task.NewService(taskRepo, zapLogger)
taskHandler := httphandlers.NewTaskHandler(taskUsecase, zapLogger)
```

### Использование логгера в коде

```go
// В usecase
s.logger.Info("Creating task", 
    zap.String("title", input.Title),
    zap.String("status", string(input.Status)))

// В repository
r.logger.Debug("Executing query", 
    zap.String("query", "SELECT * FROM tasks"))

// В handlers
h.logger.Error("Request failed", 
    zap.Error(err),
    zap.String("path", r.URL.Path))
```

### Запись метрик

```go
import "example.com/taskservice/internal/infrastructure/metrics"

// Запись HTTP метрик
metrics.RecordHTTPRequest(r.Method, endpoint, statusCode, duration)

// Запись бизнес метрик
metrics.RecordTaskOperation("create")

// Запись ошибок
metrics.RecordError("usecase", "create", "validation")

// Запись метрик БД
metrics.RecordDatabaseQuery("select", "tasks", duration)
```

---

## 🧪 Тестирование

### Проверка логирования

```bash
# Посмотреть логи приложения
docker compose logs -f app

# Фильтровать по ключевому слову
docker compose logs -f app | grep "error"

# Посмотреть логи Loki
docker compose logs -f loki

# Посмотреть логи Promtail
docker compose logs -f promtail
```

### Проверка метрик

```bash
# Получить метрики
curl -s http://localhost:8080/metrics | grep http_requests_total

# Фильтровать по типу
curl -s http://localhost:8080/metrics | grep "# HELP"
```

### Проверка Grafana

1. Открыть http://localhost:3000
2. Перейти в раздел "Logs"
3. Написать LogQL запрос
4. Увидеть логи в реальном времени

---

## 📋 Чек-лист развёртывания

- [ ] Docker и Docker Compose установлены
- [ ] Запущена команда `docker compose up --build`
- [ ] Все контейнеры в статусе "healthy"
- [ ] Приложение доступно на http://localhost:8080
- [ ] Grafana доступна на http://localhost:3000
- [ ] Лоkи доступны на http://localhost:3100
- [ ] Метрики доступны на http://localhost:8080/metrics
- [ ] Data sources в Grafana настроены
- [ ] Создан первый dashboard

---

## 🛠️ Трубл-шутинг

### Контейнеры не запускаются

```bash
# Проверить статус
docker compose ps

# Посмотреть логи
docker compose logs app

# Перестартовать
docker compose restart
```

### Логи не собираются

```bash
# Проверить Promtail
docker compose logs promtail

# Проверить Loki
docker compose logs loki

# Проверить соединение
docker compose exec app curl -s http://loki:3100/loki/api/v1/labels
```

### Метрики не видны

```bash
# Проверить эндпоинт метрик
curl -s http://localhost:8080/metrics

# Проверить VictoriaMetrics
docker compose logs victoriametrics

# Проверить соединение prometheus
curl -s http://localhost:8428/api/v1/targets
```

### Grafana не подключается к источникам

1. Проверить статус сервисов: `docker compose ps`
2. Перезагрузить Grafana: `docker compose restart grafana`
3. Проверить конфигурацию sources в `observability/grafana/provisioning/`

---

## 📚 Полезные ссылки

### Документация
- [Zap Logger](https://pkg.go.dev/go.uber.org/zap)
- [Prometheus Client](https://github.com/prometheus/client_golang)
- [Grafana Docs](https://grafana.com/docs/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [VictoriaMetrics](https://docs.victoriametrics.com/)
- [Grafana Loki](https://grafana.com/docs/loki/)

### Query Languages
- [LogQL](https://grafana.com/docs/loki/latest/logql/) - язык для Loki
- [PromQL](https://prometheus.io/docs/prometheus/latest/querying/basics/) - язык для метрик
- [MetricsQL](https://docs.victoriametrics.com/MetricsQL.html) - расширение PromQL

---

## 📝 Резюме

Успешно внедрена полная система observability с:

✅ **Логирование**
- Zap структурированное логирование
- Сбор и хранение логов через Promtail/Loki
- Визуализация в Grafana

✅ **Метрики**
- Prometheus format
- HTTP метрики через middleware
- Бизнес метрики
- Хранилище VictoriaMetrics

✅ **Визуализация**
- Grafana dashboard
- Real-time мониторинг
- LogQL и MetricsQL запросы
- Alerting возможности

✅ **Docker Compose**
- Полная автоматизация развёртывания
- Health checks
- Persistent volumes
- Network для internal communication

Система готова к production и обеспечивает полную наблюдаемость приложения.
