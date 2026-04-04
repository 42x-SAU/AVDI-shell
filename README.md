# AVDI-Shell

<img src="./frontend/contents/logo.png" width="450" alt="Логотип AVDI-Shell">

## Описание

**AVDI-Shell** - это распределённая система диагностики и мониторинга инфраструктуры на базе архитектуры "сервер-агенты". Система позволяет централизованно управлять диагностическими проверками на множестве удалённых машин, собирать результаты и хранить их в базе данных.

Проект написан на **Go 1.25** с использованием:
- **PostgreSQL** для хранения данных
- **Docker** для контейнеризации
- **REST API** для взаимодействия компонентов

## Ключевые возможности

- ✅ Распределённая архитектура с несколькими агентами
- ✅ Безопасная аутентификация агентов через токены
- ✅ Выполнение диагностических команд на удалённых системах
- ✅ Система очереди на базе PostgreSQL
- ✅ Мониторинг статуса агентов через heartbeat
- ✅ Гибкая система конфигурации команд (YAML)
- ✅ Подробное логирование результатов
- ✅ Поддержка таймаутов для выполнения команд
- ✅ Переиспользование и повторные попытки выполнения задач

## Архитектура

### Компоненты системы

```mermaid
graph TD
    Client["🖥️ Client (API Requests)"]
    Server["🔌 REST Server<br/>(Golang :8080)"]
    Queue["📋 Task Queue<br/>(PostgreSQL)"]
    DB[("🗄️ PostgreSQL DB<br/>(Storage & Queue)")]
    
    A1["🤖 Agent 1"]
    A2["🤖 Agent 2"]
    A3["🤖 Agent 3"]
    AN["🤖 Agent N"]
    
    Client -->|POST/GET| Server
    Server <-->|Read/Write| DB
    Server -->|Task Queue| Queue
    Queue -->|Pull Tasks| A1
    Queue -->|Pull Tasks| A2
    Queue -->|Pull Tasks| A3
    Queue -->|Pull Tasks| AN
    A1 -->|Heartbeat| Server
    A2 -->|Heartbeat| Server
    A3 -->|Heartbeat| Server
    AN -->|Heartbeat| Server
    A1 -->|Results| DB
    A2 -->|Results| DB
    A3 -->|Results| DB
    AN -->|Results| DB
```

### Основные компоненты

| Компонент | Описание | Путь |
|-----------|---------|------|
| **Server** | Основной REST сервер, управляет задачами и агентами | `avdi/cmd/server/main.go` |
| **Agent** | Агент, выполняющий диагностические команды | `avdi/cmd/agent/main.go` |
| **Storage** | Слой доступа к PostgreSQL | `avdi/internal/storage/` |
| **Queue** | Система очереди задач на базе PostgreSQL | `avdi/internal/queue/` |
| **Diagnostics** | Модуль для выполнения команд | `avdi/internal/diagnostics/` |
| **Logger** | Логирование системы | `avdi/pkg/logger/` |

## Структура проекта

```
avdi/
├── cmd/                              # Точки входа приложений
│   ├── server/main.go               # REST сервер
│   ├── agent/main.go                # Диагностический агент
│   ├── ctl/main.go                  # CLI утилита управления
│   └── shell/main.go                # Интерактивная shell
│
├── internal/                         # Внутренние пакеты
│   ├── server/                       # Логика сервера
│   │   ├── server.go                # Инициализация и маршруты
│   │   ├── handlers.go              # HTTP обработчики
│   │   └── types.go                 # Типы данных
│   │
│   ├── agent/                        # Логика агента
│   │   ├── agent.go                 # Основной агент
│   │   ├── checks.go                # Диагностические проверки
│   │   └── client.go                # HTTP клиент
│   │
│   ├── storage/                      # Работа с БД
│   │   ├── models.go                # Модели данных
│   │   └── postgres.go              # PostgreSQL драйвер
│   │
│   ├── queue/                        # Система очереди
│   │   └── pgqueue.go               # Queue на базе PG
│   │
│   ├── diagnostics/                  # Выполнение команд
│   │   ├── config.go                # Парсинг конфигурации
│   │   └── executor.go              # Исполнитель команд
│   │
│   └── deploy/                       # Развёртывание
│       └── ssh.go                   # SSH функции
│
├── pkg/                              # Публичные пакеты
│   └── logger/                       # Логирование
│       └── logger.go
│
├── docker/                           # Docker конфигурация
│   ├── agent.Dockerfile             # Образ агента
│   └── server.Dockerfile            # Образ сервера
│
├── diagnostics.yaml                  # Конфигурация команд
├── go.mod                            # Go зависимости
└── go.sum                            # Хеши зависимостей
```

