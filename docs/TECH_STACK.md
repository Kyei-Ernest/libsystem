# LibSystem Technology Stack

Complete reference of all technologies, frameworks, and tools used in LibSystem.

---

## Backend Technologies

### Core Language

#### Go 1.21+
- **Purpose**: Primary backend programming language
- **Why**: High performance, excellent concurrency, strong typing, fast compilation
- **Usage**: All 6 microservices
- **Key Features Used**:
  - Goroutines for concurrent processing
  - Context for request cancellation
  - Channels for communication
  - Native HTTP server

---

## Microservices Architecture

### 1. API Gateway (Port 8088)
**Technology Stack:**
- **Framework**: Gin Web Framework
- **Authentication**: JWT (golang-jwt/jwt v5)
- **Rate Limiting**: Redis + Custom middleware
- **Logging**: Zap (structured logging)
- **Metrics**: Prometheus client
- **CORS**: gin-contrib/cors

**Dependencies:**
```go
github.com/gin-gonic/gin              // Web framework
github.com/golang-jwt/jwt/v5          // JWT authentication
github.com/redis/go-redis/v9          // Redis client
go.uber.org/zap                       // Structured logging
github.com/prometheus/client_golang   // Metrics
```

**Responsibilities:**
- Request routing
- Authentication & authorization
- Rate limiting (100 req/min general, 20 req/min upload)
- Request logging
- Metrics collection
- CORS handling
- Reverse proxy to backend services

---

### 2. User Service (Port 8086)
**Technology Stack:**
- **Framework**: Gin
- **Database**: PostgreSQL + GORM
- **Password Hashing**: bcrypt
- **JWT**: golang-jwt/jwt v5
- **Validation**: go-playground/validator

**Dependencies:**
```go
github.com/gin-gonic/gin
gorm.io/gorm
gorm.io/driver/postgres
github.com/lib/pq                     // PostgreSQL driver
golang.org/x/crypto/bcrypt            // Password hashing
github.com/google/uuid                // UUID generation
```

**Database Schema:**
- Users table (id, email, password_hash, role, is_active)
- Supports roles: patron, librarian, admin
- Email uniqueness constraint
- Automatic timestamps

**Responsibilities:**
- User registration
- Login/authentication
- JWT token generation
- User profile management
- Role-based access control
- User deactivation

---

### 3. Document Service (Port 8081)
**Technology Stack:**
- **Framework**: Gin
- **Database**: PostgreSQL + GORM
- **Object Storage**: MinIO (S3-compatible)
- **Message Queue**: Kafka
- **File Hashing**: SHA-256
- **Deduplication**: Hash-based

**Dependencies:**
```go
github.com/minio/minio-go/v7          // MinIO client
github.com/segmentio/kafka-go         // Kafka producer
github.com/google/uuid
crypto/sha256                          // File hashing
```

**Storage Strategy:**
- Files stored in MinIO S3 buckets
- Metadata in PostgreSQL
- Deduplication via SHA-256 hash
- Organized by collection

**Responsibilities:**
- Document upload (PDF, DOCX, TXT, images)
- File storage in MinIO
- Metadata management
- Document download/streaming
- Permission management
- Deduplication
- Kafka event publishing

---

### 4. Collection Service (Port 8082)
**Technology Stack:**
- **Framework**: Gin
- **Database**: PostgreSQL + GORM

**Database Schema:**
- Collections (id, name, description, owner_id, is_public)
- Collection-Document relationships

**Responsibilities:**
- Collection CRUD operations
- Document-collection associations
- Access control (public/private)
- Collection sharing

---

### 5. Search Service (Port 8084)
**Technology Stack:**
- **Framework**: Gin
- **Search Engine**: Elasticsearch 8.x
- **Client**: elastic/go-elasticsearch v8

**Dependencies:**
```go
github.com/elastic/go-elasticsearch/v8
```

