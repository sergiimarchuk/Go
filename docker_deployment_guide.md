# 🐳 Документ №2: Развёртывание в Docker на openSUSE (Рабочая инструкция)

## Что получится

После выполнения всех шагов:
- Приложение работает в Docker контейнерах
- Доступно по адресу: **http://192.168.100.60**
- Nginx проксирует запросы на Go приложение
- База данных хранится в persistent volume
- Автоматический перезапуск контейнеров

---

## Структура проекта (финальная)

```
/opt/dev-py/tempo_Go/my-tracker/
├── main.go
├── models.go
├── database.go
├── auth.go
├── handlers.go
├── api.go
├── middleware.go
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
├── .dockerignore
├── nginx.conf
├── nginx-site.conf
├── .env
├── templates/
│   ├── index.html
│   ├── login.html
│   ├── register.html
│   ├── dashboard.html
│   ├── new_worklog.html
│   ├── edit_worklog.html
│   ├── worklog_list.html
│   └── reports.html
├── static/
│   └── style.css
└── data/
    └── database.db
```

---

## Шаг 1: Подготовка на локальной машине

### 1.1 Создаём Dockerfile

```dockerfile
# Файл: Dockerfile

# Multi-stage build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o worklog-tracker .

# Production image
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs tzdata && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/worklog-tracker .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

RUN mkdir -p /app/data && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./worklog-tracker"]
```

### 1.2 Создаём docker-compose.yml

```yaml
# Файл: docker-compose.yml

version: '3.8'

services:
  worklog-tracker:
    build: .
    container_name: worklog-tracker
    restart: always
    ports:
      - "127.0.0.1:8080:8080"
    volumes:
      - ./data:/app/data
    environment:
      - GIN_MODE=release
      - TZ=Europe/Berlin
      - DATABASE_PATH=/app/data/database.db
    networks:
      - worklog-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  nginx:
    image: nginx:alpine
    container_name: worklog-nginx
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx-site.conf:/etc/nginx/conf.d/default.conf:ro
    depends_on:
      - worklog-tracker
    networks:
      - worklog-network
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

networks:
  worklog-network:
    driver: bridge
```

### 1.3 Создаём .dockerignore

```
# Файл: .dockerignore

*.db
*.db-journal
*.tar.gz
*.back*
*.bak
.git
.gitignore
.env
data/
go.sum.bak
README.md
docker-compose.yml
Dockerfile
nginx*.conf
```

### 1.4 Создаём nginx.conf

```nginx
# Файл: nginx.conf

user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;

    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 20M;

    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml text/javascript 
               application/json application/javascript application/xml+rss;

    include /etc/nginx/conf.d/*.conf;
}
```

### 1.5 Создаём nginx-site.conf

```nginx
# Файл: nginx-site.conf

upstream worklog_backend {
    server worklog-tracker:8080;
}

server {
    listen 80;
    server_name _;

    access_log /var/log/nginx/worklog-access.log;
    error_log /var/log/nginx/worklog-error.log;

    client_max_body_size 10M;

    location / {
        proxy_pass http://worklog_backend;
        proxy_http_version 1.1;
        
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_cache_bypass $http_upgrade;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    location /static/ {
        proxy_pass http://worklog_backend;
        proxy_cache_valid 200 1d;
        add_header Cache-Control "public, immutable";
    }

    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
```

### 1.6 Обновляем database.go (важно!)

```go
// Файл: database.go

package main

import (
    "database/sql"
    "os"
    _ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() error {
    var err error
    
    // Путь к базе из переменной окружения
    dbPath := os.Getenv("DATABASE_PATH")
    if dbPath == "" {
        dbPath = "./database.db"
    }
    
    db, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        return err
    }

    // Создаём таблицы
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL
        )
    `)
    if err != nil {
        return err
    }

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS worklogs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            date TEXT NOT NULL,
            description TEXT,
            hours REAL NOT NULL,
            FOREIGN KEY (user_id) REFERENCES users(id)
        )
    `)
    
    return err
}
```

---

## Шаг 2: Перенос на сервер

### 2.1 На локальной машине (Ubuntu)

