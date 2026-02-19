# Tech Spec: User & Organization Management — Backend

**Type:** Engineering-initiated
**Author:** Claude
**Last updated:** 2026-02-18
**Status:** Draft
**Related:** [Frontend spec](./user-org-management-frontend.md)

---

## 1. Overview

### Summary

Implement the backend for user authentication (JWT-based), multi-organization support with role-based access control, and email-based invitation flow for the Agenteur platform. This is prerequisite infrastructure — no product features (agent deployment, skill management) can be built until users can authenticate and belong to organizations.

### Motivation & Requirements

**Problem:** Agenteur has no authentication, user management, or organization system. The backend has a single unused DB connection and an empty `/api` route group. No product features can be shipped without this foundation.

**Why now:** This is a blocking dependency for all product work. Agent deployment, skill management, and billing all require knowing who the user is and which organization they belong to.

**Out of scope:**
- OAuth/SSO providers (Google, GitHub login) — future enhancement
- Password reset flow — future enhancement
- Email verification on signup — future enhancement
- Billing/subscription management
- Agent/skill CRUD (depends on this spec)
- Audit logging
- Frontend implementation (see [frontend spec](./user-org-management-frontend.md))

| Req ID | Requirement | Success Criteria | Priority |
|--------|------------|-----------------|----------|
| ER-001 | Users can sign up with email and password | `POST /api/auth/signup` returns 201 with user data and sets httpOnly JWT cookies. Password stored as bcrypt hash. | must |
| ER-002 | Users can log in and receive JWT tokens | `POST /api/auth/login` validates credentials, returns 200 with user data and sets access + refresh token cookies. | must |
| ER-003 | JWT access tokens auto-refresh transparently | `POST /api/auth/refresh` accepts a refresh token cookie, rotates both tokens. | must |
| ER-004 | Users can log out and tokens are revoked | `POST /api/auth/logout` deletes all refresh tokens for the user from DB and clears cookies. Works even with expired access tokens. | must |
| ER-005 | Users can create organizations | `POST /api/organizations` creates an org and assigns the creator as admin. Org slug is auto-generated from name. | must |
| ER-006 | Users can belong to multiple organizations | `org_memberships` table supports many-to-many. `GET /api/organizations` returns only the user's orgs. | must |
| ER-007 | Organization admins can invite users by email | `POST /api/organizations/{orgID}/invitations` generates a unique invite token, stores its hash, and logs the invite URL to console (email stub). | must |
| ER-008 | Invited users can accept invitations | `POST /api/invitations/{token}/accept` creates an org membership if the authenticated user's email matches the invitation. | must |
| ER-009 | Role-based access control with three tiers | Superadmin (user flag, cross-org), Admin (org role, manage org), User (org role, use platform). Middleware enforces each. | must |
| ER-010 | Superadmins can access all organizations and manage users | `GET /api/admin/users` lists all users. `PUT /api/admin/users/{userID}/superadmin` toggles the flag. Superadmins bypass org membership checks. | must |

**Requirement-to-Task Mapping:**

| Req ID | Implementing Tasks |
|--------|-------------------|
| ER-001 | T-001, T-002, T-003, T-004, T-005, T-006, T-007 |
| ER-002 | T-005, T-006, T-007 |
| ER-003 | T-004, T-005, T-006, T-007 |
| ER-004 | T-005, T-006, T-007 |
| ER-005 | T-008, T-009, T-010, T-011 |
| ER-006 | T-008, T-009, T-010, T-011 |
| ER-007 | T-012, T-013, T-014 |
| ER-008 | T-012, T-013, T-014 |
| ER-009 | T-007, T-010, T-015 |
| ER-010 | T-015 |

---

## 2. Architecture & Design Decisions

### System Context

```
                    ┌─────────────┐
                    │  Frontend   │
                    │ React/Vite  │
                    │ :5173       │
                    └──────┬──────┘
                           │ fetch (credentials: include)
                           │ httpOnly cookies (access_token, refresh_token)
                    ┌──────▼──────┐
                    │  Backend    │
                    │  Go/chi     │
                    │  :8080      │
                    └──────┬──────┘
                           │ pgxpool
                    ┌──────▼──────┐
                    │  PostgreSQL │
                    │  :5432      │
                    └─────────────┘
```

The backend is the single API server. Auth state lives in the database (refresh tokens) and in short-lived JWTs (access tokens). The frontend communicates via REST with httpOnly cookies — no tokens in localStorage.

### Key Decisions

| Decision | Alternatives Considered | Rationale |
|----------|------------------------|-----------|
| JWT in httpOnly cookies (not localStorage) | localStorage, Authorization header | XSS-resistant. Cookies sent automatically by browser. No JS token handling. |
| Superadmin as a boolean on users table | Separate superadmin table, org membership role | Superadmins transcend orgs. A user-level flag is simplest and avoids a "system" org hack. |
| Refresh token hash in DB (not the raw token) | Store raw token | If DB is compromised, attacker cannot forge refresh tokens. Same pattern as password hashing. |
| Org membership roles and superadmin NOT in JWT claims for authz | Include roles/superadmin in JWT | Org roles and superadmin status can change (user promoted/demoted). Embedding in JWT means stale data until token refresh. Loading per-request from DB is acceptable at this scale and ensures immediate revocation. `is_superadmin` remains in JWT claims for convenience (e.g., frontend display) but authorization decisions always check the DB. |
| `pgxpool.Pool` instead of `pgx.Conn` | Keep single connection | Single connection cannot handle concurrent requests. Pool is required for production. |
| Console email stub (not real email) | Integrate SendGrid/SES now | Avoids external dependency complexity. Invite URLs logged to stdout for development. Interface allows easy swap later. |
| Repository pattern with `DBTX` interface | Direct pool queries in handlers | Enables unit testing handlers without a live DB. Supports transactions across repositories. |
| Content-Type enforcement middleware for CSRF | SameSite cookies alone, double-submit CSRF tokens | Cookie-based auth requires CSRF protection. HTML forms cannot send `Content-Type: application/json`, so enforcing JSON on state-changing methods (`POST/PUT/PATCH/DELETE`) ensures only JS callers (subject to CORS) can reach the API. Simpler than token-based CSRF — no client coordination needed. |
| Shared `database` package with `DBTX` and `WithTx` | Define `DBTX` in `auth/types`, pass `pgx.Tx` manually | Avoids cross-domain dependency (`administration` → `auth`). `WithTx` helper centralizes begin/commit/rollback so services declare transaction boundaries explicitly without boilerplate. |

### New Patterns or Deviations

This is the first feature spec, so all patterns established here become the baseline:

1. **Domain-driven package organization**: Feature logic lives under domain packages, each with sub-packages for `types/`, `services/`, and `handlers/`. Authentication lives in `backend/internal/auth/` and organization management in `backend/internal/administration/`. Cross-cutting infrastructure (CORS, request ID, request logging) remains in `backend/internal/middleware/`. Shared HTTP helpers live in `backend/internal/httputil/`. This pattern should be followed for future domains (e.g., `backend/internal/agents/`, `backend/internal/deployments/`).
2. **Dependency injection via constructors**: Each layer receives dependencies as interfaces through `New*()` functions.
3. **Standard JSON response envelope**: `{"data": {...}}` for success, `{"error": {"code": "...", "message": "..."}}` for errors. JSON fields use camelCase (e.g., `firstName`, `createdAt`).
4. **`DBTX` interface and `WithTx` helper in shared `database` package**: `DBTX` is the shared interface implemented by both `*pgxpool.Pool` and `pgx.Tx`, allowing repository methods to work in both transactional and non-transactional contexts. `WithTx` wraps begin/commit/rollback so services declare transaction boundaries explicitly. Both live in `backend/internal/database/` to avoid cross-domain imports.
5. **Context-based auth propagation**: JWT claims stored in `context.Context` via middleware, retrieved by downstream handlers.
6. **Content-Type enforcement on `/api` routes**: A global middleware rejects `POST/PUT/PATCH/DELETE` requests that don't carry `Content-Type: application/json`. This is the CSRF mitigation strategy — HTML forms cannot send JSON, so cross-origin form submissions are blocked before reaching any handler.

