# Raevtar Design System v2

> **Note:** This document is the single source of truth for Raevtar UI.
> Supersedes all previous design docs. If an older file says something else, this wins.

---

## 1. Design Philosophy

Raevtar is a **personal dev console, not a SaaS brochure.** Every visual choice serves three goals:

1. **Readability first** — dense technical content (code, telemetry, notes) must be easy to scan.
2. **One identity** — consistent color language across all pages, no per-page surprises.
3. **Brutalist without chaos** — thick borders, offset shadows, sharp corners. But restrained: not every section needs a different background.

**Tone:** Opinionated, warm, slightly irreverent. Retro utility aesthetic meets modern server-rendered stack (Go + Templ + HTMX).

---

## 2. Layout System

### Core Pattern: Outer/Inner Container

Every page section follows the **two-container pattern:** full-width outer for background, constrained inner for content.

```
┌─ section.w-full ──────────────────────────────────┐
│  (background: var(--background) or variant)        │
│  ┌── div.max-w-6xl.mx-auto.px-4/8 ──────────────┐ │
│  │  (content: text, cards, forms...)              │ │
│  └────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

**Rules:**
- **Outer** (`w-full`) — controls background color, top/bottom border, top/bottom padding.
- **Inner** (`max-w-6xl mx-auto px-4 sm:px-8`) — constrains content width. Never change max-width per-section (except for prose content which can be `max-w-4xl`).
- **Never** nest a constrained container inside another constrained container.
- **Never** use negative margins to fake full-width bleed.

### Layout Hierarchy

```
<body class="min-h-screen bg-background text-foreground">

  <!-- Nav: full-width bar -->
  <section class="w-full bg-card border-b-2 border-foreground sticky top-0 z-50">
    <div class="max-w-6xl mx-auto px-4 sm:px-8">
      @components.Nav(...)
    </div>
  </section>

  <!-- Hero section: unique background or default -->
  <section class="w-full bg-background border-b-2 border-foreground py-14 md:py-20">
    <div class="max-w-6xl mx-auto px-4 sm:px-8">
      <h1>...</h1>
    </div>
  </section>

  <!-- Content section: default background -->
  <section class="w-full bg-background py-12 md:py-16">
    <div class="max-w-6xl mx-auto px-4 sm:px-8">
      ...cards, grid, prose...
    </div>
  </section>

  <!-- CTA section: accent background (rare) -->
  <section class="w-full bg-accent border-y-2 border-foreground py-12 md:py-16">
    <div class="max-w-6xl mx-auto px-4 sm:px-8">
      ...
    </div>
  </section>

  <!-- Footer -->
  <section class="w-full bg-background border-t-2 border-foreground py-12">
    <div class="max-w-6xl mx-auto px-4 sm:px-8">
      @components.Footer(...)
    </div>
  </section>

</body>
```

### Section Background Usage Rules

| Section Type | Background | Border | When |
|---|---|---|---|
| Default content | `bg-background` (cream) | `border-b-2 border-foreground` optional | 90% of sections |
| Card blocks | `bg-card` (white) | `border-2 border-foreground` + shadow | Feature cards, post lists |
| Highlight bar | `bg-foreground` (black) text white | `border-y-2 border-foreground` | Marquee, announcement bars |
| CTA block | `bg-accent` (yellow) | `border-y-2 border-foreground` | Max 1 per page. **Rare.** |

**Hard rules:**
- **Do not** use `bg-success`, `bg-destructive`, `bg-secondary` as section backgrounds.
  Reserve them for **badges, status indicators, and inline tags only.**
- **Max 2 background colors per page**: `bg-background` (default) + optionally 1 accent section.
- Hero section is `bg-background` by default — differentiate via **padding size, decorative elements, and typography scale**, not color.
- `bg-foreground` (black full-width bar) is reserved for the marquee component only.

---

## 3. Color System

### 3.1 Base Palette

| Token | Hex | Usage |
|---|---|---|
| `retro.cream` | `#F5F2ED` | Page background |
| `retro.paper` | `#FFFFFF` | Card and form surfaces |
| `retro.ink` | `#000000` | Primary text, borders, shadows, nav |
| `retro.yellow` | `#FACC15` | Primary accent (singleton highlight) |
| `retro.sage` | `#6A9B7D` | Positive status, online indicator |
| `retro.sageLight` | `#DDEBDD` | Subtle positive background |
| `retro.wheat` | `#EFE2B8` | Neutral highlight, hover state |
| `retro.blush` | `#E9B7A5` | Destructive, offline, warning |
| `retro.muted` | `#6B7280` | Secondary copy, metadata |

