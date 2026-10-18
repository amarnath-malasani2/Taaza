FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o main .
RUN go build -o bootstrap-admin ./cmd/bootstrap-admin

FROM alpine:3.23
RUN apk --no-cache add ca-certificates && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/bootstrap-admin .

RUN mkdir -p uploads && chown -R app:app /app
USER app
CMD ["./main"]
