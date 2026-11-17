# Мутационное тестирование domain и usecase — before / after

Сравниваются результаты мутационного тестирования пакетов
`internal/domain/...` и `internal/usecase/...` для бонусного задания ЛР2.
Отчёты обоих инструментов `before` фиксируют исходный набор unit-тестов без
дополнительных тестов конструкторов и проверок usecase. Для общего исходного
состояния тесты `NewTrip` и `NewTripStatusHistory` временно исключены из
доменного пакета, после чего прогон `go-mutesting` выполнен повторно.

Для `after` тесты конструкторов возвращены в domain, добавлены проверки ошибок
usecase и вызовов бизнес-recorder. Производственный код между прогонами не
менялся. Версии инструментов и параметры запуска сохранены, исходные артефакты
`before` не перезаписаны.

## Gremlins: инструмент и запуск

Использован Gremlins `v0.6.0`, собранный на Go `1.27.1`. Версия инструмента
зафиксирована в Makefile, бинарник устанавливается в `bin/gremlins`.

Исходный запуск выполнен из корня репозитория до усиления тестов:

```bash
make mutation MUTATION_STAGE=before
```

Для проверки текущего набора тестов и сохранения отдельных результатов:

```bash
make mutation MUTATION_STAGE=after
```

Настройки сверх включённых по умолчанию мутаций:

```text
--invert-logical
--invert-assignments
--invert-bitwise
--invert-bwassign
--invert-loopctrl
--remove-self-assignments
--workers 2
--timeout-coefficient 10
```

Каждая область анализируется отдельным запуском. Используются unit-тесты пакета,
в котором находится мутант; режим `--integration` и build-тег `integration`
не включены. JSON форматируется через `jq` после завершения каждого запуска.

Отчёты:

- domain: [before](domain-before.json), [after](domain-after.json);
- usecase: [before](usecase-before.json), [after](usecase-after.json).

## Результаты Gremlins

| Показатель | domain before | domain after | usecase before | usecase after |
|---|---:|---:|---:|---:|
| Обнаружено мутаций, включая непокрытые | 33 | 33 | 19 | 19 |
| `KILLED` по классификации Gremlins | 31 | 33 | 19 | 19 |
| `LIVED` | 0 | 0 | 0 | 0 |
| `NOT COVERED` | 2 | 0 | 0 | 0 |
| `NOT VIABLE` по классификации Gremlins | 0 | 0 | 0 | 0 |
| `TIMED OUT` | 0 | 0 | 0 | 0 |
| `SKIPPED` | 0 | 0 | 0 | 0 |
| Test efficacy | 100.00% | 100.00% | 100.00% | 100.00% |
| Mutator coverage | 93.94% | 100.00% | 100.00% | 100.00% |
| Время выполнения, секунды | 7.257 | 7.297 | 4.090 | 4.273 |

Поле `mutants_total` в JSON Gremlins `v0.6.0` не включает `NOT COVERED`:
для domain в `before` оно равно 31, хотя в списке `files[].mutations`
содержатся 33 мутации.
В таблице число обнаруженных мутаций учитывает все записи списка.

## Разбор непокрытых мутаций Gremlins

Выживших мутантов со статусом `LIVED` в исходном прогоне нет. Обнаружены два
мутанта со статусом `NOT COVERED`, оба в домене.

### NewTrip: проверка ошибки валидации

Расположение: [internal/domain/trip/trip.go](../../../internal/domain/trip/trip.go),
строка 40, столбец 34. Вид: `CONDITIONALS_NEGATION`.

Исходный код:

```go
if err := input.Validate(); err != nil {
    return Trip{}, err
}
```

Мутация заменяет `err != nil` на `err == nil`. При корректных данных
конструктор возвращал бы пустую поездку без ошибки, а при некорректных —
продолжал бы создание поездки.

В исходном состоянии прямых тестов `NewTrip` в доменном пакете не было. Вызовы
конструктора из usecase-тестов не учитываются при отдельном запуске domain
в выбранном режиме Gremlins.

Добавлены `TestNewTripSuccess` и `TestNewTripInvalidInput` в
[trip_test.go](../../../internal/domain/trip/trip_test.go).
Успешный сценарий проверяет входные поля, переданные ID и время,
статус `active` и отсутствие `FinishedAt`. Ошибочный сценарий проверяет
`ValidationError`, поле `user_id` и пустой результат. Статус мутации изменился
с `NOT COVERED` на `KILLED`.