```bash
cd /opt/dev-py/tempo_Go

# Копируем весь проект на сервер
rsync -av my-tracker/ root@192.168.100.60:/opt/dev-py/tempo_Go/my-tracker/

# Копируем базу данных (если нужны существующие пользователи)
scp my-tracker/database.db root@192.168.100.60:/opt/dev-py/tempo_Go/my-tracker/data/database.db
```

---

## Шаг 3: Установка Docker на openSUSE

### 3.1 Подключаемся к серверу

```bash
ssh root@192.168.100.60
```

### 3.2 Устанавливаем Docker

```bash
# Установка Docker и Docker Compose
sudo zypper install docker docker-compose

# Добавляем пользователя в группу docker (опционально)
sudo usermod -aG docker $USER

# Включаем и запускаем Docker
sudo systemctl enable docker
sudo systemctl start docker

# Проверяем
docker --version
docker-compose --version
```

---

## Шаг 4: Запуск на сервере

### 4.1 Переходим в папку проекта

```bash
cd /opt/dev-py/tempo_Go/my-tracker
```

### 4.2 Создаём папки для данных

```bash
mkdir -p data
chmod 777 data
```

### 4.3 Останавливаем старые контейнеры (если были)

```bash
docker-compose down
docker stop worklog-tracker 2>/dev/null || true
docker rm worklog-tracker 2>/dev/null || true
docker stop worklog-nginx 2>/dev/null || true
docker rm worklog-nginx 2>/dev/null || true
```

### 4.4 Собираем образы

```bash
docker-compose build --no-cache
```

Это займёт 5-10 минут. Должно быть:
```
Successfully built xxx
Successfully tagged my-tracker-worklog-tracker:latest
```

### 4.5 Запускаем контейнеры

```bash
docker-compose up -d
```

### 4.6 Проверяем что запустились

```bash
docker-compose ps
```

Должно быть:
```
NAME                  STATUS    PORTS
worklog-tracker       Up        127.0.0.1:8080->8080/tcp
worklog-nginx         Up        0.0.0.0:80->80/tcp
```

### 4.7 Смотрим логи

```bash
# Все контейнеры
docker-compose logs -f

# Только приложение
docker-compose logs -f worklog-tracker

# Только Nginx
docker-compose logs -f nginx
```

Должно быть:
```
🚀 Сервер запущен на http://localhost:8080
📡 API доступно на http://localhost:8080/api/v1
```

---

## Шаг 5: Проверка работы

### 5.1 На сервере

```bash
# Проверяем что отвечает
curl http://localhost
curl http://192.168.100.60

# Проверяем health check
curl http://localhost/health
```

### 5.2 С локальной машины

Открой браузер: **http://192.168.100.60**

Должна открыться главная страница с кнопкой "Войти"

### 5.3 Создание первого пользователя

**Вариант 1:** Регистрация через веб
- Открой http://192.168.100.60/register
- Создай пользователя

**Вариант 2:** Скопировать базу с локальной машины (уже сделали в Шаге 2.1)

---

## Управление контейнерами

### Основные команды

```bash
# Запустить
docker-compose up -d

# Остановить
docker-compose down

# Перезапустить
docker-compose restart

# Перезапустить один контейнер
docker-compose restart worklog-tracker

# Пересобрать и запустить
docker-compose up -d --build

# Посмотреть статус
docker-compose ps

# Посмотреть логи
docker-compose logs -f

# Посмотреть использование ресурсов
docker stats

# Зайти внутрь контейнера
docker exec -it worklog-tracker sh

# Удалить всё и начать заново
docker-compose down -v
docker system prune -a
```

---

## Управление данными

### Бэкап базы данных

```bash
# На сервере
cd /opt/dev-py/tempo_Go/my-tracker
cp data/database.db data/database.db.backup-$(date +%Y%m%d)

# Скопировать на локальную машину
scp root@192.168.100.60:/opt/dev-py/tempo_Go/my-tracker/data/database.db ./backup-$(date +%Y%m%d).db
```

### Восстановление базы

```bash
# На сервере
docker-compose down
cp data/database.db.backup-20251126 data/database.db
chown 1000:1000 data/database.db
chmod 664 data/database.db
docker-compose up -d
```

