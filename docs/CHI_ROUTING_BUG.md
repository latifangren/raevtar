# Chi routing oddity notes (2026-05-31)

## Symptom
- POST `/admin/users` returned **404** (falls through to NotFound), while:
  - POST `/admin/posts` ✅
  - POST `/admin/servers` ✅
  - POST `/admin/login` ✅
- Also saw strange behavior with a root-level POST test route returning 404.

This looked like a **path-specific registration / matching** issue, not auth middleware.

## Workaround applied
Renamed user management routes:
- `/admin/users` → `/admin/manage-users`

Changes:
- `internal/handler/routes.go` route registrations updated to `manage-users`
- `internal/handler/admin.go` links + form action updated to `manage-users`

## Current state
- Systemd service restarted with new binary.
- `GET https://raevtar.tech/admin/manage-users` returns HTTP **200**.

## Next steps (if you want root cause)
1. Add a minimal route-walk debug dump (chi.Walk) at boot (or keep test) to confirm which routes are registered.
2. Make a tiny repro with chi v5.3.0 and two sibling POST routes differing only by path, confirm if tree bug reproducible.
3. If reproducible: bump chi version (v5.3.1+), or avoid ambiguous static segments.