### NewTripStatusHistory: взятие адреса reason

Расположение: [internal/domain/trip/history.go](../../../internal/domain/trip/history.go),
строка 32, столбец 15. Вид: `INVERT_BITWISE`.

Исходный код:

```go
Reason: &reason,
```

Gremlins рассматривает токен `&` как кандидат на замену на `|`, хотя здесь
это оператор взятия адреса. Результат `Reason: |reason` не является допустимым
Go-кодом. Это не эквивалентный мутант: мутация приводит к ошибке компиляции.

В исходном прогоне конструктор не выполняется тестами domain, поэтому Gremlins
останавливается на статусе `NOT COVERED` и не проверяет компилируемость мутации.

Добавлен табличный `TestNewTripStatusHistory` в
[history_test.go](../../../internal/domain/trip/history_test.go) со сценариями
создания и завершения поездки. Он проверяет `TripID`, `FromStatus`, `ToStatus`,
значение `Reason` и `ChangedAt`. Статус мутации изменился с `NOT COVERED` на
`KILLED`, но отклонение этой некомпилируемой мутации не доказывает наличие
проверки бизнес-поведения: сами проверки полей находятся в тесте.

## Ограничения интерпретации

Часть уже помеченных `KILLED` мутаций `INVERT_BITWISE` также затрагивает взятие
адреса: например, `&ValidationError`, `&at`, `&Service` и `&fromStatus`.
Замена `&` на `|` в этих выражениях создаёт некомпилируемый код.

Gremlins `v0.6.0` определяет статус по коду завершения `go test`: код 1
классифицируется как `KILLED`, код 2 — как `NOT VIABLE`. Ошибка сборки через
`go test` может завершиться кодом 1. Поэтому таблица воспроизводит
данные инструмента, но не позволяет утверждать, что каждый `KILLED` обнаружен
именно assertion теста. Значение `NOT VIABLE = 0` также не доказывает, что все
мутации компилировались.

Результат относится к набору операторов и настройкам этого прогона.
Отсутствие `LIVED` не означает, что тесты обнаруживают любые возможные ошибки.

## go-mutesting: инструмент и исходный запуск

Дополнительная проверка выполнена через `github.com/zimmski/go-mutesting`,
версия `v0.0.0-20210610104036-6d9217011a00`, собранная на Go `1.27.1`.
Для совместимости с этой версией Go инструмент собирается с
`golang.org/x/tools v0.51.0` в отдельном модуле, как задано в Makefile.
Использованы включённые по умолчанию мутации `branch/case`, `branch/else`,
`branch/if`, `expression/comparison`, `expression/remove` и `statement/remove`.
Они включают удаление инструкций и содержимого веток, дополняя проверку
операторов через Gremlins.

Команда запуска `go-mutesting` с сохранением исходного результата:

```bash
./bin/go-mutesting --exec-timeout=60 \
  ./internal/domain/... ./internal/usecase/... \
  2>&1 | tee docs/lab2/mutation/go-mutesting-before.log
```

Вывод сохранён как [go-mutesting-before.log](go-mutesting-before.log), повторный
прогон — как [go-mutesting-after.log](go-mutesting-after.log).
Логи содержат результаты каждой мутации; в `before` также приведены diff
для выживших мутантов.
В актуальном Makefile запуск обоих инструментов объединён в `make mutation`,
а `MUTATION_STAGE` определяет суффикс JSON-отчётов и лога.

### Результаты go-mutesting

| Показатель | before | after |
|---|---:|---:|
| Протестировано уникальных мутаций | 43 | 43 |
| Убито (`PASS`) | 38 | 43 |
| Выжило (`FAIL`) | 5 | 0 |
| Дубликатов, исключённых из total | 2 | 2 |
| Пропущено (`SKIP`) | 0 | 0 |
| Mutation score | 88.37% | 100.00% |

Значения `PASS` и `FAIL` относятся к результату обнаружения мутации:
`PASS` означает, что мутант убит, `FAIL` — что тесты не обнаружили изменение.
Mutation score вырос с `38 / 43 × 100%` до `43 / 43 × 100%`.
Дубликаты в знаменатель не входят. Наборы checksum уникальных мутаций
в обоих логах совпадают: все пять прежних `FAIL` стали `PASS`.

Распределение записей лога по пакетам:

| Область | Убито before | Выжило before | Убито after | Выжило after | Всего уникальных |
|---|---:|---:|---:|---:|---:|
| `internal/domain/...` | 19 | 1 | 20 | 0 | 20 |
| `internal/usecase/...` | 19 | 4 | 23 | 0 | 23 |

