# TripGo

TripGo — HTTP-сервис управления поездками на Go и PostgreSQL. Он поддерживает
создание, получение и завершение поездки, идемпотентные запросы на создание и
историю изменения статусов.

Текущая версия разделена на domain, usecase, transport, repository и
infrastructure, покрыта unit- и интеграционными тестами и экспортирует логи,
метрики и трейсы через OpenTelemetry.

## Быстрый запуск

```bash
tripgoctl cluster start
tripgoctl environment start
tripgoctl connect
make migrate
make run SERVICE_VERSION=local
```

## Документация

- [Лабораторная работа 1 — HTTP и PostgreSQL](docs/lab1/README.md)
- [Лабораторная работа 2 — архитектура, тестирование и observability](docs/lab2/README.md)
- [Архитектура и направления зависимостей](docs/lab2/README.md#архитектура)
- [Результаты observability](docs/lab2/README.md#результат-observability)
