# chirpy

API server for creating/viewing chirps (twitter-like)

## Setup

- Go >= 1.25
- Postgresql
- Goose
- SQLC

1. Create `.env`

```sh
cp .env.example .env
```

2. Start DB

```sh
docker compose up [-d]
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

## Docs

### Static files

- `/app/[...]`: HTML files

### Metrics

- `GET /admin/metrics`: view metrics page
- `POST /admin/reset`: reset metrics (dev only)

### API

- `GET /api/healthz`: health status

### API users

Types:

```ts
type UserInput = {
  email: string;
  password: string;
};

type User = {
  id: string;
  created_at: string;
  updated_at: string;
  email: string;
  is_chirpy_red: boolean;
};

type Tokens = {
  token: string; // access_token
  refresh_token: string;
};
```

- `POST /api/users`: create new user

  ```ts
  type RequestBody = UserInput;
  type ResponseBody = User;
  ```

- `PUT /api/users`: update user

  Authentication: `Bearer {access_token}`

  ```ts
  type RequestBody = UserInput;
  type ResponseBody = User;
  ```

- `POST /api/login`: login

  ```ts
  type RequestBody = UserInput;
  type ResponseBody = User & Tokens;
  ```

- `POST /api/refresh`: get new access token

  Authentication: `Bearer {refresh_token}`

  ```ts
  type ResponseBody = {
    token: string;
  };
  ```

- `POST /api/revoke`: revoke refresh token

  Authentication: `Bearer {refresh_token}`

- `POST /api/polka/webhooks`: upgrade user

  Authentication: `ApiKey {api_key}`

  ```ts
  type RequestBody = {
    event: "user.upgraded";
    data: {
      user_id: string;
    };
  };
  ```

### API chirps

Types:

```ts
type Chirp = {
  id: string;
  created_at: string;
  updated_at: string;
  body: string;
  user_id: string;
};
```

- `POST /api/chirps`: create new chirp

  Authentication: `Bearer {access_token}`

  ```ts
  type RequestBody = {
    body: string;
    user_id: string;
  };

  type ResponseBody = Chirp;
  ```

- `GET /api/chirps`: return all chirps (or filtered result)

  ```ts
  type QueryParams = {
    author_id?: string;
    sort?: "asc" | "desc";
  };

  type ResponseBody = Chirp[];
  ```

- `GET /api/chirps/{id}`: get chirp by ID

  ```ts
  type ResponseBody = Chirp;
  ```

- `DELETE /api/chirps/{id}`: delete chirp by ID

  Authentication: `Bearer {access_token}`