### Dependencies

| Dependency | Type | Version / Constraint | Notes |
|-----------|------|---------------------|-------|
| `golang.org/x/crypto` | Go module | latest | bcrypt password hashing |
| `github.com/golang-jwt/jwt/v5` | Go module | v5.x | JWT creation and validation |
| `github.com/jackc/pgx/v5/pgxpool` | Go module (already in pgx) | v5.8.0 | Connection pool (same module, different import) |

---

## 3. Data Model

### New Tables

#### users

| Column | Type | Constraints | Notes |
|--------|------|------------|-------|
| id | UUID | PK, DEFAULT gen_random_uuid() | |
| email | TEXT | NOT NULL | Case-insensitive uniqueness via index |
| password_hash | TEXT | NOT NULL | bcrypt hash |
| first_name | TEXT | NOT NULL, DEFAULT '' | |
| last_name | TEXT | NOT NULL, DEFAULT '' | |
| is_superadmin | BOOLEAN | NOT NULL, DEFAULT FALSE | Cross-org access flag |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | |

**Indexes:**
- `idx_users_email` UNIQUE on (LOWER(email)) — case-insensitive email uniqueness

#### organizations

| Column | Type | Constraints | Notes |
|--------|------|------------|-------|
| id | UUID | PK, DEFAULT gen_random_uuid() | |
| name | TEXT | NOT NULL | Display name |
| slug | TEXT | NOT NULL | URL-safe identifier |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | |

**Indexes:**
- `idx_organizations_slug` UNIQUE on (LOWER(slug)) — URL-safe unique identifier

#### org_memberships

| Column | Type | Constraints | Notes |
|--------|------|------------|-------|
| id | UUID | PK, DEFAULT gen_random_uuid() | |
| user_id | UUID | NOT NULL, FK → users(id) ON DELETE CASCADE | |
| organization_id | UUID | NOT NULL, FK → organizations(id) ON DELETE CASCADE | |
| role | org_role (ENUM) | NOT NULL, DEFAULT 'user' | 'admin' or 'user' |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | |

**Indexes:**
- `idx_org_memberships_user_org` UNIQUE on (user_id, organization_id) — prevents duplicate memberships
- `idx_org_memberships_org` on (organization_id) — fast member listing by org

**Relationships:**
- org_memberships.user_id → users.id (FK, ON DELETE CASCADE)
- org_memberships.organization_id → organizations.id (FK, ON DELETE CASCADE)

#### refresh_tokens

| Column | Type | Constraints | Notes |
|--------|------|------------|-------|
| id | UUID | PK, DEFAULT gen_random_uuid() | |
| user_id | UUID | NOT NULL, FK → users(id) ON DELETE CASCADE | |
| token_hash | TEXT | NOT NULL | SHA-256 hash of raw token |
| expires_at | TIMESTAMPTZ | NOT NULL | |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | |

**Indexes:**
- `idx_refresh_tokens_hash` UNIQUE on (token_hash) — fast lookup by hashed token
- `idx_refresh_tokens_user` on (user_id) — delete all tokens for a user

**Relationships:**
- refresh_tokens.user_id → users.id (FK, ON DELETE CASCADE)

#### invitations

| Column | Type | Constraints | Notes |
|--------|------|------------|-------|
| id | UUID | PK, DEFAULT gen_random_uuid() | |
| organization_id | UUID | NOT NULL, FK → organizations(id) ON DELETE CASCADE | |
| invited_by | UUID | NOT NULL, FK → users(id) ON DELETE CASCADE | |
| email | TEXT | NOT NULL | Invitee email |
| token_hash | TEXT | NOT NULL | SHA-256 hash of raw invite token |
| role | org_role (ENUM) | NOT NULL, DEFAULT 'user' | Role granted on acceptance |
| status | invitation_status (ENUM) | NOT NULL, DEFAULT 'pending' | pending/accepted/expired/revoked |
| expires_at | TIMESTAMPTZ | NOT NULL | |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ | NOT NULL, DEFAULT NOW() | |

**Indexes:**
- `idx_invitations_token_hash` UNIQUE on (token_hash) — fast lookup by hashed token
- `idx_invitations_org` on (organization_id) — list invitations by org
- `idx_invitations_email` on (LOWER(email)) — find invitations for an email
- `idx_invitations_pending_email_org` UNIQUE on (organization_id, LOWER(email)) WHERE status = 'pending' — prevents duplicate pending invites at the DB level (app logic catches the unique violation and returns 409)

**Relationships:**
- invitations.organization_id → organizations.id (FK, ON DELETE CASCADE)
- invitations.invited_by → users.id (FK, ON DELETE CASCADE)

### Goose Migration

#### `backend/migrations/00002_create_users_and_orgs.sql`

```sql
-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    first_name    TEXT NOT NULL DEFAULT '',
    last_name     TEXT NOT NULL DEFAULT '',
    is_superadmin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_users_email ON users (LOWER(email));

CREATE TABLE organizations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_organizations_slug ON organizations (LOWER(slug));

CREATE TYPE org_role AS ENUM ('admin', 'user');

CREATE TABLE org_memberships (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role            org_role NOT NULL DEFAULT 'user',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_org_memberships_user_org ON org_memberships (user_id, organization_id);
CREATE INDEX idx_org_memberships_org ON org_memberships (organization_id);

CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_refresh_tokens_hash ON refresh_tokens (token_hash);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);

CREATE TYPE invitation_status AS ENUM ('pending', 'accepted', 'expired', 'revoked');

CREATE TABLE invitations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    invited_by      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email           TEXT NOT NULL,
    token_hash      TEXT NOT NULL,
    role            org_role NOT NULL DEFAULT 'user',
    status          invitation_status NOT NULL DEFAULT 'pending',
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_invitations_token_hash ON invitations (token_hash);
CREATE INDEX idx_invitations_org ON invitations (organization_id);
CREATE INDEX idx_invitations_email ON invitations (LOWER(email));
CREATE UNIQUE INDEX idx_invitations_pending_email_org ON invitations (organization_id, LOWER(email)) WHERE status = 'pending';

INSERT INTO schema_migrations_audit (migration_name) VALUES ('00002_create_users_and_orgs');

-- +goose Down
DROP TABLE IF EXISTS invitations;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS org_memberships;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS invitation_status;
DROP TYPE IF EXISTS org_role;
DELETE FROM schema_migrations_audit WHERE migration_name = '00002_create_users_and_orgs';
```

**Migration compatibility:**

| # | Migration File | Backward Compatible | Notes |
|---|---------------|-------------------|-------|
| 1 | `backend/migrations/00002_create_users_and_orgs.sql` | Yes | All new tables, no existing schema changes |

**Verification:** `make migrate-up` applies cleanly. `make migrate-down && make migrate-up` round-trips without errors. `make migrate-status` shows migration applied.

---

## 4. API Design

### POST /api/auth/signup

**Description:** Register a new user account.

**Handler:** `handlers.AuthHandler.Signup` in `backend/internal/auth/handlers/auth_handler.go`

**Auth:** None (public)

**Request struct:**
```go
type SignupRequest struct {
    Email     string `json:"email"`
    Password  string `json:"password"`
    FirstName string `json:"firstName"`
    LastName  string `json:"lastName"`
}
```

**Response struct:**
```go
type UserResponse struct {
    ID           string `json:"id"`
    Email        string `json:"email"`
    FirstName    string `json:"firstName"`
    LastName     string `json:"lastName"`
    IsSuperadmin bool   `json:"isSuperadmin"`
    CreatedAt    string `json:"createdAt"`
}
```

