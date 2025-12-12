# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Install kubebuilder tools
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.mod
COPY go.sum go.sum

# Download dependencies
RUN go mod download

# Copy source code
COPY api/ api/
COPY controllers/ controllers/
COPY cmd/ cmd/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /workspace/manager .

# Use non-root user
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
USER nonroot:nonroot

ENTRYPOINT ["/manager"]

