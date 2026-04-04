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

## Оглавление

- [Описание](#описание)
- [Ключевые возможности](#ключевые-возможности)
- [Архитектура](#архитектура)
  - [Компоненты системы](#компоненты-системы)
  - [Основные компоненты](#основные-компоненты)
- [Структура проекта](#структура-проекта)
- [Установка и настройка](#установка-и-настройка)
- [Использование](#использование)
  - [REST API](#rest-api)
  - [Интерактивная Shell](#интерактивная-shell)
- [Конфигурация диагностических команд](#конфигурация-диагностических-команд)
- [Модели данных](#модели-данных)
- [Поток выполнения](#поток-выполнения)
- [Docker](#docker)
- [Система мониторинга и логирование](#система-мониторинга-и-логирование)
- [Расширение функционала](#расширение-функционала)
- [Безопасность](#безопасность)
- [Разработка](#разработка)
- [Troubleshooting](#troubleshooting)
- [Производительность](#производительность)
- [Лицензия](#лицензия)
- [Контакты и поддержка](#контакты-и-поддержка)
- [Благодарности](#благодарности)

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

# Shell (интерактивный интерфейс)
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

### Интерактивная Shell

Интерактивная shell - это основной интерфейс для работы с AVDI системой, предоставляющий удобный командный интерфейс:

```bash
./bin/shell
```

**Доступные команды в shell:**

| Команда | Описание |
|---------|---------|
| `health` | Проверить статус сервера |
| `agents` | Показать список активных агентов |
| `tasks` | Показать список задач |
| `results [--logs]` | Показать результаты выполнения задач |
| `deploy-agent --ssh-host <хост> --ssh-user <пользователь> --server-url <url> --agent-name <имя> --image <образ>` | **Развернуть агента на удалённом хосте по SSH (автоматическая регистрация)** |
| `create-task --agent <id> --check <тип> [--payload <данные>]` | Создать новую диагностическую задачу |
| `diagnostic-list` | Показать доступные диагностические команды |
| `diagnostic-run --command <имя> [--var ключ=значение]` | Выполнить диагностическую команду локально |
| `server <url>` | Изменить URL активного сервера |
| `help` | Показать справку |
| `exit` или `quit` | Выход из shell |

**Примеры использования deploy-agent:**

```bash
# Развернуть агент с интерактивным вводом пароля
deploy-agent --ssh-host 192.168.1.100 \
  --ssh-user root \
  --server-url http://localhost:8081 \
  --agent-name prod-agent1 \
  --image avdi:latest

# С явно заданным паролем
deploy-agent --ssh-host 192.168.1.100 \
  --ssh-user root \
  --ssh-password mypassword \
  --ssh-port 2222 \
  --server-url http://10.0.0.1:8081 \
  --agent-name agent1 \
  --image localhost:5000/avdi:latest \
  --container my-agent \
  --skip-pull
```

**Переменные окружения для SSH развёртывания:**

```bash
export AVDI_SSH_PASSWORD="пароль"  # Пароль SSH (альтернатива --ssh-password)
export AVDI_SSH_PORT=2222          # Порт SSH (если не 22)

./bin/shell
# теперь deploy-agent не будет запрашивать пароль
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

### Подготовка Docker images

Перед использованием `deploy-agent` для развёртывания агентов на удалённых хостах необходимо создать и опубликовать Docker images.

#### 1. Сборка образа агента локально

```bash
cd /path/to/avdi

# Базовая сборка с локальным тегом
docker build -f docker/agent.Dockerfile -t avdi-agent:latest .

# Проверка что образ собран
docker images | grep avdi-agent
```

#### 2. Работа с Docker Registry

**Вариант A: Docker Hub**

```bash
# Логин в Docker Hub
docker login

# Пересоздаём образ с тегом Docker Hub (username->ваше имя пользователя)
docker build -f docker/agent.Dockerfile -t username/avdi-agent:latest .

# Отправляем образ в Docker Hub
docker push username/avdi-agent:latest

# При deploy-agent используем:
# deploy-agent ... --image username/avdi-agent:latest
```

**Вариант B: Приватный Docker Registry**

```bash
# Пересоздаём образ с тегом приватного registry (registry.example.com->ваш registry)
docker build -f docker/agent.Dockerfile -t registry.example.com/avdi-agent:latest .

# Отправляем образ в приватный registry
docker push registry.example.com/avdi-agent:latest

# При deploy-agent используем:
# deploy-agent ... --image registry.example.com/avdi-agent:latest
```

**Вариант C: Локальная загрузка образа**

```bash
# Сохраняем образ в файл tar
docker save avdi-agent:latest -o avdi-agent.tar

# На удалённом хосте загружаем образ
docker load -i avdi-agent.tar

# При deploy-agent используем флаг --skip-pull (так как образ уже на хосте):
# deploy-agent ... --image avdi-agent:latest --skip-pull
```

#### 3. Сборка образа сервера (опционально)

```bash
# Образ сервера для тестирования локально или в Docker Compose
docker build -f docker/server.Dockerfile -t avdi-server:latest .

# Запуск сервера локально в контейнере
docker run -p 8080:8080 \
  -e DATABASE_URL=postgres://user:pass@db:5432/avdi \
  avdi-server:latest
```

### Структура Dockerfile

#### Agent (agent.Dockerfile)

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /agent ./cmd/agent

FROM alpine:3.20
RUN apk add --no-cache iputils iproute2 net-tools nmap curl bind-tools
WORKDIR /app
COPY --from=builder /agent /agent
COPY diagnostics.yaml /etc/avdi/diagnostics.yaml
CMD ["/agent"]
```

**Multi-stage build преимущества:**
- ✅ Минимальный размер финального образа (~10MB вместо 500MB+)
- ✅ Финальный образ содержит только runtime, без Go компилятора
- ✅ Быстрый запуск контейнера

**Установленные утилиты в контейнере агента:**
- `iputils` - ping, traceroute
- `iproute2` - ip адреса, маршруты
- `net-tools` - netstat, ifconfig
- `nmap` - сканирование портов
- `curl` - HTTP запросы
- `bind-tools` - DNS утилиты (nslookup, dig)

#### Server (server.Dockerfile)

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /server /server
EXPOSE 8080
CMD ["/server"]
```

### Полный workflow: от сборки до deploy

```bash
# 1. Собираем образ агента
docker build -f docker/agent.Dockerfile -t myregistry.com/avdi-agent:v1.0 .

# 2. Отправляем в registry
docker push myregistry.com/avdi-agent:v1.0

# 3. Запускаем shell
./bin/shell

# 4. В shell развёртываем агента на удалённом хосте
deploy-agent \
  --ssh-host 192.168.1.100 \
  --ssh-user ubuntu \
  --server-url http://10.0.0.1:8080 \
  --agent-name prod-agent-1 \
  --image myregistry.com/avdi-agent:v1.0

# 5. На удалённом хосте:
# - Агент скачивается из registry (docker pull)
# - Запускается контейнер (docker run)
# - Автоматически регистрируется на сервере
# - Начинает выполнять задачи
```

### Версионирование образов

Рекомендуется использовать semantic versioning:

```bash
# Разработка
docker build -f docker/agent.Dockerfile -t avdi-agent:dev .
docker push avdi-agent:dev

# Тестирование
docker build -f docker/agent.Dockerfile -t avdi-agent:1.0.0-beta .
docker push avdi-agent:1.0.0-beta

# Production
docker build -f docker/agent.Dockerfile -t avdi-agent:1.0.0 .
docker push avdi-agent:1.0.0

# Latest
docker build -f docker/agent.Dockerfile -t avdi-agent:latest .
docker push avdi-agent:latest
```

### Запуск через Docker Compose

```bash
# Локальный запуск системы для тестирования
docker-compose up -d

# Просмотр логов
docker-compose logs -f server
docker-compose logs -f agent

# Остановка
docker-compose down
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
