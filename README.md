# Address Book Server

## Предварительные требования

Для начала убедитесь, что у вас установлен [Docker](https://docs.docker.com/engine/install/).
```
user@pc:~$ docker --version
Docker version 24.0.7, build afdd53b
user@pc:~$
```

### Запуск PostgreSQL с использованием Docker

#### Запуск контейнера

Для создания контейнера с [`postgresql:16`](https://hub.docker.com/layers/library/postgres/16/images/sha256-c5f76d46d12230623ddccc341e3d11227258e07ce1c8a11f2d30e3a5aa15627a?context=explore) используйте следующую команду:
```bash
docker run -d --name address_book_db13 -p 5432:5432 -e POSTGRES_PASSWORD=qwerty -e POSTGRES_USER=postgres -e POSTGRES_DB=postgres postgres:16;
```
_Если возникает ошибка `Bind for 0.0.0.0:5432 failed: port is already allocated.`, то изменить `-p 5432:5432`, например, на `-p 5438:5432`_

#### Создание таблицы в базе данных

Для создания нужной таблицы в новой базе данных используем следующую команду:
```bash
docker exec -i address_book_db13 psql -U postgres -d postgres -c 'CREATE TABLE address_book (id SERIAL PRIMARY KEY, name VARCHAR(255), last_name VARCHAR(255), middle_name VARCHAR(255), address VARCHAR(255), phone VARCHAR(20));';
```

Эта команда запустит контейнер PostgreSQL с указанными конфигурациями. Сервер будет доступен на порту 5432.

## Запуск Address Book Server

Когда контейнер работает, вы можете запустить `addressBookServer`.

Выполните следующую команду в директории проекта.
```bash
go run addressBookServer
```

- Хост базы данных: `localhost` (если запущено на том же компьютере)
- Порт базы данных: `5432` (или, например, `5438`)
- Имя базы данных: `postgres`
- Пользователь базы данных: `postgres`
- Пароль пользователя: `qwerty`

## Использование с помощью Postman

Можно импортировать в [Postman](https://www.postman.com/downloads/) коллекцию запросов из файла [`addressBook.postman_collection.json`](addressBook.postman_collection.json).

В данной коллекции представлены только положительные сценарии.

## Завершение

Чтобы остановить и удалить контейнер используйте следующую команду:

```bash
docker stop address_book_db13 && docker rm address_book_db13;
```