# === build

FROM golang:alpine3.23 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd cmd 
COPY docs docs
COPY internal internal

RUN go build ./cmd/chirpy

# === final

FROM alpine:3.22.4

WORKDIR /app

COPY --from=build /app/chirpy /bin

EXPOSE 8080

CMD ["chirpy"]
