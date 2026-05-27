FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/service \
    ./cmd/service

FROM gcr.io/distroless/static-debian12:nonroot

USER nonroot:nonroot

COPY --from=builder /out/service /service

EXPOSE 8080

ENTRYPOINT ["/service"]