## Установка и настройка

### Требования

- **Go 1.25+** (для разработки)
- **PostgreSQL 13+** (для базы данных)
- **Docker & Docker Compose** (опционально, для контейнеризации)

### Локальная установка

#### 1. Клонирование репозитория

```bash
git clone https://github.com/42x-SAU/AVDI-shell.git
cd AVDI-shell/avdi
```

#### 2. Установка зависимостей

```bash
go mod download
```

#### 3. Настройка базы данных

```bash
# Создание базы данных PostgreSQL
createdb avdi_db

# Инициализация схемы (выполнить SQL скрипты)
psql -d avdi_db -f schema.sql
```

#### 4. Переменные окружения

Создайте файл `.env`:

```env
# Server
SERVER_ADDR=:8080
DATABASE_URL=postgres://user:password@localhost:5432/avdi_db

# Agent
AGENT_NAME=agent-1
SERVER_URL=http://localhost:8080
AGENT_TOKEN=your-secret-token
AGENT_ID=1
```

#### 5. Сборка приложений

```bash
# Сервер
go build -o bin/server ./cmd/server

# Агент
go build -o bin/agent ./cmd/agent

# CLI утилита
go build -o bin/ctl ./cmd/ctl

# Shell
go build -o bin/shell ./cmd/shell
```

#### 6. Запуск

```bash
# В отдельных терминалах:

# Сервер
./bin/server

# Агент (несколько копий)
./bin/agent
./bin/agent
```

### Docker Compose

```bash
# Запуск всей системы
docker-compose up -d

# Просмотр логов
docker-compose logs -f

# Остановка
docker-compose down
```

## Использование

### REST API

#### Health Check

```bash
curl http://localhost:8080/health
```

**Ответ:**
```json
{
  "status": "ok"
}
```

#### Управление агентами

**Список активных агентов:**
```bash
curl http://localhost:8080/agents
```

**Регистрация нового агента:**
```bash
curl -X POST http://localhost:8080/agents/register \
  -H "Content-Type: application/json" \
  -d '{"name":"agent-1"}'
```

**Heartbeat (отправка статуса агентом):**
```bash
curl -X POST http://localhost:8080/agents/heartbeat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{"agent_id":1}'
```

#### Управление задачами

**Создание новой задачи:**
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": 1,
    "check_type": "hostname",
    "payload": ""
  }'
```

**Список задач:**
```bash
curl http://localhost:8080/tasks
```

**Получить следующую задачу для агента:**
```bash
curl http://localhost:8080/agents/tasks/next \
  -H "Authorization: Bearer TOKEN"
```

**Отправить результат выполнения:**
```bash
curl -X POST http://localhost:8080/agents/tasks/result \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    "task_id": 1,
    "exit_code": 0,
    "result_json": "{\"hostname\":\"server1\"}",
    "stdout": "server1",
    "stderr": "",
    "logs": "Execution completed"
  }'
```

#### Просмотр результатов

```bash
curl http://localhost:8080/results
```

### CLI утилита (ctl)

```bash
# Получить список агентов
./bin/ctl agents list

# Создать задачу
./bin/ctl tasks create --agent-id 1 --check-type hostname

# Посмотреть результаты
./bin/ctl results list
```

### Интерактивная Shell

```bash
./bin/shell
```

## Конфигурация диагностических команд

Диагностические команды конфигурируются в файле `diagnostics.yaml`:

```yaml
version: "1.0"

commands:
  - name: "hostname"
    description: "Получить имя хоста"
    command: "hostname"
    args: []
    parse_output: "text"

  - name: "disk-usage"
    description: "Использование дискового пространства"
    command: "df"
    args: ["-h"]
    parse_output: "text"
    timeout: 30

  - name: "memory-usage"
    description: "Использование памяти"
    command: "free"
    args: ["-m"]
    parse_output: "text"

  - name: "ping-target"
    description: "Пинг целевого хоста"
    command: "ping"
    args: ["-c", "4", "{{target}}"]
    variables:
      - name: "target"
        required: true
        default: "8.8.8.8"
