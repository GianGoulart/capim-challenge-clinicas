FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api

FROM alpine:3.20
WORKDIR /
COPY --from=builder /bin/api /bin/api
COPY api/openapi.yaml /api/openapi.yaml
ENV OPENAPI_PATH=/api/openapi.yaml
EXPOSE 8080
ENTRYPOINT ["/bin/api"]