**Cookies set:** `access_token` (path `/`, 15min), `refresh_token` (path `/api/auth`, 7d). Both httpOnly, Secure in prod, SameSite=Lax.

**Error Responses:**

| Status | Condition | Body |
|--------|-----------|------|
| 400 | Invalid JSON | `{"error": {"code": "INVALID_JSON", "message": "..."}}` |
| 409 | Email already registered | `{"error": {"code": "CONFLICT", "message": "Email already registered"}}` |
| 422 | Validation failure | `{"error": {"code": "VALIDATION_ERROR", "message": "...", "details": {"email": "...", "password": "..."}}}` |

**Validation rules:** email required + valid format, password required + min 8 chars, firstName required.

---

### POST /api/auth/login

**Description:** Authenticate an existing user.

**Handler:** `handlers.AuthHandler.Login` in `backend/internal/auth/handlers/auth_handler.go`

**Auth:** None (public)

**Request struct:**
```go
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}
```

**Response:** Same `UserResponse` as signup. Same cookies set.

**Error Responses:**

| Status | Condition | Body |
|--------|-----------|------|
| 401 | Invalid email or password | `{"error": {"code": "UNAUTHORIZED", "message": "Invalid email or password"}}` |

---

### POST /api/auth/refresh

**Description:** Rotate access and refresh tokens using a valid refresh token cookie.

**Handler:** `handlers.AuthHandler.Refresh` in `backend/internal/auth/handlers/auth_handler.go`

**Auth:** Requires valid `refresh_token` cookie (NOT the access token)

**Request:** No body. Reads `refresh_token` from cookie.

**Response:** `200` with `UserResponse`. Sets new access_token and refresh_token cookies. Old refresh token deleted from DB.

**Error Responses:**

| Status | Condition | Body |
|--------|-----------|------|
| 401 | Missing/invalid/expired refresh token | `{"error": {"code": "UNAUTHORIZED", "message": "Invalid or expired refresh token"}}` |

---

### POST /api/auth/logout

**Description:** Log out the current user and revoke all their refresh tokens.

**Handler:** `handlers.AuthHandler.Logout` in `backend/internal/auth/handlers/auth_handler.go`

**Auth:** None (public). The handler reads the `access_token` cookie directly and parses it **without validating expiry** (using `jwt.WithoutClaimsValidation()`) to extract the user ID. This allows logout to work even after the access token has expired. If the token is missing or unparseable (wrong signature), return 401.

**Request:** No body.

**Response:** `200 {"data": {"message": "logged out"}}`

**Side effects:** Deletes **all** refresh token rows for the user from DB (via `RefreshTokenRepository.DeleteAllByUser`). Clears both cookies (MaxAge=0).

---

### GET /api/users/me

**Description:** Get the current authenticated user's profile.

**Handler:** `handlers.UserHandler.GetMe` in `backend/internal/auth/handlers/user_handler.go`

**Auth:** Requires valid access token

**Response:** `200` with `UserResponse`

---

### PUT /api/users/me

**Description:** Update the current user's profile.

**Handler:** `handlers.UserHandler.UpdateMe` in `backend/internal/auth/handlers/user_handler.go`

**Auth:** Requires valid access token

**Request struct:**
```go
type UpdateUserRequest struct {
    FirstName string `json:"firstName"`
    LastName  string `json:"lastName"`
}
```

**Response:** `200` with updated `UserResponse`

---

### POST /api/organizations

**Description:** Create a new organization. The creating user becomes its admin.

**Handler:** `handlers.OrgHandler.Create` in `backend/internal/administration/handlers/org_handler.go`

**Auth:** Requires valid access token

**Request struct:**
```go
type CreateOrgRequest struct {
    Name string `json:"name"`
}
```

**Response struct:**
```go
type OrgResponse struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Slug      string `json:"slug"`
    CreatedAt string `json:"createdAt"`
}
```

**Logic:** Auto-generate slug from name (lowercase, hyphens for non-alphanumeric). On slug collision, append random suffix. Create org + admin membership in a DB transaction.

**Error Responses:**

| Status | Condition | Body |
|--------|-----------|------|
| 422 | Name is empty | `{"error": {"code": "VALIDATION_ERROR", "message": "...", "details": {"name": "Name is required"}}}` |

---

### GET /api/organizations

**Description:** List organizations the current user belongs to (or all orgs for superadmins).

**Handler:** `handlers.OrgHandler.List` in `backend/internal/administration/handlers/org_handler.go`

**Auth:** Requires valid access token

**Response:** `200 {"data": {"organizations": [OrgResponse, ...]}}`

---

### GET /api/organizations/{orgID}

**Description:** Get a single organization's details.

**Handler:** `handlers.OrgHandler.Get` in `backend/internal/administration/handlers/org_handler.go`

**Auth:** Requires valid access token + org membership (or superadmin). Enforced by `RequireOrgMember` middleware.

**Response:** `200` with `OrgResponse`

---

### PUT /api/organizations/{orgID}

**Description:** Update an organization's name.

**Handler:** `handlers.OrgHandler.Update` in `backend/internal/administration/handlers/org_handler.go`

**Auth:** Requires org admin role (or superadmin). Enforced by `RequireOrgAdmin` middleware.

**Request:** Same `CreateOrgRequest` (name field).

**Response:** `200` with updated `OrgResponse`

---

### GET /api/organizations/{orgID}/members

**Description:** List all members of an organization with their roles.

**Handler:** `handlers.OrgHandler.ListMembers` in `backend/internal/administration/handlers/org_handler.go`

**Auth:** Requires org membership (any role) or superadmin.

**Response struct:**
```go
type MemberResponse struct {
    UserID    string `json:"userId"`
    Email     string `json:"email"`
    FirstName string `json:"firstName"`
    LastName  string `json:"lastName"`
    Role      string `json:"role"`
    JoinedAt  string `json:"joinedAt"`
}
```

**Response:** `200 {"data": {"members": [MemberResponse, ...]}}`

---

### DELETE /api/organizations/{orgID}/members/{userID}

**Description:** Remove a member from an organization.

**Handler:** `handlers.OrgHandler.RemoveMember` in `backend/internal/administration/handlers/org_handler.go`

**Auth:** Requires org admin role (or superadmin).

**Error Responses:**

| Status | Condition | Body |
|--------|-----------|------|
| 400 | Attempting to remove the last admin | `{"error": {"code": "VALIDATION_ERROR", "message": "Cannot remove the last admin"}}` |
| 404 | User is not a member | `{"error": {"code": "NOT_FOUND", "message": "Member not found"}}` |

---

### POST /api/organizations/{orgID}/invitations

**Description:** Invite a user to the organization by email.

**Handler:** `handlers.InvitationHandler.Create` in `backend/internal/administration/handlers/invitation_handler.go`

**Auth:** Requires org admin role (or superadmin).

**Request struct:**
```go
type CreateInvitationRequest struct {
    Email string `json:"email"`
    Role  string `json:"role"` // "admin" or "user"
}
```

**Response struct:**
```go
type InvitationResponse struct {
    ID        string `json:"id"`
    Email     string `json:"email"`
    Role      string `json:"role"`
    Status    string `json:"status"`
    ExpiresAt string `json:"expiresAt"`
    CreatedAt string `json:"createdAt"`
}
```

**Logic:** Generate random token via `crypto/rand`, store SHA-256 hash in DB, log invite URL (`{InviteBaseURL}/{rawToken}`) to console via `ConsoleEmailService`. 72-hour expiry.

**Error Responses:**

| Status | Condition | Body |
|--------|-----------|------|
| 409 | Pending invite already exists for this email+org | `{"error": {"code": "CONFLICT", "message": "Invitation already pending for this email"}}` |
| 409 | User is already an org member | `{"error": {"code": "CONFLICT", "message": "User is already a member"}}` |

