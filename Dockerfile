FROM golang:1.22 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/xpert ./cmd/server

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/xpert /app/xpert

ENV XPERT_HOST=0.0.0.0
ENV XPERT_PORT=8080
ENV XPERT_STORAGE_BACKEND=sqlite
ENV XPERT_DATA_PATH=/app/data/xpert.sqlite

EXPOSE 8080

CMD ["/app/xpert"]
