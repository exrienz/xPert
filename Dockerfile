FROM golang:1.22 AS builder

WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/docgen ./cmd/server

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/docgen /app/docgen

ENV DOCGEN_HOST=0.0.0.0
ENV DOCGEN_PORT=8080
ENV DOCGEN_STORAGE_BACKEND=sqlite
ENV DOCGEN_DATA_PATH=/app/data/docgen.sqlite

EXPOSE 8080

CMD ["/app/docgen"]
