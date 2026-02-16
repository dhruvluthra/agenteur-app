---
name: create-prd
description: Create high-quality Product Requirements Documents (PRDs) for new features and product changes. Use when the user asks for a PRD, product spec, requirements document, feature definition, or discovery-to-scope writeup before technical implementation.
---

# Create PRD

## Purpose

Produce a clear, decision-ready PRD that aligns product, design, and engineering before implementation starts.

The PRD must:
1. Define the user/problem and business outcome.
2. Specify scope, requirements, and acceptance criteria.
3. Call out constraints, risks, and open questions.
4. Be actionable enough to hand off into a technical specification.
5. Be created in the `/prds` directory.

## When to apply this skill

Apply this skill when the user asks to:
- write a PRD
- define requirements for a feature
- scope a new product initiative
- clarify product direction before engineering work
- produce a product spec prior to a tech spec

Do not apply this skill when the user is asking for:
- a technical implementation plan only (use tech-spec workflow)
- architecture-level design without product context
- direct code changes

## Operating principles

- Be concise and specific; avoid generic product language.
- Prefer explicit decisions over ambiguous prose.
- Separate facts, assumptions, and open questions.
- Use measurable outcomes and testable acceptance criteria.
- Express all functional requirements as user stories — never as loose bullets or prose descriptions.
- State out-of-scope items to prevent scope creep.
- Bias toward drafting with flagged assumptions rather than blocking on missing info. Only ask clarifying questions when the gap would lead to a fundamentally wrong direction.

## User story format

All functional requirements in the PRD must be expressed as user stories. This is the primary unit of scope — never describe functionality as loose bullet points or prose paragraphs.

### Structure

Every user story follows this format:

```
As a [persona], I [must/should/may] be able to [action] so that [outcome].
```

Each story must include:
- **ID**: sequential identifier (US-001, US-002, ...) for traceability into tech specs and tickets.
- **Persona**: a specific user role, not "the user." Use the personas defined in the Problem Statement.
- **Action**: what the user does, described as behavior — not implementation.
- **Outcome**: the value delivered. Answers "why does this matter to the user?"
- **Priority**: must / should / may (see definitions in Section 4).
- **Acceptance criteria**: a set of testable conditions that confirm the story is complete. Written as "Given / When / Then" or as a concise checklist.

### Writing good acceptance criteria

Acceptance criteria define the boundary of "done." They must be verifiable by QA without needing to ask a product manager for interpretation.

**Use Given / When / Then for interaction-driven criteria:**

> **US-003:** As a workspace admin, I must be able to invite users by email so that I can onboard my team without filing a support ticket.
>
> _Acceptance criteria:_
> - Given I am a workspace admin on the Members page, when I enter a valid email and click "Send Invite," then the invitee receives an email within 60 seconds and appears in the member list with status "Pending."
> - Given I enter an email that already belongs to a workspace member, when I click "Send Invite," then the system displays an inline error: "This user is already a member" and no duplicate invite is sent.
> - Given the email delivery service is unavailable, when I click "Send Invite," then the invite is queued for retry and I see a notice: "Invite will be delivered shortly."

**Use a checklist for state-based or data criteria:**

> **US-007:** As an end user, I must see a notification badge on the sidebar when I have unread messages so that I don't miss new activity.
>
> _Acceptance criteria:_
> - Badge displays the count of unread messages across all channels.
> - Badge updates within 2 seconds of a new message arriving (WebSocket push, not polling).
> - Badge clears when the user opens the channel containing the unread messages.
> - If count exceeds 99, badge displays "99+".

### Decomposing features into stories

A single feature often maps to multiple user stories. Break features down until each story represents a single user-facing behavior that can be independently tested and demoed.

**Signs a story is too large (an epic in disguise):**
- It contains the word "and" joining two distinct behaviors.
- It would take more than one sprint to implement.
- QA would need multiple test scenarios with unrelated setups.

**Signs a story is too small:**
- It has no independent user value (e.g., "As a user, I can see a loading spinner").
- It's an implementation task masquerading as a story (e.g., "As a developer, I must create a database migration").

When a story is too large, split it along user-facing boundaries:

> **Too large:** "As an admin, I must be able to manage team members so that I can control workspace access."
>
> **Split into:**
> - US-010: As an admin, I must be able to invite new members by email so that I can grow my team.
> - US-011: As an admin, I must be able to remove a member so that I can revoke access when someone leaves.
> - US-012: As an admin, I must be able to change a member's role so that I can delegate permissions appropriately.

### Stories for edge cases and error states

Don't limit stories to happy paths. Important error states, empty states, and boundary conditions deserve their own stories when the expected behavior requires a product decision:

> **US-015:** As an end user, when I attempt to send a message while offline, I must see my message marked as "pending" in the conversation so that I know it hasn't been delivered yet.
>
> _Acceptance criteria:_
> - Message appears in the conversation thread with a "pending" indicator immediately.
> - When connectivity is restored, the message is sent automatically and the indicator updates to "sent" within 5 seconds.
> - If delivery fails after 3 retries, the indicator changes to "failed" with a "Retry" action.

