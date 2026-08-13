# Roadmap

This scaffold implements the core volunteer lifecycle end-to-end:
**signup → event discovery → service log draft/submit → supervisor
verification → VSR generation/export**, including RBAC, the state
machine, conflict-of-interest checks, and audit trails for that path.

Everything below is *modeled* (tables exist in `backend/internal/models`)
but not yet wired up with handlers/routes/UI. Rough build order,
grouped by the original spec sections:

## 1. Organizations & Supervisors (spec: "Event Coordinator", "Supervisor")
- [ ] Organization registration + admin-approval workflow (`OrganizationStatus`)
- [ ] Org profile editing with locked-field + audit-log enforcement
- [ ] `OrgMembership` management endpoints (assign/revoke/expire sub-roles)
- [ ] Supervisor dashboard (pending queue exists via `/verification/queue`;
      needs recently-verified / rejected / flagged views added)
- [ ] Suspension cascade: disable event creation, freeze verifications,
      read-only mode (models support this; enforcement TODO in handlers)

## 2. Conflict Dashboard (spec: "Conflict Dashboard")
- [ ] Case CRUD on top of `models.Case` / `CaseNote` / `CaseResolution`
- [ ] Case queue filtering/sorting (severity, status, org, overdue)
- [ ] Resolution action handlers (uphold / reverse / require re-verification /
      freeze VSR / restrict access / escalate) — each must re-run
      `recomputeVSR` when it touches verified hours

## 3. Content & Policy Management (spec: "Content & Policy Management Dashboard")
- [ ] Policy editor + versioning endpoints on `models.Policy` / `PolicyVersion`
- [ ] Scheduled activation / expiration jobs (good fit for a Redis-backed
      queue given Redis is already in the stack)
- [ ] Acknowledgment tracking + login-gate middleware
      (`PolicyAcknowledgment`)

## 4. Admin Oversight Dashboard (spec: "Admin Data & Oversight Dashboard")
- [ ] System-wide metrics endpoints (registered/active volunteers, hours
      logged vs. verified, VSRs issued, backlog size)
- [ ] Trend/anomaly detection (spike/drop flags) — likely a scheduled
      job writing pre-aggregated rows rather than computing live
- [ ] Verification integrity indicators (turnaround time, approval
      ratios, clustering/outlier detection per supervisor)

## 5. Search & Recommendations (spec: "Recommended Events")
- [ ] Index published events into Meilisearch (already in docker-compose)
- [ ] Personalized "Suggested for You" ranking against volunteer
      skills/interests/availability

## 6. Notifications
- [ ] Resend integration for email (verification results, reminders,
      policy updates) — TODOs are already marked at each trigger point
      in `handlers/servicelog.go` and `handlers/verification.go`
- [ ] Twilio integration for SMS reminders (24h before event)
- [ ] In-dashboard notification center + preferences

## 7. Auth hardening
- [ ] Real Auth0 JWT verification against JWKS in
      `middleware/auth.go` (currently a dev-only header stand-in —
      see the TODO comment there)
- [ ] `@auth0/nextjs-auth0` session wiring in `frontend/src/lib/api.ts`
      (`getAccessToken()` TODO)

## 8. File handling
- [ ] Azure Blob Storage upload for service-log evidence photos/docs
- [ ] PDF/CSV rendering for VSR export (`handlers/vsr.go` has the
      authorization + audit-log contract already; only the render step
      is missing)

## 9. Analytics
- [ ] PostHog event instrumentation across key actions (signup, log
      submitted, verification decided, VSR exported)

## 10. Observability / ops
- [ ] Structured logging + request tracing
- [ ] Rate limiting via Redis (client is already connected in `main.go`)
- [ ] Cloudflare in front of both services (infra-level, no app code)
