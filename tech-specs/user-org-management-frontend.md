# Tech Spec: User & Organization Management — Frontend

**Type:** Engineering-initiated
**Author:** Claude
**Last updated:** 2026-02-18
**Status:** Draft
**Related:** [Backend spec](./user-org-management-backend.md)

---

## 1. Overview

### Summary

Implement the frontend for user authentication, organization management, and role-based access control in the Agenteur platform. This spec covers the React/TypeScript UI that consumes the backend API defined in the [backend spec](./user-org-management-backend.md). It includes login/signup flows, protected routing, organization CRUD, member management, invitation acceptance, and superadmin pages.

### Motivation & Requirements

**Problem:** The frontend has placeholder Login/Signup pages with no functionality. There is no auth state management, no API client, no protected routes, and no organization UI. Users cannot interact with the platform until these are built.

**Why now:** The backend API (auth, organizations, invitations, superadmin) is being built in parallel. The frontend must be ready to consume it. This is a blocking dependency for all product UI work.

**Out of scope:**
- OAuth/SSO login buttons (Google, GitHub) — future enhancement
- Password reset UI — future enhancement
- Email verification UI — future enhancement
- Agent/skill management UI (depends on this spec)
- Backend implementation (see [backend spec](./user-org-management-backend.md))

| Req ID | Requirement | Success Criteria | Priority |
|--------|------------|-----------------|----------|
| ER-011 | Frontend auth flow with protected routes | Login/Signup forms, AuthContext, ProtectedRoute component. Unauthenticated users redirected to `/login`. `npm run build` succeeds. | must |
| ER-012 | Frontend org management UI | Org creation, org switcher, members list, invite dialog, invitation acceptance page. `npm run build` succeeds. | must |

**Requirement-to-Task Mapping:**

| Req ID | Implementing Tasks |
|--------|-------------------|
| ER-011 | T-016, T-017, T-018, T-019 |
| ER-012 | T-020, T-021, T-022, T-023 |

---

## 2. Architecture & Design Decisions

### System Context

The frontend is a React SPA served by Vite on `:5173` in development. It communicates with the Go backend on `:8080` via REST. Authentication uses httpOnly cookies — no tokens are stored in JavaScript-accessible storage.

```
┌──────────────────────────────────────────────────────┐
│  Frontend (React/Vite :5173)                         │
│                                                       │
│  App.tsx                                              │
│    ├── AuthProvider (AuthContext)                      │
│    │   ├── OrgProvider (OrgContext)                    │
│    │   │   ├── Layout (nav bar, org switcher, menu)   │
│    │   │   │   ├── ProtectedRoute                     │
│    │   │   │   │   ├── Dashboard                      │
│    │   │   │   │   ├── Org pages                      │
│    │   │   │   │   ├── Settings                       │
│    │   │   │   │   └── Admin pages (superadmin only)  │
│    ├── Login (public)                                 │
│    ├── Signup (public)                                │
│    └── InviteAccept (public/auth)                     │
└───────────────────────┬──────────────────────────────┘
                        │ fetch (credentials: include)
                        │ httpOnly cookies
                ┌───────▼───────┐
                │  Backend API  │
                │  :8080        │
                └───────────────┘
```

### Key Decisions

| Decision | Alternatives Considered | Rationale |
|----------|------------------------|-----------|
| `credentials: "include"` on all fetch calls | Authorization header with token from localStorage | Backend uses httpOnly cookies. No token handling in JS — XSS resistant. |
| Auto-refresh on 401 in API client | Manual token refresh, redirect immediately | Transparent to user. API client retries once after refreshing, only redirects if refresh also fails. |
| AuthContext + OrgContext as separate providers | Single combined context, Zustand/Redux store | Separation of concerns. Auth state is independent of org state. Context is sufficient at this scale. |
| Org selection persisted in localStorage | URL-based org routing, cookie | LocalStorage survives page reloads. No server roundtrip needed to restore selection. |
| shadcn/ui for all components | Custom components, Material UI, Chakra | Project standard. Composable, unstyled primitives built on Radix. Already partially installed. |
| React Router `<Outlet>` pattern for protected routes | HOC wrapper, render props | Cleaner nested routing. Works naturally with React Router 7 layout routes. |