**Index Structure:**
```json
{
  "mappings": {
    "properties": {
      "id": {"type": "keyword"},
      "title": {"type": "text"},
      "content": {"type": "text"},
      "file_type": {"type": "keyword"},
      "uploader_id": {"type": "keyword"},
      "collection_id": {"type": "keyword"},
      "created_at": {"type": "date"}
    }
  }
}
```

**Responsibilities:**
- Full-text search
- Faceted search (by file_type, collection, uploader)
- Result ranking
- Aggregations
- Query suggestions

---

### 6. Indexer Service (Background Worker)
**Technology Stack:**
- **Message Queue**: Kafka consumer
- **Search**: Elasticsearch client
- **Storage**: MinIO client
- **OCR**: Tesseract 5.3.4
- **Document Parsing**: Multiple extractors

**Dependencies:**
```go
github.com/segmentio/kafka-go          // Kafka consumer
github.com/elastic/go-elasticsearch/v8
github.com/minio/minio-go/v7
```

**Text Extraction:**
- PDF: pdftotext, pdf package
- DOCX: XML parsing
- TXT: Direct read
- Images: Tesseract OCR
- HTML: HTML parser

**Responsibilities:**
- Consume document.uploaded events
- Download files from MinIO
- Extract text content
- OCR for images (automatic fallback)
- Index to Elasticsearch
- Retry with exponential backoff
- Dead Letter Queue for failures

---

## Infrastructure & Data Stores

### PostgreSQL 14+
**Purpose**: Primary relational database
**Usage**: User, Document, Collection metadata
**Configuration**:
- Port: 5432
- Separate database per service:
  - `libsystem_users`
  - `libsystem_documents`
  - `libsystem_collections`

**Features Used:**
- JSONB for flexible metadata
- Foreign keys for referential integrity
- Indexes on frequently queried columns
- UUID primary keys
- Timestamps (created_at, updated_at)

---

### Redis 7+
**Purpose**: Distributed rate limiting & caching
**Usage**: 
- Rate limiting state (sorted sets)
- Future: Session caching, query caching

**Port**: 6379

**Data Structures:**
```
Rate Limiting:
  Key: ratelimit:user:{uuid}
  Type: Sorted Set
  Members: Timestamp scores
  TTL: Window size + 1 second
```

**Configuration:**
```conf
maxmemory 2gb
maxmemory-policy allkeys-lru
```

---

### Elasticsearch 8.x
**Purpose**: Full-text search engine
**Usage**: Document content indexing & search

**Port**: 9200

**Indices:**
- `documents` - Main document index
- Shards: 3 primary, 1 replica
- Refresh interval: 1s

**Features Used:**
- Full-text search with relevance scoring
- Faceted navigation (aggregations)
- Highlighting
- Fuzzy matching
- Boolean queries

---

### Apache Kafka 3.x
**Purpose**: Event streaming & async processing
**Usage**: Document upload → indexing workflow

**Port**: 9092, 9093

**Topics:**
```
document.uploaded (main)
  - Partitions: 3
  - Retention: 7 days
  - Messages: Document upload events

document.uploaded-dlq (dead letter queue)
  - Partitions: 1
  - Retention: 30 days
  - Messages: Failed indexing operations
```

**Consumer Groups:**
- `indexer-service-group` - Indexer workers

---

### MinIO
**Purpose**: S3-compatible object storage
**Usage**: Document file storage

**Ports**: 9000 (API), 9001 (Console)

**Buckets:**
- `documents` - All uploaded files
- Organization: `/{collection_id}/{document_id}/{filename}`

**Access:**
- Access Key: minioadmin
- Secret Key: minioadmin123
- SSL: false (enable in production)

---

## Supporting Technologies

### Tesseract OCR 5.3.4
**Purpose**: Optical Character Recognition (text from images)
**Usage**: Extract text from scanned PDFs and images

**Languages Installed:**
- English (tesseract-ocr-eng)

**Trigger**: Automatic fallback when standard extraction yields no text

