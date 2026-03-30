# Project Info

> Updated: 2026-03-29

## Repository
| Item | Value |
|------|-------|
| Upstream | Wei-Shaw/sub2api |
| Fork | bayma888/sub2api-bmai |
| Go module | github.com/Wei-Shaw/sub2api |
| Go version | 1.25.7 |
| Frontend | Vue3 + TypeScript, pkg name: sub2api-frontend |

## Local Environment (Windows 11)
| Item | Value |
|------|-------|
| PostgreSQL | 16, Windows service, port 5432 |
| psql path | `C:\Program Files\PostgreSQL\16\bin\psql.exe` |
| pg_hba.conf | `C:\Program Files\PostgreSQL\16\data\pg_hba.conf` |
| DB creds | user=sub2api, pw=sub2api, db=sub2api |
| PG superuser | user=postgres, pw=postgres |
| Redis | localhost:6379, no password |
| golangci-lint | v2.7 |

## Project Structure
```
sub2api-bmai/
├── backend/
│   ├── cmd/server/          # entry point
│   ├── ent/schema/          # Ent ORM schema definitions
│   ├── internal/
│   │   ├── handler/         # HTTP handlers (admin/, dto/)
│   │   ├── service/         # business logic
│   │   ├── repository/      # data access layer
│   │   └── server/routes/   # route registration
│   └── migrations/          # SQL migrations
├── frontend/
│   ├── src/
│   │   ├── api/             # API calls (admin/)
│   │   ├── components/      # Vue components (common/)
│   │   ├── views/           # pages (admin/, user/)
│   │   ├── types/           # TypeScript types
│   │   └── i18n/            # i18n translations
│   ├── package.json
│   └── pnpm-lock.yaml       # MUST commit this
└── .claude/CLAUDE.md
```

## CI/CD (GitHub Actions)
| Workflow | Trigger | Content |
|----------|---------|---------|
| backend-ci.yml | push, PR | unit test + integration test + golangci-lint v2.7 |
| security-scan.yml | push, PR, weekly | govulncheck + gosec + pnpm audit |
| release.yml | tag v* | build release |

## Common Commands
```bash
# Backend tests
cd backend && go test -tags=unit ./...
cd backend && go test -tags=integration ./...
cd backend && golangci-lint run ./...

# Ent codegen (after schema changes)
cd backend && go generate ./ent

# Frontend
cd frontend && pnpm install
cd frontend && pnpm dev
cd frontend && pnpm build

# Database
psql -U sub2api -h 127.0.0.1 -d sub2api
```