### Dependencies

| Dependency | Type | Version / Constraint | Notes |
|-----------|------|---------------------|-------|
| shadcn `input` | npm (via shadcn CLI) | latest | Form input fields |
| shadcn `label` | npm (via shadcn CLI) | latest | Form labels |
| shadcn `card` | npm (via shadcn CLI) | latest | Auth form cards, org cards |
| shadcn `dialog` | npm (via shadcn CLI) | latest | Invite modal |
| shadcn `dropdown-menu` | npm (via shadcn CLI) | latest | User menu, org switcher |
| shadcn `avatar` | npm (via shadcn CLI) | latest | User avatar in nav |
| shadcn `badge` | npm (via shadcn CLI) | latest | Role badges |
| shadcn `table` | npm (via shadcn CLI) | latest | Members list, admin user list |
| shadcn `select` | npm (via shadcn CLI) | latest | Role selection in invite form |
| shadcn `sonner` | npm (via shadcn CLI) | latest | Toast notifications |
| shadcn `separator` | npm (via shadcn CLI) | latest | Visual separators |
| shadcn `skeleton` | npm (via shadcn CLI) | latest | Loading states |

(Button already installed in `frontend/src/components/ui/button.tsx`)

---

## 3. API Contracts

This section defines the TypeScript types that mirror the Go response structs from the [backend spec](./user-org-management-backend.md). All JSON fields use camelCase.

### TypeScript Types

```typescript
// frontend/src/types/index.ts

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  isSuperadmin: boolean;
  createdAt: string;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  createdAt: string;
}

export interface OrgWithRole extends Organization {
  role: "admin" | "user";
}

export interface Member {
  userId: string;
  email: string;
  firstName: string;
  lastName: string;
  role: "admin" | "user";
  joinedAt: string;
}

export interface Invitation {
  id: string;
  email: string;
  role: "admin" | "user";
  status: "pending" | "accepted" | "expired" | "revoked";
  expiresAt: string;
  createdAt: string;
}

export interface InvitationDetail {
  organizationName: string;
  email: string;
  role: "admin" | "user";
  invitedByName: string;
  expiresAt: string;
}
```

### Request Types

```typescript
export interface SignupRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface UpdateUserRequest {
  firstName: string;
  lastName: string;
}

export interface CreateOrgRequest {
  name: string;
}

export interface CreateInvitationRequest {
  email: string;
  role: "admin" | "user";
}

export interface ToggleSuperadminRequest {
  isSuperadmin: boolean;
}
```

### Response Envelope

All API responses follow this envelope:

```typescript
// Success
interface ApiSuccessResponse<T> {
  data: T;
}

// Error
interface ApiErrorResponse {
  error: {
    code: string;
    message: string;
    details?: Record<string, string>;
  };
}
```

### Endpoints Consumed

| Method | Path | Request Body | Response Data | Auth |
|--------|------|-------------|---------------|------|
| POST | `/api/auth/signup` | `SignupRequest` | `User` | Public |
| POST | `/api/auth/login` | `LoginRequest` | `User` | Public |
| POST | `/api/auth/refresh` | — | `User` | Refresh cookie |
| POST | `/api/auth/logout` | — | `{message: string}` | Access cookie |
| GET | `/api/users/me` | — | `User` | Access cookie |
| PUT | `/api/users/me` | `UpdateUserRequest` | `User` | Access cookie |
| POST | `/api/organizations` | `CreateOrgRequest` | `Organization` | Access cookie |
| GET | `/api/organizations` | — | `{organizations: OrgWithRole[]}` | Access cookie |
| GET | `/api/organizations/{orgID}` | — | `Organization` | Access cookie + member |
| PUT | `/api/organizations/{orgID}` | `CreateOrgRequest` | `Organization` | Access cookie + admin |
| GET | `/api/organizations/{orgID}/members` | — | `{members: Member[]}` | Access cookie + member |
| DELETE | `/api/organizations/{orgID}/members/{userID}` | — | — | Access cookie + admin |
| POST | `/api/organizations/{orgID}/invitations` | `CreateInvitationRequest` | `Invitation` | Access cookie + admin |
| GET | `/api/invitations/{token}` | — | `InvitationDetail` | Public |
| POST | `/api/invitations/{token}/accept` | — | `{membership: Member}` | Access cookie |
| GET | `/api/admin/users?page&perPage&search` | — | `{users: User[], total, page, perPage}` | Access cookie + superadmin |
| PUT | `/api/admin/users/{userID}/superadmin` | `ToggleSuperadminRequest` | `User` | Access cookie + superadmin |

