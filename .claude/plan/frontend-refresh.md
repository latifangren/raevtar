# Frontend Refresh Plan — Solusi A (Complete Refresh)

**Reference**: RETROUI.dev design system
**Goal**: Raevtar frontend konsisten dengan pola RETROUI: semantic tokens, proper typography hierarchy, accessible components, standardized interaction patterns.

---

## Phase 1: Foundation (Config + Assets)

### 1.1 Tailwind Config Update
- Add semantic color tokens (`primary`, `secondary`, `accent`, `muted`, `destructive`, `card`, `background`, `foreground`, `border`, `ring`, `primary-foreground`, `secondary-foreground`, `accent-foreground`, `destructive-foreground`, `card-foreground`, `muted-foreground`)
- Map existing `retro-*` colors to semantic tokens
- Add `font-head` to fontFamily config
- Add `borderRadius` default for `DEFAULT` (ke 4px dulu, evaluasi nanti)
- Keep `raevtar-*` palette but mark as deprecated

### 1.2 Font Setup
- Download Archivo Black WOFF2 + WOFF
- Place in `static/fonts/`
- Add `@font-face` declaration in `tailwind.src.css`
- Create `font-head` utility class

### 1.3 CSS System Update (`tailwind.src.css`)
- Update `.nb-*` classes to match RETROUI patterns:
  - Border: 4px → 2px
  - Shadow shrink on hover (`shadow-md hover:shadow active:shadow-none`)
  - Add `rounded` default
  - Interactive lift: `hover:translate-y-1 active:translate-y-2 active:translate-x-1`
- Add focus-visible pattern to all interactive classes
- Define missing CSS classes: `.nb-border-b`, `.nb-h3`, `.custom-scrollbar`
- Add semantic CSS variable tokens

### 1.4 New Component Classes
- Add button variants: `.nb-button-outline`, `.nb-button-ghost`, `.nb-button-link`, `.nb-button-icon`
- Add badge variants: `.nb-badge-outline`, `.nb-badge-solid`, `.nb-badge-surface`
- Add badge sizes: `.nb-badge-sm`, `.nb-badge-lg`
- Add card sizes: `.nb-card-sm`, `.nb-card-lg`

### 1.5 Icon System Foundation
- Create `static/icons/` with essential Lucide SVGs (or inline SVG helpers)
- Add `data-icon` attribute pattern for icons

**Files touched**: `tailwind.config.js`, `static/css/tailwind.src.css`, `static/fonts/`

---

## Phase 2: Accessibility Pass

### 2.1 Focus-Visible
- `base.templ`: Add focus-visible to body
- `nav.templ`: Add focus-visible to all links
- `footer.templ`: Add focus-visible to all links
- `post_card.templ`: Add focus-visible to article links
- `pagination.templ`: Add focus-visible to page links
- All admin templates: add focus-visible to buttons, links, inputs
- All page templates: add focus-visible to CTAs, links

### 2.2 ARIA Attributes
- `nav.templ`: Add `aria-current="page"` to active nav link
- `footer.templ`: Add `aria-label` to badge links
- `pagination.templ`: Add `aria-current="page"` to active page
- `server_detail_charts.templ`: Add `role="img"` + `aria-label` to canvas
- `dashboard.templ`: Add `aria-live="polite"` to HTMX refresh targets
- All forms: Add `aria-required` to required fields
- All decorative elements: Add `aria-hidden="true"`
- Admin layout: Add `aria-current="page"` to sidebar links

### 2.3 HTMX Accessibility
- Add `aria-busy` indicators during HTMX requests
- Ensure HTMX targets have `aria-live` regions

**Files touched**: All `.templ` files (~33 files)

---

## Phase 3: Component Standardization

### 3.1 Navigation (`nav.templ`)
- Update nav shadow from `shadow-[6px_6px_0px_0px_#000000]` to RETROUI clean pattern
- Update active state to RETROUI: `bg-retro-ink text-retro-paper`
- Add proper focus ring
- Add aria-current

