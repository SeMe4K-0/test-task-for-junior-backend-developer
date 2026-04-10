# pgAdmin 4 — подключение к PostgreSQL

Эта инструкция описывает, как получить доступ к `pgAdmin` и подключить его к базе данных из Docker Compose.

## Запуск сервисов

В корне проекта выполните:

```bash
docker compose up --build
```

После запуска будут доступны:

- `pgAdmin`: http://localhost:5050
- `PostgreSQL`: порт `5433` на хосте

## Вход в pgAdmin

Откройте в браузере:

```text
http://localhost:5050
```

Используйте учётные данные:

- Email: `admin@local`
- Password: `admin`

## Как добавить сервер PostgreSQL в pgAdmin

1. Нажмите на `Add New Server` или щёлкните правой кнопкой по `Servers` и выберите `Create > Server`.
2. На вкладке **General** введите любое имя, например `taskservice-postgres`.
3. На вкладке **Connection** укажите:
   - Host name/address: `postgres`
   - Port: `5432`
   - Maintenance database: `taskservice`
   - Username: `postgres`
   - Password: `postgres`
4. Нажмите `Save`.

> Внутри Docker-сети `pgAdmin` видит сервис `postgres` по имени хоста `postgres`, а не `localhost`.

## Альтернативное подключение из внешнего pgAdmin

Если вы используете установленный локально `pgAdmin` на вашем компьютере, подключайтесь к базе так:

- Host name/address: `localhost`
- Port: `5433`
- Maintenance database: `taskservice`
- Username: `postgres`
- Password: `postgres`

## Частые проблемы и решение

- Если форма `Add New Server` не открывается, попробуйте:
  - обновить страницу;
  - открыть `http://localhost:5050` в другом браузере или в режиме инкогнито;
  - проверить консоль браузера на ошибки JavaScript;
  - перезапустить контейнер `taskservice-pgadmin`.

- Если подключение не проходит, проверьте состояние контейнеров:

```bash
docker compose ps
```

- Если база не инициализируется, перезапустите стек с очисткой volume:

```bash
docker compose down -v
docker compose up --build
```
