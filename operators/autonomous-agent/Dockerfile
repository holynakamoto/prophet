FROM golang:1.24-alpine AS builder
WORKDIR /workspace
RUN apk add --no-cache git make
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY api/ api/
COPY controllers/ controllers/
COPY cmd/ cmd/
COPY mcp-server/ mcp-server/
COPY llm-inference/ llm-inference/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager cmd/main.go

FROM alpine:latest
WORKDIR /
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /workspace/manager .
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
USER nonroot:nonroot
ENTRYPOINT ["/manager"]

