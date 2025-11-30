# 📚 

## 1.
2. [structure files](#structure)
3. [descript. Go files](#go-files)
4. [decr. templ.](#templates)
5. [db](#db)
6. [API (#api)
7. [security](#security)

---

## arc. proj.

**pattern:** MVC (Model-View-Controller)
**framework:** Gin (Go)
**db:** SQLite

**parts:**
- Model: `models.go`, `database.go`
- View: HTML `templates/`
- Controller: `handlers.go`, `api.go`
- Auth: `auth.go`, `middleware.go`

**tech:**
- Go 1.23+, Gin framework
- SQLite (файл database.db)
- Session cookies + JWT
- ECharts 5.4.3 (charts)
- Excelize v2 (Excel)
- bcrypt (crypt)

---

## struct files

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

## descr. files

### 1. main.go

**:**
- `func main()` - initions application

**:**
1. initions db
2. Creation Gin engine
3. (session cookie)
4. Load templates
5. registered routes
6. Run server :8080

**Configured session:**
```go
MaxAge: 0          // Session cookie
HttpOnly: true     // XSS защита
Secure: false      // true для HTTPS
SameSite: Lax      // CSRF защита
```

**Route:**

Public:
- `GET /` - main page
- `GET/POST /login` - easy understand what is it and for for what
- `GET/POST /register` - registration

Secured:
- `GET /dashboard` - 
- `GET /worklog/new` - 
- `POST /worklog/create` - 
- `GET /worklog/list` - 
- `GET /worklog/edit/:id` - 
- `POST /worklog/update/:id` - 
- `POST /worklog/delete/:id` - 
- `GET /worklog/export` - Excel
- `GET /reports` - 
- `GET /logout` - 

API:
- `POST /api/v1/auth/login` - JWT 
- `POST /api/v1/auth/register` - 
- `GET /api/v1/worklogs` -  (JWT)
- `POST /api/v1/worklogs` -  (JWT)
- `PUT /api/v1/worklogs/:id` -  (JWT)
- `DELETE /api/v1/worklogs/:id` -  (JWT)
- `GET /api/v1/stats` -  (JWT)

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

**functions:**
- `InitDB()` - 

**global ver:**
```go
var db *sql.DB
```

**tables:**

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

**path:** `DATABASE_PATH` env или `./database.db`

---

### 4. auth.go

**functions:**

`HashPassword(password string) (string, error)`
- bcrypt hash (cost 14)

`CheckPassword(password, hash string) bool`
- check password

`CreateUser(username, password string) error`
- creation user

`GetUserByUsername(username string) (*User, error)`
- get usernanme

`AuthRequired() gin.HandlerFunc`
- Middleware secured routes

`GetCurrentUserID(c *gin.Context) int`
- ID from session

`GetCurrentUsername(c *gin.Context) string`
- Username from session

---

### 5. middleware.go

`CheckInactivity() gin.HandlerFunc`
- logout in 30 miniutes
- update last_activity
- redirect /login?timeout=1

---

### 6. handlers.go

**public:**

`HomePage` - 
`LoginPage` - 
`LoginHandler` - 
`RegisterPage` - 
`RegisterHandler` - 

**secured:**

`DashboardPage` - 
`NewWorkLogPage` - 
`CreateWorkLogHandler` -
`WorkLogListPage` - 
`EditWorkLogPage` - 
`UpdateWorkLogHandler` - 
`DeleteWorkLogHandler` - 
`ExportWorkLogHandler` - 
`ReportsPage` - 
`LogoutHandler` - 

**Фильтры в WorkLogListPage:**
- date_from -
- date_to - 
- search - 

---

### 7. api.go

**JWT:**
- Secret: `jwtSecret`
- algoritm: HS256
- term: 24 часа

**Структура:**
```go
type Claims struct {
    UserID   int
    Username string
    jwt.RegisteredClaims
}
```

**functions:**

`GenerateJWT(userID, username) (string, error)`
`ValidateJWT(tokenString) (*Claims, error)`
`JWTAuthMiddleware() gin.HandlerFunc`

**API Handlers:**

`APILogin` -  JWT
`APIRegister` - JWT
`APIGetWorkLogs` -  
`APICreateWorkLog` - 
`APIUpdateWorkLog` - 
`APIDeleteWorkLog` - 
`APIGetStats` - 

---

## updated 

All HTML (without base.html).

**public:**
- `index.html` -  (purpul background)
- `login.html` - форма 
- `register.html` - форма 

**secured:**
- `dashboard.html` - 
- `new_worklog.html` -
- `edit_worklog.html` - 
- `worklog_list.html` - list + filter + Excel
- `reports.html` - 4 ECharts

---

## 

**:** SQLite
**:** database.db

** users:**
- id (PK)
- username (UNIQUE)
- password (bcrypt hash)

** worklogs:**
- id (PK)
- user_id (FK)
- date (TEXT: YYYY-MM-DD)
- description
- hours (REAL)

---

## API 

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
 register

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


### DELETE /worklogs/:id


### GET /stats
Response:
```json
{
  "total_hours": 120,
  "days_count": 15,
  "avg_hours": 8.0
}
```

**:**
- 400 - Bad Request
- 401 - Unauthorized
- 404 - Not Found
- 409 - Conflict
- 500 - Server Error

---

## 

1. **Bcrypt** -  (cost 14)
2. **HttpOnly cookies** - XSS 
3. **SameSite: Lax** - CSRF 
4. **Session cookie** - 
5. **autologaout** - 30 
6. **JWT** - 24 
7. **Prepared statements** - SQL 
8. **Validation** -  + 

**TODO:**
- Rate limiting
- HTTPS (Secure cookies)
- 2FA
- Email verification
- Password reset

---

