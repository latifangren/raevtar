```markdown
# raevtar Development Patterns

> Auto-generated skill from repository analysis

## Overview

This skill teaches you the core development patterns, coding conventions, and workflows used in the `raevtar` Go codebase. You'll learn how to implement new features, expand test coverage, refresh the frontend design, update CI/CD pipelines, and maintain documentation—all following the repository's established standards. Example code and step-by-step instructions are provided to help you contribute effectively.

---

## Coding Conventions

### File Naming

- Use **snake_case** for file names.
  - Example: `user_service.go`, `admin_routes.go`
- Test files follow the pattern: `*_test.go`
  - Example: `repo_test.go`

### Import Style

- Use **relative imports** within the module.
  - Example:
    ```go
    import (
        "internal/model"
        "internal/service"
    )
    ```

### Export Style

- **Capitalize** the first letter of exported functions, types, and variables. Lowercase first letter = package-private.
  - Example:
    ```go
    // Exported function (capitalized)
    func NewUserService() *UserService {
        // ...
    }
    // Unexported function (lowercase)
    func parseConfig() error {
        // ...
    }
    ```

### Templates & Assets

- Templates are stored as `.templ` files and their generated counterparts as `*_templ.go`.
- Static assets (CSS, JS, fonts) are organized under `static/`.

---

## Workflows

### Feature Development & Implementation

**Trigger:** When adding a significant new feature or enhancement  
**Command:** `/new-feature`

1. Create or update handler files in `internal/handler/` (e.g., `handlers.go`, `admin.go`, `api.go`, `routes.go`).
2. Add or update model files in `internal/model/` and repo files in `internal/repo/` as needed.
3. Implement service logic in `internal/service/`.
4. Update or create templates in:
    - `internal/view/pages/`
    - `internal/view/admin/`
    - `internal/view/components/`
    - And their corresponding `*_templ.go` files.
5. Update or add static assets (`static/css/*.css`, `static/js/*.js`) if frontend changes are required.
6. Add or update tests in `internal/handler/`, `internal/service/`, `internal/repo/` as appropriate.

**Example:**
```go
// internal/handler/user.go
package handler

import (
    "internal/service"
)

func RegisterUserHandler(svc *service.UserService) {
    // Handler logic
}
```

---

### Test Coverage Phase

**Trigger:** When increasing code coverage or after major refactors  
**Command:** `/increase-coverage`

1. Identify coverage gaps in repo, service, model, and handler code.
2. Write new `*_test.go` files or add tests to existing ones in:
    - `internal/repo/`
    - `internal/service/`
    - `internal/model/`
    - `internal/handler/`
3. Fix minor bugs or test issues discovered during coverage expansion.
4. Update `.gitignore` to exclude coverage artifacts if needed.

**Example:**
```go
// internal/service/user_service_test.go
package service

import "testing"

func TestUserCreation(t *testing.T) {
    // Test logic
}
```

---

### Frontend Design Refresh

**Trigger:** When updating the visual style or implementing a new design system  
**Command:** `/design-refresh`

1. Update `static/css/style.css` and/or `static/css/tailwind.src.css` with new design tokens, patterns, or utilities.
2. Modify `tailwind.config.js` for new colors, fonts, or scale.
3. Update or refactor templates in:
    - `internal/view/components/`
    - `internal/view/layouts/`
    - `internal/view/pages/`
    - And their `*_templ.go` files.
4. Add or update font files in `static/fonts/` if changing typography.
5. Test UI changes in the browser and adjust as needed.

**Example:**
```css
/* static/css/style.css */
:root {
    --primary-color: #4f46e5;
}
```

---

### CI/CD Pipeline Setup or Update

**Trigger:** When automating testing, building, or releasing the project  
**Command:** `/setup-ci-cd`

1. Add or update `.github/workflows/*.yml` for CI and release workflows.
2. Add or update `.goreleaser.yaml` for release configuration.
3. Test the pipeline by pushing changes or tags.

**Example:**
```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.20
      - run: go test ./...
```

---

### Documentation Update

**Trigger:** When reflecting new features, architecture changes, or roadmap progress  
**Command:** `/update-docs`

1. Edit `README.md` to reflect the latest features and usage.
2. Update or add:
    - `docs/ARCHITECTURE.md`
    - `docs/PRD.md`
    - `docs/ROADMAP.md`
    - `docs/CODEMAPS/*`
3. Commit and push documentation changes.

---

## Testing Patterns

- Test files are named with the pattern `*_test.go`.
- Tests are written using Go's standard `testing` package.
- Place tests alongside the code they cover, e.g., `internal/service/user_service_test.go`.
- Test coverage is often increased in organized phases, sometimes after major feature pushes or refactors.

**Example:**
```go
// internal/repo/user_repo_test.go
package repo

import "testing"

func TestFindUserByID(t *testing.T) {
    // Test logic
}
```

---

## Commands

| Command           | Purpose                                              |
|-------------------|------------------------------------------------------|
| /new-feature      | Start a new feature or major enhancement             |
| /increase-coverage| Add or improve automated test coverage               |
| /design-refresh   | Perform a major UI/UX design update                  |
| /setup-ci-cd      | Add or update CI/CD pipelines and release automation |
| /update-docs      | Update project documentation                         |
```
