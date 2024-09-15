## Структура проекта
В данном проекте находится решение тестового задания на стажировку Авито 2024.

## Задание
В папке "задание" размещена задача.

## Сбор и развертывание приложения

Для развертывания приложения необходимо установить следующие переменные окружения

- `SERVER_ADDRESS` — адрес и порт, который будет слушать HTTP сервер при запуске. Пример: 0.0.0.0:8080.
- `POSTGRES_CONN` — URL-строка для подключения к PostgreSQL в формате postgres://{username}:{password}@{host}:{5432}/{dbname}.
- `POSTGRES_JDBC_URL` — JDBC-строка для подключения к PostgreSQL в формате jdbc:postgresql://{host}:{port}/{dbname}.
- `POSTGRES_USERNAME` — имя пользователя для подключения к PostgreSQL.
- `POSTGRES_PASSWORD` — пароль для подключения к PostgreSQL.
- `POSTGRES_HOST` — хост для подключения к PostgreSQL (например, localhost).
- `POSTGRES_PORT` — порт для подключения к PostgreSQL (например, 5432).
- `POSTGRES_DATABASE` — имя базы данных PostgreSQL, которую будет использовать приложение.

Для сборки Docker-контейнера приложения используется Dockerfile, расположенный в корневой директории проекта. Следуйте этим шагам для сборки и запуска контейнера:

1. Выполните команду для сборки Docker-образа:
    ```bash
    docker build -t avito-internship-app:latest .
    ```
    Здесь `-t avito-internship-app:latest` задает имя (`avito-internship-app`) и тег (`latest`) для создаваемого образа. Тег можно изменить по своему усмотрению.

2. После успешной сборки образа запустите контейнер:
    ```bash
    docker run -d --name avito-internship-app \
    -e SERVER_ADDRESS=0.0.0.0:8080 \
    -e POSTGRES_CONN=postgres://user:password@localhost:5432/dbname \
    -e POSTGRES_JDBC_URL=jdbc:postgresql://localhost:5432/dbname \
    -e POSTGRES_USERNAME=user \
    -e POSTGRES_PASSWORD=password \
    -e POSTGRES_HOST=localhost \
    -e POSTGRES_PORT=5432 \
    -e POSTGRES_DATABASE=dbname \
    -p 8080:8080 \
    avito-internship-app:latest
    ```