### Разбор пяти выживших мутантов

1. **Ошибка валидации при создании доменной поездки.**
   В [trip.go](../../../internal/domain/trip/trip.go), строка 41,
   удалён `return Trip{}, err` после неуспешного `input.Validate()`.
   Конструктор мог бы возвращать поездку без ошибки даже при невалидном вводе.
   В исходном доменном пакете прямых тестов конструктора не было, поэтому мутация
   выжила. Её убивает `TestNewTripInvalidInput`: проверяет `ValidationError`
   через `require.ErrorAs` и пустой результат. `TestNewTripSuccess` дополнительно
   проверяет исходные поля и начальное состояние поездки.
   Оба теста находятся в [trip_test.go](../../../internal/domain/trip/trip_test.go).

2. **Ошибка получения поездки при idempotency replay.**
   В [create.go](../../../internal/usecase/trip/create/create.go), строка 45,
   удалён возврат ошибки `GetTrip`. При совпадающем ключе и hash выполнение
   может продолжиться с `replayed == true`, несмотря на ошибку репозитория.
   Добавлен `TestCreateTripIdempotencyGetTripError` в
   [create_test.go](../../../internal/usecase/trip/create/create_test.go):
   `ReserveOrGet` возвращает существующую запись, а `GetTrip` — ошибку.
   Тест проверяет `require.ErrorIs`, пустую поездку и `replayed == false`.
   Вызовы создания, истории и recorder не ожидаются.

3. **Запись бизнес-события при replay.**
   В [create.go](../../../internal/usecase/trip/create/create.go), строка 63,
   условие `!replayed && s.recorder != nil` заменено на
   `true && s.recorder != nil`. Повторный запрос мог бы увеличивать счётчик
   созданных поездок и повторно логировать создание.
   В исходных usecase-тестах recorder был равен `nil`, поэтому изменение оставалось
   незамеченным. В `TestCreateTripIdempotencyReplay` передан mock recorder,
   а вызов `Create` запрещён через `Times(0)`. В `TestCreateTripSuccess` и
   `TestCreateTripWithNewIdempotencyKey` проверяется ровно один вызов `Create`
   с ID созданной поездки. Все проверки находятся в
   [create_test.go](../../../internal/usecase/trip/create/create_test.go).

4. **Ошибка обновления поездки при завершении.**
   В [finish.go](../../../internal/usecase/trip/finish/finish.go), строка 29,
   удалён возврат ошибки репозитория `FinishTrip`. Обработка может продолжиться
   после неуспешного обновления. Добавлен `TestFinishTripRepositoryError` в
   [finish_test.go](../../../internal/usecase/trip/finish/finish_test.go):
   проверяет `require.ErrorIs`, пустой результат и отсутствие записи истории.
   Неожиданный вызов истории приводит к падению mock-теста.

5. **Ошибка записи истории завершения.**
   В [finish.go](../../../internal/usecase/trip/finish/finish.go), строка 39,
   удалён возврат ошибки `AddInTripHistory`. Callback транзакции мог бы
   завершиться успешно, несмотря на ошибку истории.
   Добавлен `TestFinishTripHistoryError` в
   [finish_test.go](../../../internal/usecase/trip/finish/finish_test.go):
   обновление поездки проходит, а запись истории возвращает ошибку. Mock
   транзакционного менеджера возвращает ошибку callback без изменений;
   `require.ErrorIs` проверяет её возврат из usecase, результат поездки пустой.

Все пять мутаций меняют наблюдаемое поведение domain или usecase и закрыты
тестами. Они не являются эквивалентными. Дополнительно `TestCreateTripWithoutRecorder`
сохраняет проверку успешного создания с `nil` recorder после замены зависимости
на mock в остальных тестах. Проценты Gremlins и go-mutesting
относятся к разным наборам мутаций и сравниваются отдельно для каждого
инструмента.

## Итог

В Gremlins число непокрытых мутаций domain снизилось с 2 до 0,
Mutator coverage выросло с 93.94% до 100.00%; показатели usecase сохранились.
В go-mutesting число выживших снизилось с 5 до 0, Mutation score вырос
с 88.37% до 100.00%. Необъяснённых выживших мутантов не осталось.
Ограничение Gremlins для некомпилируемых `INVERT_BITWISE`-мутаций сохраняется
и для `after`; результаты относятся только к проверенным наборам мутаций.
