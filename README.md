# O11y platform

O11y Platform — комплексная демонстрационная платформа для практики в области Observability: метрик, алертинга и нагрузочного тестирования.

## Стек технологий

1. Сервис – Elixir + Phoenix
2. БД – CocroachDB
3. Нагрузочный сервис – Go + Vegeta
4. o11y-platform – HTML, CSS, JavaScript
5. o11y – Prometheus, Grafana, AlertManager, Portainer
6. Infra – Docker, Docker Compose
7. Tracing – Jaeger


## Ссылки на доступные сервисы

| Название                      | URL                                                           | Описание                                                       |
| ----------------------------- | ------------------------------------------------------------- | -------------------------------------------------------------- |
| **Todo Task Manager Swagger** | [Swagger](http://localhost:4000/swagger/index.html#/)         | REST API для задач, регистрация, логин, документация           |
| **Load Generator**            | [Load Generator](http://localhost:8081/)                      | UI для генерации нагрузки на любой сервис, профили нагрузок    |
| **Database Admin Panel**      | [Database Admin Panel](http://localhost:8080/#/overview/list) | Веб-интерфейс CockroachDB: мониторинг состояния БД, статистика |
| **Grafana**                   | [Grafana](http://localhost:3000/)                             | Дашборды для всех метрик и алертов платформы                   |
| **Prometheus**                | [Prometheus](http://localhost:9090/query)                     | Сбор и хранение всех метрик, PromQL-запросы                    |
| **AlertManager**              | [AlertManager](http://localhost:9093/#/alerts)                | Менеджер алертов, интеграция с Telegram                        |
| **O11y Platform Main Page**   | [O11y-platform](http://localhost/#)                        | Портал платформы и навигация                                   |
| **Elixir Metrics**            | [Elixir /metrics](http://localhost:4000/metrics)              | Метрики приложения (todo)                                      |
| **Database Metrics**          | [DB /\_status/vars](http://localhost:8080/_status/vars)       | Метрики базы данных CockroachDB                                |
| **Load Generator Metrics**    | [LoadGen /metrics](http://localhost:8081/metrics)             | Метрики генератора нагрузки                                    |
| **Telegram алерты**           | [@o11y\_alerts](https://t.me/o11y_alerts)                     | Уведомления по алертам (latency, RPS)                          |

## Мониторинг и метрики
- Все сервисы экспортируют метрики по /metrics

- В Grafana настроены отдельные дашборды для приложения, базы, нагрузочного сервиса

- Алерты настраиваются в Prometheus/AlertManager, уведомления — в Telegram

## Доступы к сервисам

1. Grafana admin/admin
2. Portainer admin/qwertyasdfgh

## Проверить алертинг

### ❗️ Обязательно: в заголовках должен быть Content-Type: application/json

🔴 HighLatency p99 > 500ms 

1. Выбрать в качестве нагружаемой ручке GET /api/tasks
2. В заголовках поставить:
```
Content-Type: application/json
Authorization: Bearer <token>
```
token взять из swagger /login, пользователь уже создан, поэтому достаточно нажать try out. 
3. В нагрузочном сервисе указать: 100 RPS, длительность 120 секунд, профиль Постоянная

🟡 HighDBLoad rps в БД > 100

1. Выбрать в качестве нагружаемой ручке POST /api/tasks
2. В payload поставить 
```
{
  "title": "Load2",
  "status": "pending",
  "description": "Testing..."
}
```
В заголовках поставить:
```
Content-Type: application/json
Authorization: Bearer <token>
```
token взять из swagger /login, пользователь уже создан, поэтому достаточно нажать try out. 
3. В нагрузочном сервисе указать: 100 RPS, длительность 120 секунд, профиль Постоянная

## Профили нагрузки в Load Generator
| Профиль           | Описание                                                                           |
| ----------------- | ---------------------------------------------------------------------------------- |
| **Постоянная**    | Строго одинаковый RPS всё время (рекомендуется для стабильных нагрузочных тестов)  |
| **Умеренная**     | RPS в 2 раза ниже, равномерная, мягкая нагрузка                                    |
| **Хаотичная**     | Рандомные паузы, нестабильный поток запросов, имитирует дерганую реальную нагрузку |
| **Спайковая**     | Периодические всплески: RPS ×2 каждые 5 секунд                                     |
| **Волнообразная** | Плавное чередование высокой и низкой нагрузки (синусоподобный трафик)              |
| **Нагрев**        | Постепенное увеличение RPS от 50% до 100%                                          |
| **Ночной режим**  | Очень низкий RPS — 10% от заданного (симуляция ночного времени/low traffic)        |

## Сборка и запуск

```bash
git clone https://github.com/Kin1599/phoenix-todo-o11y
cd phoenix-todo-o11y

# заполнить по шаблону .env.sample
touch .env 

# если не на локалке, то нужно поменять localhost на реальный ip/domain в файлах

docker compose up --build -d
```

## ⚠️ Важные комментарии
Профили нагрузки «Спайковая», «Волнообразная» и «Нагрев» могут работать не так, как описано, извините, не починил(

При использовании постоянной нагрузки, если возникают ошибки (5xx, timeouts), фактический RPS может колебаться — это связано с тем, как Vegeta обрабатывает недоставленные запросы.

Метрики в нагрузочном сервисе и в Grafana/Prometheus могут немного расходиться. Это нормально и объясняется разным окном агрегации. Можно поиграться с окном агрегации в edit panel, и поставить свое значение. В целом там почти все идеально😁
