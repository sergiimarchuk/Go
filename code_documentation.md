# 📚 Документ №1: Полное описание кода проекта "Трекер рабочего времени"

## Оглавление
1. [Архитектура проекта](#архитектура)
2. [Структура файлов](#структура)
3. [Описание Go файлов](#go-файлы)
4. [Описание шаблонов](#шаблоны)
5. [База данных](#база-данных)
6. [API документация](#api)
7. [Безопасность](#безопасность)

---

## Архитектура проекта

**Паттерн:** MVC (Model-View-Controller)
**Фреймворк:** Gin (Go)
**База данных:** SQLite

**Компоненты:**
- Model: `models.go`, `database.go`
- View: HTML шаблоны `templates/`
- Controller: `handlers.go`, `api.go`
- Auth: `auth.go`, `middleware.go`

**Технологии:**
- Go 1.23+, Gin framework
- SQLite (файл database.db)
- Session cookies + JWT
- ECharts 5.4.3 (графики)
- Excelize v2 (Excel)
- bcrypt (шифрование)

---

## Структура файлов

```
my-tracker/
├── main.go              # Точка входа, роуты
├── models.go            # Структуры данных
├── database.go          # Работа с БД
├── auth.go              # Аутентификация
├── handlers.go          # Web обработчики
├── api.go               # REST API
├── middleware.go        # Middleware
├── go.mod               # Зависимости
├── database.db          # SQLite БД
├── templates/           # HTML шаблоны
│   ├── index.html
│   ├── login.html
│   ├── register.html
│   ├── dashboard.html
│   ├── new_worklog.html
│   ├── edit_worklog.html
│   ├── worklog_list.html
│   └── reports.html
└── static/              # CSS, JS
```

---

## Описание Go файлов

### 1. main.go

**Функции:**
- `func main()` - инициализация приложения

**Что делает:**
1. Инициализирует БД
2. Создаёт Gin engine
3. Настраивает сессии (session cookie)
4. Загружает шаблоны
5. Регистрирует роуты
6. Запускает сервер :8080

**Настройки сессий:**
```go
MaxAge: 0          // Session cookie
HttpOnly: true     // XSS защита
Secure: false      // true для HTTPS
SameSite: Lax      // CSRF защита
```

**Роуты:**

Публичные:
- `GET /` - главная
- `GET/POST /login` - вход
- `GET/POST /register` - регистрация

Защищённые:
- `GET /dashboard` - дашборд
- `GET /worklog/new` - новая запись
- `POST /worklog/create` - создать
- `GET /worklog/list` - список
- `GET /worklog/edit/:id` - редактировать
- `POST /worklog/update/:id` - обновить
- `POST /worklog/delete/:id` - удалить
- `GET /worklog/export` - Excel
- `GET /reports` - отчёты
- `GET /logout` - выход

API:
- `POST /api/v1/auth/login` - JWT логин
- `POST /api/v1/auth/register` - регистрация
- `GET /api/v1/worklogs` - список (JWT)
- `POST /api/v1/worklogs` - создать (JWT)
- `PUT /api/v1/worklogs/:id` - обновить (JWT)
- `DELETE /api/v1/worklogs/:id` - удалить (JWT)
- `GET /api/v1/stats` - статистика (JWT)

---

### 2. models.go

```go
type User struct {
    ID       int
    Username string
    Password string  // bcrypt hash
}

type WorkLog struct {
    ID          int
    UserID      int
    Date        time.Time
    Description string
    Hours       float64
}
```

---

### 3. database.go

**Функции:**
- `InitDB()` - инициализация БД

**Глобальная переменная:**
```go
var db *sql.DB
```

**Таблицы:**

users:
```sql
id INTEGER PRIMARY KEY
username TEXT UNIQUE NOT NULL
password TEXT NOT NULL
```

worklogs:
```sql
id INTEGER PRIMARY KEY
user_id INTEGER NOT NULL
date TEXT NOT NULL
description TEXT
hours REAL NOT NULL
FOREIGN KEY (user_id) REFERENCES users(id)
```

**Путь к БД:** `DATABASE_PATH` env или `./database.db`

---

### 4. auth.go

**Функции:**

`HashPassword(password string) (string, error)`
- bcrypt hash (cost 14)

`CheckPassword(password, hash string) bool`
- Проверка пароля

`CreateUser(username, password string) error`
- Создание пользователя

`GetUserByUsername(username string) (*User, error)`
- Получить пользователя

`AuthRequired() gin.HandlerFunc`
- Middleware защиты роутов

`GetCurrentUserID(c *gin.Context) int`
- ID из сессии

`GetCurrentUsername(c *gin.Context) string`
- Username из сессии

---

### 5. middleware.go

`CheckInactivity() gin.HandlerFunc`
- Автологаут через 30 минут
- Обновляет last_activity
- Редирект на /login?timeout=1

---

### 6. handlers.go

**Публичные:**

`HomePage` - главная страница
`LoginPage` - форма входа
`LoginHandler` - обработка входа
`RegisterPage` - форма регистрации
`RegisterHandler` - обработка регистрации

**Защищённые:**

`DashboardPage` - дашборд
`NewWorkLogPage` - форма записи
`CreateWorkLogHandler` - создание
`WorkLogListPage` - список (с фильтрами)
`EditWorkLogPage` - форма редактирования
`UpdateWorkLogHandler` - обновление
`DeleteWorkLogHandler` - удаление
`ExportWorkLogHandler` - экспорт Excel
`ReportsPage` - отчёты с графиками
`LogoutHandler` - выход

**Фильтры в WorkLogListPage:**
- date_from - от даты
- date_to - до даты
- search - поиск по description

---

### 7. api.go

**JWT:**
- Secret: `jwtSecret`
- Алгоритм: HS256
- Срок: 24 часа

**Структура:**
```go
type Claims struct {
    UserID   int
    Username string
    jwt.RegisteredClaims
}
```

**Функции:**

`GenerateJWT(userID, username) (string, error)`
`ValidateJWT(tokenString) (*Claims, error)`
`JWTAuthMiddleware() gin.HandlerFunc`

**API Handlers:**

`APILogin` - вход, возврат JWT
`APIRegister` - регистрация, возврат JWT
`APIGetWorkLogs` - список записей
`APICreateWorkLog` - создать
`APIUpdateWorkLog` - обновить
`APIDeleteWorkLog` - удалить
`APIGetStats` - статистика

---

## Описание шаблонов

Все HTML самодостаточные (без base.html).

**Публичные:**
- `index.html` - главная (фиолетовый фон)
- `login.html` - форма входа
- `register.html` - форма регистрации

**Защищённые:**
- `dashboard.html` - 3 карточки
- `new_worklog.html` - форма новой записи
- `edit_worklog.html` - форма редактирования
- `worklog_list.html` - список + фильтры + Excel
- `reports.html` - 4 графика ECharts

Все защищённые страницы имеют кнопку "🚪 Выйти"

---

## База данных

**Тип:** SQLite
**Файл:** database.db

**Таблица users:**
- id (PK)
- username (UNIQUE)
- password (bcrypt hash)

**Таблица worklogs:**
- id (PK)
- user_id (FK)
- date (TEXT: YYYY-MM-DD)
- description
- hours (REAL)

---

## API документация

**Base:** `/api/v1`
**Auth:** `Authorization: Bearer <JWT>`

### POST /auth/register
Request:
```json
{
  "username": "user",
  "password": "pass123"
}
```
Response:
```json
{
  "token": "eyJ...",
  "user": {"id": 1, "username": "user"}
}
```

### POST /auth/login
Аналогично register

### GET /worklogs
Query: date_from, date_to, search
Response:
```json
{
  "data": [
    {"id": 1, "date": "2025-11-20", "description": "...", "hours": 8}
  ]
}
```

### POST /worklogs
Request:
```json
{
  "date": "2025-11-20",
  "description": "Работа",
  "hours": 8.5
}
```

### PUT /worklogs/:id
Обновление записи

### DELETE /worklogs/:id
Удаление записи

### GET /stats
Response:
```json
{
  "total_hours": 120,
  "days_count": 15,
  "avg_hours": 8.0
}
```

**Ошибки:**
- 400 - Bad Request
- 401 - Unauthorized
- 404 - Not Found
- 409 - Conflict
- 500 - Server Error

---

## Безопасность

1. **Bcrypt** - пароли (cost 14)
2. **HttpOnly cookies** - XSS защита
3. **SameSite: Lax** - CSRF защита
4. **Session cookie** - умирает при закрытии браузера
5. **Автологаут** - 30 минут неактивности
6. **JWT** - 24 часа срок
7. **Prepared statements** - SQL инъекции
8. **Validation** - клиент + сервер

**TODO:**
- Rate limiting
- HTTPS (Secure cookies)
- 2FA
- Email verification
- Password reset

---

**Версия:** 1.0
**Дата:** 2025-11-26