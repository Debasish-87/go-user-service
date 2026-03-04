# Go User Service (Docker + PostgreSQL)

A simple **Go REST API** that demonstrates:

* PostgreSQL integration using `pgx`
* Docker + Docker Compose setup
* Middleware (logging, auth, JWT)
* Graceful server shutdown
* Goroutines and WaitGroups
* Background email simulation
* Connection pooling
* Load testing using JSON data

---

# Project Structure

```
go-user-service/
│
├── main.go
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── data.json
└── README.md
```

---

# Features

### REST API

* Create user
* Get all users
* Delete user

### Middleware

* Request logging
* Admin authentication
* JWT authentication

### Database

* PostgreSQL
* Automatic table creation
* Connection pool tuning

### Concurrency

* Email sending simulation
* Goroutines
* WaitGroup
* Context timeout

### DevOps

* Docker container for app
* PostgreSQL container
* Docker Compose orchestration

---

# Requirements

Install the following:

* Go 1.24+
* Docker
* Docker Compose
* curl (for testing)

---

# Run the Project

## 1. Clone Repo

```
git clone <repo-url>
cd go-user-service
```

---

## 2. Start with Docker

Build and run services:

```
docker compose up --build
```

This will start:

| Service    | Port |
| ---------- | ---- |
| Go API     | 8080 |
| PostgreSQL | 5432 |

---

# Environment Variables

Configured in `docker-compose.yml`

```
DATABASE_URL=postgres://postgres:pass@db:5432/godb?sslmode=disable
```

Database:

```
User: postgres
Password: pass
Database: godb
```

---

# API Endpoints

## 1. Create User

```
POST /signup
```

Example:

```
curl -X POST http://localhost:8080/signup \
-H "Content-Type: application/json" \
-d '{"name":"John","email":"john@example.com"}'
```

Response:

```
{
 "message":"User Created Successfully & Login Success!",
 "token":"valid-token-123",
 "user":{
   "id":1,
   "name":"John",
   "email":"john@example.com"
 },
 "reports":[
   "Success: Sent to john@example.com"
 ]
}
```

---

## 2. Get All Users

```
GET /signup
```

Example:

```
curl http://localhost:8080/signup
```

Response:

```
[
 {
   "id":1,
   "name":"John",
   "email":"john@example.com"
 }
]
```

---

## 3. Delete User

```
DELETE /signup?id=1
```

Example:

```
curl -X DELETE "http://localhost:8080/signup?id=1"
```

---

## 4. Admin Protected Route

```
GET /
```

Example:

```
curl "http://localhost:8080/?user=admin"
```

Response:

```
Welcome to Home Page!
```

---

## 5. JWT Protected Route

```
GET /dashboard
```

Example:

```
curl http://localhost:8080/dashboard \
-H "Authorization: Bearer valid-token-123"
```

---

# Load Testing Using data.json

`data.json` contains sample user payloads.

Example load test using `curl` + `xargs`:

```
cat data.json | while read line; do
curl -X POST http://localhost:8080/signup \
-H "Content-Type: application/json" \
-d "$line" &
done
wait
```

This sends **60 concurrent signup requests**.

---

# Graceful Shutdown

The server listens for interrupt signals:

```
CTRL + C
```

When triggered:

* Stops accepting new requests
* Waits for active requests to complete
* Closes database pool
* Shuts down cleanly

---

# Database Table

Created automatically:

```
CREATE TABLE users (
 id SERIAL PRIMARY KEY,
 name TEXT NOT NULL,
 email TEXT UNIQUE NOT NULL
);
```

---

# Docker Services

### App

```
golang:1.24-alpine
```

Builds the Go binary and runs it.

### Database

```
postgres:15-alpine
```

---

# Example Logs

```
Connected to PostgreSQL
Users table ready.
Server running at http://localhost:8080/signup
URL: /signup | Method: POST
Signup process started...
User created in Database.
Token generated!
All background tasks finished.
```

---

# Future Improvements

* Proper JWT generation
* Password hashing
* User login endpoint
* Email queue (RabbitMQ / Kafka)
* Pagination for user listing
* Rate limiting
* Structured logging
* Prometheus metrics
* Unit tests

---

# Author

Go backend demo project for learning:

* REST APIs
* Concurrency
* Docker
* PostgreSQL
* Middleware
* Context management