---

## Required workflow

Follow this sequence every time.

### 1) Discovery and context gathering

Collect or infer the following. If the user provides a brief, draft with reasonable assumptions and **flag each assumption with `[ASSUMPTION]`** inline so the user can correct them.

**Must have before drafting (ask if missing):**
1. Who is the target user and what specific pain or job-to-be-done are we solving?
2. What business outcome matters most, and how will we measure success?
3. What is explicitly out of scope for this iteration?

**Ask only if answers to the above reveal ambiguity:**
4. What timeline or launch constraint is non-negotiable?
5. Are there compliance, security, or platform constraints?
6. What dependencies or teams could block delivery?
7. What existing behavior must remain unchanged?

Do not ask more than 5 questions before producing a first draft. Remaining gaps should be surfaced as open questions inside the PRD itself.

### 2) Problem framing

Document:
- **Problem statement**: one or two sentences, specific and falsifiable.
- **Why now**: what changed (market, user feedback, technical unlock, business priority) that makes this urgent.
- **Who is affected**: user segment, estimated reach, and severity of impact.
- **Cost of inaction**: what happens if we ship nothing in the next quarter.

### 3) Solution framing

Define:
- **Proposed solution overview**: describe the experience change from the user's perspective, not the implementation.
- **Alternatives considered**: list at least two alternatives. For each, include one sentence on why it was rejected (cost, complexity, user impact, timeline, etc.).
- **Rationale for chosen direction**: tie back to the problem statement and success metrics.
- **Key user flows**: describe the primary happy path and the most important error/edge case.

### 4) Scope and requirements

**In-scope features** — express every functional requirement as a user story following the format defined in the "User story format" section above. Never describe functionality as loose bullets or prose.

For each story, provide:
- A unique ID (US-001, US-002, ...)
- The story in "As a [persona], I [must/should/may]..." format
- Acceptance criteria using Given/When/Then or a testable checklist
- Priority classification (must/should/may)

Decompose large features into multiple stories. Include stories for important error states and edge cases where the expected behavior requires a product decision (see "Stories for edge cases and error states" above).

Classify each story:
- **must** = required for MVP; launch is blocked without it.
- **should** = important, expected soon after launch, but not blocking.
- **may** = desirable for a future iteration; explicitly deferred.

**Out-of-scope boundaries** — list specific features or behaviors that will NOT be addressed. Be concrete:

> **Good:** "Bulk CSV import of users is out of scope for V1."
>
> **Bad:** "Advanced features are out of scope." ← too vague to prevent scope creep.

**Non-functional requirements** (include where relevant):
- Performance: target latencies, throughput (e.g., "invite API responds in < 200ms p95").
- Security: auth model, data access controls, audit logging.
- Reliability: uptime target, degradation behavior.
- Accessibility: WCAG level, keyboard nav, screen reader support.
- Privacy: data retention, PII handling, consent flows.

### 5) Validation and rollout planning

- **Acceptance criteria**: restate the top-level criteria that confirm the feature is shippable. These must be testable by QA without product interpretation.
- **Instrumentation/analytics plan**: list the specific events or metrics to track, and where they feed (e.g., Mixpanel, Datadog, internal dashboard).
- **Rollout strategy**: phases, feature flags, percentage ramps, migration steps, and rollback/fallback plan.
- **Risks and mitigations**: each risk must be specific and paired with a concrete mitigation or contingency.

> **Good:** "Risk: email delivery latency from third-party provider exceeds 60s SLA during peak. Mitigation: implement retry queue with exponential backoff; alert if p95 > 45s."
>
> **Bad:** "Risk: things might be slow. Mitigation: monitor performance." ← not actionable.

- **Open questions**: each must have a designated owner and a target resolution date.

### 6) Self-evaluation and revision

Before presenting the final PRD, evaluate it against every item below. Revise any section that fails. Only present the PRD once all items pass.

- [ ] Problem and target users are unambiguous — someone outside the team could restate them.
- [ ] Scope and non-goals are explicit — each out-of-scope item names a specific feature or behavior.
- [ ] Every functional requirement is expressed as a user story with a unique ID, persona, action, outcome, and priority.
- [ ] Every user story has acceptance criteria written as Given/When/Then or a testable checklist.
- [ ] No story is an epic in disguise — each represents a single user-facing behavior that can be independently tested.
- [ ] Important error states and edge cases have their own stories, not just happy paths.
- [ ] Story personas match the users defined in the Problem Statement — no generic "the user" references.
- [ ] Non-functional requirements include numeric targets where applicable.
- [ ] Success metrics include targets and timeframes (not just metric names).
- [ ] Every risk is specific and paired with a mitigation that could actually be executed.
- [ ] Open questions have owners and resolution dates.
- [ ] No vague terms remain ("fast", "easy", "better", "seamless", "intuitive") — all replaced with measurable definitions.
- [ ] The document contains enough detail to begin writing a technical spec without further product clarification.

