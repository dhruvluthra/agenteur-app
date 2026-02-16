# Agenteur 

Agenteur is a deployment platform for AI agents that utilize coding harnesses like the codex sdk, claude agent sdk, and opencode. 

# Instructions
1. Always create a new branch when implementing a feature. 
2. Use conventional commits whenever commiting gots to you git branch. 

# Directory Overview 

1. This is a monorepo. 
2. The backend api is in the `/backend` directory. The backend is written in go.  
3. The frontend lives in the `/frontend` directory. The frontend is a React/TypeScript project that uses vite. 
4. The `/tech-specs` direcotry holds implementation plans created by breaking down product features into discrete engineering tasks. 

# Backend Database Workflow

1. Use Postgres for local, dev, and production databases.
2. Run local database and migration commands from the `/backend` directory.
3. Use `backend/docker-compose.yml` for local Postgres lifecycle:
   - `make db-up`
   - `make db-create`
   - `make db-down`
   - `make db-reset`
   - `make db-logs`
4. Use `backend/migrations` for all schema changes.
5. Use backend migration commands:
   - `make migrate-up`
   - `make migrate-down`
   - `make migrate-status`
   - `make migrate-create name=<migration_name>`
6. Keep local defaults in `backend/.env.example` and copy to `backend/.env.local`.
7. Never commit real dev/production credentials.
8. If `POSTGRES_DB` changes after initial container setup, run `make db-reset` to reinitialize the volume.
