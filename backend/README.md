# Backend

## Local Database Workflow

Run all commands from the `backend` directory.

```bash
cd backend
cp .env.example .env.local
```

Start and inspect the local Postgres database:

```bash
make db-up
make db-create
make db-logs
```

Apply and inspect migrations:

```bash
make migrate-up
make migrate-status
```

Run the API:

```bash
go run ./cmd/api
```

Stop/reset local DB:

```bash
make db-down
make db-reset
```

## Environment Notes

- Local uses `.env.local` loaded by the backend app when `ENV=local`.
- Dev and production must provide `DATABASE_URL` via deployment environment/secret manager.
- Configure browser origins with `CORS_ALLOWED_ORIGINS` as a comma-separated list.
- Keep credentials out of git.
- If you change `POSTGRES_DB` after data already exists, run `make db-reset` once to reinitialize.
