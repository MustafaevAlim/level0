Микросервис для обработки заказов: получение из Kafka, хранение в PostgreSQL и быстрый доступ через LRU-кеш. В комплекте есть пример фронтенда на HTML+JS.
## Возможности
 - Чтение заказов из Kafka
 - Сохраняет структуру заказа с платежом, доставкой и товарами
 - Быстрый in-memory LRU-кеш
 - HTTP API для получения заказа по ID
 - Простой веб-интерфейс

## Технологии
 - Golang
 - Kafka (segmentio/kafka-go), Kafka UI
 - PostgreSQL (sqlx)
 - Docker, Docker Compose
 - HTML+JS (frontend)

## Быстрый старт
  ### Клонирование
    git clone https://github.com/MustafaevAlim/level0.git
    cd level0

  ### Настрой .env
Создать .env на основе .env_example и заполнить значения:
    
    POSTGRES_USER=postgres
    POSTGRES_PASSWORD=yourpassword
    POSTGRES_DB=ordersdb
    KAFKA_BROKERS=localhost:9092
    KAFKA_TOPIC=orders-topic
    KAFKA_GROUP=orders-group
    CACHE_SIZE=100
    HTTP_PORT=8082

   ### Запуск через Docker Compose
    docker compose up --build

Чтобы управлять миграциями локально нужно скачать migrate:

    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz \
    | tar xvz -C /usr/local/bin

Используется makefile:

    make migrate-up # применить миграции
    make migrate-down # отменить миграции
    make migrate-version # посмотреть версию миграции

Это поднимет контейнеры с сервисом, Kafka, Kafka UI и PostgreSQL.

Создать топик в кафке (настроить по желанию):

    docker exec -it kafka /bin/bash # войти в контейнер с кафкой 
    kafka-topics --create --topic orders-topic --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1


## HTTP API
Получить заказ:

    GET /order/{order_uid}
Пример:

    curl http://localhost:8082/order/OOBmrfkDRphyYFiH

## Статичная HTML-страница
Открыть /info/ или (при нужной конфигурации) /order.html в браузере, чтобы воспользоваться веб-интерфейсом поиска заказа.
## Структура проекта
    ├── cmd/
    │   └── myapp/          # main.go — точка входа
    ├── internal/
    │   ├── config/         # Настройки приложения
    │   ├── app/            # Жизненный цикл приложения
    │   ├── api/            # HTTP API и маршруты
    │   ├── repository/     # Работа с БД, Kafka, Cache
    │   ├── model/          # модели данных
    ├── web/                # фронтенд
    ├── migrations/         # миграции БД
    ├── scripts/            # скрипты(имитация записи сообщений в кафку)
    ├── vendor/             # зависимости
    ├── Dockerfile
    ├── docker-compose.yml
    ├── .env.example
    ├── Makefile
    └── README.md