---

### GET /api/invitations/{token}

**Description:** View invitation details (public — the token itself authenticates the request).

**Handler:** `handlers.InvitationHandler.GetByToken` in `backend/internal/administration/handlers/invitation_handler.go`

**Auth:** None (public). The raw token in the URL is the auth.

**Response struct:**
```go
type InvitationDetailResponse struct {
    OrganizationName string `json:"organizationName"`
    Email            string `json:"email"`
    Role             string `json:"role"`
    InvitedByName    string `json:"invitedByName"`
    ExpiresAt        string `json:"expiresAt"`
}
```

**Error Responses:**

| Status | Condition | Body |
|--------|-----------|------|
| 404 | Token not found, expired, or already accepted | `{"error": {"code": "NOT_FOUND", "message": "Invitation not found or expired"}}` |

---

### POST /api/invitations/{token}/accept

**Description:** Accept an invitation and join the organization.

**Handler:** `handlers.InvitationHandler.Accept` in `backend/internal/administration/handlers/invitation_handler.go`

**Auth:** Requires valid access token. The authenticated user's email must match the invitation email (case-insensitive).

**Logic:** Hash token, look up invitation, verify status is `pending` and not expired, verify email match using `strings.EqualFold()`, create org_membership, update invitation status to `accepted`.

**Response:** `200 {"data": {"membership": MemberResponse}}`

**Error Responses:**

| Status | Condition | Body |
|--------|-----------|------|
| 403 | Authenticated user's email doesn't match invitation | `{"error": {"code": "FORBIDDEN", "message": "Email does not match invitation"}}` |
| 404 | Invalid/expired/already-accepted token | `{"error": {"code": "NOT_FOUND", "message": "Invitation not found or expired"}}` |
| 409 | User already a member | `{"error": {"code": "CONFLICT", "message": "Already a member of this organization"}}` |

---

### GET /api/admin/users

**Description:** List all users in the system (paginated).

**Handler:** `handlers.AdminHandler.ListUsers` in `backend/internal/administration/handlers/admin_handler.go`

**Auth:** Requires superadmin.

**Query params:** `?page=1&perPage=20&search=<email or name>`

**Response:** `200 {"data": {"users": [UserResponse, ...], "total": 42, "page": 1, "perPage": 20}}`

---

### PUT /api/admin/users/{userID}/superadmin

**Description:** Toggle superadmin status on a user.

**Handler:** `handlers.AdminHandler.ToggleSuperadmin` in `backend/internal/administration/handlers/admin_handler.go`

**Auth:** Requires superadmin.

**Request struct:**
```go
type ToggleSuperadminRequest struct {
    IsSuperadmin bool `json:"isSuperadmin"`
}
```

**Response:** `200` with updated `UserResponse`

---

## 5. Component & Module Design

### Package Structure

```
backend/internal/
  database/                       # Shared database primitives (no domain logic)
    database.go                   # DBTX interface, WithTx helper
  auth/                           # Domain package for authentication & user management
    types/                        # Auth domain types, repository interfaces
      user.go                     # User, CreateUserParams, UpdateUserParams
      token.go                    # RefreshToken, TokenClaims
      repositories.go             # UserRepository, RefreshTokenRepository (imports database.DBTX)
    services/                     # Auth business logic, data access, crypto
      auth_service.go             # AuthService — Signup, Login, Logout, Refresh
      user_service.go             # UserService — GetByID, Update
      user_repo.go                # pgxUserRepository (implements types.UserRepository)
      token_repo.go               # pgxRefreshTokenRepository
      password.go                 # HashPassword, CheckPassword (bcrypt)
      jwt.go                      # GenerateAccessToken, ValidateAccessToken
      token.go                    # GenerateRandomToken (crypto/rand + SHA-256)
    handlers/                     # Auth HTTP handlers + auth middleware
      auth_handler.go             # AuthHandler — Signup, Login, Logout, Refresh
      user_handler.go             # UserHandler — GetMe, UpdateMe
      auth_middleware.go           # Authenticate middleware, GetUserClaims(ctx)
  administration/                 # Domain package for org management & platform admin
    types/                        # Org domain types, repository interfaces
      organization.go             # Organization, CreateOrgParams, UpdateOrgParams
      membership.go               # OrgMembership, MemberWithUser, OrgWithRole
      invitation.go               # Invitation, InvitationWithOrg, CreateInvitationParams
      repositories.go             # OrganizationRepository, MembershipRepository, InvitationRepository, EmailService
    services/                     # Org business logic, data access
      org_service.go              # OrgService — Create, List, Get, Update, ListMembers, RemoveMember
      invitation_service.go       # InvitationService — Create, GetByToken, Accept
      email_service.go            # ConsoleEmailService (implements types.EmailService)
      org_repo.go                 # pgxOrganizationRepository
      membership_repo.go          # pgxMembershipRepository
      invitation_repo.go          # pgxInvitationRepository
    handlers/                     # Org HTTP handlers, role middleware, admin
      org_handler.go              # OrgHandler — Create, List, Get, Update, ListMembers, RemoveMember
      invitation_handler.go       # InvitationHandler — Create, GetByToken, Accept
      admin_handler.go            # AdminHandler — ListUsers, ToggleSuperadmin
      role_middleware.go           # RequireOrgMember, RequireOrgAdmin, RequireSuperadmin
  httputil/                       # Shared HTTP response helpers
    response.go                   # JSON(), Error(), ValidationError()
  app/                            # Application setup & routing (existing)
  config/                         # Configuration (existing)
  middleware/                     # Cross-cutting middleware: CORS, RequestID, RequestLogger (existing), ContentType (new)
```

### New Files

**Shared:**

| File Path | Package | Exported Interface |
|-----------|---------|-------------------|
| `backend/internal/database/database.go` | database | `DBTX` interface, `WithTx()` helper |
| `backend/internal/httputil/response.go` | httputil | `JSON()`, `Error()`, `ValidationError()` |
| `backend/internal/middleware/content_type.go` | middleware | `RequireJSONContentType()` middleware |

**Auth domain (`backend/internal/auth/`):**

| File Path | Package | Exported Interface |
|-----------|---------|-------------------|
| `backend/internal/auth/types/user.go` | types | `User`, `CreateUserParams`, `UpdateUserParams` |
| `backend/internal/auth/types/token.go` | types | `RefreshToken`, `TokenClaims` |
| `backend/internal/auth/types/repositories.go` | types | `UserRepository`, `RefreshTokenRepository` (imports `database.DBTX`) |
| `backend/internal/auth/services/password.go` | services | `HashPassword`, `CheckPassword` |
| `backend/internal/auth/services/jwt.go` | services | `GenerateAccessToken`, `ValidateAccessToken` |
| `backend/internal/auth/services/token.go` | services | `GenerateRandomToken` |
| `backend/internal/auth/services/user_repo.go` | services | `pgxUserRepository` |
| `backend/internal/auth/services/token_repo.go` | services | `pgxRefreshTokenRepository` |
| `backend/internal/auth/services/auth_service.go` | services | `AuthService` — `Signup`, `Login`, `Logout`, `Refresh` |
| `backend/internal/auth/services/user_service.go` | services | `UserService` — `GetByID`, `Update` |
| `backend/internal/auth/handlers/auth_handler.go` | handlers | `AuthHandler` — `Signup`, `Login`, `Logout`, `Refresh` |
| `backend/internal/auth/handlers/user_handler.go` | handlers | `UserHandler` — `GetMe`, `UpdateMe` |
| `backend/internal/auth/handlers/auth_middleware.go` | handlers | `AuthMiddleware.Authenticate`, `GetUserClaims(ctx)` |

**Administration domain (`backend/internal/administration/`):**

