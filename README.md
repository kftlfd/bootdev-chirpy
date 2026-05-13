# chirpy

API server for creating/viewing chirps (twitter-like)

- Go
- Postgresql
- [Goose](https://github.com/pressly/goose)
- [SQLC](https://docs.sqlc.dev/en/latest/index.html)
- [Swagger](https://github.com/swaggo/swag), [HTTP-Swagger](https://github.com/swaggo/http-swagger)

## Setup

1. Create `.env`

```sh
cp .env.example .env
```

2. Start DB

```sh
docker compose up postgres [-d]
```

3. Migrate DB

```sh
cd sql/schema
export DB="postgres://[...]"
goose postgres $DB up
```

4. Start server

```sh
# dev move
go run ./cmd/api

# build
go build -o chirpy ./cmd/api
./chirpy
```

## Develop

- New DB migration

  ```bash
  # create up+down SQL migration in `sql/schema`
  sql/schema/ $ goose postgres $DB up
  ```

- New DB queries

  ```bash
  # create new SQL query in `sql/queries`
  $ sqlc generate
  ```

- Update swagger docs

  ```bash
  # install swaggo/swag
  $ swag fmt
  $ swag init -g cmd/api/api.go
  ```