### 3.2 Buttons (across all templates)
- Replace `shadow-[3px_3px_0px_0px_#2D3748] hover:shadow-[1px_1px_0px_0px_#2D3748]` with RETROUI pattern: `shadow-md hover:shadow active:shadow-none hover:translate-y-1 active:translate-y-2 active:translate-x-1`
- Replace hardcoded shadow colors with semantic tokens
- Apply `.nb-button-outline` for secondary actions

### 3.3 Cards (PostCard, ProjectCard, ServerCard)
- Update border: 4px → 2px
- Update shadow: pake RETROUI shrinking pattern
- Add `rounded` where appropriate
- `PostCard`: hover → `bg-muted` instead of `bg-retro-yellow/10`
- `ServerCard`: standardize dengan card lain

### 3.4 Badge (TagBadges, footer badges, etc.)
- Update to RETROUI badge variants (default, outline, solid, surface)
- Add size variants

### 3.5 Form Inputs
- Standardize input styling across all forms
- Use consistent focus ring pattern
- Add proper labels and aria attributes

### 3.6 Pagination
- Update to match RETROUI style

**Files touched**: All `.templ` files (~33 files)

---

## Phase 4: Shadow & Interaction Standardization

### 4.1 Shadow Pattern
- Replace all hardcoded `shadow-[*_*_0px_0px_#2D3748]` with Tailwind semantic shadow classes
- Standardize on: `shadow-md` (default), `hover:shadow` (hover), `active:shadow-none` (press)
- Remove all hardcoded `#2D3748` from shadow values

### 4.2 Hover/Active Pattern
- Standardize: `hover:translate-y-1 active:translate-y-2 active:translate-x-1`
- Update in `tailwind.src.css` class definitions

### 4.3 Border Standardization
- All borders: 2px (was 4px)
- Consistent `border-2 border-retro-ink`
- Add `rounded` where appropriate

**Files touched**: `tailwind.src.css`, `tailwind.config.js`, all templates

---

## Phase 5: Font Self-Hosting (Optional)

### 5.1 Download Fonts
- Inter (400, 500, 600, 700, 900) — WOFF2
- JetBrains Mono (400, 500, 700) — WOFF2
- Archivo Black (400) — WOFF2

### 5.2 @font-face Declarations
- Add to `tailwind.src.css`
- Use `font-display: swap`

### 5.3 Remove CDN Links
- Remove Google Fonts `<link>` from `base.templ` and admin `layout.templ`
- Update font-family fallback

**Files touched**: `base.templ`, `admin/layout.templ`, `tailwind.src.css`, `tailwind.config.js`

---

## Execution Order

| Step | Phase | Effort | Risk | Impact |
|------|-------|--------|------|--------|
| 1 | 1.1 Tailwind semantic tokens | Low | Low | High |
| 2 | 1.3 CSS system update | Medium | Medium | High |
| 3 | 1.4 New component classes | Medium | Low | High |
| 4 | 2.1 Focus-visible | Medium | Low | High |
| 5 | 2.2 ARIA attributes | Medium | Low | Medium |
| 6 | 3.1 Nav update | Low | Low | Medium |
| 7 | 3.2 Button standardization | High | Medium | High |
| 8 | 3.3 Card standardization | High | Medium | High |
| 9 | 3.4 Badge update | Medium | Low | Medium |
| 10 | 4.1 Shadow pattern | High | Medium | High |
| 11 | 5 Font self-hosting | Medium | Low | Medium |

**Total estimated files touched**: ~35 files (config, CSS, 33+ templates)
**Implementation approach**: Per-phase commits for safety

---

## Rollback Consideration
- Each phase committed separately for easy rollback
- Phase 1 (config + CSS) does NOT affect templates — safe foundation
- Visual changes in Phase 3-4 can be reverted by rolling back CSS + template commits