```

## Модели данных

### Agent

| Поле | Тип | Описание |
|------|-----|---------|
| `id` | int64 | Уникальный идентификатор |
| `name` | string | Имя агента |
| `token` | string | Токен аутентификации |
| `last_heartbeat` | timestamp | Время последнего heartbeat |
| `created_at` | timestamp | Время создания |

### Task

| Поле | Тип | Описание |
|------|-----|---------|
| `id` | int64 | Уникальный идентификатор |
| `agent_id` | int64 | ID агента-исполнителя |
| `check_type` | string | Тип проверки |
| `payload` | string | Параметры задачи (JSON) |
| `status` | string | Статус (pending, running, completed, failed) |
| `retry_count` | int | Текущее количество повторов |
| `max_retries` | int | Максимальное количество повторов |
| `created_at` | timestamp | Время создания |
| `started_at` | timestamp | Время начала выполнения |
| `finished_at` | timestamp | Время завершения |

### TaskResult

| Поле | Тип | Описание |
|------|-----|---------|
| `id` | int64 | Уникальный идентификатор |
| `task_id` | int64 | ID задачи |
| `exit_code` | int | Код выхода команды |
| `result_json` | string | Распарсенный результат (JSON) |
| `stdout` | string | Стандартный вывод |
| `stderr` | string | Стандартная ошибка |
| `logs` | string | Логи выполнения |
| `created_at` | timestamp | Время создания |

## Поток выполнения

### Жизненный цикл задачи

```mermaid
sequenceDiagram
    participant Client as 📱 Client
    participant Server as 🔌 Server
    participant Queue as 📋 Queue
    participant Agent as 🤖 Agent
    participant DB as 🗄️ Database
    
    Client->>Server: POST /tasks (создание)
    Server->>DB: INSERT task
    Server->>Queue: Add to queue
    Queue-->>Server: Task enqueued
    
    Agent->>Server: GET /agents/tasks/next
    Server->>Queue: Get next task
    Queue-->>Server: Task data
    Server-->>Agent: Task (check_type, payload)
    
    Agent->>Agent: Выполнение команды
    Agent->>Agent: Сбор результата
    
    Agent->>Server: POST /agents/tasks/result
    Server->>DB: INSERT result
    
    Client->>Server: GET /results
    Server->>DB: SELECT results
    DB-->>Server: Results data
    Server-->>Client: JSON results
```

### Регистрация агента

```mermaid
sequenceDiagram
    participant Agent as 🤖 Agent
    participant Server as 🔌 Server
    participant DB as 🗄️ Database
    
    Agent->>Agent: Agent.Run() запуск
    Agent->>Server: POST /agents/register (name)
    Server->>DB: INSERT agent
    DB-->>Server: id, token
    Server-->>Agent: id, token
    
    Note over Agent: Сохранение credentials
    
    loop Every 15 seconds
        Agent->>Server: POST /agents/heartbeat
        Server->>DB: UPDATE last_heartbeat
        DB-->>Server: OK
    end
    
    Agent->>Server: GET /agents/tasks/next
    Server->>DB: SELECT pending task
    DB-->>Server: Task
    Server-->>Agent: Task ready
    Note over Agent: Agent готов к работе
```

## Docker

### Структура Dockerfile

#### Agent (agent.Dockerfile)

```dockerfile
# Multi-stage build для минимизации размера
FROM golang:1.25-alpine AS builder
# Сборка приложения

FROM alpine:3.20
# Базовый образ с необходимыми утилитами:
# - iputils (ping, traceroute)
# - iproute2 (ip команды)
# - net-tools (netstat и т.д.)
# - nmap (сканирование портов)
# - curl (HTTP запросы)
# - bind-tools (DNS утилиты)
```

### Запуск через Docker

```bash
# Сборка образа агента
docker build -f docker/agent.Dockerfile -t avdi-agent .

# Запуск контейнера
docker run --rm \
  -e SERVER_URL=http://server:8080 \
  -e AGENT_NAME=agent-1 \
  avdi-agent
