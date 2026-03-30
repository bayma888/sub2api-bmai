# Pitfalls & Debugging Notes

> Updated: 2026-03-08
> Source: CLAUDE.md + GROUP_SORT_PLAN.md experience

## 1. pnpm-lock.yaml Must Be Committed
- CI uses `pnpm install --frozen-lockfile`
- After `pnpm install` or adding deps, always `git add pnpm-lock.yaml`

## 2. npm vs pnpm node_modules Conflict
- If previously used npm, `pnpm install` gets EPERM errors
- Fix: `rm -rf node_modules` then `pnpm install`

## 3. PowerShell Escapes `$` in Bcrypt Hashes
- `$2a$10$...` gets mangled by PowerShell variable interpolation
- Fix: write SQL to file, use `psql -f file.sql`

## 4. psql Can't Handle Chinese Paths
- `psql -f "D:\中文路径\file.sql"` fails
- Fix: copy to English path first, e.g. `C:\temp.sql`

## 5. PostgreSQL Password Reset
1. Edit pg_hba.conf: `scram-sha-256` → `trust`
2. Restart service: `Restart-Service postgresql-x64-16`
3. Reset passwords via psql
4. Revert pg_hba.conf and restart

## 6. Go Interface → Test Stubs Must Be Updated
- Adding methods to an interface → ALL test stubs/mocks must implement them
- Search: `grep -r "type.*Stub.*struct" internal/` and `grep -r "type.*Mock.*struct" internal/`
- This hit us hard in features #4 and #5 (multiple stub files)

## 7. psql: Use 127.0.0.1 Not localhost
- Windows psql tries IPv6 (::1) first → may fail
- Always use `-h 127.0.0.1`

## 8. Windows Has No make
- Use raw commands: `go test -tags=unit ./...` etc.

## 9. Ent Schema → Must Regenerate
- After modifying `ent/schema/*.go`: `go generate ./ent`
- Generated files must be committed too

## 10. PR Checklist
- [ ] `go test -tags=unit ./...`
- [ ] `go test -tags=integration ./...`
- [ ] `golangci-lint run ./...`
- [ ] `pnpm-lock.yaml` synced (if package.json changed)
- [ ] All test stubs updated (if interface changed)
- [ ] Ent generated code committed (if schema changed)

## 11. Anthropic vs OpenAI API Protocol
- Anthropic accounts use `/v1/messages` endpoint with `anthropic-version` header
- OpenAI accounts use `/v1/chat/completions` or `/v1/responses` endpoint
- If backend returns "no available OpenAI accounts", check if the upstream account is actually Anthropic
- The backend routes to different providers based on account type, not endpoint

## 12. zh.ts / en.ts Editing — Watch Adjacent Lines
- When editing i18n locale files, the `old_string` match may accidentally consume adjacent sections
- Always verify surrounding context is intact after editing nested objects
- Especially dangerous near section boundaries (e.g., end of `usage:` section → start of `cleanup:` section)

## 13. Force-Push After Rebase on Feature Branches
- After `git rebase upstream/main`, push will be rejected (non-fast-forward)
- Safe to `git push --force-with-lease` on **own feature branches only**
- Never force-push shared branches (main/master)

## 14. vue-tsc Transient Errors During HMR
- After editing Vue files, HMR may show type errors about newly added refs/functions
- Run `npx vue-tsc --noEmit` directly to verify — usually passes clean
- Don't trust HMR error output for type checking; always run the CLI check
