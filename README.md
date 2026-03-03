# Go User Service

Dockerized Go REST API with PostgreSQL.

## Features

- REST API
- PostgreSQL (pgxpool)
- Docker & Docker Compose
- Retry logic for DB startup
- Connection pool tuning
- Graceful shutdown
- Middleware (Auth + Logging)
- Goroutines with WaitGroup

## Run with Docker

```bash
docker compose up --build
````

Server runs on:

[http://localhost:8080/signup](http://localhost:8080/signup)

````

---

# 🚀 How To Push To GitHub

Inside `D:\go-user-service`

---

### 1️⃣ Initialize Git

```bash
git init
````

---

### 2️⃣ Add Files

```bash
git add .
```

---

### 3️⃣ Commit

```bash
git commit -m "Initial commit - Go user service with Docker"
```

---

### 4️⃣ Create Repo on GitHub

Go to:

👉 [https://github.com/new](https://github.com/new)

Repo name:

```
go-user-service
```

DO NOT initialize with README (since you already have one).

---

### 5️⃣ Connect Remote

GitHub will show commands like:

```bash
git remote add origin https://github.com/yourusername/go-user-service.git
git branch -M main
git push -u origin main
```

Run them.

Done ✅

---
