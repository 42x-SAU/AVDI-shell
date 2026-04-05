package main

import (
	"fmt"
	"strings"
)

// printHelpTopic prints detailed help for a command name. Returns false if unknown.
func printHelpTopic(topic string) bool {
	t := normalizeHelpTopic(topic)
	text, ok := helpTopics[t]
	if !ok {
		return false
	}
	fmt.Println(text)
	return true
}

// tryHelpCommands resolves help for args after "help" (e.g. ["create-task"] or ["create", "task"]).
func tryHelpCommands(parts []string) bool {
	if len(parts) == 0 {
		return false
	}
	for i := len(parts); i >= 1; i-- {
		seg := strings.Join(parts[:i], "-")
		if printHelpTopic(seg) {
			return true
		}
		seg = strings.Join(parts[:i], " ")
		if printHelpTopic(seg) {
			return true
		}
	}
	return false
}

func normalizeHelpTopic(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

// helpTopics: keys must be normalized (lowercase, hyphens).
var helpTopics = map[string]string{
	"health": helpHealth,
	"agents": helpAgents,
	"tasks":  helpTasks,

	"results":     helpResults,
	"export-logs": helpExportLogs,

	"create-task": helpCreateTask,

	"deploy-agent": helpDeployAgent,

	"diagnostic-list": helpDiagnosticList,
	"diagnostic-run":  helpDiagnosticRun,

	"get":  helpGet,
	"post": helpPost,

	"server":     helpServer,
	"stats":      helpStats,
	"stats-all":  helpStatsAll,
	"server-add": helpServerAdd,
	"server-list": helpServerList,

	"alias":   helpAlias,
	"unalias": helpUnalias,

	"recurring-list":    helpRecurringList,
	"recurring-add":     helpRecurringAdd,
	"recurring-delete":  helpRecurringDelete,
	"recurring-enable":  helpRecurringEnable,
	"recurring-disable": helpRecurringDisable,

	"clear": helpClear,
	"exit":  helpExit,
	"quit":  helpExit,
	"help": helpHelpMeta,

	"!":        helpBang,
	"bang":     helpBang,
	"shell-cmd": helpBang,
}

const helpHelpMeta = `help — справка по shell

  help              Общий список команд и примеры
  help <команда>    Подробности по одной команде (например: help create-task)

Имя команды можно писать через дефис или подчёркивание: help create_task`

const helpHealth = `health — проверка API

  GET /health

Флагов нет. Показывает JSON со статусом сервера.`

const helpAgents = `agents — список агентов

  GET /agents

Флагов нет. Таблица: id, имя, последний heartbeat, дата создания.`

const helpTasks = `tasks — список задач

  GET /tasks

Флагов нет. Таблица: id, агент, тип проверки, статус, ретраи, время создания.`

const helpResults = `results — результаты выполнения задач

  GET /results с опциональными query-параметрами (сервер)

Флаги:
  --logs              Полный вывод: stdout, stderr, logs для каждого результата
  --export            Записать каждый результат в файл (см. --output-dir)
  --output-dir <path> Каталог для файлов при --export (по умолчанию: output)
  --limit N           Запросить не более N последних результатов (0 = все, сервер ограничивает сверху)
  --task ID           Только результаты для задачи с этим id
  --result-id ID      Только строка результата с этим id
  --exit-code N       Только результаты с данным кодом выхода

Примеры:
  results
  results --logs
  results --export --limit 10 --output-dir ./out
  results --task 5 --exit-code 0`

const helpExportLogs = `export-logs — то же, что results --export

Эквивалент: results --export [остальные флаги как у results]

Пример:
  export-logs --limit 20 --output-dir output`

const helpCreateTask = `create-task — поставить задачу в очередь

  POST /tasks

Обязательные флаги:
  --agent <id>        ID агента (число > 0)
  --check <type>      Тип: hostname | ping | ports | diagnostic

Необязательные:
  --payload <string>  Для ping — адрес (например 8.8.8.8); для diagnostic — JSON, например:
                      '{"command":"disk-usage"}'
  --retries <n>       max_retries для задачи (по умолчанию 0)

Примеры:
  create-task --agent 2 --check hostname
  create-task --agent 2 --check ping --payload 8.8.8.8
  create-task --agent 2 --check ports
  create-task --agent 2 --check diagnostic --payload '{"command":"disk-usage"}'`

const helpDeployAgent = `deploy-agent — развернуть агент в Docker на удалённом хосте по SSH

Обязательные флаги:
  --ssh-host <host>     Хост или IP
  --ssh-user <user>     Пользователь SSH
  --server-url <url>    URL API AVDI, доступный с машины агента (лучше с http:// или https://)
  --agent-name <name>   Имя агента на сервере
  --image <ref>         Образ Docker (registry/имя:тег)

Необязательные:
  --ssh-port <n>        Порт SSH (по умолчанию 22; если флаг не задан, можно AVDI_SSH_PORT)
  --ssh-password <pwd>  Пароль SSH (иначе переменная AVDI_SSH_PASSWORD или запрос в консоль)
  --container <name>    Имя контейнера на удалённой машине (по умолчанию avdi-agent)
  --skip-pull           Не выполнять docker pull (образ уже есть локально на хосте)

Переменные окружения: AVDI_SSH_PASSWORD, AVDI_SSH_PORT`

const helpDiagnosticList = `diagnostic-list — список диагностических команд из YAML

Флаги:
  --config <path>  Путь к diagnostics.yaml (если не задан — авто-поиск)
  --json           Вывести список команд в JSON`

const helpDiagnosticRun = `diagnostic-run — выполнить диагностику локально (на машине shell)

Флаги:
  --command <name>   Имя команды из конфига (обязательно)
  --config <path>    Путь к YAML (если нужен не стандартный)
  --var key=value    Переменная (можно повторять; в текущей реализации удобнее один --var)
  --json             Формат вывода (по умолчанию true)

Выполняется на локальной машине, не на агенте.`

const helpGet = `get — произвольный GET к текущему серверу

  get /path

Пример:
  get /agents
  get /stats`

const helpPost = `post — произвольный POST с JSON телом

  post /path '{"key":"value"}'

Пример:
  post /tasks '{"agent_id":1,"check_type":"ping","payload":"8.8.8.8","max_retries":0}'`

const helpServer = `server — сменить базовый URL API в этой сессии shell

  server <url>

Пример:
  server http://192.168.1.10:8081

Переменная окружения при старте: AVDI_SERVER`

const helpStats = `stats — сводная статистика текущего сервера (JSON)

  GET /stats

Флагов нет. Поля: agents_total, tasks_*, results_total, recurring_*.`

const helpStatsAll = `stats-all — таблица /stats для текущего URL и всех URL из server-list

Флагов нет. Список дополнительных серверов: ~/.config/avdi/servers.txt (см. server-add).`

const helpServerAdd = `server-add — добавить URL в файл списка серверов

  server-add <url>

Файл: ~/.config/avdi/servers.txt (для stats-all).`

const helpServerList = `server-list — показать содержимое ~/.config/avdi/servers.txt

Строки с # в начале считаются комментариями.`

const helpAlias = `alias — пользовательские сокращения команд

  alias                    Список всех алиасов
  alias <name>             Показать значение одного алиаса
  alias <name>=<команда>   Задать алиас (сохраняется в ~/.config/avdi/aliases.json)

Пример:
  alias p=create-task --agent 1 --check ping --payload 8.8.8.8`

const helpUnalias = `unalias — удалить алиас

  unalias <name>`

const helpRecurringList = `recurring-list — список периодических заданий (JSON)

  GET /recurring

Флагов нет. Требуется миграция БД с таблицей recurring_jobs.`

const helpRecurringAdd = `recurring-add — периодически ставить одну и ту же проверку в очередь

  POST /recurring

Флаги:
  --agent <id>         ID агента (обязательно)
  --check <type>       hostname | ping | ports | diagnostic (обязательно)
  --payload <string>   Как у create-task (для diagnostic — JSON)
  --interval <sec>     Период в секундах (10–86400, по умолчанию 60)
  --retries <n>        max_retries для каждой создаваемой задачи (по умолчанию 0)
  --start-in <sec>     Задержка до первого постановления в очередь (по умолчанию 0)

Пример (ping раз в минуту):
  recurring-add --agent 2 --check ping --payload 8.8.8.8 --interval 60`

const helpRecurringDelete = `recurring-delete — удалить расписание

  DELETE /recurring/{id}

Флаги:
  --id <n>   ID записи recurring_jobs`

const helpRecurringEnable = `recurring-enable — включить расписание

  PATCH /recurring/{id}  {"enabled":true}

Флаги:
  --id <n>`

const helpRecurringDisable = `recurring-disable — выключить расписание

  PATCH /recurring/{id}  {"enabled":false}

Флаги:
  --id <n>`

const helpClear = `clear — очистить экран и снова показать баннер (с логотипом)`

const helpExit = `exit | quit — выход из shell`

const helpBang = `Префикс ! — команда операционной системы (не API AVDI)

  !<текст>   Выполняется через cmd /C (Windows) или $SHELL -c (Unix)

Примеры:
  !dir
  !echo test`
