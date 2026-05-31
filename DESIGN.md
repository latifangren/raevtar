# Raevtar Design System

## Source of Truth

Design public Raevtar UI follows RetroUI:

https://github.com/Logging-Studio/RetroUI

Use RetroUI as visual and interaction reference only. Do not import React/Next code from RetroUI. Raevtar stays Go + `a-h/templ` + Tailwind + HTMX.

The old dark-first slate/green design is deprecated. Do not use this file as evidence to bring back dark cards, rounded SaaS panels, or terminal-green styling.

## Visual Direction

Raevtar is now bright, blocky, and retro-brutalist:

- White or near-white page background.
- Black 2px borders.
- Hard offset shadows like `shadow-[4px_4px_0px_0px_#000]`.
- Bold uppercase labels and headings.
- High-contrast accent blocks.
- Minimal radius; prefer square corners.
- Mono text for API paths, timestamps, metrics, and technical labels.

The UI should feel like an opinionated personal dev console, not generic dark dashboard software.

## Palette

Use Tailwind tokens directly in templ classes.

Primary surfaces:

- `bg-neutral-100` for page background.
- `bg-white` for cards and content panels.
- `bg-black text-white` for strong bars, nav, and primary actions.
- `border-2 border-black` for most containers.

Accents:

- `bg-emerald-400` for positive/online/primary CTA.
- `bg-rose-400` or `bg-rose-300` for destructive/offline.
- `bg-yellow-200` for hover highlights.
- `bg-purple-300`, `bg-blue-300`, `bg-orange-300`, `bg-cyan-300`, `bg-amber-300` for category/role badges.
- `text-neutral-500` / `text-neutral-600` for secondary copy.

Avoid:

- Dark-first full-page themes.
- Subtle gray borders instead of black borders.
- Soft glassmorphism.
- Rounded SaaS cards.
- Purple default AI-app styling.

## Typography

Current pages load:

- `Inter` for primary UI text.
- `JetBrains Mono` for technical text.

Guidelines:

- Headings: bold or black weight, often uppercase.
- Labels: small, bold, uppercase, high contrast.
- Technical values: `font-mono font-bold`.
- Body copy: readable but still assertive; avoid faint low-contrast text.

## Components

### Cards

Default card:

```html
<div class="bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#000]">
```

Interactive card hover:

```html
hover:shadow-[2px_2px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all
```

### Buttons

Primary:

```html
border-2 border-black bg-black text-white font-bold shadow-[4px_4px_0px_0px_#000]
```

Positive action:

```html
border-2 border-black bg-emerald-400 text-black font-bold shadow-[3px_3px_0px_0px_#000]
```

Destructive action:

```html
border-2 border-black bg-rose-300 text-black font-bold
```

### Badges

Badges should use black borders, bold mono text, and strong fill colors:

```html
<span class="text-xs px-2 py-0.5 font-bold border-2 border-black bg-blue-300 text-black font-mono">
```

### Navigation

Public navigation uses simple text links with bold weight and highlight hover. Admin navigation remains inline legacy HTML for now but should visually match black sidebar + high-contrast panels.

## Implementation Rules

- Public UI must be implemented in `internal/view/**/*.templ`.
- Reusable public components live in `internal/view/components/`.
- Layout shell lives in `internal/view/layouts/base.templ`.
- Page components live in `internal/view/pages/`.
- Generated `_templ.go` files are committed but never hand-edited.
- Run `make templ-gen` or `go run github.com/a-h/templ/cmd/templ@v0.3.906 generate` after editing `.templ` files.
- Tailwind scans `internal/view/**/*.templ` and admin handler inline HTML.

## Current Scope

Public pages are the primary design surface:

- Landing page.
- Blog list/detail.
- Dashboard.
- Server detail.
- 404 page.
- RSS remains XML, not visual UI.

Admin panel is still legacy inline HTML in `internal/handler/admin.go`. Keep it secure and roughly aligned with this visual style, but do not treat admin templ migration as already done.

## Do / Don't

Do:

- Follow https://github.com/Logging-Studio/RetroUI for component feel.
- Use bright surfaces, black borders, hard shadows.
- Keep UI server-rendered with templ.
- Use HTMX only for lightweight interactivity.
- Preserve Handler -> Service -> Repo.

Don't:

- Reintroduce the old dark slate/green system.
- Add React, Next, Vue, or a JS framework.
- Copy RetroUI runtime code.
- Make rounded, soft, generic SaaS cards.
- Hand-edit generated templ Go files.
