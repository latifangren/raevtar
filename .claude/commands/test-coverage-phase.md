---
name: test-coverage-phase
description: Workflow command scaffold for test-coverage-phase in raevtar.
allowed_tools: ["Bash", "Read", "Write", "Grep", "Glob"]
---

# /test-coverage-phase

Use this workflow when working on **test-coverage-phase** in `raevtar`.

## Goal

Adds or improves automated test coverage for repo, service, model, and handler layers, often in organized phases.

## Common Files

- `internal/repo/*_test.go`
- `internal/service/*_test.go`
- `internal/model/*_test.go`
- `internal/handler/*_test.go`
- `.gitignore`

## Suggested Sequence

1. Understand the current state and failure mode before editing.
2. Make the smallest coherent change that satisfies the workflow goal.
3. Run the most relevant verification for touched files.
4. Summarize what changed and what still needs review.

## Typical Commit Signals

- Identify coverage gaps in repo, service, model, and handler code.
- Write new *_test.go files or add tests to existing ones in internal/repo/, internal/service/, internal/model/, internal/handler/.
- Fix minor bugs or test issues discovered during coverage expansion.
- Update .gitignore to exclude coverage artifacts if needed.

## Notes

- Treat this as a scaffold, not a hard-coded script.
- Update the command if the workflow evolves materially.