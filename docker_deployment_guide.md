# 🐳  №2: Docker  openSUSE 

## 

- Docker 
- **http://192.168.100.60**
- Nginx Go 
- persistent volume

---

## 
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

## 1: 

### 1.1  Dockerfile

```dockerfile
# : Dockerfile

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

### 1.2  docker-compose.yml

```yaml
# docker-compose.yml

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

### 1.3  .dockerignore

```
# .dockerignore

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

### 1.4  nginx.conf

```nginx
# nginx.conf

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

### 1.5  nginx-site.conf

```nginx
# : nginx-site.conf

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

### 1.6  database.go (!!!)

```go
// : database.go

package main

import (
    "database/sql"
    "os"
    _ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() error {
    var err error
    
    // oath to db in environment 
    dbPath := os.Getenv("DATABASE_PATH")
    if dbPath == "" {
        dbPath = "./database.db"
    }
    
    db, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        return err
    }

    // tables
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

## 2: move to openSuse into docker  

### 2.1 this is on local machine (Ubuntu)

```bash
cd /opt/dev-py/tempo_Go

# to srv
rsync -av my-tracker/ root@192.168.100.60:/opt/dev-py/tempo_Go/my-tracker/

# if needs copy db
scp my-tracker/database.db root@192.168.100.60:/opt/dev-py/tempo_Go/my-tracker/data/database.db
```

---

## 3:  Docker на openSUSE

### 3.1

```bash
ssh root@192.168.100.60
```

### 3.2  Docker

```bash
#  Docker Docker Compose
sudo zypper install docker docker-compose

# docker (optional)
sudo usermod -aG docker $USER

#  Docker
sudo systemctl enable docker
sudo systemctl start docker

# 
docker --version
docker-compose --version
```

---

## 

4: 

### 4.1 

```bash
cd /opt/dev-py/tempo_Go/my-tracker
```

### 4.2 
```bash
mkdir -p data
chmod 777 data
```

### 4.3 

```bash
docker-compose down
docker stop worklog-tracker 2>/dev/null || true
docker rm worklog-tracker 2>/dev/null || true
docker stop worklog-nginx 2>/dev/null || true
docker rm worklog-nginx 2>/dev/null || true
```

### 4.4 Creation images !!!! 

```bash
docker-compose build --no-cache
```


```
Successfully built xxx
Successfully tagged my-tracker-worklog-tracker:latest
```

### 4.5 

```bash
docker-compose up -d
```

### 4.6 

```bash
docker-compose ps
```

 :
```
NAME                  STATUS    PORTS
worklog-tracker       Up        127.0.0.1:8080->8080/tcp
worklog-nginx         Up        0.0.0.0:80->80/tcp
```

### 4.7  

```bash
#  
docker-compose logs -f

#  
docker-compose logs -f worklog-tracker

#   Nginx
docker-compose logs -f nginx
```

 :
```
🚀  http://localhost:8080
📡 API http://localhost:8080/api/v1
```

---

## 5:  

### 5.1  

```bash
#  
curl http://localhost
curl http://192.168.100.60

#   health check
curl http://localhost/health
```

### 5.2  

 : **http://192.168.100.60**

 

### 5.3  

** 1:**  
-  http://192.168.100.60/register

---

### 

```bash
# 
docker-compose up -d

# 
docker-compose down

# 
docker-compose restart

# 
docker-compose restart worklog-tracker

#
docker-compose up -d --build

# 
docker-compose ps

# 
docker-compose logs -f

# 
docker stats

# 
docker exec -it worklog-tracker sh

# Remove all and start again from begining 
docker-compose down -v
docker system prune -a
```

---

### backup

```bash
# 
cd /opt/dev-py/tempo_Go/my-tracker
cp data/database.db data/database.db.backup-$(date +%Y%m%d)

# backup if needs just in case you are not sure and true admin as usual do not do this never
scp root@192.168.100.60:/opt/dev-py/tempo_Go/my-tracker/data/database.db ./backup-$(date +%Y%m%d).db
```

### restore db

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

### port 80 busy

```bash
# check port 80 80
ss -tlnp | grep :80
lsof -i :80

# just f example
systemctl stop apache2
systemctl stop httpd
systemctl disable apache2
```

### in case not running app

```bash
# as you in book write be good to answer on interview 
docker-compose logs worklog-tracker

# not bad take a look 
docker-compose build --no-cache
```

### 

```bash
# 
ls -la data/database.db

# 
chown 1000:1000 data/database.db
chmod 664 data/database.db

# 
docker-compose restart worklog-tracker
```

### login does not work

```bash
# guk mal in container 
docker exec -it worklog-tracker sh

# check db 
ls -la /app/data/

# check env
env | grep DATABASE

# leave without any message 
exit
```

### issue: Nginx not do proxy

```bash
# seems has to check logs this app  -  Nginx
docker-compose logs nginx

# config exam
docker exec worklog-nginx nginx -t

# here we go to restart - Nginx
docker-compose restart nginx
```

---

## like monitoring aa..

### 
```bash
# Health check
curl http://localhost/health

# 
docker-compose ps

# check used resources
docker stats --no-stream
```

### again and again logs 

```bash
# last 100 lines
docker-compose logs --tail=100

# logs with time 
docker-compose logs -f -t

# guess what is it 
docker-compose logs --since 1h

# Nur Fehlerprotokolle
docker-compose logs | grep -i error
```

### logs Nginx

```bash
# Access log
docker exec worklog-nginx tail -f /var/log/nginx/access.log

# Error log
docker exec worklog-nginx tail -f /var/log/nginx/error.log
```

---

## how to upgrade do

### so directly process 

**local machine Ubuntu :**
```bash
# 1. new code
# 2. trying to run: go run .
# 3. move to new server with docker 
cd /opt/dev-py/tempo_Go
rsync -av --exclude='data/' --exclude='*.db' my-tracker/ root@192.168.100.60:/opt/dev-py/tempo_Go/my-tracker/
```

**docker server:**
```bash
cd /opt/dev-py/tempo_Go/my-tracker

# let\s try to rebuild new one app 
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# unglaublich / incredible 
docker-compose logs -f
```

---

## Firewall (depends)

if ON firewall on openSUSE:

```bash
# just open ports
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload

# check it 
sudo firewall-cmd --list-all
```

---

## auto run this on server reboot etc 

Docker Compose - (`restart: always`).

check :
```bash
# init 6
sudo reboot

#after 
docker-compose ps

# containers has to be "Up"
```

---

## check list

- [ ] `docker-compose ps`
- [ ] `docker-compose logs`
- [ ]  http://192.168.100.60
- [ ]  login 
- [ ] new entry into db
- [ ] charts
- [ ] export Excel 
- [ ] API return JSON: `curl http://192.168.100.60/api/v1/auth/login`
- [ ] `ls -la data/database.db`
- [ ] autologaout in 30 
- [ ] logout

---

## so

**server:** openSUSE Leap 15.6  
**IP:** 192.168.100.60  
**ports:**
- 80 - Nginx (HTTP)
- 8080 - Go app (localhost only)

**URL:**
- Web: http://192.168.100.60
- API: http://192.168.100.60/api/v1
- Health: http://192.168.100.60/health

**data:**
- db: `/opt/dev-py/tempo_Go/my-tracker/data/database.db`
- logs: `docker-compose logs`

---