---

## 4. Component & Module Design

### New Files

| File Path | Type | shadcn Components Used | Props Interface |
|-----------|------|----------------------|----------------|
| `frontend/src/lib/api.ts` | Utility | — | — |
| `frontend/src/types/index.ts` | Types | — | — |
| `frontend/src/contexts/AuthContext.tsx` | Context provider | — | `AuthContextValue` |
| `frontend/src/contexts/OrgContext.tsx` | Context provider | — | `OrgContextValue` |
| `frontend/src/hooks/useAuth.ts` | Hook | — | — |
| `frontend/src/hooks/useOrg.ts` | Hook | — | — |
| `frontend/src/components/ProtectedRoute.tsx` | Reusable | — | `{ requireSuperadmin?: boolean }` |
| `frontend/src/components/Layout.tsx` | Page shell | DropdownMenu, Avatar, Separator | — |
| `frontend/src/components/OrgSwitcher.tsx` | Feature | DropdownMenu, Button | — |
| `frontend/src/components/UserMenu.tsx` | Feature | DropdownMenu, Avatar, Button | — |
| `frontend/src/pages/Dashboard.tsx` | Page | Card, Button | — |
| `frontend/src/pages/organizations/OrgList.tsx` | Page | Card, Button, Badge | — |
| `frontend/src/pages/organizations/OrgCreate.tsx` | Page | Card, Input, Label, Button | — |
| `frontend/src/pages/organizations/OrgSettings.tsx` | Page | Card, Input, Label, Button | — |
| `frontend/src/pages/organizations/OrgMembers.tsx` | Page | Table, Dialog, Input, Label, Select, Button, Badge | — |
| `frontend/src/pages/invitations/InviteAccept.tsx` | Page | Card, Button | — |
| `frontend/src/pages/settings/Profile.tsx` | Page | Card, Input, Label, Button | — |
| `frontend/src/pages/admin/UserList.tsx` | Page | Table, Badge, Button | — |

### Modified Files

| File Path | Change Description |
|-----------|-------------------|
| `frontend/src/App.tsx` | Wrap with `AuthProvider` and `OrgProvider`, add all new routes with `ProtectedRoute`, add `Toaster` from sonner |
| `frontend/src/pages/Home.tsx` | Redirect to `/dashboard` if user is authenticated |
| `frontend/src/pages/Login.tsx` | Replace placeholder with functional login form |
| `frontend/src/pages/Signup.tsx` | Replace placeholder with functional signup form |

### shadcn/ui Components to Install

```bash
npx shadcn@latest add input label card dialog dropdown-menu avatar badge table select sonner separator skeleton
```

### API Client (`frontend/src/lib/api.ts`)

Thin fetch wrapper with the following behavior:

- Base URL from `VITE_API_URL` env var (default `http://localhost:8080`)
- All requests include `credentials: "include"` for cookie auth
- All requests with a body include `Content-Type: application/json` header (required — the backend rejects `POST/PUT/PATCH/DELETE` without it as a CSRF mitigation, returning 415)
- On 401 response (not on `/auth/refresh`), automatically attempt `POST /api/auth/refresh`, then retry original request once
- If refresh also fails, redirect to `/login`
- Methods: `get<T>(path)`, `post<T>(path, data)`, `put<T>(path, data)`, `del<T>(path)`

```typescript
// Conceptual interface — not full implementation
const api = {
  get<T>(path: string): Promise<T>;
  post<T>(path: string, data?: unknown): Promise<T>;
  put<T>(path: string, data?: unknown): Promise<T>;
  del<T>(path: string): Promise<T>;
};
```

