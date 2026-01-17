# LibSystem - Document Management Platform

Enterprise-grade document management system with full-text search, OCR capabilities, and distributed architecture.

## Features

✅ **User Management** - JWT authentication with role-based access control (RBAC)  
✅ **Document Upload** - Support for PDF, DOCX, TXT, images with deduplication  
✅ **Full-Text Search** - Elasticsearch-powered search with faceted navigation  
✅ **OCR Support** - Automatic text extraction from images using Tesseract  
✅ **Collections** - Organize documents into collections  
✅ **Rate Limiting** - Redis-based distributed rate limiting  
✅ **Health Monitoring** - Comprehensive dependency health checks  
✅ **Async Processing** - Kafka-based event-driven indexing  
✅ **Dead Letter Queue** - Reliable failure handling  
✅ **Metrics** - Prometheus-ready monitoring  

---

## Architecture

### Microservices
- **Frontend** (3000) - Next.js 14 Dashboard
- **API Gateway** (8088) - Entry point, rate limiting, auth
- **User Service** (8086) - User management, authentication
- **Document Service** (8081) - Document CRUD, storage
- **Collection Service** (8082) - Collection management
- **Search Service** (8084) - Elasticsearch integration
- **Indexer Service** - Async document processing

### Infrastructure
- **PostgreSQL** - Relational data storage
- **Redis** - Rate limiting, caching
- **Elasticsearch** - Full-text search
- **MinIO** - Object storage
- **Kafka** - Event streaming

See [Architecture Overview](architecture.md) for details.

---

## Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Node.js 18+ (for Frontend)
- Tesseract OCR (optional, for local OCR)

### Automated Setup (Recommended)
We provide a setup script to initialize the environment, check dependencies, and build services.

```bash
# Clone repository
git clone https://github.com/your-org/libsystem.git
cd libsystem

# Run the setup script
./libsystem_setup.sh

# Start services using Make
make up
```

### Manual Installation

If you prefer to run services individually or use Docker Compose directly:

```bash
# Start infrastructure (Postgres, Redis, Kafka, Elastic, MinIO)
docker-compose up -d

# Build and run services
./run-services.sh
```

### Frontend Setup

To start the Next.js frontend:

```bash
cd frontend
npm install
npm run dev
# Access at http://localhost:3000
```

---

## Documentation

- [Frontend Documentation](frontend/README.md)
- [API Documentation](docs/API.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [Architecture Overview](architecture.md)
- [Tech Stack](docs/TECH_STACK.md)

---

## API Usage

### Register User
```bash
curl -X POST http://localhost:8088/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123",
    "username": "johndoe",
    "full_name": "John Doe"
  }'
```

### Search
```bash
curl "http://localhost:8088/api/v1/search?q=machine+learning" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Project Structure

```
libsystem/
├── services/               # Go Microservices
│   ├── api-gateway/
│   ├── document-service/
│   ├── search-service/
│   └── ...
├── frontend/               # Next.js Web App
├── infrastructure/         # Terraform, K8s, Docker
├── shared/                 # Shared Go modules
├── docs/                   # Documentation
├── scripts/                # Helper scripts
└── architecture.md         # Architecture Diagram/Desc
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
