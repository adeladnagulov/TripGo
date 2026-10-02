## Trip Service

## Запуск
# 1.Поднять окружение
tripgoctl cluster start
tripgoctl environment start
tripgoctl connect
# 2.Применить миграции
make migrate
Откат
make migrate-down 
# 3.Сгенерировать код из OpenAPI
make generate
# 4.Запустить сервис
make run

## env переменные
HTTP_ADDR
LOG_LEVEL
SHUTDOWN_TIMEOUT
DATABASE_URL
DATABASE_MAX_CONNS
DATABASE_MIN_CONNS
DATABASE_MAX_CONN_LIFETIME
DATABASE_CONNECT_TIMEOUT
DATABASE_QUERY_TIMEOUT

## Решения
# Уровень изоляции 
Используется `Read Committed` - дефолтный уровень для PostgreSQL

Потому что:
- Для операций в этой лабе (создание поездки, завершение) достаточно `Read Committed`. Мы не читаем одни и те же строки дважды внутри транзакции.

# Менеджер транзауций
Реализован интерфейс
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
Реализован следующим образом:
- Проверяет, есть ли уже транзакция в контексте. Если есть - просто вызывает переданную функцию, не создавая новую транзакцию, что защищает от вложенных вызовов
- Если транзакции нет - открывает новую через pool.Begin(ctx)
- Кладёт pgx.Tx в контекст через context.WithValue
- Вызывает переданную функцию fn, передавая ей новый контекст
- В defer проверяет результат: если функция вернула ошибку или была паника - делает Rollback, иначе Commit

# Запрет двух активных поездок на одного водителя
Решение на уровне схемы БД - частичный уникальный индекс, он покрывает только строки со статусом active. PostgreSQL сам гарантирует, что вставить вторую активную поездку для того же водителя не получится, даже при параллельных запросах

CREATE UNIQUE INDEX trips_driver_active_idx ON trips (driver_id) WHERE status = 'active';