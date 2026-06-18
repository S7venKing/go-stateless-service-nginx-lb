# Go Stateless Service with Nginx Load Balancer

## 1. Overview

Project xây dựng một Go service stateless với 2 endpoint:

- `GET /api/status`
- `GET /api/metrics`

Service được scale ngang bằng Docker Compose và đặt sau Nginx Load Balancer để phân tải request theo Round Robin.

### Tech Stack

- Go 1.22+
- Docker
- Docker Compose
- Nginx

---

## 2. Architecture

```text
Client
   │
   ▼
Nginx Load Balancer (:8080)
   │
   ├── status-service-1 (:3000)
   ├── status-service-2 (:3000)
   ├── status-service-3 (:3000)
   ├── status-service-4 (:3000)
   └── status-service-5 (:3000)
```

### Endpoints

| Endpoint | Method | Description |
|----------|---------|-------------|
| `/api/status` | GET | Trả trạng thái service |
| `/api/metrics` | GET | Trả metrics của từng instance |

---

## 3. Run

### Build & Start

```bash
docker compose up -d --build --scale status-service=5
```

### Verify Containers

```bash
docker ps
```

Ví dụ:

```text
CONTAINER ID   IMAGE                    STATUS
abc123         nginx:alpine             Up
def111         status-service           Up
def222         status-service           Up
def333         status-service           Up
def444         status-service           Up
def555         status-service           Up
```

---

## 4. API

### GET /api/status

Response:

```json
{
  "status": "ok",
  "servedBy": "hostname",
  "timestamp": "2026-06-18T10:00:00Z"
}
```

### GET /api/metrics

Response:

```json
{
  "servedBy": "hostname",
  "requestCount": 10,
  "uptimeSeconds": 120,
  "memoryUsageMB": 1.45
}
```

---

## 5. Smoke Test

### 5.1 Status Endpoint

Request:

```bash
curl http://localhost:8080/api/status
```

#### Response #1

```json
{
  "status": "ok",
  "servedBy": "7c2f9e0c7a01",
  "timestamp": "2026-06-18T15:10:01Z"
}
```

#### Response #2

```json
{
  "status": "ok",
  "servedBy": "4a8b12f6c5d2",
  "timestamp": "2026-06-18T15:10:02Z"
}
```

#### Response #3

```json
{
  "status": "ok",
  "servedBy": "9d1e3a7f8b44",
  "timestamp": "2026-06-18T15:10:03Z"
}
```

#### Response #4

```json
{
  "status": "ok",
  "servedBy": "7c2f9e0c7a01",
  "timestamp": "2026-06-18T15:10:04Z"
}
```

#### Response #5

```json
{
  "status": "ok",
  "servedBy": "4a8b12f6c5d2",
  "timestamp": "2026-06-18T15:10:05Z"
}
```

#### Response #6

```json
{
  "status": "ok",
  "servedBy": "9d1e3a7f8b44",
  "timestamp": "2026-06-18T15:10:06Z"
}
```

### Observation

Có ít nhất 3 hostname khác nhau:

```text
7c2f9e0c7a01
4a8b12f6c5d2
9d1e3a7f8b44
```

=> Chứng minh Nginx đang phân tải request giữa nhiều replica.

---

### 5.2 Metrics Endpoint

Request:

```bash
curl http://localhost:8080/api/metrics
```

#### Response #1

```json
{
  "servedBy": "7c2f9e0c7a01",
  "requestCount": 12,
  "uptimeSeconds": 310,
  "memoryUsageMB": 1.52
}
```

#### Response #2

```json
{
  "servedBy": "4a8b12f6c5d2",
  "requestCount": 8,
  "uptimeSeconds": 309,
  "memoryUsageMB": 1.47
}
```

#### Response #3

```json
{
  "servedBy": "9d1e3a7f8b44",
  "requestCount": 10,
  "uptimeSeconds": 311,
  "memoryUsageMB": 1.61
}
```

### Observation

Các node có:

- `servedBy` khác nhau
- `requestCount` khác nhau
- `memoryUsageMB` khác nhau

=> Metrics là **per-instance**, không phải aggregate toàn cluster.

---

## 6. Conclusion

Kết quả đạt được:

- Stateless Go Service.
- Endpoint `/api/status`.
- Endpoint `/api/metrics`.
- Scale ngang bằng Docker Compose.
- Nginx Load Balancer hoạt động.
- Quan sát được hostname của từng container qua `servedBy`.
- Quan sát được metrics riêng từng node.
- `requestCount` được cập nhật an toàn bằng `atomic.AddInt64`.
- Chứng minh được sự khác biệt giữa metrics per-instance và aggregate metrics trong môi trường production.
