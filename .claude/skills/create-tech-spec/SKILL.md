---
name: create-tech-spec
description: Create detailed Technical Specifications for product features or engineering initiatives. Use when the user asks for a tech spec, implementation plan, engineering design doc, or system design — whether driven by a PRD or by engineering-initiated work such as refactors, performance improvements, infrastructure changes, tech debt reduction, or developer tooling. The output is optimized for handoff to a coding agent for implementation.
---

# Create Tech Spec

## Purpose

Produce a precise, implementation-ready technical specification that a coding agent can execute against without further clarification from a human.

A tech spec can originate from two sources:

- **PRD-driven**: implements a product feature defined by an approved PRD. Every implementation task traces back to a user story.
- **Engineering-initiated**: addresses a technical need that doesn't have (or need) a PRD — refactors, performance work, infrastructure changes, tech debt, migrations, developer tooling, observability, or architectural improvements.

Regardless of source, the tech spec must:
1. State a clear motivation — what problem this solves and why it matters now.
2. Define data models, API contracts, and component architecture as needed.
3. Specify the exact file paths, functions, and interfaces to create or modify.
4. Order tasks so a coding agent can execute them sequentially with clear done-conditions.
5. Be created in the `/tech-specs` directory.

For PRD-driven specs, additionally:
- Map every user story from the PRD to concrete implementation tasks.

For engineering-initiated specs, additionally:
- Define engineering requirements with measurable success criteria that substitute for user stories.

## Reference directories

- **`/prds`** — Source PRDs that define product requirements. PRD-driven tech specs must reference the PRD they implement.
- **`/tech-specs/completed`** — Features and initiatives that have already been implemented. Before writing any tech spec, review completed specs to understand existing architecture, patterns, naming conventions, and data models. New specs must be consistent with established patterns unless the spec explicitly documents a deviation and its rationale.

## Technology stack

All tech specs must target this stack. Do not introduce alternative frameworks, ORMs, or UI libraries without explicitly documenting the deviation and its rationale in the "New Patterns or Deviations" section.

### Backend: Go + PostgreSQL

