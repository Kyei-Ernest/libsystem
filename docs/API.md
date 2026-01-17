# LibSystem API Documentation

## Base URL
```
http://localhost:8088/api/v1
```

## Authentication

All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

---

## Endpoints

### Authentication

#### Register User
```http
POST /api/v1/users/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securePassword123",
  "username": "johndoe",
  "full_name": "John Doe"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "johndoe",
    "role": "patron",
    "token": "eyJhbGc..."
  }
}
```

#### Login
```http
POST /api/v1/users/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGc...",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "role": "patron"
    }
  }
}
```

---

### Collections

#### Create Collection
```http
POST /api/v1/collections
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Research Papers",
  "description": "Collection of academic research",
  "is_public": false
}
```

#### List Collections
```http
GET /api/v1/collections?page=1&limit=20
Authorization: Bearer <token>
```

---

### Documents

#### Upload Document
```http
POST /api/v1/documents
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <binary>
collection_id: uuid
title: "Document Title"
description: "Optional description"
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "title": "Document Title",
    "file_type": "application/pdf",
    "size_bytes": 1048576,
    "status": "pending_indexing"
  }
}
```

#### Search Documents
```http
GET /api/v1/search?q=query&collection_id=uuid&file_type=pdf&page=1&limit=10
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "hits": [
      {
        "id": "uuid",
        "title": "Matching Document",
        "file_type": "pdf",
        "score": 0.95
      }
    ],
    "total": 42,
    "facets": {
      "file_type": [
        {"value": "pdf", "count": 25},
        {"value": "docx", "count": 17}
      ]
    }
  }
}
```

#### Download Document
```http
GET /api/v1/documents/:id/download
Authorization: Bearer <token>
```

Returns binary file with appropriate Content-Type header.

---

## Rate Limiting

**Limits:**
- General API: 100 requests/minute
- Authentication: 10 requests/minute
- Document Upload: 20 requests/minute
- Search: 50 requests/minute
- Download: 30 requests/minute

**Response on limit exceeded (429):**
```json
{
  "success": false,
  "error": {
    "code": "TOO_MANY_REQUESTS",
    "message": "Rate limit exceeded. Please try again later."
  }
}
```

---

## Error Responses

### Standard Error Format
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

### Common Status Codes

| Code | Meaning |
|------|---------|
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing or invalid token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource doesn't exist |
| 409 | Conflict - Resource already exists |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error |

---

## Health & Monitoring

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy"
}
```

### Readiness Check
```http
GET /ready
```

**Response:**
```json
{
  "status": "ready",
  "timestamp": "2026-01-10T14:02:00Z",
  "service": "api-gateway"
}
```

### Metrics (Prometheus)
```http
GET /metrics
```

Returns Prometheus-formatted metrics.