### 3.2 Semantic Tokens

Semantic tokens connect palette to Tailwind utility classes. Defined in `tailwind.config.js` and `static/css/tailwind.src.css`.

| Token | Maps To | Usage |
|---|---|---|
| `bg-background` | `retro.cream` | Page background |
| `text-foreground` | `retro.ink` | Body text |
| `bg-card` / `text-card-foreground` | `retro.paper` / `retro.ink` | Cards, content panels |
| `bg-primary` / `text-primary-foreground` | `retro.ink` / `retro.cream` | Primary buttons, nav bars |
| `bg-primary-hover` | `#1F2937` | Primary button hover |
| `bg-secondary` / `text-secondary-foreground` | `retro.wheat` / `retro.ink` | Secondary buttons, neutral tags |
| `bg-accent` / `text-accent-foreground` | `retro.yellow` / `retro.ink` | CTAs, highlights, featured badges |
| `bg-destructive` / `text-destructive-foreground` | `retro.blush` / `retro.ink` | Delete, offline, danger |
| `bg-success` / `text-success-foreground` | `retro.sage` / `retro.paper` (or `retro.ink`) | Online, positive, green badges |
| `text-muted` | `retro.muted` | Secondary copy, timestamps |
| `bg-muted` / `text-muted-foreground` | `retro.muted` / `retro.paper` | Inactive badges, neutral indicators |
| `border-foreground` | `retro.ink` | All borders |
| `border-background` | `retro.cream` | Borders on dark bg (marquee, etc) |

### 3.3 Color Rules

**Do:**
- Background is **always** `bg-background` (cream) unless there's a specific reason otherwise.
- Borders are **always** `border-2 border-foreground` (black 2px).
- Shadows use `var(--retro-ink)` (`#000000`) or the foreground of the context.
- Accent (yellow) used sparingly: 1 element per viewport.
- Status colors used **inline** only (badges, server indicators), never as section backgrounds.

**Don't:**
- Don't use green (`bg-success`) as a section hero background.
- Don't use blush (`bg-destructive`) as a section background.
- Don't use wheat (`bg-secondary`) as a section background.
- Don't layer multiple colored sections on one page.
- Don't use `bg-accent` for more than 1 element per page section.

> **Past sin:** `about.templ` used `bg-success` for hero, `blog_post.templ` used `bg-success` for header,
> `about.templ` had 5 background colors on one page. This violates the design system.

---

## 4. Typography

| Role | Font | Weight | Case | Size Scale |
|---|---|---|---|---|
| Display/Headings | `Archivo Black` | 400 (black) | UPPERCASE | `text-4xl` to `text-7xl` |
| Subheadings | `Space Grotesk` | 700 (bold) | UPPERCASE | `text-2xl` to `text-3xl` |
| Body | `Space Grotesk` | 400-700 | sentence | `text-base` to `text-lg` |
| Labels | `Space Grotesk` | 900 (black) | UPPERCASE, tracking-widest | `text-xs` to `text-sm` |
| Technical/Mono | `JetBrains Mono` | 400-700 | as-is | `text-xs` to `text-sm` |
| Buttons | `Space Grotesk` | 900 (black) | UPPERCASE, tracking-widest | `text-base` |

**CSS Variables:**
```css
--font-sans: 'Space Grotesk', 'Inter', system-ui, sans-serif;
--font-mono: 'JetBrains Mono', 'Fira Code', monospace;
--font-head: 'Archivo Black', 'Inter', system-ui, sans-serif;
```

### Typography Patterns