### Auth Context (`frontend/src/contexts/AuthContext.tsx`)

Provides:
- `user: User | null` — current authenticated user
- `loading: boolean` — true during initial auth check on mount
- `login(email: string, password: string): Promise<void>` — calls `POST /api/auth/login`, sets user state
- `signup(email: string, password: string, firstName: string, lastName: string): Promise<void>` — calls `POST /api/auth/signup`, sets user state
- `logout(): Promise<void>` — calls `POST /api/auth/logout`, clears user state
- `refreshUser(): Promise<void>` — calls `GET /api/users/me`, updates user state

**Mount behavior:** Calls `GET /api/users/me`. If successful, user is authenticated (cookie was valid). If 401, the API client tries refresh. If that also fails, `user` stays null and `loading` becomes false.

```typescript
interface AuthContextValue {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  signup: (email: string, password: string, firstName: string, lastName: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}
```

### Org Context (`frontend/src/contexts/OrgContext.tsx`)

Provides:
- `organizations: OrgWithRole[]` — user's orgs with roles
- `currentOrg: OrgWithRole | null` — selected org
- `setCurrentOrg(org: OrgWithRole): void` — switch org, persist ID to localStorage
- `refreshOrgs(): Promise<void>` — re-fetch orgs list from API

**Mount behavior:** When user is authenticated (from AuthContext), calls `GET /api/organizations`. Restores selected org from localStorage by ID, defaulting to first org if stored ID is not found.

```typescript
interface OrgContextValue {
  organizations: OrgWithRole[];
  currentOrg: OrgWithRole | null;
  setCurrentOrg: (org: OrgWithRole) => void;
  refreshOrgs: () => Promise<void>;
}
```

### ProtectedRoute (`frontend/src/components/ProtectedRoute.tsx`)

Uses React Router's `<Outlet>` pattern:
- If `loading` (from AuthContext): show skeleton loading state
- If no `user`: redirect to `/login`, preserving intended destination in React Router `state`
- If `requireSuperadmin` prop is true and user is not superadmin: show 403 forbidden page
- Otherwise: render `<Outlet />`

### Layout (`frontend/src/components/Layout.tsx`)

Top navigation bar with:
- Left: Agenteur logo/name, `OrgSwitcher` component
- Right: "Admin" nav link (visible only to superadmins), `UserMenu` component
- Below nav: `<Outlet />` for page content

### OrgSwitcher (`frontend/src/components/OrgSwitcher.tsx`)

Dropdown menu showing:
- Current org name as trigger button
- List of user's orgs (clicking switches `currentOrg` via OrgContext)
- Separator
- "Create organization" link to `/organizations/new`

### UserMenu (`frontend/src/components/UserMenu.tsx`)

Dropdown menu showing:
- Avatar with user initials as trigger
- User's name and email (non-clickable)
- Link to `/settings` (Profile)
- Separator
- Logout button (calls `logout()` from AuthContext)

### Error Taxonomy

| HTTP Status | Error Code | Frontend Handling | User-Facing Message |
|-------------|------------|------------------|---------------------|
| 401 | `UNAUTHORIZED` (on login) | Inline form error | "Invalid email or password" |
| 409 | `CONFLICT` (on signup) | Inline form error | "An account with this email already exists" |
| 422 | `VALIDATION_ERROR` | Per-field form errors | Field-specific messages from `details` object |
| 401 | `UNAUTHORIZED` (on API calls) | Auto-refresh via API client, then redirect to `/login` if refresh fails | — |
| 403 | `FORBIDDEN` | Toast notification | "You don't have permission to perform this action" |
| 404 | `NOT_FOUND` | Toast notification | "Resource not found" |
| 409 | `CONFLICT` (on invite/membership) | Toast notification | Context-specific (invite exists, already member, etc.) |
| 400 | `VALIDATION_ERROR` (last admin) | Toast notification | "Cannot remove the last admin from the organization" |

---

## 5. Implementation Tasks

### Phase 5: Frontend Auth

