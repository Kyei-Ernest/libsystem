# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy shared module first
COPY shared/ ./shared/

# Copy service files
COPY services/document-service/go.mod services/document-service/go.sum ./services/document-service/
WORKDIR /build/services/document-service
RUN go mod download

# Copy service source code
COPY services/document-service/ ./

# Sync dependencies
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /build/services/document-service/main .

EXPOSE 8081

CMD ["./main"]