- **Badge/label:** `text-xs font-black uppercase tracking-widest text-muted`
- **Metric value:** `font-mono font-bold text-foreground`
- **Section header:** `font-head text-4xl md:text-5xl font-black uppercase leading-none`
- **Card title:** `font-black uppercase text-xl md:text-2xl`

---

## 5. Component System (Neo-Brutalist)

### 5.1 Cards

```html
<!-- Default -->
<div class="bg-card border-2 border-foreground p-6 nb-shadow transition-all duration-200
            hover:shadow-[1px_1px_0px_0px_var(--retro-ink)] hover:translate-x-[2px] hover:translate-y-[2px]">

<!-- Full-width (no shadow) -->
<div class="bg-card border-2 border-foreground p-6">
```

### 5.2 Buttons

All buttons: `border-2 border-foreground font-black uppercase tracking-widest transition-all duration-200 cursor-pointer`

| Variant | Classes |
|---|---|
| Primary | `bg-foreground text-background nb-shadow` |
| Secondary | `bg-card text-foreground nb-shadow` |
| Accent (CTA) | `bg-accent text-foreground nb-shadow` |
| Destructive | `bg-destructive text-foreground nb-shadow` |

**Interaction:**
- Hover: `shadow-[1px_1px_0px_0px_var(--retro-ink)] translate-x-[2px] translate-y-[2px]`
- Active: `shadow-none translate-x-[3px] translate-y-[3px]`

### 5.3 Badges

```html
<!-- Default -->
<span class="text-xs px-2 py-0.5 font-bold border-2 border-foreground bg-card text-foreground">

<!-- Accent -->
<span class="text-xs px-2 py-0.5 font-bold border-2 border-foreground bg-accent text-foreground">

<!-- Status: positive -->
<span class="text-xs px-2 py-0.5 font-bold border-2 border-foreground bg-success text-background">

<!-- Status: destructive -->
<span class="text-xs px-2 py-0.5 font-bold border-2 border-foreground bg-destructive text-foreground">

<!-- Outline (on dark bg) -->
<span class="text-xs px-2 py-0.5 font-bold border-2 border-background bg-transparent text-background">
```

### 5.4 Navigation

```html
<!-- Inactive -->
<a class="text-sm font-black uppercase tracking-widest text-foreground no-underline hover:bg-accent px-2 py-1">

<!-- Active -->
<a class="text-sm font-black uppercase tracking-widest bg-foreground text-background px-2 py-1">
```

### 5.5 Marquee

```html
<div class="w-full bg-foreground border-y-2 border-foreground py-3 overflow-hidden">
  <div class="animate-marquee whitespace-nowrap">
    <span class="inline-block text-background text-2xl font-black uppercase mx-8"> ... </span>
  </div>
</div>
```

Full-width only. Always `bg-foreground text-background`. Always the only full-width bar on the page.

---

## 6. Section Hierarchy Per Page

Every page follows this section structure:

| Order | Type | Background | Notes |
|---|---|---|---|
| 1 | Nav | `bg-card` (white) | Sticky top, `border-b-2 border-foreground` |
| 2 | Hero | `bg-background` (cream) | Large padding (`py-14` to `py-20`). Differentiate via typography, badge, decorative element — NOT background color. |
| 3 | Content | `bg-background` (cream) | Standard padding (`py-12`). This is where cards, grids, prose live. |
| 4 | Optional Marquee | `bg-foreground` (black) | Announcer bar. Optional. Max 1 per page. |
| 5 | Optional CTA | `bg-accent` (yellow) | Max 1 per page. Only on index. |
| 6 | Content | `bg-background` (cream) | Returns to default. |
| 7 | Footer | `bg-background` (cream) | `border-t-2 border-foreground` |

**Key rule:** Once you use a non-default background (accent, foreground), the **next section must return to `bg-background`**. Never chain two non-default sections.

---

## 7. Shadow System

Shadows use the **offset shadow** pattern:

```css
/* Defined in tailwind.src.css */
.nb-shadow {
  box-shadow: 3px 3px 0px 0px var(--retro-ink);
}
.nb-shadow-lg {
  box-shadow: 6px 6px 0px 0px var(--retro-ink);
}
```

