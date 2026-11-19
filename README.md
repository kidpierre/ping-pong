# Ping-Pong

Два микросервиса, ping и pong, с настраиваемой задержкой шлют запросы друг другу и пишут результат в лог. Ping при успешном запросе выводит `Ping`, pong выводит `Pong`. Количество успешных запросов можно получить снаружи кластера через `GET /count`.

Go 1.25, только стандартная библиотека. В тестах testify.

## Запуск в Minikube

Нужны Docker, minikube и kubectl.

```bash
minikube start
make deploy
```

`make deploy` соберёт образы `ping:1.0` и `pong:1.0`, загрузит их в minikube, применит манифесты из `k8s/` и дождётся, пока поды поднимутся.

Логи:

```bash
make logs-ping   # {"msg":"Ping","service":"ping","count":5}
make logs-pong
```

Счётчики снаружи кластера (Service типа NodePort, порты 30080 и 30081):

```bash
make count
# {"pings":12}
# {"pongs":8}
```

Сменить задержку:

```bash
make interval PING=500ms          # только ping
make interval PING=1s PONG=250ms  # оба сразу
```

Команда меняет значение в ConfigMap `ping-pong-config` и перезапускает deployment, потому что переменные окружения читаются при старте пода.

Удалить из кластера: `make clean`.

Тесты:

```bash
go test -race ./...
```

## Переменные окружения

| Переменная | По умолчанию | |
|---|---|---|
| `PEER_URL` | | адрес соседа, в кластере `http://pong:8080` или `http://ping:8080` |
| `INTERVAL` | `2s` | задержка между запросами, в кластере берётся из ConfigMap |
| `REQUEST_TIMEOUT` | `2s` | таймаут запроса к соседу |
| `HTTP_PORT` | `8080` | порт сервиса |
| `SHUTDOWN_TIMEOUT` | `10s` | сколько ждать активные запросы при остановке |
| `LOG_LEVEL` | `info` | |

## API

| | |
|---|---|
| `GET /ping` у ping, `GET /pong` у pong | сюда стучится сосед |
| `GET /count` | `{"pings": N}` или `{"pongs": N}` |
| `GET /healthz` | для liveness и readiness проб |