| File Path | Package | Exported Interface |
|-----------|---------|-------------------|
| `backend/internal/administration/types/organization.go` | types | `Organization`, `CreateOrgParams`, `UpdateOrgParams` |
| `backend/internal/administration/types/membership.go` | types | `OrgMembership`, `MemberWithUser`, `OrgWithRole` |
| `backend/internal/administration/types/invitation.go` | types | `Invitation`, `InvitationWithOrg`, `CreateInvitationParams` |
| `backend/internal/administration/types/repositories.go` | types | `OrganizationRepository`, `MembershipRepository`, `InvitationRepository`, `EmailService` |
| `backend/internal/administration/services/org_repo.go` | services | `pgxOrganizationRepository` |
| `backend/internal/administration/services/membership_repo.go` | services | `pgxMembershipRepository` |
| `backend/internal/administration/services/invitation_repo.go` | services | `pgxInvitationRepository` |
| `backend/internal/administration/services/org_service.go` | services | `OrgService` — `Create`, `List`, `Get`, `Update`, `ListMembers`, `RemoveMember` |
| `backend/internal/administration/services/invitation_service.go` | services | `InvitationService` — `Create`, `GetByToken`, `Accept` |
| `backend/internal/administration/services/email_service.go` | services | `ConsoleEmailService` |
| `backend/internal/administration/handlers/org_handler.go` | handlers | `OrgHandler` — `Create`, `List`, `Get`, `Update`, `ListMembers`, `RemoveMember` |
| `backend/internal/administration/handlers/invitation_handler.go` | handlers | `InvitationHandler` — `Create`, `GetByToken`, `Accept` |
| `backend/internal/administration/handlers/admin_handler.go` | handlers | `AdminHandler` — `ListUsers`, `ToggleSuperadmin` |
| `backend/internal/administration/handlers/role_middleware.go` | handlers | `RoleMiddleware` — `RequireOrgMember`, `RequireOrgAdmin`, `RequireSuperadmin` |

### Modified Files

| File Path | Change Description |
|-----------|-------------------|
| `backend/internal/app/app.go` | Change `*pgx.Conn` to `*pgxpool.Pool`. Import auth and administration sub-packages. Instantiate all repos, services, handlers. Wire all routes with middleware. Add graceful shutdown. |
| `backend/internal/app/app_test.go` | Update for pool type |
| `backend/internal/config/config.go` | Add `JWTSecret`, `AccessTokenTTL`, `RefreshTokenTTL`, `InviteBaseURL`, `InviteTokenTTL`, `BcryptCost` fields |
| `backend/internal/config/config_test.go` | Add tests for new config fields |
| `backend/internal/middleware/cors.go` | Add `Access-Control-Allow-Credentials: true` header |
| `backend/.env.example` | Add `JWT_SECRET`, `ACCESS_TOKEN_TTL`, `REFRESH_TOKEN_TTL`, `INVITE_BASE_URL`, `INVITE_TOKEN_TTL`, `BCRYPT_COST` |
| `backend/go.mod` | Add `golang.org/x/crypto`, `github.com/golang-jwt/jwt/v5` |

### Interfaces

**Shared database primitives** in `backend/internal/database/database.go`:

```go
package database

import (
    "context"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgx/v5/pgxpool"
)

// DBTX — abstracts *pgxpool.Pool and pgx.Tx for repository use.
// All domain packages import this interface rather than defining their own.
type DBTX interface {
    Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// WithTx runs fn inside a database transaction. If fn returns an error or
// panics, the transaction is rolled back; otherwise it is committed.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx) // no-op after commit
    if err := fn(tx); err != nil {
        return err
    }
    return tx.Commit(ctx)
}
```

**Auth interfaces** in `backend/internal/auth/types/repositories.go`:

```go
package types

import "agenteur.ai/api/internal/database"

// UserRepository — user data access
type UserRepository interface {
    Create(ctx context.Context, db database.DBTX, params CreateUserParams) (*User, error)
    GetByID(ctx context.Context, db database.DBTX, id uuid.UUID) (*User, error)
    GetByEmail(ctx context.Context, db database.DBTX, email string) (*User, error)
    Update(ctx context.Context, db database.DBTX, id uuid.UUID, params UpdateUserParams) (*User, error)
}

// RefreshTokenRepository — refresh token data access
type RefreshTokenRepository interface {
    Create(ctx context.Context, db database.DBTX, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*RefreshToken, error)
    GetByHash(ctx context.Context, db database.DBTX, hash string) (*RefreshToken, error)
    DeleteByHash(ctx context.Context, db database.DBTX, hash string) error
    DeleteAllByUser(ctx context.Context, db database.DBTX, userID uuid.UUID) error
}
```

**Administration interfaces** in `backend/internal/administration/types/repositories.go`:

```go
package types

import "agenteur.ai/api/internal/database"

// OrganizationRepository — org data access
type OrganizationRepository interface {
    Create(ctx context.Context, db database.DBTX, params CreateOrgParams) (*Organization, error)
    GetByID(ctx context.Context, db database.DBTX, id uuid.UUID) (*Organization, error)
    GetBySlug(ctx context.Context, db database.DBTX, slug string) (*Organization, error)
    Update(ctx context.Context, db database.DBTX, id uuid.UUID, params UpdateOrgParams) (*Organization, error)
    ListAll(ctx context.Context, db database.DBTX) ([]*Organization, error)
}

// MembershipRepository — org membership data access
type MembershipRepository interface {
    Create(ctx context.Context, db database.DBTX, userID, orgID uuid.UUID, role string) (*OrgMembership, error)
    GetByUserAndOrg(ctx context.Context, db database.DBTX, userID, orgID uuid.UUID) (*OrgMembership, error)
    ListByUser(ctx context.Context, db database.DBTX, userID uuid.UUID) ([]*OrgWithRole, error)
    ListByOrg(ctx context.Context, db database.DBTX, orgID uuid.UUID) ([]*MemberWithUser, error)
    Delete(ctx context.Context, db database.DBTX, userID, orgID uuid.UUID) error
    CountAdmins(ctx context.Context, db database.DBTX, orgID uuid.UUID) (int, error)
}

// InvitationRepository — invitation data access
type InvitationRepository interface {
    Create(ctx context.Context, db database.DBTX, params CreateInvitationParams) (*Invitation, error)
    GetByTokenHash(ctx context.Context, db database.DBTX, hash string) (*InvitationWithOrg, error)
    GetPendingByEmailAndOrg(ctx context.Context, db database.DBTX, email string, orgID uuid.UUID) (*Invitation, error)
    UpdateStatus(ctx context.Context, db database.DBTX, id uuid.UUID, status string) error
}

// EmailService — email sending abstraction
type EmailService interface {
    SendInvitation(ctx context.Context, to, inviterName, orgName, inviteURL string) error
}
```

### Error Taxonomy

| Error (Go) | HTTP Status | User-Facing Message |
|------------|-------------|---------------------|
| `ErrInvalidCredentials` | 401 | "Invalid email or password" |
| `ErrEmailExists` | 409 | "Email already registered" |
| `ErrValidation` | 422 | Field-specific messages from `details` |
| `ErrUnauthorized` | 401 | "Unauthorized" |
| `ErrForbidden` | 403 | "You don't have permission to perform this action" |
| `ErrNotFound` | 404 | "Resource not found" |
| `ErrConflict` | 409 | Context-specific (invite exists, already member, etc.) |
| `ErrLastAdmin` | 400 | "Cannot remove the last admin from the organization" |

---

## 6. Implementation Tasks

### Phase 1: Database & Infrastructure

- **T-001:** Create database migration for users, orgs, memberships, tokens, invitations
  - _Implements:_ ER-001, ER-005, ER-006, ER-007, ER-008, ER-009
  - _Description:_ Create goose migration file with all 5 tables, 2 enum types, and indexes as specified in Section 3. Includes pgcrypto extension for `gen_random_uuid()`.
  - _Files:_
    - Create: `backend/migrations/00002_create_users_and_orgs.sql`
  - _Done condition:_ `make db-reset && make migrate-up` completes without error. `make migrate-status` shows migration applied. `make migrate-down && make migrate-up` round-trips cleanly.
  - _Dependencies:_ None

