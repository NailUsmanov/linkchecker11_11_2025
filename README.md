# LinkChecker

HTTP-сервис для проверки доступности ссылок и формирования отчётов в PDF.  
Сервис принимает список ссылок, проверяет их доступность, сохраняет данные и позволяет получить PDF-отчёт по номеру запроса.  
Решение соответствует требованиям задания: без Docker, БД и внешней инфраструктуры — всё хранится в файле.

---

## Архитектура

```
main.go
  ↓
App (router, graceful shutdown)
  ↓
Handlers (POST /links, GET /links_num)
  ↓
Service (CheckLink, CreatePDF)
  ↓
Storage (FileStorage + in-memory кэш)
```

- **handlers** — принимают HTTP‑запросы, валидируют входные данные, вызывают сервис.
- **service** — бизнес‑логика: проверка ссылок, создание PDF.
- **storage** — file‑based хранилище, которое переживает перезапуск сервиса.
- **graceful shutdown** — незавершённые запросы завершаются корректно.

---

## Маршруты

### POST `/links`

Проверяет список URL.

**Пример запроса**

```json
{
  "links": ["google.com", "malformedlink.gg"]
}
```

**Пример ответа**

```json
{
  "links": {
    "google.com": "available",
    "malformedlink.gg": "not available"
  },
  "links_num": 1
}
```

---

### GET `/links_num`

Получает PDF‑отчёт по номерам запросов.

**Пример запроса**

```json
{
  "links_list": [1]
}
```

**Ответ**

- `200 OK`
- `Content-Type: application/pdf`
- `Content-Disposition: attachment; filename=report.pdf`

---

## Хранение данных

### FileStorage

- Все результаты сохраняются в файл `storage.json`.
- Формат хранения:

```json
{
  "1": {
    "links_num": 1,
    "links": {
      "google.com": "available",
      "ya.ru": "not available"
    }
  }
}
```

- При старте сервиса файл загружается обратно в память.
- Используется атомарная запись через временный файл + `os.Rename()`.

Таким образом сервис **переживает перезагрузку**, данные не теряются.

---

## Graceful Shutdown

Сервер завершает работу корректно:

- перестаёт принимать новые подключения,
- ждёт завершения текущих запросов (таймаут 5 секунд),
- только потом останавливается.

Это полностью соответствует ТЗ пункту про «не потерять задачи во время остановки».

---

## Команды CURL

### Проверка ссылок

```bash
curl -X POST http://localhost:8080/links   -H "Content-Type: application/json"   -d '{"links":["google.com","ya.ru"]}'
```

### Получение PDF по номеру

```bash
curl -X GET http://localhost:8080/links_num   -H "Content-Type: application/json"   -d '{"links_list":[1]}'   --output report.pdf
```

---

## Почему выбрано файловое хранилище

- ТЗ запрещает Docker, базы данных и внешние сервисы.
- Важна устойчивость к перезапуску — значит, нужен файл.
- FileStorage — простое, понятное и надёжное решение.

---

## Тестирование

Покрыты следующие части:

- handler POST /links
- handler GET /links_num
- storage (memory)
- service.CreatePDF (генерация PDF)
- часть негативных сценариев (битый JSON, неверный Content-Type)

Планы расширения:

- тестирование CheckLink через `httptest.Server`
- e2e‑тест приложения через `httptest.NewServer`

---

## Запуск

```bash
go run ./cmd/app
```

После запуска сервер доступен по адресу:

```
http://localhost:8080
```

---

## Минимальные требования

- Go 1.22+
- Без зависимостей на внешние БД или Docker
- Кроссплатформенное решение