---

## Troubleshooting

### Проблема: Порт 80 занят

```bash
# Проверяем кто слушает порт 80
ss -tlnp | grep :80

# Останавливаем Apache/другой веб-сервер
systemctl stop apache2
systemctl stop httpd
systemctl disable apache2
```

### Проблема: Контейнер не запускается

```bash
# Смотрим логи
docker-compose logs worklog-tracker

# Проверяем ошибки сборки
docker-compose build --no-cache
```

### Проблема: База данных пустая

```bash
# Проверяем права
ls -la data/database.db

# Исправляем права
chown 1000:1000 data/database.db
chmod 664 data/database.db

# Перезапускаем
docker-compose restart worklog-tracker
```

### Проблема: Не работает логин

```bash
# Заходим в контейнер
docker exec -it worklog-tracker sh

# Проверяем базу
ls -la /app/data/

# Проверяем переменные окружения
env | grep DATABASE

# Выходим
exit
```

### Проблема: Nginx не проксирует

```bash
# Проверяем логи Nginx
docker-compose logs nginx

# Проверяем конфиг
docker exec worklog-nginx nginx -t

# Перезапускаем Nginx
docker-compose restart nginx
```

---

## Мониторинг

### Проверка здоровья

```bash
# Health check
curl http://localhost/health

# Статус контейнеров
docker-compose ps

# Использование ресурсов
docker stats --no-stream
```

### Просмотр логов

```bash
# Последние 100 строк
docker-compose logs --tail=100

# Логи с временными метками
docker-compose logs -f -t

# Логи за последний час
docker-compose logs --since 1h

# Логи только ошибок
docker-compose logs | grep -i error
```

### Логи Nginx

```bash
# Access log
docker exec worklog-nginx tail -f /var/log/nginx/access.log

# Error log
docker exec worklog-nginx tail -f /var/log/nginx/error.log
```

---

## Обновление приложения

### Процесс обновления

**На локальной машине:**
```bash
# 1. Изменяем код
# 2. Тестируем локально: go run .
# 3. Копируем на сервер
cd /opt/dev-py/tempo_Go
rsync -av --exclude='data/' --exclude='*.db' my-tracker/ root@192.168.100.60:/opt/dev-py/tempo_Go/my-tracker/
```

**На сервере:**
```bash
cd /opt/dev-py/tempo_Go/my-tracker

# Пересобираем и перезапускаем
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Проверяем логи
docker-compose logs -f
```

---

## Firewall (опционально)

Если включён firewall на openSUSE:

```bash
# Открываем порты
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload

# Проверяем
sudo firewall-cmd --list-all
```

---

## Автозапуск при перезагрузке сервера

Docker Compose автоматически настроен на перезапуск (`restart: always`).

Проверка:
```bash
# Перезагружаем сервер
sudo reboot

# После перезагрузки
docker-compose ps

# Контейнеры должны быть в статусе "Up"
```

---

## Чек-лист финальной проверки

- [ ] Контейнеры запущены: `docker-compose ps`
- [ ] Логи без ошибок: `docker-compose logs`
- [ ] Главная страница открывается: http://192.168.100.60
- [ ] Логин работает
- [ ] Создание записей работает
- [ ] Графики отображаются
- [ ] Экспорт в Excel работает
- [ ] API возвращает JSON: `curl http://192.168.100.60/api/v1/auth/login`
- [ ] База данных сохраняется: `ls -la data/database.db`
- [ ] Автологаут через 30 минут работает
- [ ] Кнопка "Выйти" работает

---

## Итоговая конфигурация

**Сервер:** openSUSE Leap 15.6  
**IP:** 192.168.100.60  
**Порты:**
- 80 - Nginx (HTTP)
- 8080 - Go приложение (localhost only)

**URL:**
- Web: http://192.168.100.60
- API: http://192.168.100.60/api/v1
- Health: http://192.168.100.60/health

**Данные:**
- База: `/opt/dev-py/tempo_Go/my-tracker/data/database.db`
- Логи: `docker-compose logs`

---

**Дата:** 2025-11-26  
**Версия:** 1.0 (Production Ready)