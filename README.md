# YVEN — Youth Volunteer Engagement Network

A platform for logging, verifying, and issuing official records (VSRs) of
youth volunteer service hours, built for four audiences: volunteers,
organization supervisors/admins, school admins, and platform admins.

## Stack

| Layer | Choice |
|---|---|
| Frontend | Next.js (TypeScript) |
| Backend | Go + Gin |
| ORM | GORM |
| Database | PostgreSQL |
| Cache/Queue | Redis |
| Auth | Auth0 |
| File storage | Azure Blob Storage |
| Search | Meilisearch |
| Email | Resend |
| SMS | Twilio |
| Analytics | PostHog |
| Cloud | Microsoft Azure |
| Containers | Docker |
| CI/CD | GitHub Actions |
| Edge/security | Cloudflare |

## What's built vs. what's scaffolded

This repo implements the **core volunteer lifecycle end-to-end** —
registration, event discovery, service-log submission with fraud
checks, supervisor verification with conflict-of-interest controls, and
automatic VSR generation/export — because that pipeline is the spine
the rest of the spec hangs off.

Every other module from the spec (org onboarding, the Conflict
Dashboard, policy versioning, the admin oversight dashboard, search,
notifications) has its data model already defined in
`backend/internal/models/` so the schema is settled, but the
handlers/routes/UI are TODOs. See **[docs/ROADMAP.md](docs/ROADMAP.md)**
for the full breakdown and suggested build order.

## Repo layout

```
yven/
├── backend/                 # Go + Gin API
│   ├── cmd/server/          # main.go entrypoint
│   └── internal/
│       ├── config/          # env-based config loader
│       ├── db/               # GORM connection + AutoMigrate
│       ├── models/           # all DB models (full spec schema)
│       ├── middleware/       # Auth0 JWT auth, RBAC
│       ├── handlers/         # request handlers
│       └── routes/           # route -> handler + middleware wiring
├── frontend/                 # Next.js (TypeScript) app
│   └── src/
│       ├── app/               # pages (home, signup, dashboard, events)
│       ├── lib/api.ts         # typed fetch wrapper for the backend
│       └── types/             # TS types mirroring backend models
├── docs/
│   └── ROADMAP.md            # what's left, mapped to spec sections
├── docker-compose.yml         # postgres + redis + meilisearch + both apps
├── .env.example
└── .github/workflows/ci.yml   # build/lint/test on push + PR
```

## Local development

**Prerequisites:** Docker + Docker Compose (simplest path), or Go 1.22+
and Node 20+ if you'd rather run services natively.

```bash
cp .env.example .env
# fill in Auth0 / Azure / Resend / Twilio / PostHog values as you get them —
# the app boots without them, those integrations just no-op until configured

docker compose up --build
```

This starts:
- Postgres on `:5432`
- Redis on `:6379`
- Meilisearch on `:7700`
- Backend API on `:8080` (health check: `GET /healthz`)
- Frontend on `:3000`

Tables are created automatically on backend startup via GORM
`AutoMigrate` — no separate migration step needed for local dev (see
the note in `backend/internal/db/db.go` about switching to versioned
migrations before production).

### Running without Docker

```bash
# backend
cd backend
go mod tidy
go run ./cmd/server

# frontend, in a second terminal
cd frontend
npm install
npm run dev
```

## Authentication note

The Auth0 JWT verification in `backend/internal/middleware/auth.go` is
currently a **scaffold** — it reads an identity from a dev-only header
rather than verifying a real signed token against Auth0's JWKS
endpoint. This is marked with `TODO` comments at the exact lines that
need the real verification code before this goes anywhere near
production. Same on the frontend side in `frontend/src/lib/api.ts`
(`getAccessToken()`).

## Deploying

Both `backend/Dockerfile` and `frontend/Dockerfile` are multi-stage and
production-ready as images; `.github/workflows/ci.yml` builds them on
every push to `main`. Wiring those images to an actual Azure target
(Container Apps, AKS, or App Service) plus Cloudflare in front is
infrastructure work not included here — happy to scaffold Azure
deployment config (Bicep/Terraform) as a next step if useful.

## License

Add a LICENSE file appropriate to your organization before making this
repository public.
