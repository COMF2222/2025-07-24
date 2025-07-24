
# Архиватор файлов по ссылкам

## Описание
Данная система позволяет пользователям создавать задачи на скачивание и упаковку в zip-архив нескольких файлов из интернета по предоставленным ссылкам. Поддерживаются файлы с расширениями `.pdf` и `.jpeg`.

Основные функции:
- Создание задачи на упаковку файлов.
- Добавление ссылок на файлы в задачу.
- Получение статуса задачи, включая ссылку на готовый архив.
- Ограничение на максимальное количество одновременно обрабатываемых задач (3).
- Ограничение на максимальное количество файлов в задаче (3).
- Отчёт об ошибках скачивания отдельных файлов, при этом архив создаётся из доступных файлов.

## Запуск сервера
Сервер запускается на порту `8080`.

```bash
cd 2025-07-24
go run /cmd/server/main.go
```

## Конфигурация
Конфигурация задаётся через файл `config.json`. Пример:

```json
{
  "port": "8080",
  "allowed_extensions": [".pdf", ".jpeg"],
  "max_tasks": 3,
  "max_files_per_task": 3
}
```

## API

### 1. Создать задачу

```
POST /tasks
```

**Ответ:**

```json
{
  "id": "uuid задачи"
}
```

### 2. Добавить ссылку в задачу

```
POST /tasks/{id}/links
Content-Type: application/json

{
  "link": "http://example.com/file.pdf"
}
```

**Ответ:**

```json
{
  "message": "сслыка добавлена"
}
```

### 3. Получить статус задачи

```
GET /tasks/{id}/status
```

**Ответ:**

```json
{
  "ID": "uuid задачи",
  "Status": "created|processing|completed|failed",
  "Links": [
    "http://example.com/file1.pdf",
    "http://example.com/file2.jpeg",
    "http://example.com/file3.pdf"
  ],
  "ArchivePath": "http://localhost:8080/archives/{id}" (если completed),
  "Errors": ["список ошибок скачивания файлов"]
}
```

### 4. Скачать архив

```
GET /archives/{id}
```

Загружает zip архив, если задача выполнена.

## Ограничения

- Одновременно обрабатывается не более 3 задач.
- В одной задаче максимум 3 файла.
- Поддерживаются только `.pdf` и `.jpeg` файлы.

## Примеры curl (Windows PowerShell)

```powershell
# Создать задачу
curl -Method POST http://localhost:8080/tasks

# Добавить ссылку в задачу
curl -Method POST http://localhost:8080/tasks/{id}/links -Body '{"link":"http://example.com/file.pdf"}' -ContentType 'application/json'

# Получить статус задачи
curl http://localhost:8080/tasks/{id}/status

# Скачать архив
curl -OutFile archive.zip http://localhost:8080/archives/{id}
```

Файл сохранится туда откуда вы запускали PowerShell

Если же зайти в браузере по ссылке `http://localhost:8080/archives/{id}` то сохранится в загрузки