- **T-002:** Upgrade to pgxpool, extend config, add shared database package, and add Content-Type middleware
  - _Implements:_ ER-001, ER-002, ER-003
  - _Description:_ Change `App.DB` from `*pgx.Conn` to `*pgxpool.Pool` in `app.go`. Change `pgx.Connect()` to `pgxpool.New()`. Add graceful shutdown (OS signal handler that calls `pool.Close()` and `server.Shutdown()`). Extend `Config` with JWT and auth fields: `JWTSecret`, `AccessTokenTTL` (default 15m), `RefreshTokenTTL` (default 168h), `InviteBaseURL` (default `http://localhost:5173/invitations`), `InviteTokenTTL` (default 72h), `BcryptCost` (default 12). No `JWTRefreshSecret` — refresh tokens are opaque random tokens (not JWTs), so only one JWT signing secret is needed. Update `.env.example` with new vars. Update CORS middleware to include `Access-Control-Allow-Credentials: true`. Create shared `database` package (`backend/internal/database/database.go`) containing the `DBTX` interface and `WithTx(ctx, pool, fn)` transaction helper — both auth and administration domains import `DBTX` from here instead of cross-importing from each other. Create `RequireJSONContentType` middleware (`backend/internal/middleware/content_type.go`) that rejects `POST/PUT/PATCH/DELETE` requests without `Content-Type: application/json` with a `415 Unsupported Media Type` response. Wire this middleware on the `/api` route group in `app.go` as the CSRF mitigation strategy.
  - _Files:_
    - Create: `backend/internal/database/database.go`
    - Create: `backend/internal/database/database_test.go`
    - Create: `backend/internal/middleware/content_type.go`
    - Create: `backend/internal/middleware/content_type_test.go`
    - Modify: `backend/internal/app/app.go` — pgxpool, graceful shutdown, wire Content-Type middleware on `/api`
    - Modify: `backend/internal/app/app_test.go` — update DB type
    - Modify: `backend/internal/config/config.go` — add new fields
    - Modify: `backend/internal/config/config_test.go` — test new fields
    - Modify: `backend/internal/middleware/cors.go` — add Allow-Credentials
    - Modify: `backend/.env.example` — add new env vars
  - _Done condition:_ `go build ./...` succeeds. `go test ./internal/config/...` passes. `go test ./internal/middleware/...` passes (including Content-Type enforcement tests: POST without JSON returns 415, POST with JSON passes, GET without JSON passes). `go test ./internal/database/...` passes. App starts and connects to DB with pool.
  - _Dependencies:_ None

- **T-003:** Add Go dependencies and create crypto utilities
  - _Implements:_ ER-001, ER-002, ER-003, ER-007
  - _Description:_ Add `golang.org/x/crypto` and `github.com/golang-jwt/jwt/v5` to go.mod. Create crypto utilities in the auth `services` sub-package with: (1) `password.go` — `HashPassword(plain string) (string, error)` using bcrypt with configurable cost, `CheckPassword(hash, plain string) error` wrapping bcrypt.CompareHashAndPassword. (2) `jwt.go` — `GenerateAccessToken(claims types.TokenClaims, secret string, ttl time.Duration) (string, error)` using HS256, `ValidateAccessToken(tokenString, secret string) (*types.TokenClaims, error)`, `ParseAccessTokenUnvalidated(tokenString, secret string) (*types.TokenClaims, error)` — verifies signature but skips expiry validation (for logout). (3) `token.go` — `GenerateRandomToken() (raw string, hash string, err error)` using 32 bytes from `crypto/rand`, hex-encoded raw, SHA-256 hash.
  - _Files:_
    - Modify: `backend/go.mod` — add dependencies
    - Create: `backend/internal/auth/services/password.go`
    - Create: `backend/internal/auth/services/password_test.go`
    - Create: `backend/internal/auth/services/jwt.go`
    - Create: `backend/internal/auth/services/jwt_test.go`
    - Create: `backend/internal/auth/services/token.go`
    - Create: `backend/internal/auth/services/token_test.go`
  - _Done condition:_ `go test ./internal/auth/services/... -run "TestPassword|TestJWT|TestToken"` passes. Tests cover: password hashing + verification + wrong password, JWT generation + validation + expired token + bad secret, token generation uniqueness + hash verification.
  - _Dependencies:_ None