- **Language:** Go. Follow idiomatic Go patterns — explicit error handling, interfaces for testability, package-level organization by domain.
- **Database:** PostgreSQL.
- **Migrations:** [goose](https://github.com/pressly/goose). All schema changes must be expressed as goose migration files.
  - Migration files live in the migrations directory (check `/tech-specs/completed` for the exact path, typically `migrations/` or `db/migrations/`).
  - File naming: `YYYYMMDDHHMMSS_description.sql` (goose's default timestamp format).
  - Every migration must include both an `-- +goose Up` and `-- +goose Down` section.
  - Specify whether each migration is backward-compatible (can run before new code deploys) or requires coordinated deployment.
  - Done conditions for migration tasks should use `goose status` to verify the migration was applied.
- **Testing:** Use Go's standard `testing` package. Use table-driven tests for functions with multiple cases. Use `testify` if already in use per `/tech-specs/completed`. Benchmarks use `testing.B`.

### Frontend: React + TypeScript + Vite

- **Framework:** React with TypeScript, bundled with Vite.
- **Component library:** [shadcn/ui](https://ui.shadcn.com/). Use shadcn components as the default for all UI elements (buttons, inputs, dialogs, tables, dropdowns, etc.) before building custom components. If a shadcn component doesn't exist for the need, document why a custom component is required.
- **Styling:** Use the styling approach established in `/tech-specs/completed` (typically Tailwind CSS via shadcn's defaults). Do not introduce additional CSS frameworks.
- **File organization:** Follow the existing project structure in `/tech-specs/completed`. Typically:
  - Pages/routes: `src/pages/` or `src/routes/`
  - Reusable components: `src/components/`
  - shadcn components: `src/components/ui/`
  - API client/hooks: `src/api/` or `src/hooks/`
  - Types: `src/types/`
- **Type safety:** All props, API responses, and state must be explicitly typed. No `any` types unless unavoidable and documented.
- **Testing:** Use the frontend testing setup from `/tech-specs/completed` (typically Vitest for unit tests, Playwright or Cypress for E2E). Done conditions should reference `npm run test` or the specific test command.

## When to apply this skill

Apply this skill when the user asks to:
- write a tech spec from a PRD
- create an implementation plan for a feature
- produce an engineering design document
- break down a PRD into development tasks
- design the technical architecture for a product requirement
- spec out a refactor, migration, or infrastructure change
- plan performance optimization work
- define a tech debt reduction initiative
- design developer tooling or internal platform work
- spec out observability, monitoring, or reliability improvements

Do not apply this skill when the user is asking for:
- product requirements or scope definition (use create-prd skill)
- a code review or bug fix with no architectural change
- direct code implementation (just write the code)

## Operating principles

- **Determine the spec type first.** If a PRD exists or is referenced, this is a PRD-driven spec. If the work is a refactor, migration, performance improvement, tooling, or infrastructure initiative with no PRD, this is an engineering-initiated spec. The spec type determines the traceability model (user stories vs engineering requirements) but does not change the rigor of the output.
- For PRD-driven specs: every decision must trace back to a user story. If a technical task has no corresponding user story, it's either missing from the PRD (flag it) or unnecessary (remove it).
- For engineering-initiated specs: every decision must trace back to an engineering requirement defined in the spec itself. If a task doesn't serve a stated requirement, it's scope creep — remove it.
- Optimize for coding agent execution: be explicit about file paths, function signatures, and expected behavior. Ambiguity in a tech spec becomes bugs in implementation.
- Reuse existing patterns. Always review `/tech-specs/completed` before proposing new patterns, abstractions, or architectural conventions.
- Prefer the simplest solution that satisfies the requirements. Call out where you're incurring complexity and why.
- Separate schema migrations, API changes, and UI changes into distinct ordered phases so they can be implemented and tested incrementally.
- Flag technical risks and unknowns early — don't bury them in implementation tasks.

## Required workflow

Follow this sequence every time.

### 1) Input validation and spec type determination

Before writing anything:

1. **Determine spec type.** If the user references a PRD or the work implements product-facing functionality, this is a **PRD-driven** spec. If the work is a refactor, migration, performance initiative, infrastructure change, tooling, or tech debt effort, this is an **engineering-initiated** spec.

2. **Review completed specs.** Regardless of spec type, read through `/tech-specs/completed` to understand:
   - Existing data models and schema conventions
   - API patterns (URL structure, auth, error format, pagination)
   - Frontend component patterns and state management approach
   - Testing patterns and coverage expectations
   - Naming conventions (files, functions, variables, database tables/columns)

3. **Gather requirements based on spec type:**

**If PRD-driven:**
- Read the PRD. Confirm it has been provided or is available at the expected path in `/prds`. If no PRD exists and the work is product-facing, stop and tell the user to create one first using the create-prd skill.
- Identify every user story in the PRD. Create a traceability list mapping each story ID (US-001, US-002, ...) to the technical work required. Every story must appear in this mapping — if a story requires no technical work, document why.
- If the PRD is missing acceptance criteria, has ambiguous requirements, or contains contradictions, surface these as blocking questions before drafting.

**If engineering-initiated:**
- Define the engineering motivation. Collect or infer the following (ask targeted questions if critical details are missing):
  - **What is the technical problem?** Specific and falsifiable — not "the code is messy" but "the message fanout path makes N+1 queries per channel, causing p95 latency of 800ms at current load."
  - **Why now?** What trigger makes this work urgent — approaching scale limits, blocking product work, reliability incidents, developer productivity loss, security exposure, etc.
  - **What does success look like?** Measurable engineering outcomes — latency targets, error rate reductions, build time improvements, lines of code removed, operational cost savings, etc.
  - **What are the constraints?** Backward compatibility requirements, migration windows, uptime expectations during the change, team bandwidth.
  - **What is out of scope?** Explicit boundaries to prevent the initiative from expanding unboundedly.
- Write **engineering requirements** (ER-001, ER-002, ...) that serve the same traceability role as user stories. Each requirement must have:
  - **ID**: sequential (ER-001, ER-002, ...).
  - **Requirement**: a specific, testable statement of what the system must do or achieve after the work is complete.
  - **Success criteria**: measurable condition(s) that confirm the requirement is met — a benchmark result, a test passing, a metric threshold, or an observable behavior change.
  - **Priority**: must / should / may (same definitions as PRD stories).

**Engineering requirement examples:**

> **ER-001** (must): Eliminate N+1 queries in the message fanout path.
> _Success criteria:_ Message broadcast to a 50-member channel executes ≤ 2 SQL queries. `BenchmarkMessageFanout` shows p95 < 50ms at 100 concurrent channels.

> **ER-002** (must): Migrate invite tokens from VARCHAR to UUID without downtime.
> _Success criteria:_ Migration completes on production data (est. 2M rows) in < 5 minutes. Zero errors in application logs during migration window. Old and new token formats are accepted during the transition period.

> **ER-003** (should): Reduce CI pipeline duration for the `api` module.
> _Success criteria:_ `api` module test suite completes in < 90 seconds (current: ~4 minutes). No reduction in test coverage.

> **ER-004** (may): Add structured logging to the WebSocket connection lifecycle.
> _Success criteria:_ Connection open, close, and error events emit JSON logs with connection_id, user_id, and duration_ms fields. Logs are queryable in the observability platform.

### 2) Architecture and design decisions

Document the high-level technical approach:

- **System context**: where does this feature sit in the existing architecture? What services, databases, and external systems does it touch?
- **Key design decisions**: for each non-trivial decision, state the decision, the alternatives you considered, and why you chose this approach. Tie each decision back to a PRD requirement (US-XXX) or engineering requirement (ER-XXX).
- **New patterns or deviations**: if introducing a new pattern that differs from `/tech-specs/completed`, explicitly call it out with rationale. This prevents a coding agent from following stale conventions.
- **Dependencies**: external services, libraries, or internal modules this feature depends on. Include version constraints where relevant.

### 3) Data model design

Define all schema changes as PostgreSQL DDL:

- **New tables**: full schema with column names, PostgreSQL types, constraints, indexes, and relationships. Use PostgreSQL-native types (`UUID`, `TIMESTAMPTZ`, `JSONB`, `TEXT`, etc.).
- **Modified tables**: specify the exact alteration — columns added/removed/changed, new indexes, constraint changes.
- **Goose migrations**: each schema change must map to a goose migration file. Specify:
  - The migration filename: `YYYYMMDDHHMMSS_description.sql`
  - The full `-- +goose Up` SQL
  - The full `-- +goose Down` SQL (must cleanly reverse the up migration)
  - Whether the migration is backward-compatible (can deploy before code) or requires coordinated deployment
  - The migration file path (e.g., `migrations/20260216120000_create_invites.sql`)
- **Seed data or backfills**: if existing data needs transformation, define the backfill logic. For large backfills, specify whether to use a goose migration (for small/fast operations) or a separate Go script (for long-running backfills that shouldn't block deployment).

Use the naming conventions established in `/tech-specs/completed`. If no convention exists, define one and document it.

**Example goose migration:**

```sql
-- +goose Up
CREATE TABLE invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    token UUID NOT NULL DEFAULT gen_random_uuid(),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'expired', 'revoked')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_invites_workspace_id ON invites(workspace_id);
CREATE UNIQUE INDEX idx_invites_token ON invites(token);
CREATE INDEX idx_invites_email_workspace ON invites(email, workspace_id);

-- +goose Down
DROP TABLE IF EXISTS invites;
```

### 4) API design

Define every new or modified endpoint. Handlers are implemented in Go — specify the handler function signature and the package it belongs to.

For each endpoint provide:
- **Method and path** (e.g., `POST /api/v1/workspaces/{workspace_id}/invites`)
- **Handler**: Go function name and package (e.g., `handlers.CreateInvite`)
- **Request schema**: fields, types, validation rules, required vs optional. Define the Go struct used for request binding.
- **Response schema**: fields, types, status codes for success and each error case. Define the Go struct used for JSON serialization.
- **Authentication and authorization**: what roles/permissions are required
- **Rate limiting**: if applicable
- **Idempotency**: behavior on duplicate requests

For WebSocket or real-time events, define:
- **Event name and payload schema**
- **Subscription/channel model**
- **Delivery guarantees** (at-least-once, exactly-once, best-effort)

Match the API style from `/tech-specs/completed`. Consistency in error format, pagination style, and auth headers matters for coding agents.

### 5) Component and module design

Define the implementation structure for both backend and frontend:

**Backend (Go):**
- **New files to create**: exact file paths, package name, and exported types/functions. Follow Go conventions — one responsibility per file, package name matches directory name.
- **Existing files to modify**: exact file paths, what changes, and why.
- **Key functions/methods**: provide the function signature with full Go types (e.g., `func (s *InviteService) Create(ctx context.Context, req CreateInviteRequest) (*Invite, error)`), input/output types, and a brief description of the behavior and edge cases. Do not write full implementations.
- **Interfaces**: define any new interfaces for dependency injection and testability (e.g., `type InviteStore interface { ... }`).
- **Error handling**: define the error taxonomy using typed errors or sentinel errors. What errors can occur, how they propagate through the handler chain, and what HTTP status + response body the user sees for each.

**Frontend (React/TypeScript):**
- **New components**: exact file paths, props interface (TypeScript), and which shadcn/ui components they compose. State whether each component is a page, a reusable component, or a feature-specific component.
- **Existing components to modify**: exact file paths, what changes, and why.
- **shadcn components to add**: list any shadcn/ui components that need to be installed (e.g., `npx shadcn-ui@latest add dialog`). Check the existing `src/components/ui/` to avoid re-adding components already installed.
- **TypeScript types**: define new interfaces/types for API request/response shapes, component props, and shared state. These should mirror the Go structs where applicable to maintain contract alignment.
- **API client layer**: define new API functions or hooks (e.g., `useInvites()`, `createInvite()`), their input/output types, and error handling behavior.
- **State management**: how data flows — what's fetched from the API, what's stored in component state vs context vs a store, and how components communicate.

**Cross-cutting:**
- **State management**: how data flows end-to-end — what's stored where (PostgreSQL, in-memory cache, React state) and how backend events (e.g., WebSocket messages) update the frontend.
- **Error propagation**: how a Go error becomes an HTTP response becomes a TypeScript error becomes a user-facing toast or inline message. Define this chain explicitly.

### 6) Implementation task breakdown

This is the core deliverable for a coding agent. Break the spec into an ordered list of discrete tasks that can be executed sequentially.

Each task must include:
- **Task ID**: sequential (T-001, T-002, ...) for reference.
- **Requirement mapping**: which requirement(s) this task implements — PRD stories (e.g., "Implements US-001, US-003") for PRD-driven specs, or engineering requirements (e.g., "Implements ER-001") for engineering-initiated specs.
- **Description**: what to build, in precise terms.
- **Files to create or modify**: exact paths.
- **Done condition**: how to verify this task is complete — a test to run, a behavior to observe, or an assertion to check. Must be objectively verifiable, not subjective.
- **Dependencies**: which prior tasks must be complete before starting this one.

**Task ordering rules:**
1. Schema migrations and data model changes first.
2. Backend logic and API endpoints second.
3. Real-time/event infrastructure third.
4. Frontend components and UI fourth.
5. Integration wiring (connecting frontend to backend) fifth.
6. Tests last (or interleaved per phase if the project uses TDD).

Within each phase, order by dependency — a task should never reference code or schema that hasn't been created by a prior task.

**Task sizing guidance:**
- Each task should represent roughly 1 meaningful commit — a single coherent change that could be reviewed independently.
- If a task description exceeds 8-10 sentences, it's probably too large. Split it.
- If a task only changes a single line or adds an import, it's too small. Merge it with a related task.

**Example task (PRD-driven):**

> **T-003**: Create the invite API endpoint
> - _Implements:_ US-001, US-002
> - _Description:_ Create `POST /api/v1/workspaces/{workspace_id}/invites` endpoint. Validate that the request body contains a valid email. Check that the caller has the `workspace:admin` role. Check that the email is not already an active member (return 409 if so). Create an `invites` row with status `pending` and a generated token with 72-hour expiry. Emit an `invite.created` event to the event bus for the email service to consume.
> - _Files:_
>   - Create: `src/api/handlers/invites.go`
>   - Create: `src/api/handlers/invites_test.go`
>   - Modify: `src/api/router.go` — register the new route
> - _Done condition:_ `go test ./src/api/handlers/... -run TestCreateInvite` passes. Manual test: POST request with valid admin token returns 201 with invite object; POST with duplicate email returns 409; POST without admin role returns 403.
> - _Dependencies:_ T-001 (schema migration), T-002 (invite model)

**Example task (engineering-initiated):**

> **T-002**: Replace N+1 channel member queries with batch load
> - _Implements:_ ER-001
> - _Description:_ Refactor `broadcastToChannel()` in `src/messaging/fanout.go` to replace the per-member `SELECT` inside the loop with a single `SELECT ... WHERE channel_id = ? AND status = 'active'` batch query. Extract the batch query into `src/messaging/members.go` as `getActiveMembers(channelID uuid.UUID) ([]Member, error)`. Update the fanout loop to iterate over the pre-fetched slice.
> - _Files:_
>   - Create: `src/messaging/members.go`
>   - Create: `src/messaging/members_test.go`
>   - Modify: `src/messaging/fanout.go` — replace member lookup loop with call to `getActiveMembers`
> - _Done condition:_ `go test ./src/messaging/... -run TestBroadcastToChannel` passes. `BenchmarkMessageFanout` with 50 members shows ≤ 2 SQL queries (verified via query counter in test harness). p95 latency < 50ms.
> - _Dependencies:_ T-001 (add index on channel_members.channel_id)

### 7) Testing strategy

Define the testing approach for both backend and frontend:

**Backend (Go):**
- **Unit tests**: use Go's `testing` package with table-driven tests. Define which functions/methods need tests, key edge cases, and expected coverage. Test files live alongside the code they test (e.g., `invites_test.go` next to `invites.go`). Use interfaces and dependency injection to mock external dependencies (database, external APIs).
- **Integration tests**: API-level tests that exercise the handler → service → database chain. Define the setup (test database, seed data via goose migrations) and key assertions. Specify the test command (e.g., `go test ./... -tags=integration`).
- **Benchmarks**: for performance-sensitive code, define Go benchmarks using `testing.B`. Specify the benchmark name and the target thresholds.

**Frontend (React/TypeScript):**
- **Unit tests**: component-level tests using the project's testing framework (check `/tech-specs/completed` — typically Vitest + React Testing Library). Test component rendering, user interactions, and state changes. Mock API calls at the fetch/hook layer.
- **E2E tests**: critical user flows using the project's E2E framework (check `/tech-specs/completed` — typically Playwright or Cypress). Map these directly to the happy-path and error user flows from the PRD (PRD-driven) or to engineering requirements' success criteria (engineering-initiated).

**Cross-cutting:**
- **Manual testing checklist**: for behaviors that are hard to automate or need human judgment (e.g., visual regressions, real-time WebSocket behavior under network conditions, shadcn component styling edge cases).

### 8) Rollout and operational concerns

Translate the rollout plan into technical steps. For PRD-driven specs, this comes from the PRD's rollout section. For engineering-initiated specs, define the rollout strategy here based on the risk profile of the change:

- **Feature flags**: flag name, default state, how to enable/disable, cleanup plan after GA.
- **Monitoring and alerting**: specific metrics to track, alert thresholds, dashboards to create or update.
- **Backward compatibility**: can the new code be deployed without the new schema? Can old clients coexist with new endpoints?
- **Rollback procedure**: exact steps to revert — does rolling back the deploy suffice, or do schema/data changes need reversal too?

### 9) Self-evaluation and revision

Before presenting the final tech spec, verify every item below. Revise any section that fails. Only present the spec once all items pass.

**Traceability (PRD-driven specs):**
- [ ] Every user story from the PRD is mapped to at least one implementation task.
- [ ] No implementation task exists that doesn't trace back to a user story — if it does, either the PRD is missing a story or the task is unnecessary.
- [ ] The testing strategy covers happy paths, error cases, and edge cases from the PRD acceptance criteria.

**Traceability (engineering-initiated specs):**
- [ ] Every engineering requirement (ER-XXX) has measurable success criteria.
- [ ] Every engineering requirement is mapped to at least one implementation task.
- [ ] No implementation task exists that doesn't trace back to an engineering requirement.
- [ ] The testing strategy validates each engineering requirement's success criteria.

**Universal (all specs):**
- [ ] Data model changes are fully specified — PostgreSQL column types, constraints, indexes, and migration order.
- [ ] Every schema change has a corresponding goose migration file with both `-- +goose Up` and `-- +goose Down` sections.
- [ ] API contracts include request/response Go structs, HTTP status codes, auth requirements, and error cases.
- [ ] Frontend components use shadcn/ui where applicable — custom components are only introduced when no shadcn equivalent exists, with documented rationale.
- [ ] All TypeScript types are explicitly defined — no `any` types unless documented as unavoidable.
- [ ] API response types in TypeScript mirror the corresponding Go response structs.
- [ ] Every implementation task has a clear done condition that a coding agent can verify programmatically (a `go test` command, an `npm run test` command, or a verifiable assertion).
- [ ] Tasks are ordered so no task references code, schema, or infrastructure created by a later task.
- [ ] File paths are exact — no placeholders like "somewhere in the api folder."
- [ ] Patterns are consistent with `/tech-specs/completed` — any deviation is explicitly documented with rationale.
- [ ] A coding agent could execute this spec top-to-bottom without asking a human for clarification.

---

## Tech spec output template

Use exactly this structure and headings when producing the final document:

```markdown
# Tech Spec: [Feature Name]

**Type:** PRD-driven | Engineering-initiated
**PRD:** [path to source PRD — omit for engineering-initiated specs]
**Author:** [name]
**Last updated:** [date]
**Status:** Draft | In Review | Approved

---

## 1. Overview

### Summary

[One to two sentences: what this spec implements and why.]

### Requirement Mapping

_Use the section that matches the spec type. Delete the other._

#### PRD-driven: Story Mapping

| Story ID | Story Summary | Implementing Tasks |
|----------|--------------|-------------------|
| US-001 | [Brief description from PRD] | T-001, T-003, T-005 |
| US-002 | [Brief description from PRD] | T-003, T-004 |

#### Engineering-initiated: Motivation & Requirements

**Problem:** [Specific, falsifiable description of the technical problem.]

**Why now:** [What trigger makes this urgent — approaching limits, incidents, blocked work, etc.]

**Out of scope:** [Explicit boundaries.]

| Req ID | Requirement | Success Criteria | Priority |
|--------|------------|-----------------|----------|
| ER-001 | [Specific, testable requirement] | [Measurable condition] | must/should/may |
| ER-002 | [Specific, testable requirement] | [Measurable condition] | must/should/may |

**Requirement-to-Task Mapping:**

| Req ID | Implementing Tasks |
|--------|-------------------|
| ER-001 | T-001, T-002 |
| ER-002 | T-003 |

---

## 2. Architecture & Design Decisions

### System Context

[Where this feature fits in the existing architecture. What services, databases, and systems it touches. A brief diagram in Mermaid or ASCII if helpful.]

### Key Decisions

| Decision | Alternatives Considered | Rationale |
|----------|------------------------|-----------|
| [e.g., Store invites in PostgreSQL, not Redis] | [Redis with TTL expiry] | [Need query history; TTL-only doesn't support status transitions] |

### New Patterns or Deviations

[If introducing anything that differs from established patterns in /tech-specs/completed, document it here. If none, state "No new patterns — follows existing conventions."]

### Dependencies

| Dependency | Type | Version / Constraint | Notes |
|-----------|------|---------------------|-------|
| [e.g., github.com/google/uuid] | Go module | [>= 1.6] | [UUID generation for invite tokens] |
| [e.g., @radix-ui/react-dialog] | npm (via shadcn) | [installed with shadcn Dialog] | [Invite confirmation modal] |

---

## 3. Data Model

### New Tables

#### [table_name]

| Column | Type | Constraints | Notes |
|--------|------|------------|-------|
| id | UUID | PK, DEFAULT gen_random_uuid() | |
| [column] | [type] | [constraints] | [notes] |
| created_at | TIMESTAMPTZ | NOT NULL, DEFAULT now() | |

**Indexes:**
- `idx_[table]_[column]` on ([column]) — [reason for index]

**Relationships:**
- [table_name].[column] → [other_table].[column] (FK, ON DELETE CASCADE)

### Modified Tables

| Table | Change | Migration Notes |
|-------|--------|----------------|
| [table] | Add column [name] [type] | Backward-compatible; nullable, backfill in T-XXX |

### Goose Migrations

#### `migrations/[YYYYMMDDHHMMSS]_[description].sql`

```sql
-- +goose Up
[SQL statements]

-- +goose Down
[SQL statements that cleanly reverse the Up migration]
```

**Migration order and compatibility:**

| # | Migration File | Backward Compatible | Notes |
|---|---------------|-------------------|-------|
| 1 | `migrations/[timestamp]_[description].sql` | yes/no | [deployment notes] |
| 2 | `migrations/[timestamp]_[description].sql` | yes/no | [deployment notes] |

**Verification:** `goose status` shows all migrations applied. `goose down` followed by `goose up` completes without errors.

---

## 4. API Design

### [METHOD] [path]

**Description:** [What this endpoint does]

**Handler:** `[package].[FunctionName]` in `[file path]`

**Auth:** [Required role or permission]

**Request struct:**
```go
type CreateInviteRequest struct {
    Email string `json:"email" validate:"required,email"`
}
```

**Response struct:**
```go
type InviteResponse struct {
    ID          uuid.UUID `json:"id"`
    Email       string    `json:"email"`
    Status      string    `json:"status"`
    ExpiresAt   time.Time `json:"expires_at"`
    CreatedAt   time.Time `json:"created_at"`
}
```

**TypeScript request/response types:**
```typescript
interface CreateInviteRequest {
  email: string;
}

interface InviteResponse {
  id: string;
  email: string;
  status: "pending" | "accepted" | "expired" | "revoked";
  expires_at: string;
  created_at: string;
}
```

**Error Responses:**

| Status | Condition | Body |
|--------|-----------|------|
| 400 | [Validation failure] | `{"error": "description"}` |
| 403 | [Insufficient permissions] | `{"error": "description"}` |
| 409 | [Conflict condition] | `{"error": "description"}` |

---

### WebSocket / Real-Time Events

#### [event_name]

**Channel:** [channel/topic pattern]
**Payload:**
```json
{
  "field": "type"
}
```
**Delivery:** [at-least-once / best-effort]
**Trigger:** [What causes this event to fire]

---

## 5. Component & Module Design

### Backend (Go) — New Files

| File Path | Package | Exported Interface |
|-----------|---------|-------------------|
| [exact/path/to/file.go] | [package name] | [Key exported functions/types] |

### Backend (Go) — Modified Files

| File Path | Change Description |
|-----------|-------------------|
| [exact/path/to/file.go] | [What changes and why] |

### Backend Interfaces

```go
// [InterfaceName] — [purpose, used for dependency injection/testing]
type InviteStore interface {
    Create(ctx context.Context, invite *Invite) error
    GetByToken(ctx context.Context, token uuid.UUID) (*Invite, error)
}
```

### Frontend (React/TypeScript) — New Files

| File Path | Type | shadcn Components Used | Props Interface |
|-----------|------|----------------------|----------------|
| [src/components/feature/Component.tsx] | Page / Reusable / Feature | [Button, Dialog, Input] | [ComponentProps] |

### Frontend (React/TypeScript) — Modified Files

| File Path | Change Description |
|-----------|-------------------|
| [src/path/to/file.tsx] | [What changes and why] |

### shadcn/ui Components to Install

```bash
npx shadcn-ui@latest add [component-name]
```

[List only components not already in src/components/ui/]

### Frontend TypeScript Types

```typescript
// [src/types/invites.ts]
interface Invite {
  id: string;
  email: string;
  status: "pending" | "accepted" | "expired" | "revoked";
  // ...
}
```

### API Client / Hooks

| Hook / Function | File Path | Description |
|----------------|-----------|-------------|
| [useInvites()] | [src/hooks/useInvites.ts] | [Fetches invite list for current workspace] |

### Error Taxonomy

| Error (Go) | HTTP Status | TypeScript Error | User-Facing Message | Log Level |
|------------|-------------|-----------------|---------------------|-----------|
| [ErrInviteExists] | 409 | [InviteConflictError] | [This user has already been invited] | warn |

---

## 6. Implementation Tasks

### Phase 1: Data Model

- **T-001:** [Task title]
  - _Implements:_ [US-XXX or ER-XXX]
  - _Description:_ [Precise description of what to build]
  - _Files:_
    - Create: [path]
    - Modify: [path] — [what changes]
  - _Done condition:_ [Test command or verifiable assertion]
  - _Dependencies:_ None

### Phase 2: Backend Logic & API

- **T-002:** [Task title]
  - _Implements:_ [US-XXX or ER-XXX]
  - _Description:_ [Precise description]
  - _Files:_
    - Create: [path]
    - Modify: [path] — [what changes]
  - _Done condition:_ [Test command or verifiable assertion]
  - _Dependencies:_ T-001

### Phase 3: Real-Time / Events

[If applicable]

### Phase 4: Frontend

- **T-00X:** [Task title]
  - _Implements:_ [US-XXX or ER-XXX]
  - _Description:_ [Precise description]
  - _Files:_
    - Create: [path]
    - Modify: [path] — [what changes]
  - _Done condition:_ [Test command or verifiable assertion]
  - _Dependencies:_ [Prior tasks]

### Phase 5: Integration

[Wiring frontend to backend, connecting services]

### Phase 6: Tests

[Additional test tasks not covered inline above]

---

## 7. Testing Strategy

### Backend Unit Tests (Go)

| Package / Function | Key Cases | Test Command |
|-------------------|-----------|--------------|
| [package.Function] | [edge cases to cover] | `go test ./[path]/... -run [TestName]` |

### Backend Integration Tests (Go)

| Test | Setup | Key Assertions | Command |
|------|-------|---------------|---------|
| [test description] | [test DB, goose migrations, seed data] | [what to assert] | `go test ./... -tags=integration` |

### Backend Benchmarks (Go)

| Benchmark | Target Threshold | Command |
|-----------|-----------------|---------|
| [BenchmarkName] | [e.g., p95 < 50ms] | `go test ./[path]/... -bench=[BenchmarkName] -benchmem` |

### Frontend Unit Tests (React/TypeScript)

| Component / Hook | Key Cases | Test Command |
|-----------------|-----------|--------------|
| [ComponentName] | [rendering, interactions, state] | `npm run test -- [path]` |

### End-to-End Tests

| User Flow / Scenario | Steps | Maps to Requirement | Framework |
|---------------------|-------|---------------------|-----------|
| [flow name] | [brief steps] | [US-XXX or ER-XXX] | [Playwright/Cypress] |

### Manual Testing Checklist

- [ ] [Behavior to verify manually and why it can't be automated]

---

## 8. Rollout & Operations

### Feature Flags

| Flag Name | Default | Cleanup Target |
|-----------|---------|---------------|
| [flag_name] | off | [date or milestone] |

### Monitoring & Alerts

| Metric | Alert Threshold | Dashboard |
|--------|----------------|-----------|
| [metric name] | [threshold] | [dashboard name/link] |

### Backward Compatibility

[Can the code deploy independently of schema changes? Can old and new clients coexist?]

### Rollback Procedure

[Exact steps to revert. State whether deploy rollback is sufficient or if `goose down` is needed to reverse schema changes. If goose down is required, specify which migrations to revert and confirm the down migrations have been tested.]

---

## 9. Open Questions & Risks

### Technical Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| [Specific technical risk] | [What breaks] | [Concrete action] |

### Open Questions

| Question | Owner | Target Resolution |
|----------|-------|-------------------|
| [Question] | [Name] | [Date] |

---

## 10. Assumptions

- [ASSUMPTION] [Technical assumption that needs validation]

---

_Traceability verified: Every requirement (US-XXX or ER-XXX) maps to at least one task. Every task maps to at least one requirement. See self-evaluation (Section 9 of skill) for full criteria._
```

---

## Writing standards

- Be precise about file paths, function names, and types. A coding agent cannot infer "the right place" — it needs an exact path.
- Use the codebase's existing naming conventions. When in doubt, check `/tech-specs/completed`.
- **Go conventions**: use idiomatic Go — exported names are PascalCase, unexported are camelCase, packages are lowercase single words, errors are returned not thrown. Define interfaces for dependencies to enable testing with mocks.
- **TypeScript conventions**: explicitly type all props, API responses, and state. Use union types for known string enums (e.g., `"pending" | "accepted"`). Never use `any` unless documented as unavoidable.
- **shadcn-first**: always check if a shadcn/ui component exists before creating a custom UI element. If using shadcn, reference the component name and its import path (e.g., `import { Button } from "@/components/ui/button"`).
- **Contract alignment**: Go response structs and TypeScript response interfaces must stay in sync. When defining an API, always specify both the Go struct and the corresponding TypeScript type.
- Describe behavior, not just structure. For complex functions, state what the function does, what inputs it expects, what it returns, and how it handles errors — but do not write the full implementation. The coding agent writes the code; the spec defines the contract.
  - **Good:** "Create `func (s *InviteService) Create(ctx context.Context, req CreateInviteRequest) (*Invite, error)` — validates email format, checks workspace membership via `MemberStore.ExistsByEmail()`, creates invite row with 72-hour expiry, returns `ErrInviteExists` if duplicate. Emits `invite.created` event on success."
  - **Bad:** "Write a function to create invites." ← too vague for a coding agent.
  - **Also bad:** Full 40-line function implementation in the spec. ← the coding agent should write the code; the spec defines the contract and edge cases.
- Keep done conditions objectively verifiable. "It works" is not a done condition. "`go test ./... -run TestInviteCreate` passes with 0 failures" is. "`npm run test -- src/components/InviteDialog.test.tsx` passes" is.
- When referencing requirements, use the exact ID (US-001 or ER-001) — never paraphrase the requirement or use approximate descriptions.