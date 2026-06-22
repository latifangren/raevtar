# Frontend Gap Analysis: Raevtar vs RETROUI.dev

> **Tujuan**: Membandingkan konfigurasi frontend Raevtar saat ini dengan design system
> RETROUI.dev (https://retroui.dev) sebagai referensi konsistensi warna, layout,
> dan komponen antar halaman/menu.

---

## 1. Font & Typography

| Aspek | RETROUI.dev | Raevtar (Current) | Gap |
|-------|-------------|--------------------|-----|
| **Sans font** | Space Grotesk (variable, 300–700) | Inter | Berbeda font. RETROUI pake Space Grotesk yg lebih "grotesk" vibes neo-brutalist |
| **Headings font** | Archivo Black (heading专用, `font-head`) | Inter 900 (sama dgn body weight) | **Missing**. Raevtar gak punya dedicated heading font. RETROUI pake Archivo Black yg chunky buat headline |
| **Mono font** | Space Mono | JetBrains Mono | OK. Bedanya Space Mono vs JetBrains Mono — minor, dua-duanya bagus |
| **Font loading** | Self-hosted WOFF2, local fallback + `size-adjust` | Google Fonts CDN | Raevtar masih dependen ke Google Fonts CDN (render-blocking). RETROUI self-host + fallback metrics |
| **`font-head` utility** | Ada — class `font-head` pake Archivo Black via CSS variable `--font-head` | Tidak ada | **Missing utility class**. Raevtar perlu tambah `font-head` di tailwind config |

---

## 2. Color Palette

### RETROUI.dev (Tailwind v4)
Menggunakan CSS variable-based theming via `@layer theme`:

| Token | Value |
|-------|-------|
| `--color-primary` | #000000 (black) — solid, bold |
| `--color-secondary` | #374151 (gray-700) |
| `--color-background` | #FFFFFF (white) |
| `--color-foreground` | #000000 |
| `--color-muted` | gray-500 (#6A7282) |
| `--color-card` | #FFFFFF |
| `--color-accent` | #FACC15 (yellow) — almost same as our retro-yellow |
| `--color-destructive` | red-500 (#FB2C36) |
| `--color-red-*` | 100, 300, 500, 600, 800 — lengkap |
| `--color-yellow-*` | 300, 800 | 
| `--color-green-*` | 300, 400, 500, 600, 800 |
| `--color-blue-*` | 300, 400, 500, 600, 800 |
| `--color-purple-*` | 300, 500 |
| `--color-gray-*` | 50, 100, 200, 300, 400, 500, 600, 700, 800, 900 — **lengkap 10 shade** |

### Raevtar (Current Tailwind v3)

| Token | Value | Notes |
|-------|-------|-------|
| `retro-cream` | #F5F2ED | Vibe unik — gak ada di RETROUI |
| `retro-paper` | #FFFFFF | Sama |
| `retro-ink` | #000000 | Sama |
| `retro-yellow` | #FACC15 | Sama persis dgn `accent` RETROUI |
| `retro-sage` | #6A9B7D | **Unique** — gak ada di RETROUI |
| `retro-sageLight` | #DDEBDD | **Unique** |
| `retro-wheat` | #EFE2B8 | **Unique** |
| `retro-blush` | #E9B7A5 | **Unique** |
| `retro-muted` | #6B7280 | Sama dgn gray-500 RETROUI |
| `raevtar-50–950` | palette hijau (kelabu) | **Unique** — mirip gray scale dgn nuance hijau |

### Gap Warna

| No | Gap | Detail |
|----|-----|--------|
| 1 | **Missing semantic tokens** | Raevtar gak punya semantic tokens: `primary`, `secondary`, `accent`, `card`, `destructive`, `primary-foreground`, `secondary-foreground`, `card-foreground`, `accent-foreground`, `background`, `foreground`, `border`, `ring` |
| 2 | **Missing state colors** | RETROUI punya red/green/blue/amber/purple scale (min 3–5 shade per warna). Raevtar cuma punya yellow 1 shade, sage 1 shade, blush 1 shade, gak ada red/green/blue/purple scale |
| 3 | **Hover states** | RETROUI pake `primary-hover`, `secondary-hover`, `destructive-hover` via opacity (`/90`) atau explicit darkening. Raevtar hardcode hover di CSS manual |
| 4 | **Dark mode** | RETROUI pake `class="dark"` toggle + `localStorage`. Raevtar **tidak ada** dark mode sama sekali |
| 5 | **No raevtar-* palette needed** | `raevtar-50`–`950` hampir gak dipakai di template. Bisa disederhanakan atau dihapus |

---

## 3. Layout System

| Aspek | RETROUI.dev | Raevtar | Gap |
|-------|-------------|---------|-----|
| **Max width container** | 72rem (container-6xl) pages, atau full-width blocks | `max-w-6xl mx-auto` (72rem) | Sama |
| **Spacing scale** | Standard Tailwind v4 (0.25rem base) | Standard Tailwind v3 | OK |
| **Responsive breakpoints** | sm/md/lg/xl/2xl | sm/md/lg/xl/2xl | OK |
| **Grid** | Tailwind grid, flex, beberapa komponen pake `bento-grid` | Tailwind grid (`grid-cols-1 md:grid-cols-2 lg:grid-cols-3`) | OK |
| **Dark mode** | Ada (class strategy + localStorage) | **Tidak ada** | Gap besar untuk user experience modern |

---

## 4. Komponen Design System

### 4.1 Button

| Kategori | RETROUI.dev | Raevtar | Gap |
|----------|-------------|---------|-----|
| `primary` | `bg-primary text-primary-foreground border-2 border-black shadow-md hover:shadow active:shadow-none hover:translate-y-1 active:translate-y-2 active:translate-x-1` | `nb-button nb-button-primary` | RETROUI punya **shadow transition** + **active state translate** yg lebih detail |
| `secondary` | `bg-secondary shadow-primary` dgn border sama | `nb-button nb-button-secondary` | Mirip, tp RETROUI pake `shadow-primary` (tinted shadow) |
| `outline/ghost` | `bg-transparent border-2 transition` + hover underline | **Tidak ada** | Raevtar gak punya variant outline button |
| `link` | `bg-transparent hover:underline px-4 py-1.5` | **Tidak ada** | Raevtar gak punya link-style button |
| `icon` | `p-2` (square) dgn icon Lucide | **Tidak ada** | Raevtar gak punya icon-only button variant |
| `with-icon` | Button dgn icon + text, `mr-2` spacing | **Tidak ada** | Gap — icon Lucide gak ada di project kita |
| `disabled` | `disabled:opacity-60 disabled:cursor-not-allowed` | **Tidak ada explicit** | Hanya di CSS class manual |
| **Size variants** | Default, lg, sm (via padding/text size) | Single size | Gap — Raevtar gak punya size variant pada button |
| **Focus ring** | `focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary` | **Tidak ada** | Aksesibilitas gap — keyboard focus style hilang |

### 4.2 Card

| Kategori | RETROUI.dev | Raevtar | Gap |
|----------|-------------|---------|-----|
| Default | `<div class="bg-card border text-foreground">` + shadow/hover shadow-sm | `nb-card` (heavy neo-brutalist shadow) | Agak beda approach — RETROUI pake border + shadow, Raevtar pake heavy shadow solid |
| Hover effect | Shadow reduce (`shadow-sm`) | Shadow increase (`6px 6px`) | Kebalikan — RETROUI:ngecilin shadow pas hover, Raevtar: ngegedein |
| Border style | `rounded` (default radius) + `border-2` | `border-4 border-retro-ink` | Raevtar pake border lebih tebal (4px vs 2px). Raevtar juga gak pake rounding |
| Padding | Bervariasi (px-4 lg:px-12 py-24 di demo) | `p-6` konsisten | RETROUI demo pake padding lebih besat |

### 4.3 Badge

| Kategori | RETROUI.dev | Raevtar | Gap |
|----------|-------------|---------|-----|
| Color | `bg-muted text-muted-foreground` (default), outline, solid, surface | `nb-badge` + color class custom | Mirip approach-nya |
| Border | Tidak explicit — pake `rounded` +bg | `nb-border` (4px ink border) | Raevtar pake border tebal (neo-brutalist) |
| Rounded | `rounded-sm`, `rounded-full`, etc | Tidak ada rounding | **Gap** — Raevtar gak punya rounded variant |
| Size | px-2 py-1 text-xs (sm), px-2.5 py-1.5 text-sm (md), px-3 py-2 text-base (lg) | Single size | **Gap** — Raevtar gak punya size variant |
| Variants | Default (muted), Outline (transparent + border), Solid (foreground), Surface (primary) | Single style + color class | **Gap** — Raevtar bisa lebih fleksibel |

### 4.4 Form Inputs & Fields (gap total)

| Item | RETROUI.dev | Raevtar | Gap |
|------|-------------|---------|-----|
| **Text input** | Ada (di authentication page) | **Tidak ada** | Belum ada komponen input |
| **Select dropdown** | Ada | **Tidak ada** | Belum ada |
| **Checkbox/Radio** | Kemungkinan ada | **Tidak ada** | Belum ada |
| **Auth forms** | Ada halaman sign-in, authentication templates | Login form di admin panel (inline HTML) | Minimal — perlu standardisasi |

### 4.5 Navigation (Navbar)

| Aspek | RETROUI.dev | Raevtar | Gap |
|-------|-------------|---------|-----|
| Style | Flat bg, border-b, tabs | Sticky nav, bg-paper + nb-border + heavy shadow | Berbeda — Raevtar pake shadow 6px tebal, RETROUI lebih clean |
| Active state | `data-selected:bg-black data-selected:text-white` | `bg-retro-yellow text-retro-ink border-2` | Berbeda approach — yellow vs black |
| Mobile | Scrollable (overflow-x-auto) | Scrollable (overflow-x-auto) | Sama |

### 4.6 Accordion (gap total)

RETROUI punya accordion component dengan:

```html
<div data-orientation="vertical" dir="ltr" role="region" class="space-y-4 w-full">
  <div data-orientation="vertical" class="border-2 bg-background rounded text-foreground shadow-md">
```

- **Raevtar: Tidak ada** accordion component

### 4.7 Breadcrumb (gap total)

Ada halaman breadcrumb.html — Raevtar: **tidak ada**.

### 4.8 Avatar (gap total)

Ada halaman avatar.html — Raevtar: **tidak ada**.

---

## 5. Komponen Khusus Raevtar (tidak ada di RETROUI)

| Komponen | Keterangan |
|----------|-----------|
| `ServerCard` | Server monitoring card — unik untuk Raevtar |
| `Pagination` | Halaman blog/projects |
| `PostCard` | Blog post card |
| `ProjectCard` | Project card dgn cover image + grid layout |
| Charts (Chart.js) | Grafik monitoring server — via CDN |
| Marquee animation | Footer style |
| `highlight.go` | Text highlight component |

---

## 6. CSS Patterns

### RETROUI.dev
- **Tailwind v4** (CSS variable-based theming via `@layer theme`)
- Class-based component pattern (shadcn/ui style)
- Modern approach: `data-selected:`, `data-open:`, `data-closed:` attribute selectors
- Semantic color tokens: `bg-primary`, `text-primary-foreground`, `bg-card`, `bg-muted`
- Hover states via Tailwind utility (`hover:translate-y-1`, `active:translate-y-2`)
- `shadow-md hover:shadow active:shadow-none` — shadow shrinks on interaction
- No custom CSS classes — 100% utility-first

### Raevtar (Current)
- **Tailwind v3** (config-based)
- Custom CSS classes: `.nb-card`, `.nb-button`, `.nb-badge`, `.nb-border`, `.nb-shadow`, `.nb-shadow-lg`
- Neo-brutalist heavy style: 4px border, 4-8px solid shadow
- `!important` hover di CSS: hover effect via class (`nb-card:hover`)
- Shadow grows on hover (inverse of RETROUI)
- Mix of utility classes + custom CSS (hybrid approach)

---

## 7. Icon System

| Aspek | RETROUI.dev | Raevtar | Gap |
|-------|-------------|---------|-----|
| Source | **Lucide icons** (SVG inline) | **Tidak ada icon system** | Raevtar pake text/arrow characters |
| Usage | `lucide lucide-pen`, `lucide lucide-trash`, etc | Tidak ada icon | Beberapa link pake `&rarr;` HTML entity |
| Admin panel | Icon di button, form, auth | Icon minimal | Gap besar — admin panel tanpa icon |

---

## 8. Shadow & Border Consistency

### RETROUI approach
- `shadow-md hover:shadow active:shadow-none` — shadow dimulai besar, mengecil saat interaksi
- `border-2 border-black` — tipis, konsisten di semua komponen
- `rounded` — ada rounding default

### Raevtar approach
- `nb-shadow`: `4px 4px 0px 0px black`, `nb-shadow-lg`: `8px 8px 0px 0px black`
- `nb-border`: `4px solid black` — **sangat tebal**
- Hover: shadow nambah **6px 6px** (kebalikan RETROUI)
- **Tidak ada border radius** — sharp 0px everywhere

### Gap
Perbedaan approach shadow & border bisa membuat Raevtar terasa lebih "heavy" dari RETROUI.
Jika ingin konsisten dgn RETROUI, perlu reducer border width (4px → 2px) dan tambah `rounded`.

---

## 9. Aksesibilitas Gaps

| Item | RETROUI.dev | Raevtar | Status |
|------|-------------|---------|--------|
| `focus-visible:outline-2` | Ada di button, tab, link | **Tidak ada** di semua komponen | **Gap kritis** |
| `disabled:opacity-60` | Ada di button variant | **Tidak ada** | Gap |
| `aria-expanded`, `aria-selected` | Ada di accordion, tabs | Tidak konsisten | Gap |
| `role="tab", role="tabpanel"` | Lengkap | Tidak ada tab system | Gap |
| Dark mode support | localStorage toggle | **Tidak ada** | Gap |

---

## 10. Ringkasan Prioritas Gap

### P0 — Critical (design inconsistency)

| # | Gap | Impact |
|---|-----|--------|
| 1 | **Tidak ada semantic color tokens** (`primary`, `secondary`, `accent`, dll) | Setiap halaman pake warna hardcode — sulit maintain konsistensi |
| 2 | **Focus ring aksesibilitas hilang** | Keyboard users gak bisa navigate — accessibility issue |
| 3 | **Tidak ada button variants** (outline, ghost, link, icon, disabled) | UX monoton, gak bisa bedain action hierarchy |

### P1 — High

| # | Gap | Impact |
|---|-----|--------|
| 4 | **Tidak ada dedicated heading font** | Heading gak beda dari body — visual hierarchy kurang |
| 5 | **Tidak ada dark mode** | User experience gap — gak ada preferensi malam |
| 6 | **Tidak ada icon system** | Admin panel & UI terasa "plain" |
| 7 | **Border terlalu tebal (4px vs 2px)** | Komponen terasa berat, kurang modern |

### P2 — Medium

| # | Gap | Impact |
|---|-----|--------|
| 8 | **Google Fonts CDN dependency** | Render-blocking, gak bisa offline |
| 9 | **Tidak ada badge size variant** | Badge selalu ukuran sama |
| 10 | **Tidak ada card size variant** | Komponen kaku |
| 11 | **Missing state colors** (red/green/blue scale) | Alert & status indicator terbatas |
| 12 | **Tidak ada form component standard** | Form login, contact, search belum terstandardisasi |

### P3 — Low

| # | Gap | Impact |
|---|-----|--------|
| 13 | **Tidak ada accordion** | Docs page bisa lebih terstruktur |
| 14 | **Tidak ada breadcrumb** | Navigasi depth kurang |
| 15 | **Tidak ada avatar** | Kurang personal |
| 16 | **`raevtar-*` palette jarang dipakai** | Bisa disederhanakan |

---

## 11. Rekomendasi Quick Wins (dengan effort rendah)

1. **Tambah semantic colors ke tailwind.config.js** → mapping `retro-*` → `primary`, `secondary`, dll
2. **Self-host fonts** → download Inter + Archivo Black, serve dari `/static/fonts/`
3. **Tambah `font-head` utility** → font-family Archivo Black (atau Inter 900 sebagai fallback)
4. **Tambah `focus-visible:outline-2`** ke semua interactive elements
5. **Tambah Lucide icons** → CDN atau self-host SVG sprite

---

*Dibuat: 22 Juni 2026 — Perbandingan RETROUI.dev (scraped) vs Raevtar frontend config.*