- **T-016:** Install shadcn components and create API client + types
  - _Implements:_ ER-011
  - _Description:_ Install all needed shadcn components: `input`, `label`, `card`, `dialog`, `dropdown-menu`, `avatar`, `badge`, `table`, `select`, `sonner`, `separator`, `skeleton`. Create TypeScript types in `frontend/src/types/index.ts` matching Go response structs (all camelCase). Create API client in `frontend/src/lib/api.ts` with fetch wrapper: base URL from `VITE_API_URL`, `credentials: "include"`, auto-refresh on 401, redirect to `/login` on auth failure. Create `frontend/.env.local` with `VITE_API_URL=http://localhost:8080`.
  - _Files:_
    - Create: `frontend/src/types/index.ts`
    - Create: `frontend/src/lib/api.ts`
  - _Done condition:_ `npm run build` succeeds (TypeScript compilation passes). shadcn components appear in `frontend/src/components/ui/`.
  - _Dependencies:_ None (frontend tasks can start after backend Phase 2 is done for auth testing)

- **T-017:** Create AuthContext and ProtectedRoute
  - _Implements:_ ER-011
  - _Description:_ Create `AuthContext` with `AuthProvider` that: on mount calls `GET /api/users/me` to check auth state, provides `user`, `loading`, `login()`, `signup()`, `logout()`, `refreshUser()`. Create `useAuth` hook that calls `useContext(AuthContext)` and throws if used outside provider. Create `ProtectedRoute` component using React Router's `<Outlet>` pattern: if `loading`, show skeleton. If no `user`, redirect to `/login` (preserving intended destination in `state`). If `requireSuperadmin` prop and user is not superadmin, show 403 page.
  - _Files:_
    - Create: `frontend/src/contexts/AuthContext.tsx`
    - Create: `frontend/src/hooks/useAuth.ts`
    - Create: `frontend/src/components/ProtectedRoute.tsx`
  - _Done condition:_ `npm run build` succeeds.
  - _Dependencies:_ T-016

- **T-018:** Build Login and Signup pages
  - _Implements:_ ER-011
  - _Description:_ Replace Login.tsx placeholder with form: email + password inputs, submit button, error display (inline for validation, toast for server errors), link to signup. On success, redirect to `/dashboard` (or to the preserved location from ProtectedRoute). Replace Signup.tsx placeholder with form: email, password, first name, last name inputs, same error handling, link to login. Both use shadcn Card, Input, Label, Button components. If an `invite` query param is present, preserve it through the auth flow and redirect to `/invitations/{token}` after auth.
  - _Files:_
    - Modify: `frontend/src/pages/Login.tsx`
    - Modify: `frontend/src/pages/Signup.tsx`
  - _Done condition:_ `npm run build` succeeds. Visual: login form renders with fields and submit button. Signup form renders with all fields.
  - _Dependencies:_ T-017

- **T-019:** Create Layout shell and wire App.tsx
  - _Implements:_ ER-011
  - _Description:_ Create `Layout` component with: top navigation bar showing Agenteur logo/name on left, user menu on right. Create `UserMenu` dropdown with user's name/email, link to `/settings`, logout button. Create `Dashboard` page (placeholder content — "Welcome to Agenteur" + prompt to create org if no orgs). Update `App.tsx`: wrap with `BrowserRouter`, `AuthProvider`. Add routes — public (`/`, `/login`, `/signup`, `/invitations/:token`), protected with Layout (`/dashboard`, `/settings`). Update `Home.tsx` to redirect to `/dashboard` if authenticated. Add `Toaster` from sonner for toast notifications.
  - _Files:_
    - Create: `frontend/src/components/Layout.tsx`
    - Create: `frontend/src/components/UserMenu.tsx`
    - Create: `frontend/src/pages/Dashboard.tsx`
    - Modify: `frontend/src/App.tsx`
    - Modify: `frontend/src/pages/Home.tsx`
  - _Done condition:_ `npm run build` succeeds. Visual: signup → redirected to dashboard with nav bar. User menu shows name and logout works. Unauthenticated visit to `/dashboard` redirects to `/login`.
  - _Dependencies:_ T-017, T-018

### Phase 6: Frontend Organization Management