---

## PRD output template

Use exactly this structure and headings when producing the final document:

```markdown
# PRD: [Feature Name]

**Author:** [name]
**Last updated:** [date]
**Status:** Draft | In Review | Approved
**Stakeholders:** [list of people/roles who must review]

---

## 1. Problem Statement

[One to two sentences. Specific and falsifiable.]

### Why now

[What changed that makes this the right time.]

### Who is affected

[User segment, estimated reach, severity.]

### Cost of inaction

[What happens if we do nothing this quarter.]

---

## 2. Success Metrics

| Metric | Current Baseline | Target | Timeframe |
|--------|-----------------|--------|-----------|
| [e.g., Onboarding completion rate] | [e.g., 34%] | [e.g., 60%] | [e.g., 90 days post-launch] |

---

## 3. Proposed Solution

### Overview

[Describe the experience change from the user's perspective.]

### Key User Flows

**Happy path:**
1. [Step]
2. [Step]
3. [Step]

**Primary error/edge case:**
1. [Step]
2. [Step]

### Alternatives Considered

| Alternative | Why rejected |
|-------------|-------------|
| [Option A] | [One sentence reason] |
| [Option B] | [One sentence reason] |

---

## 4. Scope

### In Scope (Must Have)

- **US-001:** As a [persona], I must be able to [action] so that [outcome].
  - _Acceptance criteria:_
    - Given [precondition], when [action], then [expected result].
    - Given [precondition], when [action], then [expected result].

- **US-002:** As a [persona], I must be able to [action] so that [outcome].
  - _Acceptance criteria:_
    - [Testable condition]
    - [Testable condition]

### In Scope (Should Have)

- **US-[number]:** As a [persona], I should be able to [action] so that [outcome].
  - _Acceptance criteria:_
    - Given [precondition], when [action], then [expected result].

### In Scope (May — Future Consideration)

- **US-[number]:** As a [persona], I may [action] so that [outcome]. [Explicitly deferred — rationale.]

### Edge Case & Error State Stories

- **US-[number]:** As a [persona], when [error condition occurs], I must [expected behavior] so that [outcome].
  - _Acceptance criteria:_
    - Given [error precondition], when [trigger], then [graceful handling].

### Out of Scope

- [Specific feature or behavior not addressed in this iteration]
- [Another specific exclusion]

---

## 5. Non-Functional Requirements

| Category | Requirement |
|----------|------------|
| Performance | [e.g., API response < 200ms p95] |
| Security | [e.g., RBAC enforced; actions audit-logged] |
| Reliability | [e.g., 99.9% uptime; graceful degradation if email provider is down] |
| Accessibility | [e.g., WCAG 2.1 AA; full keyboard navigation] |
| Privacy | [e.g., PII encrypted at rest; 30-day retention for invite tokens] |

---

## 6. Instrumentation & Analytics

| Event / Metric | Trigger | Destination |
|----------------|---------|-------------|
| [e.g., invite_sent] | [User clicks Send Invite] | [Mixpanel] |

---

## 7. Rollout Plan

| Phase | Audience | Criteria to advance |
|-------|----------|-------------------|
| 1 | [e.g., Internal dogfood] | [e.g., Zero P0 bugs for 48 hrs] |
| 2 | [e.g., 10% of workspaces] | [e.g., Error rate < 0.1%] |
| 3 | [e.g., GA] | [e.g., Success metrics trending toward target] |

**Rollback plan:** [How to revert if something goes wrong.]

---

## 8. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| [Specific risk] | High/Med/Low | High/Med/Low | [Concrete action] |

---

## 9. Open Questions

| Question | Owner | Target resolution date |
|----------|-------|----------------------|
| [Specific question] | [Name] | [Date] |

---

## 10. Assumptions

[List any assumptions made during drafting. Flag those that need validation.]

- [ASSUMPTION] [description]

---

_Hand-off checklist: All items verified prior to submission. See Self-evaluation (Section 6 of skill) for criteria._
```

---

## Writing standards

- Use short sections and bullet points where possible.
- Avoid implementation details unless needed for a product tradeoff decision. The PRD describes **what** and **why**, not **how**.
  - **Good:** "The system must deliver invite emails within 60 seconds of the admin action."
  - **Bad:** "Use a Redis-backed Sidekiq queue to process invite emails asynchronously." ← this belongs in the tech spec.
- Use must/should/may language consistently (see Section 4).
- User stories describe **user behavior**, not implementation tasks. The persona must be an end user or product role, never a developer or system.
  - **Good:** "As a workspace admin, I must be able to revoke an invite so that I can correct mistakes."
  - **Bad:** "As a developer, I must create an API endpoint for invite revocation." ← this is a task, not a story.
- Replace every vague term with a measurable definition:
  - "fast" → "< 200ms p95"
  - "reliable" → "99.9% uptime measured weekly"
  - "easy to use" → "new user completes flow in < 2 minutes without help text"