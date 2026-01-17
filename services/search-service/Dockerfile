# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy shared module first
COPY shared/ ./shared/

# Copy service files
COPY services/search-service/go.mod ./services/search-service/
# No go.sum yet, or maybe there is if I initialize it
WORKDIR /build/services/search-service
# We need to create go.mod first or copy it. It doesn't exist yet in my workspace? 
# Wait, list_dir showed go.mod in search-service.

# Copy service source code
COPY services/search-service/ ./

# Initialize module and tidy (since I added main.go imports)
# This will update go.mod and generate go.sum inside the container build
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /build/services/search-service/main .

EXPOSE 8084

CMD ["./main"]