- **T-004:** Create domain types for both packages
  - _Implements:_ ER-001, ER-005, ER-006, ER-007, ER-008
  - _Description:_ Create Go structs matching the database schema across both domain type packages. **Auth types:** `User` (all columns), `CreateUserParams` (email, password_hash, first_name, last_name), `UpdateUserParams` (first_name, last_name), `RefreshToken`, `TokenClaims` (extends `jwt.RegisteredClaims` with `UserID`, `Email`, `IsSuperadmin`), `UserRepository` and `RefreshTokenRepository` interfaces (import `database.DBTX` from the shared `database` package created in T-002). **Administration types:** `Organization`, `CreateOrgParams` (name, slug), `UpdateOrgParams` (name), `OrgMembership`, `MemberWithUser` (join of membership + user for member listing), `OrgWithRole` (join of org + role for user's org listing), `Invitation`, `InvitationWithOrg` (join with org name + inviter name), `CreateInvitationParams`, `OrganizationRepository`, `MembershipRepository`, `InvitationRepository`, and `EmailService` interfaces (all repository methods use `database.DBTX`). Also create `httputil/response.go` with shared response helpers.
  - _Files:_
    - Create: `backend/internal/auth/types/user.go`
    - Create: `backend/internal/auth/types/token.go`
    - Create: `backend/internal/auth/types/repositories.go`
    - Create: `backend/internal/administration/types/organization.go`
    - Create: `backend/internal/administration/types/membership.go`
    - Create: `backend/internal/administration/types/invitation.go`
    - Create: `backend/internal/administration/types/repositories.go`
    - Create: `backend/internal/httputil/response.go`
  - _Done condition:_ `go build ./internal/auth/types/...` and `go build ./internal/administration/types/...` and `go build ./internal/httputil/...` succeed.
  - _Dependencies:_ T-002 (for `database.DBTX`), T-003 (for jwt.RegisteredClaims import)

### Phase 2: Backend Auth

- **T-005:** Create repository implementations for auth
  - _Implements:_ ER-001, ER-002, ER-003, ER-004
  - _Description:_ Implement `UserRepository` and `RefreshTokenRepository` (defined in `auth/types/repositories.go`) as pgx-backed structs in the auth `services` sub-package.
  - _Files:_
    - Create: `backend/internal/auth/services/user_repo.go`
    - Create: `backend/internal/auth/services/token_repo.go`
  - _Done condition:_ `go build ./internal/auth/services/...` succeeds.
  - _Dependencies:_ T-004

- **T-006:** Create auth service and auth middleware
  - _Implements:_ ER-001, ER-002, ER-003, ER-004
  - _Description:_ Create `AuthService` with methods: `Signup(ctx, email, password, firstName, lastName) (*types.User, rawRefresh string, accessJWT string, error)` — normalizes email to lowercase via `strings.ToLower()`, validates, checks uniqueness, hashes password, inserts user, creates refresh token, generates access JWT. `Login(ctx, email, password)` — normalizes email to lowercase, finds user, compares password, same return as signup. `Logout(ctx, userID uuid.UUID)` — deletes all refresh tokens for the user from DB via `DeleteAllByUser`. `Refresh(ctx, refreshTokenRaw string)` — hashes, looks up, verifies expiry, deletes old, creates new, generates new access JWT. Create `AuthMiddleware` with `Authenticate` method — reads `access_token` cookie, calls `ValidateAccessToken`, stores `TokenClaims` in context. Export `GetUserClaims(ctx)` helper.
  - _Files:_
    - Create: `backend/internal/auth/services/auth_service.go`
    - Create: `backend/internal/auth/handlers/auth_middleware.go`
  - _Done condition:_ `go build ./internal/auth/services/...` and `go build ./internal/auth/handlers/...` succeed.
  - _Dependencies:_ T-003, T-005

- **T-007:** Create auth + user handlers and wire routes
  - _Implements:_ ER-001, ER-002, ER-003, ER-004
  - _Description:_ Create `AuthHandler` with HTTP handlers for signup, login, logout, refresh. Each handler: parses JSON body, validates input, calls auth service, sets/clears httpOnly cookies, writes JSON response. Cookie config: `access_token` (path `/`, httpOnly, SameSite=Lax, Secure in prod, MaxAge from config), `refresh_token` (path `/api/auth`, same flags, MaxAge from config). Create `UserHandler` with `GetMe` (reads claims from context, queries user by ID) and `UpdateMe`. Create `UserService` for user business logic. Wire all routes in `app.go`: public auth routes (signup, login, refresh, logout), authenticated user routes. Logout is a public route (no auth middleware) — it parses the access token directly without expiry validation to extract the user ID. Instantiate repos, services, handlers in `NewApp()` passing the pool.
  - _Files:_
    - Create: `backend/internal/auth/handlers/auth_handler.go`
    - Create: `backend/internal/auth/handlers/user_handler.go`
    - Create: `backend/internal/auth/services/user_service.go`
    - Modify: `backend/internal/app/app.go` — instantiate deps, wire auth + user routes
  - _Done condition:_ `go build ./...` succeeds. Manual curl test: `curl -v -X POST localhost:8080/api/auth/signup -d '{"email":"test@test.com","password":"password123","firstName":"Test","lastName":"User"}'` returns 201 with Set-Cookie headers. `curl -v -b "access_token=<token>" localhost:8080/api/users/me` returns user data. `curl -v -X POST localhost:8080/api/auth/login -d '{"email":"test@test.com","password":"password123"}'` returns 200. `curl -v -X POST -b "refresh_token=<token>" localhost:8080/api/auth/refresh` rotates tokens. `curl -v -X POST -b "access_token=<token>" localhost:8080/api/auth/logout` deletes all refresh tokens and clears cookies (works even with expired access token).
  - _Dependencies:_ T-001, T-002, T-006

### Phase 3: Backend Organizations & Memberships

- **T-008:** Create organization and membership repository implementations
  - _Implements:_ ER-005, ER-006
  - _Description:_ Implement `OrganizationRepository` (Create, GetByID, GetBySlug, Update, ListAll) and `MembershipRepository` (Create, GetByUserAndOrg, ListByUser with org join, ListByOrg with user join, Delete, CountAdmins) using pgx queries against the schema from T-001. Both implement interfaces defined in `types/repositories.go`.
  - _Files:_
    - Create: `backend/internal/administration/services/org_repo.go`
    - Create: `backend/internal/administration/services/membership_repo.go`
  - _Done condition:_ `go build ./internal/administration/services/...` succeeds.
  - _Dependencies:_ T-004

- **T-009:** Create organization service
  - _Implements:_ ER-005, ER-006, ER-009
  - _Description:_ Create `OrgService` that receives `*pgxpool.Pool`, `OrganizationRepository`, and `MembershipRepository` via constructor. Methods: `Create(ctx, userID, name)` — generates slug from name (lowercase, replace non-alphanumeric with hyphens, trim, append random suffix on collision), uses `database.WithTx(ctx, pool, func(tx) { ... })` to create org and admin membership atomically. `List(ctx, userID, isSuperadmin)` — returns user's orgs (via membership) or all orgs if superadmin. `Get(ctx, orgID)`. `Update(ctx, orgID, name)`. `ListMembers(ctx, orgID)`. `RemoveMember(ctx, orgID, userID)` — checks admin count, prevents removing last admin. Non-transactional methods pass the pool directly to repos (it satisfies `database.DBTX`).
  - _Files:_
    - Create: `backend/internal/administration/services/org_service.go`
  - _Done condition:_ `go build ./internal/administration/services/...` succeeds.
  - _Dependencies:_ T-008

- **T-010:** Create role middleware
  - _Implements:_ ER-009, ER-010
  - _Description:_ Create `RoleMiddleware` struct that holds `types.MembershipRepository`, `authtypes.UserRepository`, and `*pgxpool.Pool`. Imports `auth/handlers.GetUserClaims(ctx)` to read JWT claims from context. Methods: `RequireOrgMember` — reads `{orgID}` from chi URL params, queries membership for the authenticated user, allows through if membership exists OR user is superadmin (checked via `UserRepository.GetByID`), stores membership/role in context. Returns 403 if neither. `RequireOrgAdmin` — checks context for role=admin or is_superadmin (from DB). Returns 403 otherwise. `RequireSuperadmin` — reads user ID from JWT claims, queries `UserRepository.GetByID` to check current `is_superadmin` value from DB. Returns 403 if not superadmin. This ensures immediate revocation when a user is demoted — unlike JWT-claims-only checks, there is no stale-claim window.
  - _Files:_
    - Create: `backend/internal/administration/handlers/role_middleware.go`
  - _Done condition:_ `go build ./internal/administration/handlers/...` succeeds.
  - _Dependencies:_ T-005 (user repo for superadmin DB check), T-006 (auth middleware for claims), T-008 (membership repo)

- **T-011:** Create organization handler and wire routes
  - _Implements:_ ER-005, ER-006, ER-009
  - _Description:_ Create `OrgHandler` with HTTP handlers: `Create`, `List`, `Get`, `Update`, `ListMembers`, `RemoveMember`. Wire org routes in `app.go` with role middleware: org-scoped routes use `RequireOrgMember`, admin-only actions use `RequireOrgAdmin`.
  - _Files:_
    - Create: `backend/internal/administration/handlers/org_handler.go`
    - Modify: `backend/internal/app/app.go` — instantiate org deps, wire org routes
  - _Done condition:_ `go build ./...` succeeds. Manual curl: create org (returns 201 with slug), list orgs (returns created org), get org by ID, list members (shows creator as admin), try accessing org as non-member (returns 403).
  - _Dependencies:_ T-007, T-009, T-010

### Phase 4: Backend Invitations & Superadmin

- **T-012:** Create invitation repository and email service stub
  - _Implements:_ ER-007, ER-008
  - _Description:_ Implement `InvitationRepository` (Create, GetByTokenHash with org+inviter joins, GetPendingByEmailAndOrg, UpdateStatus) as a pgx-backed struct in the `services` sub-package. Create `ConsoleEmailService` that implements `types.EmailService` and logs the invite URL to stdout using `slog`.
  - _Files:_
    - Create: `backend/internal/administration/services/invitation_repo.go`
    - Create: `backend/internal/administration/services/email_service.go`
  - _Done condition:_ `go build ./internal/administration/services/...` succeeds.
  - _Dependencies:_ T-004

- **T-013:** Create invitation service
  - _Implements:_ ER-007, ER-008
  - _Description:_ Create `InvitationService` that receives `*pgxpool.Pool`, `InvitationRepository`, `MembershipRepository`, and `EmailService` via constructor. Methods: `Create(ctx, orgID, invitedByUserID, email, role)` — checks for existing pending invite (409 if exists), checks if email is already a member (409), generates random token, stores hash, calls EmailService.SendInvitation with constructed URL, returns invitation. The `idx_invitations_pending_email_org` unique partial index provides a DB-level safety net — the repo layer catches unique violation errors (`pgconn.PgError` code `23505`) and surfaces them as a conflict error. `GetByToken(ctx, rawToken)` — hashes token, looks up with org/inviter joins, verifies status is pending and not expired. `Accept(ctx, rawToken, authenticatedUser)` — hashes token, looks up, verifies email match using `strings.EqualFold()` (403 if mismatch), verifies pending+not expired, uses `database.WithTx(ctx, pool, func(tx) { ... })` to create org membership and update invitation status to accepted atomically. Non-transactional methods pass the pool directly to repos.
  - _Files:_
    - Create: `backend/internal/administration/services/invitation_service.go`
  - _Done condition:_ `go build ./internal/administration/services/...` succeeds.
  - _Dependencies:_ T-008, T-012

- **T-014:** Create invitation handler and wire routes
  - _Implements:_ ER-007, ER-008
  - _Description:_ Create `InvitationHandler` with HTTP handlers: `Create` (parses email+role, calls service), `GetByToken` (reads token from URL, calls service), `Accept` (reads token from URL, gets user from context, calls service). Wire routes: `POST /api/organizations/{orgID}/invitations` (RequireOrgAdmin), `GET /api/invitations/{token}` (public), `POST /api/invitations/{token}/accept` (authenticated).
  - _Files:_
    - Create: `backend/internal/administration/handlers/invitation_handler.go`
    - Modify: `backend/internal/app/app.go` — wire invitation routes
  - _Done condition:_ `go build ./...` succeeds. Manual curl: create invite as org admin (check console for URL), GET invite token (returns org name + details), accept invite as another user (creates membership), verify new member appears in org members list.
  - _Dependencies:_ T-011, T-013

- **T-015:** Create superadmin handler and wire routes
  - _Implements:_ ER-010
  - _Description:_ Create `AdminHandler` with HTTP handlers: `ListUsers` (paginated, with search by email/name), `ToggleSuperadmin` (sets is_superadmin on user). Extend `UserRepository` interface in `auth/types/repositories.go` with `ListAll(ctx, db, page, perPage, search)` and `SetSuperadmin(ctx, db, userID, isSuperadmin)` methods, and add implementations in `auth/services/user_repo.go`. Wire routes under `/api/admin` with `RequireSuperadmin` middleware. Add a Makefile target `seed-superadmin` that runs: `psql $(DATABASE_URL) -c "UPDATE users SET is_superadmin = true WHERE LOWER(email) = LOWER('$(email)')"`.
  - _Files:_
    - Create: `backend/internal/administration/handlers/admin_handler.go`
    - Modify: `backend/internal/auth/types/repositories.go` — extend UserRepository interface
    - Modify: `backend/internal/auth/services/user_repo.go` — add ListAll and SetSuperadmin methods
    - Modify: `backend/internal/app/app.go` — wire admin routes
    - Modify: `backend/Makefile` — add seed-superadmin target
  - _Done condition:_ `go build ./...` succeeds. Manual: promote user via `make seed-superadmin email=test@test.com`, call `GET /api/admin/users` with superadmin token (returns user list), call `PUT /api/admin/users/{id}/superadmin` to toggle another user.
  - _Dependencies:_ T-007, T-010

---

## 7. Testing Strategy

### Unit Tests

| Package / Function | Key Cases | Test Command |
|-------------------|-----------|--------------|
| `database.WithTx` | Commits on success, rolls back on error, rolls back on panic | `go test ./internal/database/... -run TestWithTx` |
| `middleware.RequireJSONContentType` | POST without JSON returns 415, POST with JSON passes, GET without Content-Type passes | `go test ./internal/middleware/... -run TestContentType` |
| `services.HashPassword` / `CheckPassword` | Hash succeeds, verify correct password, reject wrong password | `go test ./internal/auth/services/... -run TestPassword` |
| `services.GenerateAccessToken` / `ValidateAccessToken` | Generate valid token, validate succeeds, reject expired token, reject wrong secret | `go test ./internal/auth/services/... -run TestJWT` |
| `services.GenerateRandomToken` | Tokens are unique, hash is deterministic for same input | `go test ./internal/auth/services/... -run TestToken` |
| `httputil.JSON` / `httputil.Error` | Correct envelope format, status codes, content-type header | `go test ./internal/httputil/... -run TestResponse` |
| `handlers.Authenticate` | Valid token passes, missing cookie returns 401, expired token returns 401 | `go test ./internal/auth/handlers/... -run TestAuth` |
| `handlers.RequireOrgMember` | Member passes, non-member returns 403, superadmin bypasses (DB-checked) | `go test ./internal/administration/handlers/... -run TestRole` |

### Integration Tests

| Test | Setup | Key Assertions | Command |
|------|-------|---------------|---------|
| Auth flow | Test DB with migrations applied | Signup creates user, login returns tokens, refresh rotates tokens, logout revokes all refresh tokens, logout works with expired access token | `go test ./internal/auth/handlers/... -run TestAuthIntegration` |
| Org flow | Test DB + test user | Create org adds admin membership, list returns only member orgs, remove member works | `go test ./internal/administration/handlers/... -run TestOrgIntegration` |
| Invite flow | Test DB + test org + admin user | Create invite, accept adds member, duplicate invite returns 409 | `go test ./internal/administration/handlers/... -run TestInviteIntegration` |

### Manual Testing Checklist

- [ ] Full signup → create org → invite user → accept invite flow via curl
- [ ] Refresh token rotation: wait >15min (or set short AccessTokenTTL), verify API calls still work
- [ ] Logout and verify all refresh tokens for the user are revoked
- [ ] Logout with an expired access token still succeeds
- [ ] Superadmin can access all orgs without membership
- [ ] Cannot remove the last admin from an org
- [ ] CORS works correctly with `Access-Control-Allow-Credentials: true`
- [ ] POST to `/api/*` without `Content-Type: application/json` returns 415
- [ ] GET requests pass through without Content-Type requirement

---

## 8. Rollout & Operations

### Feature Flags

None — this is foundational infrastructure, not an incremental feature behind a flag.

### Monitoring & Alerts

Not applicable for initial local development. To be defined when deploying to dev/production.

### Backward Compatibility

No backward compatibility concerns — this is the first user-facing feature. All new tables and endpoints.

### Rollback Procedure

1. Stop the application
2. Run `make migrate-down` to drop all new tables
3. Deploy previous version of the code
4. `goose down` cleanly reverses the migration (verified in T-001 done condition)

---

## 9. Open Questions & Risks

### Technical Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| JWT secret leaked | All access tokens can be forged | Use strong random secrets (32+ bytes). Rotate secrets requires all users to re-authenticate. Document rotation procedure. |
| Refresh token table grows unbounded | DB bloat, slower queries | Add a periodic cleanup job (cron or background goroutine) that deletes expired tokens. Can be added as follow-up. |
| Email stub delays real email integration | Invites only work via console log in dev | EmailService interface is designed for easy swap. Real implementation is a future task. |

### Open Questions

| Question | Owner | Target Resolution |
|----------|-------|-------------------|
| What email provider to use for production invites? | Engineering | Before production deployment |
| Should password reset be part of this spec or a follow-up? | Product | Follow-up (explicitly out of scope) |
| Rate limiting on auth endpoints? | Engineering | Follow-up (add before production) |

---

## 10. Assumptions

- [ASSUMPTION] The app will run behind HTTPS in production (required for Secure cookie flag)
- [ASSUMPTION] PostgreSQL 16 `gen_random_uuid()` is available via pgcrypto extension
- [ASSUMPTION] bcrypt cost of 12 provides sufficient security vs performance tradeoff for current scale
- [ASSUMPTION] 15-minute access token TTL and 7-day refresh token TTL are acceptable defaults
- [ASSUMPTION] Email sending can be stubbed for the MVP — console logging is sufficient for development/testing
- [ASSUMPTION] No need for email verification on signup in the initial implementation

---

_Traceability verified: Every backend engineering requirement (ER-001 through ER-010) maps to at least one task. Every task (T-001 through T-015) maps to at least one requirement._