**CLI Usage:**
```bash
tesseract input.jpg output -l eng
```

---

### Document Processing Libraries

#### PDF Processing
```go
import "github.com/ledongthuc/pdf"
```
- Extract text from PDF files
- Page-by-page processing
- Metadata extraction

#### DOCX Processing
```go
import "archive/zip"
import "encoding/xml"
```
- Unzip DOCX files
- Parse document.xml
- Extract formatted text

#### HTML Processing
```go
import "golang.org/x/net/html"
```
- Parse HTML structure
- Extract text nodes
- Skip script/style tags

---

## Shared Libraries

### Custom Packages

#### `shared/models`
**Purpose**: Common data models
**Contents**:
- User, Document, Collection structs
- Permission levels (Admin, Write, View)
- Roles (Admin, Librarian, Patron)

#### `shared/security`
**Purpose**: Security utilities
**Contents**:
- JWT generation & validation
- Password hashing
- Token parsing

#### `shared/errors`
**Purpose**: Standardized error handling
**Contents**:
- AppError struct
- Error types (NotFound, Unauthorized, Forbidden, etc.)
- HTTP status mapping

#### `shared/response`
**Purpose**: Standard API responses
**Contents**:
- Success/Error response builders
- Pagination helpers
- Consistent JSON structure

#### `shared/extraction`
**Purpose**: Text extraction from files
**Contents**:
- PDFExtractor
- DOCXExtractor
- TXTExtractor
- HTMLExtractor
- OCRExtractor
- Extractor factory

#### `shared/retry`
**Purpose**: Retry logic with backoff
**Contents**:
- Exponential backoff
- Configurable max retries
- Context-aware cancellation

#### `shared/health`
**Purpose**: Health check utilities
**Contents**:
- Database health check
- Redis health check
- Elasticsearch health check
- Comprehensive status reporting

#### `shared/kafka`
**Purpose**: Kafka producer/consumer wrappers
**Contents**:
- Producer with retry
- Consumer with offset management
- Topic creation helpers

---

## Development & Build Tools

### Go Modules
**Purpose**: Dependency management
**Version**: Go 1.21+

**Key Dependencies Version Matrix:**
```
github.com/gin-gonic/gin              v1.10.0
gorm.io/gorm                          v1.25.5
github.com/redis/go-redis/v9          v9.3.0
github.com/elastic/go-elasticsearch/v8 v8.11.0
go.uber.org/zap                       v1.26.0
github.com/golang-jwt/jwt/v5          v5.2.0
github.com/minio/minio-go/v7          v7.0.65
github.com/segmentio/kafka-go         v0.4.45
```

### Docker & Docker Compose
**Purpose**: Containerization & local development
**Version**: Docker 24+, Docker Compose v2

**Images Used:**
```yaml
postgres:14-alpine
redis:7-alpine
elasticsearch:8.11.0
confluentinc/cp-kafka:7.5.0
minio/minio:latest
```

---

## Monitoring & Observability

### Prometheus
**Purpose**: Metrics collection & alerting
**Port**: 9090

**Metrics Exposed:**
- HTTP request count
- Request duration (histogram)
- Active connections
- Error rates
- Go runtime metrics (goroutines, memory, GC)

**Endpoints:**
```
/metrics - Prometheus format metrics
```

### Zap Logger
**Purpose**: Structured logging
**Format**: JSON

**Log Levels:**
- DEBUG - Development info
- INFO - Normal operations
- WARN - Warning conditions
- ERROR - Error conditions

**Fields Logged:**
- timestamp
- level
- message
- request_id
- user_id
- latency
- status_code
- method
- path

---

## Security Technologies

### JWT (JSON Web Tokens)
**Library**: golang-jwt/jwt v5
**Algorithm**: HS256 (HMAC-SHA256)
**Expiration**: 24 hours
**Claims**: user_id, email, role

### Bcrypt
**Library**: golang.org/x/crypto/bcrypt
**Cost**: 10 (2^10 iterations)
**Purpose**: Password hashing

