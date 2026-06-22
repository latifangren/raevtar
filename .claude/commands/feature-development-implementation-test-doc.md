---
name: feature-development-implementation-test-doc
description: Workflow command scaffold for feature-development-implementation-test-doc in raevtar.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /feature-development-implementation-test-doc

Use this workflow when working on **feature-development-implementation-test-doc** in `raevtar`.

## Goal

Implements a new feature or major enhancement, including backend logic, frontend templates, and associated tests.

## Common Files

- `internal/handler/*.go`
- `internal/model/*.go`
- `internal/repo/*.go`
- `internal/service/*.go`
- `internal/view/pages/*.templ`
- `internal/view/pages/*_templ.go`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Create or update handler files in internal/handler/ (e.g., handlers.go, admin.go, api.go, routes.go).
- Add or update model files in internal/model/ and repo files in internal/repo/ as needed.
- Implement service logic in internal/service/.
- Update or create templates in internal/view/pages/, internal/view/admin/, internal/view/components/, and their corresponding *_templ.go files.
- Update or add static assets (CSS, JS) if frontend changes are needed.

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.