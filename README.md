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
- [Разработка](#разработка)
- [Лицензия](#лицензия)
- [Контакты](#контакты)

## Архитектура

```mermaid
graph TD
    Client[Web Client / API] --> Server[REST API Server]
    Server --> DB[(PostgreSQL)]
    Server --> Redis[(Redis Cache)]
    Server --> Queue[Task Queue]
    Queue --> Agent1[Agent 1]
    Queue --> Agent2[Agent 2]
    Agent1 --> Server
    Agent2 --> Server
    Server --> WS[WebSocket Hub]
    WS --> Client
```

**Компоненты:**
- **Server**: REST API сервер на Go с WebSocket поддержкой
- **Agent**: Лёгкие исполнители диагностических команд
- **PostgreSQL**: Основная БД для хранения задач, результатов и агентов
- **Redis**: Кэш для лимитера задач, retry-механизмов и pub/sub
- **Queue**: Распределённая очередь задач с поддержкой лимитов и retry

## Структура проекта

```
avdi/
├── cmd/ (точки входа)
│   ├── agent/ (исполнитель диагностических команд)
│   ├── server/ (REST API сервер)
│   └── shell/ (интерактивная консоль)
├── internal/ (бизнес-логика)
│   ├── agent/ (логика агента)
│   ├── deploy/ (SSH развертывание агентов)
│   ├── diagnostics/ (парсер YAML конфигураций)
│   ├── httputil/ (HTTP утилиты)
│   ├── queue/ (очередь задач с лимитером и retry)
│   ├── redis/ (Redis клиент)
│   ├── server/ (HTTP handlers, WebSocket)
│   └── storage/ (PostgreSQL модели и запросы)
├── pkg/ (общие пакеты)
│   └── logger/ (структурированное логирование)
├── migrations/ (SQL схемы БД)
├── docker/ (Dockerfiles для агента и сервера)
├── diagnostics.yaml (конфигурация команд диагностики)
├── docker-compose.yml (полный стек для разработки)
└── go.mod (зависимости Go)
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
# PostgreSQL
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=diag
   DB_SSLMODE=disable

   # Redis
   REDIS_URL=redis://localhost:6379

   # Server
   SERVER_ADDR=:8080
   WEBSOCKET_ENABLED=true

   # Task Management
   MAX_CONCURRENT_TASKS_PER_AGENT=3
   BASE_RETRY_DELAY_SECONDS=30
   MAX_RETRY_DELAY_SECONDS=3600

   # Agent
   AGENT_NAME=agent-1
   SERVER_URL=http://localhost:8080
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

Полная документация API endpoints:

#### Health & Status
- `GET /health` - Проверка здоровья сервера
- `GET /stats` - Статистика системы (агенты, задачи, очередь)

#### Агенты
- `GET /agents` - Список всех агентов
- `POST /agents/register` - Регистрация нового агента
### Диагностические команды

Команды настраиваются в `diagnostics.yaml`. Поддерживаются переменные, аргументы и различные типы парсинга вывода:

```yaml
version: "1.0"
commands:
  - name: "hostname"
    description: "Получить имя хоста"
    command: "hostname"
    parse_output: "text"

  - name: "ping-target"
    description: "Пинг целевого хоста"
    command: "ping"
    args: ["-c", "4", "{{target}}"]
    variables:
      - name: "target"
        required: true
        default: "8.8.8.8"
    timeout: 30

  - name: "disk-usage"
    description: "Использование дискового пространства"
    command: "df"
    args: ["-h", "{{path}}"]
    variables:
      - name: "path"
        required: false
        default: "/"
```

### Переменные окружения

См. `.env.example` для полного списка переменных. Основные категории:

- **База данных:** `DB_*` - подключение к PostgreSQL
- **Redis:** `REDIS_URL` - подключение к Redis
- **Сервер:** `SERVER_ADDR`, `WEBSOCKET_ENABLED`
- **Очередь задач:** `MAX_CONCURRENT_TASKS_PER_AGENT`, retry настройки
- **Агент:** `AGENT_NAME`, `SERVER_URL`, `AGENT_TOKENАутентификация:** Агенты используют Bearer токен в заголовке `Authorization: Bearer <token>` или `X-Agent-ID` для идентификации.

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

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: diag
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
### Health Checks
- **Сервер:** `GET /health` возвращает `{"status": "ok"}`
- **База данных:** Автоматическая проверка подключения при старте
- **Redis:** Проверка доступности для дополнительных функций

### Метрики
- **Heartbeat:** Агенты отправляют статус каждые 15 секунд
- **Очередь:** Мониторинг количества задач в очереди
- **Лимиты:** Отслеживание одновременных задач на агента

### Логирование
- Структурированные логи в JSON формате
- Уровни: debug, info, warn, error
- Вывод в stdout/stderr для контейнеров

### WebSocket Real-time
Подключение к `/ws` для получения обновлений:
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Update:', data);
};
```

###Разработка

MIT License - см. [LICENSE](LICENSE) файл для деталей.

## Контакты

- **GitHub:** [42x-SAU/AVDI-shell](https://github.com/42x-SAU/AVDI-shell)
- **Issues:** [GitHub Issues](https://github.com/42x-SAU/AVDI-shell/issues)
- **Discussions:** [GitHub Discussions](https://github.com/42x-SAU/AVDI-shell/discussions)
- **Email:** [your-email@example.com]

---

**AVDI-Shell** — надёжное решение для распределённой диагностики инфраструктуры с поддержкой real-time мониторинга и автоматического масштабирования.
# Сборка всех компонентов
go build -o bin/server ./cmd/server
go build -o bin/agent ./cmd/agent
go build -o bin/shell ./cmd/shell

# Запуск тестов
go test ./...

# Линтинг
golangci-lint run
```

### Добавление новых команд диагностики

1. Добавьте команду в `diagnostics.yaml`
2. Протестируйте на агенте
3. Обновите документацию

### Contributing

1. Fork репозиторий
2. Создайте feature branch: `git checkout -b feature/amazing-feature`
3. Commit изменения: `git commit -m 'Add amazing feature'`
4. Push branch: `git push origin feature/amazing-feature`
5. Создайте Pull Request

### Roadmap

- [ ] Web UI интерфейс
- [ ] Поддержка периодических задач (cron-like)
- [ ] Метрики и графики (Prometheus/Grafana)
- [ ] Распределённое развертывание агентов
- [ ] Поддержка Windows агентов
- [ ] API токены для пользователей
- [ ] Ролевая модель доступа
- Убедитесь, что агент живой (heartbeat)
- Проверьте логи агента на ошибки выполнения команд

#### Redis недоступен
- Система работает без Redis, но без лимитов и retry
- Проверьте `REDIS_URL` и доступность Redis

#### База данных
- Выполните миграции: `migrations/*.sql`
- Проверьте подключение: `DB_*` переменные postgres -d diag"]
      interval: 3s
      timeout: 5s
      retries: 20

  redis:
    image: redis:7-alpine
    ports:
      - "6380:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  server:
    build:
      context: .
      dockerfile: docker/server.Dockerfile
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: diag
      DB_SSLMODE: disable
      REDIS_URL: redis://redis:6379
      WEBSOCKET_ENABLED: "true"
      MAX_CONCURRENT_TASKS_PER_AGENT: "3"
      BASE_RETRY_DELAY_SECONDS: "30"
      MAX_RETRY_DELAY_SECONDS: "3600"
      SERVER_ADDR: :8080
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

volumes:
  pgdata:
```

**Запуск:**
```bash
docker-compose up -d
docker-compose logs -f server
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