### SHA-256
**Library**: crypto/sha256
**Purpose**: File deduplication hashing

### Rate Limiting
**Strategy**: Token bucket with sliding window
**Storage**: Redis sorted sets
**Granularity**: Per user/IP
**Limits**: Configurable per endpoint

---

## API & Communication

### REST API
**Format**: JSON
**Version**: /api/v1
**Authentication**: Bearer JWT
**CORS**: Enabled with credentials

**Standard Response Format:**
```json
{
  "success": true/false,
  "data": {...},
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message"
  }
}
```

### HTTP Status Codes
```
200 OK - Success
201 Created - Resource created
400 Bad Request - Invalid input
401 Unauthorized - Missing/invalid token
403 Forbidden - Insufficient permissions
404 Not Found - Resource not found
409 Conflict - Duplicate resource
429 Too Many Requests - Rate limit exceeded
500 Internal Server Error - Server error
```

---

## Testing

### E2E Testing
**Framework**: Go testing package
**Location**: `tests/e2e/`
**Coverage**: Complete user workflow

**Scenarios Tested:**
- User registration → login
- Collection creation
- Document upload → indexing
- Search with facets
- Document download
- RBAC enforcement

**Test Duration**: ~6-7 seconds

---

## File Formats Supported

### Documents
| Format | Extension | Extraction Method |
|--------|-----------|-------------------|
| PDF | .pdf | pdftotext library |
| Word | .docx | XML parsing |
| Text | .txt | Direct read |
| HTML | .html, .htm | HTML parser |
| Images | .jpg, .jpeg, .png, .gif | Tesseract OCR |

### Storage
| Type | Size Limit | Storage |
|------|------------|---------|
| Documents | 100MB | MinIO |
| Metadata | N/A | PostgreSQL |
| Search Index | N/A | Elasticsearch |

---

## Performance Characteristics

### Throughput
- Document uploads: 20 req/min (rate limited)
- Search queries: 50 req/min (rate limited)
- General API: 100 req/min (rate limited)
- Indexing: ~1000 docs/hour

### Latency
- API Gateway: <10ms
- Database queries: <50ms
- Search queries: <100ms (p95)
- Document upload: <500ms (excluding file transfer)

---

## Environment Requirements

### Development
```
OS: Ubuntu 20.04+ / macOS 12+
RAM: 8GB minimum
CPU: 4 cores
Disk: 20GB
Go: 1.21+
Docker: 24+
Tesseract: 5.x
```

### Production
```
OS: Ubuntu 22.04 LTS
RAM: 16GB minimum (32GB recommended)
CPU: 8 cores minimum
Disk: 100GB+ SSD
Network: 1Gbps
Load Balancer: Nginx/HAProxy
```

---

## Deployment Technologies

### Container Orchestration
**Options:**
- Docker Compose (development)
- Kubernetes (production)
- Docker Swarm (alternative)

### Load Balancing
**Technology**: Nginx
**Algorithm**: Least connections
**Health Checks**: /health, /ready

### Process Management
**Options:**
- systemd (Linux services)
- supervisord
- Docker restart policies

---

## Summary

**Total Technologies:** 25+
**Programming Languages:** Go (primary)
**Databases:** 3 (PostgreSQL, Redis, Elasticsearch)
**Message Queue:** 1 (Kafka)
**Object Storage:** 1 (MinIO)
**Microservices:** 6
**Ports Used:** 8081, 8082, 8084, 8086, 8088, 5432, 6379, 9092, 9093, 9200, 9000

**Architecture Style:** 
- Microservices
- Event-driven
- CQRS (Command Query Responsibility Segregation)
- REST API
- Distributed systems

**Design Patterns:**
- API Gateway
- Circuit Breaker (planned)
- Retry with Backoff
- Dead Letter Queue
- Repository Pattern
- Factory Pattern (extractors)
- Middleware Pattern
- Publisher-Subscriber