**Interaction pattern (cards & buttons):**
```
Default:  offset 3px 3px
Hover:   offset 1px 1px  (+ translate 2px 2px)
Active:  shadow-none      (+ translate 3px 3px)
```

Shadow color **always** follows `--retro-ink` (black). Never use colored shadows.

**Note for future dark mode:** shadow color will change to `--foreground` (cream) to remain visible on dark background.

---

## 8. Dark Mode (Future)

Current system is **light-mode only**. Dark mode is planned as `class`-based strategy:

### Token Mapping (Future)

```
:root.dark {
  --background: #1a1a1a;     /* dark gray */
  --foreground: #f5f2ed;     /* warm cream (same retro cream) */
  --card: #222222;            /* dark card */
  --primary: #facc15;         /* yellow becomes primary button */
  --primary-hover: #ffd600;
  --secondary: #6a9b7d;       /* sage */
  --accent: #f33;             /* red */
  --border: #f5f2ed;          /* cream border */
  --muted: #555555;
  --muted-foreground: #aaaaaa;
  --shadow-color: #f5f2ed;    /* cream shadow on dark bg */
}
```

### Design Decisions (Recorded — Not Implemented)

- **Shadow color inverts:** black shadow on light bg → cream shadow on dark bg.
- **Card bg darkens:** `#FFFFFF` → `#222222` (not pure black, maintains depth).
- **Text stays cream:** `#f5f2ed` (same as current retro-cream). Keeps warm tone.
- **Accent adapts:** Yellow (`#FACC15`) becomes primary, red (`#F33`) becomes accent.
- **Borders invert:** black on light → cream on dark.

---

## 9. What NOT To Do

- **Don't** use `bg-success` (green) as a section/hero background.
- **Don't** use `bg-destructive` (blush) as a section/hero background.
- **Don't** use `bg-secondary` (wheat) as a section/hero background.
- **Don't** have 3+ different section background colors on one page.
- **Don't** use `full-width` class as a workaround (it doesn't work with parent constraints).
- **Don't** use negative margins (`-mx-*`) to fake full-width sections.
- **Don't** use hardcoded shadow colors (`shadow-[4px_4px_0px_0px_#2D3748]`). Use CSS classes (`nb-shadow`, `nb-shadow-lg`).
- **Don't** add a dark mode toggle until the light mode layout + color system is validated on all pages.
- **Don't** import external design systems. Raevtar owns its UI.

---

## 10. Quick Reference

### Most Common Class Combinations

```html
<!-- Page section shell -->
<section class="w-full bg-background border-b-2 border-foreground py-12 md:py-16">
  <div class="max-w-6xl mx-auto px-4 sm:px-8">

<!-- Page section shell (hero — larger padding) -->
<section class="w-full bg-background border-b-2 border-foreground py-14 md:py-20">
  <div class="max-w-6xl mx-auto px-4 sm:px-8">

<!-- Card block -->
<div class="bg-card border-2 border-foreground p-6 nb-shadow">

<!-- Primary button -->
<a class="nb-button nb-button-primary">

<!-- Badge (default) -->
<span class="nb-badge nb-badge-outline">

<!-- Badge (accent/highlight) -->
<span class="nb-badge nb-badge-surface">

<!-- Section label -->
<p class="text-xs font-black uppercase tracking-widest text-muted mb-2">

<!-- Marquee bar -->
<div class="w-full bg-foreground border-y-2 border-foreground py-3 overflow-hidden">

<!-- Content wrapper for prose (slightly narrower) -->
<div class="max-w-4xl mx-auto bg-card border-2 border-foreground p-6 nb-shadow">
```

### Page Checklist

Before considering a page "done," verify:

- [ ] Hero section is `bg-background` (unless it's the index page CTA)
- [ ] No `bg-success`/`bg-destructive`/`bg-secondary` used as section backgrounds
- [ ] All sections use the two-container pattern (`w-full` outer + `max-w-6xl` inner)
- [ ] Borders are `border-2 border-foreground` (black)
- [ ] Max 1 non-background section per page (marquee or CTA, not both)
- [ ] No negative margin hacks
- [ ] All shadows use CSS classes, not inline hardcoded values
