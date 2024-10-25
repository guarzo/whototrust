FROM golang:1.23.2 as builder
ARG CGO_ENABLED=0
ARG VERSION

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o whototrust -ldflags="-X 'main.version=${VERSION}'" .

FROM gcr.io/distroless/static AS final

COPY --from=builder /app/static /static
COPY --from=builder /app/templates /templates
COPY --from=builder /app/whototrust /whototrust

ENTRYPOINT ["/whototrust"]