- **T-020:** Create OrgContext and OrgSwitcher
  - _Implements:_ ER-012
  - _Description:_ Create `OrgContext` with `OrgProvider`: fetches user's orgs on mount (when user is authenticated), provides `organizations`, `currentOrg`, `setCurrentOrg()`, `refreshOrgs()`. Persists selected org ID in localStorage. Create `useOrg` hook. Create `OrgSwitcher` dropdown component showing current org name, list of user's orgs, divider, "Create organization" link to `/organizations/new`. Add `OrgProvider` inside `AuthProvider` in `App.tsx`. Add `OrgSwitcher` to `Layout` nav bar.
  - _Files:_
    - Create: `frontend/src/contexts/OrgContext.tsx`
    - Create: `frontend/src/hooks/useOrg.ts`
    - Create: `frontend/src/components/OrgSwitcher.tsx`
    - Modify: `frontend/src/App.tsx` — add OrgProvider
    - Modify: `frontend/src/components/Layout.tsx` — add OrgSwitcher to nav
  - _Done condition:_ `npm run build` succeeds. Visual: org switcher appears in nav, shows user's orgs, switching persists across page reload.
  - _Dependencies:_ T-019

- **T-021:** Build organization CRUD pages
  - _Implements:_ ER-012
  - _Description:_ Create `OrgList` page: card grid showing user's orgs with name, slug, role badge, link to org details. Create `OrgCreate` page: form with name input, submit creates org and redirects to org detail. Create `OrgSettings` page: form to edit org name (admin only), danger zone for org management. Add routes: `/organizations`, `/organizations/new`, `/organizations/:id`.
  - _Files:_
    - Create: `frontend/src/pages/organizations/OrgList.tsx`
    - Create: `frontend/src/pages/organizations/OrgCreate.tsx`
    - Create: `frontend/src/pages/organizations/OrgSettings.tsx`
    - Modify: `frontend/src/App.tsx` — add org routes
  - _Done condition:_ `npm run build` succeeds. Visual: can create org, see it in list, navigate to settings, edit name.
  - _Dependencies:_ T-020

- **T-022:** Build members page and invitation flow
  - _Implements:_ ER-012
  - _Description:_ Create `OrgMembers` page: table showing members (name, email, role badge, remove button for admins). Invite dialog: form with email + role select, calls `POST /api/organizations/{orgID}/invitations`. Create `InviteAccept` page: fetches invite details from `GET /api/invitations/{token}`, shows org name + inviter + role. If user authenticated: "Accept invitation" button calls `POST /api/invitations/{token}/accept`, redirects to org dashboard. If not authenticated: shows "Sign up to accept" and "Log in to accept" buttons that link to `/signup?invite={token}` and `/login?invite={token}`. Add routes: `/organizations/:id/members`, `/invitations/:token`.
  - _Files:_
    - Create: `frontend/src/pages/organizations/OrgMembers.tsx`
    - Create: `frontend/src/pages/invitations/InviteAccept.tsx`
    - Modify: `frontend/src/App.tsx` — add members and invitation routes
  - _Done condition:_ `npm run build` succeeds. Visual: members table shows org members. Invite dialog submits successfully (check backend console for URL). Invite acceptance page renders for both auth and unauth states.
  - _Dependencies:_ T-021

- **T-023:** Build profile settings and superadmin pages
  - _Implements:_ ER-012
  - _Description:_ Create `Profile` page: form to edit first name + last name, calls `PUT /api/users/me`. Create `UserList` admin page: table of all users with email, name, superadmin badge, toggle button. Paginated with search. Only accessible to superadmins (wrapped in `ProtectedRoute` with `requireSuperadmin`). Add routes: `/settings`, `/admin/users`. Update `Layout` to show "Admin" nav link only for superadmin users.
  - _Files:_
    - Create: `frontend/src/pages/settings/Profile.tsx`
    - Create: `frontend/src/pages/admin/UserList.tsx`
    - Modify: `frontend/src/App.tsx` — add settings + admin routes
    - Modify: `frontend/src/components/Layout.tsx` — conditional admin link
  - _Done condition:_ `npm run build` succeeds. Visual: profile page updates user name. Admin page shows all users (superadmin only). Non-superadmin cannot access `/admin/users`.
  - _Dependencies:_ T-019, T-022

