# AVDI-Shell

<img src="./frontend/contents/logo.png" width="450" alt="Логотип AVDI-Shell">

## Описание

**AVDI-Shell** - распределённая система диагностики инфраструктуры на базе архитектуры "сервер-агенты". Позволяет централизованно управлять диагностическими проверками на удалённых машинах.

Проект написан на **Go 1.25** с использованием PostgreSQL и Docker.

## Ключевые возможности

- Распределённая архитектура с агентами
- Автоматическая регистрация агентов
- Выполнение диагностических команд
- Система очереди задач
- Мониторинг агентов через heartbeat
- Конфигурация команд в YAML
- Поддержка Docker развертывания

## Оглавление

- [Описание](#описание)
- [Ключевые возможности](#ключевые-возможности)
- [Архитектура](#архитектура)
- [Структура проекта](#структура-проекта)
- [Установка](#установка)
- [Использование](#использование)
- [Конфигурация](#конфигурация)
- [Docker](#docker)
- [Мониторинг](#мониторинг)
- [Лицензия](#лицензия)

## Архитектура

```mermaid
graph TD
    Client --> Server
    Server --> DB[(PostgreSQL)]
    Server --> Queue[Task Queue]
    Queue --> Agent1
    Queue --> Agent2
    Agent1 --> Server
    Agent2 --> Server
```

**Компоненты:**
- **Server**: REST API сервер
- **Agent**: Выполняет диагностические команды
- **Queue**: Очередь задач на базе PostgreSQL

## Структура проекта

```
avdi/
├── cmd/ (точки входа: server, agent, shell)
├── internal/ (логика: server, agent, storage, queue, diagnostics)
├── pkg/ (logger)
├── docker/ (Dockerfiles)
├── diagnostics.yaml (конфигурация команд)
└── go.mod
```

## Установка

### Требования

- Go 1.25+
- PostgreSQL 13+
- Docker (для контейнеризации)

### Локальная установка

1. **Клонирование:**
   ```bash
   git clone https://github.com/42x-SAU/AVDI-shell.git
   cd AVDI-shell/avdi
   ```

2. **Зависимости:**
   ```bash
   go mod download
   ```

3. **База данных:**
   ```bash
   createdb avdi_db
   # Выполнить SQL скрипты для схемы
   ```

4. **Переменные окружения (.env):**
   ```env
   SERVER_ADDR=:8080
   DATABASE_URL=postgres://user:pass@localhost:5432/avdi_db
   AGENT_TOKEN=your-token
   ```

5. **Сборка:**
   ```bash
   # Сервер
   go build -o bin/server ./cmd/server

   # Агент
   go build -o bin/agent ./cmd/agent

   # Shell
   go build -o bin/shell ./cmd/shell
   ```

6. **Запуск:**
   ```bash
   ./bin/server  # в одном терминале
   ./bin/agent   # в другом
   ```

## Использование

### Интерактивная Shell

Основной интерфейс для работы с системой:

```bash
./bin/shell
```

**Команды:**
- `health` - статус сервера
- `agents` - список агентов
- `tasks` - список задач
- `results` - результаты
- `deploy-agent --ssh-host <host> --ssh-user <user> --server-url <url> --agent-name <name> --image <image>` - развернуть агента по SSH
- `create-task --agent <id> --check <type>` - создать задачу
- `diagnostic-list` - доступные команды
- `help` - справка

### REST API

Основные endpoints:
- `GET /health` - health check
- `GET /agents` - список агентов
- `POST /tasks` - создать задачу
- `GET /tasks` - список задач
- `GET /results` - результаты

## Конфигурация

Диагностические команды настраиваются в `diagnostics.yaml`:

```yaml
version: "1.0"
commands:
  - name: "hostname"
    description: "Имя хоста"
    command: "hostname"
    parse_output: "text"

  - name: "ping-target"
    description: "Пинг хоста"
    command: "ping"
    args: ["-c", "4", "{{target}}"]
    variables:
      - name: "target"
        required: true
        default: "8.8.8.8"
```

## Docker

### Сборка образов

#### Агент

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

#### Сервер

```bash
docker build -f docker/server.Dockerfile -t avdi-server:latest .
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

**Установленные утилиты:**
- `iputils` (ping, traceroute)
- `iproute2` (ip команды)
- `net-tools` (netstat, ifconfig)
- `nmap` (сканирование)
- `curl` (HTTP)
- `bind-tools` (DNS)

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

### Workflow развертывания

```bash
# 1. Собрать образ агента
docker build -f docker/agent.Dockerfile -t myregistry.com/avdi-agent:v1.0 .

# 2. Отправить в registry
docker push myregistry.com/avdi-agent:v1.0

# 3. Развернуть через shell
./bin/shell
deploy-agent --ssh-host 192.168.1.100 --ssh-user ubuntu --server-url http://server:8080 --agent-name agent1 --image myregistry.com/avdi-agent:v1.0
```

### Docker Compose

```bash
# Запуск системы
docker-compose up -d

# Логи
docker-compose logs -f

# Остановка
docker-compose down
```

## Мониторинг

- **Heartbeat**: Агенты отправляют статус каждые 15 секунд
- **Логи**: Выводятся в stdout/stderr
- **Таймауты**: Для команд и подключений

## Лицензия

Укажите лицензию проекта.

## Контакты

- **Issues:** [GitHub Issues](https://github.com/42x-SAU/AVDI-shell/issues)
- **Discussions:** [GitHub Discussions](https://github.com/42x-SAU/AVDI-shell/discussions)