```

## Система мониторинга и логирование

### Логирование

Система использует стандартный пакет `log` Go. Логи выводятся в stdout/stderr.

### Heartbeat мониторинг

- Каждый агент отправляет heartbeat каждые **15 секунд**
- Сервер обновляет `last_heartbeat` для каждого агента
- Инактивные агенты можно идентифицировать по времени последнего heartbeat

### Таймауты

- Диагностические команды выполняются с опциональным таймаутом
- При превышении таймаута процесс прерывается
- Результат содержит информацию об ошибке

## Расширение функционала

### Добавление новой диагностической команды

1. Добавьте команду в `diagnostics.yaml`:

```yaml
  - name: "my-check"
    description: "Моя проверка"
    command: "mycommand"
    args: ["arg1", "arg2"]
    parse_output: "json"
    timeout: 60
```

2. В агенте будет автоматически доступна при создании задачи с `check_type: "my-check"`

### Добавление нового HTTP endpoint

Отредактируйте `internal/server/server.go`:

```go
func (s *Server) routes() {
    // Существующие маршруты...
    
    // Новый маршрут
    s.mux.HandleFunc("POST /custom/endpoint", s.handleCustomEndpoint)
}

func (s *Server) handleCustomEndpoint(w http.ResponseWriter, r *http.Request) {
    // Ваша логика здесь
}
```

## Безопасность

### Аутентификация

- Каждый агент получает уникальный **токен** при регистрации
- Все защищённые операции требуют токена в заголовке `Authorization: Bearer TOKEN`
- Токены не логируются

### HTTPS

Для production окружения рекомендуется:
- Использовать HTTPS вместо HTTP
- Добавить TLS сертификаты
- Использовать переменные окружения для чувствительных данных

### База данных

- Пароли PostgreSQL должны храниться в секретах
- Используйте SSL подключение к БД
- Регулярные резервные копии данных

## Разработка

### Требования для разработки

```bash
# Go 1.25
go version

# PostgreSQL CLI
psql --version

# Git
git --version
```

### Запуск тестов

```bash
cd avdi
go test ./...
```

### Code Style

Проект следует стандартам Go:

```bash
# Форматирование кода
go fmt ./...

# Статический анализ
go vet ./...

# Линтинг (если установлен golangci-lint)
golangci-lint run
```

## Troubleshooting

### Агент не может подключиться к серверу

**Проблема:** `connection refused`

**Решение:**
- Проверьте, запущен ли сервер: `curl http://localhost:8080/health`
- Убедитесь, что `SERVER_URL` указывает на правильный адрес
- Проверьте файрволл

### PostgreSQL: unable to connect

**Проблема:** `FATAL: host is invalid`

**Решение:**
```bash
# Проверьте подключение
psql -h localhost -U postgres -d avdi_db

# Проверьте переменную DATABASE_URL
echo $DATABASE_URL
```

### Большое количество задач в очереди

**Решение:**
- Добавьте больше агентов
- Оптимизируйте время выполнения команд
- Проверьте логи агентов: `docker logs [agent-container]`

### Задача зависает/не выполняется

**Проверка:**
- Посмотрите статус: `curl http://localhost:8080/tasks`
- Проверьте логи агента: `docker logs -f [agent-container]`
- Убедитесь, что agentID в задаче соответствует реальному агенту

## Производительность

### Рекомендации для масштабирования

| Метрика | Рекомендация |
|---------|-------------|
| **Количество агентов** | Начните с 5-10, масштабируйте по необходимости |
| **Количество ВМ** | 1 на 100+ агентов |
| **PostgreSQL** | Используйте отдельный сервер для >500 агентов |
| **Connection pool** | Настройте для вашей нагрузки |
| **Heartbeat интервал** | По умолчанию 15 сек, оптимизируйте для вашей сети |

## Лицензия

Укажите лицензию вашего проекта.

## Контакты и поддержка

- **Issues:** [GitHub Issues](https://github.com/42x-SAU/AVDI-shell/issues)
- **Discussions:** [GitHub Discussions](https://github.com/42x-SAU/AVDI-shell/discussions)

## Благодарности

Спасибо всем контрибьютерам, помогающим развивать этот проект!