---

## 6. Testing Strategy

### Frontend Unit Tests (React/TypeScript)

| Component / Hook | Key Cases | Test Command |
|-----------------|-----------|--------------|
| `AuthContext` | Provides user after login, clears user after logout, handles loading state on mount | `npm run test -- src/contexts/AuthContext.test.tsx` |
| `ProtectedRoute` | Redirects when unauthenticated, renders children when authenticated, blocks non-superadmin from superadmin routes | `npm run test -- src/components/ProtectedRoute.test.tsx` |
| `api.ts` | Retries on 401 with refresh, redirects on failed refresh, includes credentials on all requests | `npm run test -- src/lib/api.test.ts` |
| `OrgContext` | Loads orgs on mount, persists selection to localStorage, restores selection on reload | `npm run test -- src/contexts/OrgContext.test.tsx` |

### Manual Testing Checklist

- [ ] Full signup → create org → invite user → accept invite flow across two browser sessions
- [ ] Refresh token rotation: set short AccessTokenTTL (e.g., 30s), verify API calls still work transparently after expiry
- [ ] Logout and verify navigating to `/dashboard` redirects to `/login`
- [ ] Invite URL works in incognito (shows signup/login prompt with invite query param)
- [ ] After signup/login with invite query param, user is redirected to invitation acceptance page
- [ ] Org switcher persists selection across page reload
- [ ] Non-superadmin cannot access `/admin/users` (shows 403 or redirects)
- [ ] Superadmin sees "Admin" link in nav and can access `/admin/users`
- [ ] Cannot remove the last admin from an org (toast error appears)
- [ ] CORS works correctly between frontend (`:5173`) and backend (`:8080`) with credentials
- [ ] All forms show inline validation errors for 422 responses
- [ ] Toast notifications appear for 403, 404, and 409 error responses

---

## 7. Rollout & Operations

### Feature Flags

None — this is foundational infrastructure, not an incremental feature behind a flag.

### Monitoring & Alerts

Not applicable for initial local development. To be defined when deploying to dev/production.

### Backward Compatibility

No backward compatibility concerns — this is the first user-facing frontend feature. All new pages and components.

### Rollback Procedure

1. Deploy previous version of the frontend code
2. No data rollback needed — frontend has no persistent state beyond localStorage (which is non-critical)

---

## 8. Open Questions & Risks

### Technical Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Backend API not ready when frontend work starts | Frontend cannot be tested end-to-end | Frontend tasks can be built and type-checked independently (`npm run build`). Manual testing requires backend to be running. Structure tasks so API client and types can be built first. |
| CORS misconfiguration | API calls fail silently or with opaque errors | Backend CORS middleware must include `Access-Control-Allow-Credentials: true` and the exact frontend origin (not `*`). Test early with `curl` and browser dev tools. |
| Cookie not sent cross-origin in dev | Auth fails because browser blocks cookies | Ensure `SameSite=Lax` (not `Strict`), and that frontend and backend share the same domain or are correctly configured for cross-origin. Vite proxy can be used as fallback. |

### Open Questions

| Question | Owner | Target Resolution |
|----------|-------|-------------------|
| Should we add a Vite proxy to avoid CORS in development? | Engineering | During T-016 implementation |
| What loading/skeleton patterns should be standardized? | Design | During T-017 implementation |

---

## 9. Assumptions

- [ASSUMPTION] The backend API endpoints defined in the [backend spec](./user-org-management-backend.md) are available and return the documented response shapes
- [ASSUMPTION] shadcn/ui components install cleanly with the existing Tailwind CSS v4 + Vite setup
- [ASSUMPTION] React Router 7 is already configured in the project (based on existing `frontend/src/App.tsx`)
- [ASSUMPTION] `VITE_API_URL` environment variable is sufficient for configuring the API base URL across environments
- [ASSUMPTION] localStorage is acceptable for persisting non-sensitive UI state (selected org ID)

---

_Traceability verified: Every frontend engineering requirement (ER-011, ER-012) maps to at least one task. Every task (T-016 through T-023) maps to at least one requirement. See Section 1 requirement-to-task mapping table._
